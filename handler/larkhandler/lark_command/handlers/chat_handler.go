package handlers

import (
	"context"
	"errors"
	"fmt"
	"iter"
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
		// // 先reply一条，占位。
		// resp, err := larkutils.ReplyMsgText(
		// 	ctx, "我想想应该怎么说...", *event.Event.Message.MessageId, "_chat_random", false,
		// )
		// if err != nil {
		// 	return err
		// }
		// if !resp.Success() {
		// 	return errors.New(resp.Error())
		// }
		// lastMsgID := *resp.Data.MessageId
		lastData := &doubao.ModelStreamRespReasoning{}
		// idx := 0
		// replyMsg := func(content string) {
		// 	idx++
		// 	resp, err := larkutils.ReplyMsgText(
		// 		ctx, content, lastMsgID, strconv.Itoa(idx), true,
		// 	)
		// 	if err != nil {
		// 		logging.Logger.Err(errors.New(err.Error()))
		// 		return
		// 	}
		// 	if !resp.Success() {
		// 		logging.Logger.Err(errors.New(resp.Error()))
		// 		return
		// 	}
		// }
		refernce := ""
		for data := range res {
			eot := "**回复:**"
			sor := "\n参考资料:"
			span.SetAttributes(attribute.String("lastData", data.Content))
			if idx := strings.Index(data.Content, eot); idx != -1 {
				lastData = data
				lastData.Content = strings.TrimSpace(lastData.Content[idx+len(eot):])
				if idx := strings.Index(data.Content, sor); idx != -1 {
					refernce = strings.TrimSpace(data.Content[idx:])
					lastData.Content = strings.TrimSpace(lastData.Content[:idx])
					lastData.Content += "<i>(*包含在线检索内容)</i>"
					span.SetAttributes(attribute.String("refernce", refernce))
				}
			}

			if strings.Contains(lastData.Content, "[无需回复]") {
				return err
			}
			if data.Reply2Show != nil {
				span.SetAttributes(attribute.String("reply2Show"+"."+data.Reply2Show.ID, data.Reply2Show.Content))
				// replyMsg(data.Reply2Show.Content)
			}
		}
		// if refernce != "" {
		// 	replyMsg(refernce)
		// }
		// err = larkutils.UpdateMessageText(ctx, lastMsgID, lastData.Content)
		// if err != nil {
		// 	return err
		// }

		resp, err := larkutils.ReplyMsgText(
			ctx, lastData.Content, *event.Event.Message.MessageId, "_chat_random", false,
		)
		if err != nil {
			return err
		}
		if !resp.Success() {
			return errors.New(resp.Error())
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
		promptTemplate.HistoryRecords = promptTemplate.HistoryRecords[len(promptTemplate.HistoryRecords)-*size:]
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

	iter, err := doubao.ResponseStreaming(ctx, b.String(), modelID, chatID, files...)
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
		trie := utility.BuildTrie(mentionMap)
		lastData := &doubao.ModelStreamRespReasoning{}
		for data := range iter {
			lastData = data
			if !yield(data) {
				return
			}
		}
		lastData.Content = trie.ReplaceMentionsWithTrie(lastData.Content)
		if !yield(lastData) {
			return
		}
	}, err
}
