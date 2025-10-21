package handlers

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/history"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/grouputil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkimg"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	"github.com/BetaGoRobot/BetaGo/utility/logging"
	"github.com/BetaGoRobot/BetaGo/utility/message"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/BetaGoRobot/BetaGo/utility/retriver"
	commonutils "github.com/BetaGoRobot/go_utils/common_utils"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/tmc/langchaingo/schema"
	"go.opentelemetry.io/otel/attribute"
)

func ChatHandler(chatType string) func(ctx context.Context, event *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	return func(ctx context.Context, event *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
		newChatType := chatType
		size := new(int)
		*size = 20
		argMap, input := parseArgs(args...)
		if _, ok := argMap["r"]; ok {
			newChatType = consts.MODEL_TYPE_REASON
		}
		if _, ok := argMap["c"]; ok {
			// no context
			*size = 0
		}
		return ChatHandlerInner(ctx, event, newChatType, size, input)
	}
}

func ChatHandlerInner(ctx context.Context, event *larkim.P2MessageReceiveV1, chatType string, size *int, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	var (
		res   iter.Seq[*doubao.ModelStreamRespReasoning]
		files = make([]string, 0)
	)
	if ext, err := redis.GetRedisClient().
		Exists(ctx, MuteRedisKeyPrefix+*event.Event.Message.ChatId).Result(); err != nil {
		return err
	} else if ext != 0 {
		return nil // Do nothing
	}
	urlSeq, err := larkimg.GetAllImgURLFromMsg(ctx, *event.Event.Message.MessageId)
	if err != nil {
		return err
	}
	if urlSeq != nil {
		for url := range urlSeq {
			files = append(files, url)
		}
	}
	// 看看有没有quote的消息包含图片
	urlSeq, err = larkimg.GetAllImgURLFromParent(ctx, event)
	if err != nil {
		return err
	}
	if urlSeq != nil {
		for url := range urlSeq {
			files = append(files, url)
		}
	}
	if chatType == consts.MODEL_TYPE_REASON {
		res, err = GenerateChatSeq(ctx, event, doubao.ARK_REASON_EPID, size, files, args...)
		if err != nil {
			return
		}
		err = message.SendAndUpdateStreamingCard(ctx, event.Event.Message, res)
		if err != nil {
			return
		}
	} else {
		res, err = GenerateChatSeq(ctx, event, doubao.ARK_NORMAL_EPID, size, files, args...)
		if err != nil {
			return err
		}
		// 先reply一条，占位。
		resp, err := larkutils.ReplyMsgText(
			ctx, "我想想应该怎么说...", *event.Event.Message.MessageId, "_chat_random", false,
		)
		if err != nil {
			return err
		}
		if !resp.Success() {
			return errors.New(resp.Error())
		}
		lastMsgID := *resp.Data.MessageId
		idx := 0
		lastData := &doubao.ModelStreamRespReasoning{}
		replyMsg := func(content string) {
			idx++
			resp, err := larkutils.ReplyMsgText(
				ctx, content, lastMsgID, strconv.Itoa(idx), true,
			)
			if err != nil {
				logging.Logger.Err(errors.New(err.Error()))
				return
			}
			if !resp.Success() {
				logging.Logger.Err(errors.New(resp.Error()))
				return
			}
		}
		for data := range res {
			eot := "**回复:**"
			sor := "\n参考资料:"
			span.SetAttributes(attribute.String("lastData", data.Content))
			if idx := strings.Index(data.Content, eot); idx != -1 {
				lastData = data
				lastData.Content = strings.TrimSpace(lastData.Content[idx+len(eot):])
				if idx := strings.Index(data.Content, sor); idx != -1 {
					replyMsg(strings.TrimSpace(data.Content[idx:]))
					lastData.Content = strings.TrimSpace(lastData.Content[:idx])
				}
			}

			if strings.Contains(lastData.Content, "[无需回复]") {
				return err
			}
			if data.Reply2Show != nil {
				replyMsg(data.Reply2Show.Content)
			}
		}
		err = larkutils.UpdateMessageText(ctx, lastMsgID, lastData.Content)
		if err != nil {
			return err
		}
	}
	return
}

