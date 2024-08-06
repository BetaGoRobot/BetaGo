package larkutils

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func ReBuildArgs(argName, argValue string) string {
	if trimmed := strings.Trim(argValue, "\""); trimmed != "" {
		return strings.Join([]string{"--", argName, "=", trimmed}, "")
	} else {
		return strings.Join([]string{"--", argName}, "")
	}
}

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
	atMsgRepattern       = regexp2.MustCompile(`@[^ ]+\s+(?P<content>.+)`, regexp2.RE2)
	commandMsgRepattern  = regexp2.MustCompile(`\/(?P<commands>[^--]+)( --)*`, regexp2.RE2)                                                                   // 只校验是不是合法命令
	commandFullRepattern = regexp2.MustCompile(`((@[^ ]+\s+)|^)\/(?P<commands>\w+( )*)+( )*(--(?P<arg_name>\w+)=(?P<arg_value>("[^"]*"|\S+)))*`, regexp2.RE2) // command+参数格式校验
	commandArgRepattern  = regexp2.MustCompile(`--(?P<arg_name>\w+)(=(?P<arg_value>("[^"]*"|\S+)))?`, regexp2.RE2)                                            // 提取参数
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
	// 校验合法性
	matched, err := commandFullRepattern.MatchString(content)
	if err != nil {
		log.ZapLogger.Error("GetCommand", zaplog.Error(err))
		return
	}
	if !matched {
		return nil
	}

	match, err := commandMsgRepattern.FindStringMatch(content)
	if match.GroupByName("commands") != nil { // 提取command
		commands = strings.Fields(match.GroupByName("commands").String())

		// 转换args
		match, err := commandArgRepattern.FindStringMatch(content)
		if err != nil {
			log.ZapLogger.Error("GetCommand", zaplog.Error(err))
			return
		}
		if match != nil {
			commands = append(commands, ReBuildArgs(
				match.GroupByName("arg_name").String(),
				match.GroupByName("arg_value").String()),
			)
			for {
				newMatch, err := commandArgRepattern.FindNextMatch(match)
				if err != nil {
					panic(err)
				}
				if newMatch == nil {
					break
				}
				match = newMatch
				commands = append(commands, ReBuildArgs(
					match.GroupByName("arg_name").String(),
					match.GroupByName("arg_value").String()),
				)
			}
		}
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

func AddTraceLog2DB(ctx context.Context, msgID string) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("AddTraceLog2DB", zaplog.String("msgID", msgID), zaplog.String("traceID", span.SpanContext().TraceID().String()))
	if result := database.GetDbConnection().Create(&database.MsgTraceLog{
		MsgID:   msgID,
		TraceID: span.SpanContext().TraceID().String(),
	}); result.Error != nil {
		log.ZapLogger.Error("AddTraceLog2DB", zaplog.Error(result.Error))
	}
}

func ReplyMsgRawContentType(ctx context.Context, msgID, msgType, content, suffix string, replyInThread bool) (err error) {
	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			ReplyInThread(replyInThread).
			Uuid(msgID + suffix).Build(),
	).MessageId(msgID).Build()

	resp, err := LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		log.ZapLogger.Error("ReplyMessage", zaplog.Error(err))
		return err
	}
	if resp.Code != 0 {
		log.ZapLogger.Error("ReplyMessage", zaplog.String("Error", resp.Error()))
		return errors.New(resp.Error())
	}
	AddTraceLog2DB(ctx, *resp.Data.MessageId)
	return
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

// ReplyMsgText ReplyMsgText 注意：不要传入已经Build过的文本
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyMsgText(ctx context.Context, text, msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	return ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeText, larkim.NewTextMsgBuilder().Text(text).Build(), suffix, replyInThread)
}

// ReplyMsgTextRaw ReplyMsgTextRaw 注意：必须传入已经Build的文本
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyMsgTextRaw(ctx context.Context, text, msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	return ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeText, text, suffix, replyInThread)
}

// CreateMsgText 不需要自行BuildText
func CreateMsgText(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	// TODO: Add id saving
	return CreateMsgTextRaw(ctx, larkim.NewTextMsgBuilder().Text(content).Build(), msgID, chatID)
}

// CreateMsgTextRaw 需要自行BuildText
func CreateMsgTextRaw(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	// TODO: Add id saving
	resp, err := LarkClient.Im.Message.Create(ctx,
		larkim.NewCreateMessageReqBuilder().
			Body(
				larkim.NewCreateMessageReqBodyBuilder().
					ReceiveId(chatID).
					Content(content).
					Uuid(msgID+"_create").
					MsgType(larkim.MsgTypeText).
					Build(),
			).
			Build(),
	)
	if err != nil {
		log.ZapLogger.Error("CreateMessage", zaplog.Error(err))
		return err
	}
	if resp.Code != 0 {
		log.ZapLogger.Error("CreateMessage", zaplog.String("Error", resp.Error()))
		return errors.New(resp.Error())
	}
	AddTraceLog2DB(ctx, *resp.Data.MessageId)
	return
}
