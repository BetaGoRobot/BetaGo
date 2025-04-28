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
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/message"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	commonutils "github.com/BetaGoRobot/go_utils/common_utils"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/defensestation/osquery"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func ChatHandler(chatType string) func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
	return func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
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
	urlSeq, err := larkutils.GetAllImgURLFromMsg(ctx, *event.Event.Message.MessageId)
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
		lastData := &doubao.ModelStreamRespReasoning{}
		for data := range res {
			lastData = data
		}
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

	// 默认获取最近20条消息
	if size == nil {
		size = new(int)
		*size = 20
	}
	chatID := *event.Event.Message.ChatId
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("chat_id", chatID),
			),
		).
		SourceIncludes("raw_message", "mentions", "create_time", "user_id", "chat_id", "user_name").
		Size(uint64(*size*3)).
		Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(
		context.Background(),
		"lark_msg_index",
		query)
	if err != nil {
		panic(err)
	}
	messageList := FilterMessage(resp.Hits.Hits, *size)

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
	member, err := larkutils.GetUserMemberFromChat(ctx, chatID, *event.Event.Sender.SenderId.OpenId)
	if err != nil {
		return
	}
	userName := ""
	if member == nil {
		userName = "NULL"
	} else {
		userName = *member.Name
	}

	promptTemplate.UserInput = commonutils.TransSliceWithSkip(input, func(s string) (string, bool) {
		if strings.TrimSpace(s) != "" {
			return fmt.Sprintf("[%s] <%s>: %s", utility.EpoMil2DateStr(*event.Event.Message.CreateTime), userName, strings.ReplaceAll(s, "\n", "\\n")), false
		}
		return "", true
	})
	promptTemplate.HistoryRecords = messageList
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return nil, err
	}
	res, err = doubao.SingleChatStreamingPrompt(ctx, b.String(), modelID, files...)
	if err != nil {
		return
	}
	return
}