func GenerateChatSeq(ctx context.Context, event *larkim.P2MessageReceiveV1, modelID string, size *int, files []string, input ...string) (res iter.Seq[*doubao.ModelStreamRespReasoning], err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	// 默认获取最近20条消息
	if size == nil {
		size = new(int)
		*size = 20
	}

	chatID := *event.Event.Message.ChatId
	messageList, err := history.New(ctx).
		Query(osquery.Bool().Must(osquery.Term("chat_id", chatID))).
		Source("raw_message", "mentions", "create_time", "user_id", "chat_id", "user_name", "message_type").
		Size(uint64(*size*3)).Sort("create_time", "desc").GetMsg()
	if err != nil {
		return
	}

	templateRows, _ := database.FindByCacheFunc(database.PromptTemplateArgs{PromptID: 1}, func(d database.PromptTemplateArgs) string {
		return strconv.Itoa(d.PromptID)
	})
	if len(templateRows) == 0 {
		return nil, errors.New("prompt template not found")
	}

	promptTemplate := templateRows[0]
	promptTemplateStr := promptTemplate.TemplateStr
	tp, err := template.New("prompt").Parse(promptTemplateStr)
	if err != nil {
		return nil, err
	}
	member, err := grouputil.GetUserMemberFromChat(ctx, chatID, *event.Event.Sender.SenderId.OpenId)
	if err != nil {
		return
	}
	userName := ""
	if member == nil {
		userName = "NULL"
	} else {
		userName = *member.Name
	}
	createTime := utility.EpoMil2DateStr(*event.Event.Message.CreateTime)
	promptTemplate.UserInput = []string{fmt.Sprintf("[%s](%s) <%s>: %s", createTime, *event.Event.Sender.SenderId.OpenId, userName, larkutils.PreGetTextMsg(ctx, event))}
	promptTemplate.HistoryRecords = messageList.ToLines()
	if len(promptTemplate.HistoryRecords) > *size {
		promptTemplate.HistoryRecords = promptTemplate.HistoryRecords[:*size]
	}
	docs, err := retriver.Cli.RecallDocs(ctx, chatID, *event.Event.Message.Content, 10)
	if err != nil {
		logging.Logger.Err(err)
	}
	promptTemplate.Context = commonutils.TransSlice(docs, func(doc schema.Document) string {
		if doc.Metadata == nil {
			doc.Metadata = map[string]any{}
		}
		createTime, _ := doc.Metadata["create_time"].(string)
		userID, _ := doc.Metadata["user_id"].(string)
		userName, _ := doc.Metadata["user_name"].(string)
		return fmt.Sprintf("[%s](%s) <%s>: %s", createTime, userID, userName, doc.PageContent)
	})
	promptTemplate.Topics = make([]string, 0)
	for _, doc := range docs {
		msgID, ok := doc.Metadata["msg_id"]
		if ok {
			resp, err := opensearchdal.SearchData(ctx, consts.LarkChunkIndex, osquery.
				Search().Sort("timestamp", osquery.OrderDesc).
				Query(osquery.Bool().Must(osquery.Term("msg_ids", msgID))).
				Size(1),
			)
			if err != nil {
				return nil, err
			}
			chunk := &handlertypes.MessageChunkLogV3{}
			if len(resp.Hits.Hits) > 0 {
				sonic.Unmarshal(resp.Hits.Hits[0].Source, &chunk)
				promptTemplate.Topics = append(promptTemplate.Topics, chunk.Summary)
			}
		}
	}
	promptTemplate.Topics = utility.Dedup(promptTemplate.Topics)
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return nil, err
	}
	iter, err := doubao.ResponseStreaming(ctx, b.String(), modelID, files...)
	if err != nil {
		return
	}
	return func(yield func(*doubao.ModelStreamRespReasoning) bool) {
		mentionMap := make(map[string]string)
		for _, item := range messageList {
			mentionMap[item.UserName] = larkmsgutils.AtUser(item.UserID, item.UserName)
			mentionMap[item.UserID] = larkmsgutils.AtUser(item.UserID, item.UserName)
			for _, mention := range item.MentionList {
				mentionMap[*mention.Name] = larkmsgutils.AtUser(*mention.Id, *mention.Name)
				mentionMap[*mention.Id] = larkmsgutils.AtUser(*mention.Id, *mention.Name)
			}
		}
		memberMap, err := grouputil.GetUserMapFromChatIDCache(ctx, chatID)
		if err != nil {
			return
		}
		for _, member := range memberMap {
			mentionMap[*member.Name] = larkmsgutils.AtUser(*member.MemberId, *member.Name)
			mentionMap[*member.MemberId] = larkmsgutils.AtUser(*member.MemberId, *member.Name)
		}
		trie := BuildTrie(mentionMap)
		lastData := &doubao.ModelStreamRespReasoning{}
		for data := range iter {
			lastData = data
			if !yield(data) {
				return
			}
		}
		lastData.Content = ReplaceMentionsWithTrie(lastData.Content, trie)
		if !yield(lastData) {
			return
		}
	}, err
}

