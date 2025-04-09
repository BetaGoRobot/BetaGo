package larkutils

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
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
	return getContentFromTextMsg(*event.Event.Message.Content)
}

func getContentFromTextMsg(s string) string {
	msgMap := make(map[string]interface{})
	err := sonic.UnmarshalString(s, &msgMap)
	if err != nil {
		log.ZapLogger.Error("repeatMessage", zaplog.Error(err))
		return ""
	}
	if text, ok := msgMap["text"]; ok {
		s = text.(string)
	}
	return s
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
			lastIdx := 0
			for match, err = commandArgRepattern.FindStringMatch(content); match != nil; {
				lastIdx = match.Index + len(match.String()) + 1
				commands = append(commands, ReBuildArgs(
					match.GroupByName("arg_name").String(),
					match.GroupByName("arg_value").String()),
				)
				if err != nil {
					panic(err)
				}
				match, err = commandArgRepattern.FindNextMatch(match)
			}
			if lastIdx < len(content) {
				commands = append(commands, content[lastIdx:])
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
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("msgType").String(msgType), attribute.Key("content").String(content))
	defer span.End()
	uuid := (msgID + suffix)
	if len(uuid) > 50 {
		uuid = uuid[:50]
	}
	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			ReplyInThread(replyInThread).
			Uuid(uuid).Build(),
	).MessageId(msgID).Build()

	resp, err := LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		log.ZapLogger.Error("ReplyMessage", zaplog.Error(err))
		return err
	}
	if resp.CodeError.Code != 0 {
		log.ZapLogger.Error("ReplyMessage", zaplog.String("Error", larkcore.Prettify(resp.CodeError.Err)))
		return errors.New(resp.Error())
	}
	AddTraceLog2DB(ctx, *resp.Data.MessageId)
	RecordReplyMessage2Opensearch(ctx, resp)
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
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(text))
	defer span.End()
	text = strings.ReplaceAll(text, "\n", "\\n")
	return ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeText, NewTextMsgBuilder().Text(text).Build(), suffix, replyInThread)
}

// ReplyMsgTextRaw ReplyMsgTextRaw 注意：必须传入已经Build的文本
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyMsgTextRaw(ctx context.Context, text, msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(text))
	defer span.End()

	return ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeText, text, suffix, replyInThread)
}

func RecordMessage2Opensearch(ctx context.Context, resp *larkim.CreateMessageResp) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	content := getContentFromTextMsg(utility.AddressORNil(resp.Data.Body.Content))
	msgLog := &database.MessageLog{
		MessageID:   utility.AddressORNil(resp.Data.MessageId),
		RootID:      utility.AddressORNil(resp.Data.RootId),
		ParentID:    utility.AddressORNil(resp.Data.ParentId),
		ChatID:      utility.AddressORNil(resp.Data.ChatId),
		ThreadID:    utility.AddressORNil(resp.Data.ThreadId),
		ChatType:    "",
		MessageType: utility.AddressORNil(resp.Data.MsgType),
		UserAgent:   "",
		Mentions:    utility.MustMashal(resp.Data.Mentions),
		RawBody:     utility.MustMashal(resp),
		Content:     content,
		TraceID:     span.SpanContext().TraceID().String(),
	}
	database.GetDbConnection().Create(msgLog)
	embedded, usage, err := doubao.EmbeddingText(ctx, utility.AddressORNil(resp.Data.Body.Content))
	if err != nil {
		log.ZapLogger.Error("EmbeddingText error", zaplog.Error(err))
	}
	err = opensearchdal.InsertData(ctx, "lark_msg_index", utility.AddressORNil(resp.Data.MessageId),
		&handlertypes.MessageIndex{
			MessageLog: msgLog,
			ChatName:   GetChatName(ctx, utility.AddressORNil(resp.Data.ChatId)),
			RawMessage: content,
			CreateTime: utility.EpoMil2DateStr(*resp.Data.CreateTime),
			Message:    embedded,
			UserID:     "你",
			UserName:   "你",
			TokenUsage: usage,
		},
	)
	if err != nil {
		log.ZapLogger.Error("InsertData", zaplog.Error(err))
		return
	}
}

