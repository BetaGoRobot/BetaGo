package larkutils

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// PreGetTextMsg 获取消息内容
//
//	@param ctx
//	@param event
//	@return string
func PreGetTextMsg(ctx context.Context, event *larkim.P2MessageReceiveV1) string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	msgMap := make(map[string]interface{})
	msg := *event.Event.Message.Content
	err := sonic.UnmarshalString(msg, &msgMap)
	if err != nil {
		log.ZapLogger.Error("repeatMessage", zaplog.Error(err))
		return ""
	}
	if text, ok := msgMap["text"]; ok {
		msg = text.(string)
	}
	return msg
}

var (
	atMsgRepattern      = regexp2.MustCompile(`@[^ ]+\s+(?P<content>.+)`, regexp2.RE2)
	commandMsgRepattern = regexp2.MustCompile(`((@[^ ]+\s+)|^)\/(?P<content>.+)`, regexp2.RE2)
)

// TrimAtMsg trim掉at的消息
//
//	@param ctx context.Context
//	@param msg string
//	@return string
//	@author heyuhengmatt
//	@update 2024-07-17 01:39:05
func TrimAtMsg(ctx context.Context, msg string) string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	match, err := atMsgRepattern.FindStringMatch(msg)
	if err != nil {
		return msg
	}
	if match != nil && match.Length > 0 {
		return match.GroupByName("content").String()
	}
	return msg
}

func IsMentioned(mentions []*larkim.MentionEvent) bool {
	for _, mention := range mentions {
		if *mention.Id.OpenId == consts.BotOpenID {
			return true
		}
	}
	return false
}

func GetMsgByID(ctx context.Context, msgID string) string {
	resp, err := LarkClient.Im.V1.Message.Get(ctx, larkim.NewGetMessageReqBuilder().MessageId(msgID).Build())
	if err != nil {
		log.ZapLogger.Error("GetMsgByID", zaplog.Error(err))
	}
	return *resp.Data.Items[0].Body.Content
}

func GetMsgFullByID(ctx context.Context, msgID string) *larkim.GetMessageResp {
	resp, err := LarkClient.Im.V1.Message.Get(ctx, larkim.NewGetMessageReqBuilder().MessageId(msgID).Build())
	if err != nil {
		log.ZapLogger.Error("GetMsgByID", zaplog.Error(err))
	}
	return resp
}

func GetCommandWithMatched(ctx context.Context, content string) (commands []string, isCommand bool) {
	if IsCommand(ctx, content) {
		isCommand = true
		match, err := commandMsgRepattern.FindStringMatch(content)
		if err != nil {
			log.ZapLogger.Error("GetCommand", zaplog.Error(err))
			return
		}
		if match.GroupByName("content") != nil {
			commands = strings.Fields(strings.TrimLeft(match.GroupByName("content").String(), "/"))
		}
	}

	return
}

func GetCommand(ctx context.Context, content string) (commands []string) {
	match, err := commandMsgRepattern.FindStringMatch(content)
	if err != nil {
		log.ZapLogger.Error("GetCommand", zaplog.Error(err))
		return
	}
	if match.GroupByName("content") != nil {
		commands = strings.Fields(strings.ToLower(strings.TrimLeft(match.GroupByName("content").String(), "/")))
	}

	return
}

func IsCommand(ctx context.Context, content string) bool {
	content = strings.Trim(content, " ")
	matched, err := commandMsgRepattern.MatchString(content)
	if err != nil {
		log.ZapLogger.Error("GetCommand", zaplog.Error(err))
		return matched
	}
	return matched
}

func SendMsgRawContentType(ctx context.Context, msgID, msgType, content string, replyInThread bool) error {
	stickerReq := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			ReplyInThread(replyInThread).
			Uuid(msgID + "_reply").Build(),
	).MessageId(msgID).Build()

	resp, err := LarkClient.Im.V1.Message.Reply(ctx, stickerReq)
	if err != nil {
		log.ZapLogger.Error("ReplyMessage", zaplog.Error(err))
		return err
	}
	if resp.Code != 0 {
		log.ZapLogger.Error("ReplyMessage", zaplog.String("Error", resp.Error()))
		return errors.New(resp.Error())
	}
	return nil
}

func GetMsgImages(ctx context.Context, msgID, fileKey, fileType string) (file io.Reader, err error) {
	req := larkim.NewGetMessageResourceReqBuilder().MessageId(msgID).FileKey(fileKey).Type(fileType).Build()
	resp, err := LarkClient.Im.MessageResource.Get(ctx, req)
	if err != nil {
		log.ZapLogger.Error("GetMsgImages", zaplog.Error(err))
		return nil, err
	}
	if resp.Code != 0 {
		log.ZapLogger.Error("GetMsgImages", zaplog.String("Error", resp.Error()))
		return nil, errors.New(resp.Error())
	}
	return resp.File, nil
}