// ReplaceMentionsRawText 在文本中查找所有 @xxx 格式的用户名，并根据提供的 map 进行替换。
// @ 前面必须是文本开头或空格。
// xxx 必须是一个完整的单词（后面是空格、标点符号或文本结尾）。
func ReplaceMentionsRawText(text string, replacements map[string]string) string {
	// 正则表达式: (?:^|\s)@(\w+)\b
	// (?:^|\s)  - 非捕获组，匹配字符串开头或空格
	// @         - 匹配'@'符号
	// (\w+)     - 捕获组，匹配一个或多个单词字符（这是 key）
	// \b        - 单词边界，确保我们匹配的是一个完整的 "单词"
	re := regexp.MustCompile(`@([\p{Han}\w]+)`)

	// FindAllStringSubmatchIndex 会返回所有匹配项的索引位置
	// 每个匹配项是一个数组：[完整匹配开始, 完整匹配结束, 第一个捕获组开始, 第一个捕获组结束]
	matches := re.FindAllStringSubmatchIndex(text, -1)

	// 如果没有找到匹配项，直接返回原文
	if len(matches) == 0 {
		return text
	}

	var builder strings.Builder
	lastIndex := 0

	for _, match := range matches {
		// match[2] 和 match[3] 是捕获组 (\w+) 的开始和结束索引，也就是 'xxx'
		key := text[match[2]:match[3]]

		// 从 map 中查找替换值
		value, found := replacements[key]

		// 找到要替换的整个模式 "@xxx" 的起始位置
		// 注意: 捕获组 'xxx' 的起始位置 (match[2]) 的前一个字符就是 '@'
		mentionStartIndex := match[2] - 1

		// 1. 先将上一个匹配项到当前匹配项之间的文本追加进来
		builder.WriteString(text[lastIndex:mentionStartIndex])

		if found {
			// 2. 如果在 map 中找到了 key，则追加替换后的 value
			builder.WriteString(value)
		} else {
			// 3. 如果没找到，则将原始的 "@xxx" 追加回来
			builder.WriteString(text[mentionStartIndex:match[3]])
		}

		// 4. 更新 lastIndex，为下一次循环做准备
		lastIndex = match[3]
	}

	// 追加最后一个匹配项到字符串末尾的剩余文本
	builder.WriteString(text[lastIndex:])

	return builder.String()
}

// TrieNode 定义了字典树的节点
type TrieNode struct {
	children    map[rune]*TrieNode
	isEndOfWord bool
	replacement string // 在单词结束节点存储替换文本
}

// NewTrieNode 创建一个新的字典树节点
func NewTrieNode() *TrieNode {
	return &TrieNode{
		children:    make(map[rune]*TrieNode),
		isEndOfWord: false,
		replacement: "",
	}
}

// BuildTrie 从一个 map[string]string 构建字典树。
// map 的 key 是要查找的关键词，value 是对应的替换文本。
func BuildTrie(wordList map[string]string) *TrieNode {
	root := NewTrieNode()
	for key, value := range wordList {
		node := root
		// 将关键词（rune切片）插入字典树
		for _, r := range []rune(key) {
			if _, found := node.children[r]; !found {
				node.children[r] = NewTrieNode()
			}
			node = node.children[r]
		}
		// 标记单词结束，并存储替换值
		node.isEndOfWord = true
		node.replacement = value
	}
	return root
}

// ReplaceMentionsWithTrie 使用预构建的字典树来查找并替换文本中的 @mentions。
func ReplaceMentionsWithTrie(text string, root *TrieNode) string {
	if root == nil || len(root.children) == 0 {
		return text
	}

	var builder strings.Builder
	runes := []rune(text) // 使用 rune 以正确处理多字节字符（如中文）
	lastIndex := 0

	for i := 0; i < len(runes); i++ {
		// 寻找 @ 符号
		if runes[i] != '@' {
			continue
		}

		// 从 @ 符号的下一个字符开始在字典树中搜索
		node := root
		matchEndIndex := -1
		var foundReplacement string

		for j := i + 1; j < len(runes); j++ {
			char := runes[j]
			child, found := node.children[char]
			if !found {
				// 如果当前字符在字典树中没有对应的子节点，则停止搜索
				break
			}

			node = child
			// 如果当前节点是一个单词的结尾，记录下这个可能的匹配
			// 我们继续搜索，以找到最长的匹配项
			if node.isEndOfWord {
				matchEndIndex = j
				foundReplacement = node.replacement
			}
		}

		// 如果找到了一个或多个匹配项，matchEndIndex 会记录最长匹配的结束位置
		if matchEndIndex != -1 {
			// 1. 将上一个匹配结束位置到当前 @ 符号前的文本追加进来
			builder.WriteString(string(runes[lastIndex:i]))
			// 2. 追加替换后的文本
			builder.WriteString(foundReplacement)
			// 3. 更新索引，跳过已处理的 @mention
			i = matchEndIndex
			lastIndex = i + 1
		}
	}

	// 4. 追加最后一个匹配项到字符串末尾的剩余文本
	if lastIndex < len(runes) {
		builder.WriteString(string(runes[lastIndex:]))
	}

	return builder.String()
}