func RecordReplyMessage2Opensearch(ctx context.Context, resp *larkim.ReplyMessageResp) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	content := getContentFromTextMsg(utility.AddressORNil(resp.Data.Body.Content))
	msgLog := &database.MessageLog{
		MessageID:   utility.AddressORNil(resp.Data.MessageId),
		RootID:      utility.AddressORNil(resp.Data.RootId),
		ParentID:    utility.AddressORNil(resp.Data.ParentId),
		ChatID:      utility.AddressORNil(resp.Data.ChatId),
		ThreadID:    utility.AddressORNil(resp.Data.ThreadId),
		ChatType:    "",
		MessageType: utility.AddressORNil(resp.Data.MsgType),
		UserAgent:   "",
		Mentions:    utility.MustMashal(resp.Data.Mentions),
		RawBody:     utility.MustMashal(resp),
		Content:     content,
		TraceID:     span.SpanContext().TraceID().String(),
	}

	embedded, usage, err := doubao.EmbeddingText(ctx, utility.AddressORNil(resp.Data.Body.Content))
	if err != nil {
		log.ZapLogger.Error("EmbeddingText error", zaplog.Error(err))
	}
	err = opensearchdal.InsertData(ctx, "lark_msg_index", utility.AddressORNil(resp.Data.MessageId),
		&handlertypes.MessageIndex{
			MessageLog: msgLog,
			ChatName:   GetChatName(ctx, utility.AddressORNil(resp.Data.ChatId)),
			RawMessage: content,
			CreateTime: utility.EpoMil2DateStr(*resp.Data.CreateTime),
			Message:    embedded,
			UserID:     "你",
			UserName:   "你",
			TokenUsage: usage,
		},
	)
	if err != nil {
		log.ZapLogger.Error("InsertData", zaplog.Error(err))
		return
	}
}

// CreateMsgText 不需要自行BuildText
func CreateMsgText(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(content))
	defer span.End()

	content = strings.ReplaceAll(content, "\n", "\\n")
	// TODO: Add id saving
	return CreateMsgTextRaw(ctx, NewTextMsgBuilder().Text(content).Build(), msgID, chatID)
}

// CreateMsgTextRaw 需要自行BuildText
func CreateMsgTextRaw(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(content))
	defer span.End()
	// TODO: Add id saving
	uuid := (msgID + "_create")
	if len(uuid) > 50 {
		uuid = uuid[:50]
	}
	resp, err := LarkClient.Im.Message.Create(ctx,
		larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(larkim.ReceiveIdTypeChatId).
			Body(
				larkim.NewCreateMessageReqBodyBuilder().
					ReceiveId(chatID).
					Content(content).
					Uuid(uuid).
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
	RecordMessage2Opensearch(ctx, resp)
	return
}

func AddReaction(ctx context.Context, reactionType, msgID string) (reactionID string, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()

	req := larkim.NewCreateMessageReactionReqBuilder().Body(larkim.NewCreateMessageReactionReqBodyBuilder().ReactionType(larkim.NewEmojiBuilder().EmojiType(reactionType).Build()).Build()).MessageId(msgID).Build()
	resp, err := LarkClient.Im.V1.MessageReaction.Create(ctx, req)
	if err != nil {
		log.ZapLogger.Error("AddReaction", zaplog.Error(err))
		return "", err
	}
	if resp.Code != 0 {
		log.ZapLogger.Error("AddReaction", zaplog.String("Error", resp.Error()))
		return "", errors.New(resp.Error())
	}
	AddTraceLog2DB(ctx, msgID)
	return *resp.Data.ReactionId, err
}

func AddReactionAsync(ctx context.Context, reactionType, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()

	req := larkim.NewCreateMessageReactionReqBuilder().Body(larkim.NewCreateMessageReactionReqBodyBuilder().ReactionType(larkim.NewEmojiBuilder().EmojiType(reactionType).Build()).Build()).MessageId(msgID).Build()
	go func() {
		resp, err := LarkClient.Im.V1.MessageReaction.Create(ctx, req)
		if err != nil {
			log.ZapLogger.Error("AddReaction", zaplog.Error(err))
			return
		}
		if resp.Code != 0 {
			log.ZapLogger.Error("AddReaction", zaplog.String("Error", resp.Error()))
			return
		}
		AddTraceLog2DB(ctx, msgID)
	}()
	return nil
}

func RemoveReaction(ctx context.Context, reactionID, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	req := larkim.NewDeleteMessageReactionReqBuilder().MessageId(msgID).ReactionId(reactionID).Build()
	resp, err := LarkClient.Im.V1.MessageReaction.Delete(ctx, req)
	if err != nil {
		log.ZapLogger.Error("RemoveReaction", zaplog.Error(err))
		return err
	}
	if resp.Code != 0 {
		log.ZapLogger.Error("RemoveReaction", zaplog.String("Error", resp.Error()))
		return errors.New(resp.Error())
	}
	AddTraceLog2DB(ctx, msgID)
	return
}
