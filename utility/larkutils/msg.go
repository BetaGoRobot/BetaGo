package larkutils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/go_utils/reflecting"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	return getContentFromTextMsg(*event.Event.Message.Content)
}

func getContentFromTextMsg(s string) string {
	msgMap := make(map[string]interface{})
	err := sonic.UnmarshalString(s, &msgMap)
	if err != nil {
		log.Zlog.Error("repeatMessage", zaplog.Error(err))
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
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
		log.Zlog.Error("GetMsgByID", zaplog.Error(err))
	}
	return *resp.Data.Items[0].Body.Content
}

func GetMsgFullByID(ctx context.Context, msgID string) *larkim.GetMessageResp {
	resp, err := LarkClient.Im.V1.Message.Get(ctx, larkim.NewGetMessageReqBuilder().MessageId(msgID).Build())
	if err != nil {
		log.Zlog.Error("GetMsgByID", zaplog.Error(err))
	}
	return resp
}

func GetCommandWithMatched(ctx context.Context, content string) (commands []string, isCommand bool) {
	if IsCommand(ctx, content) {
		isCommand = true
		match, err := commandMsgRepattern.FindStringMatch(content)
		if err != nil {
			log.Zlog.Error("GetCommand", zaplog.Error(err))
			return
		}
		if match.GroupByName("content") != nil {
			commands = strings.Fields(strings.TrimPrefix(match.GroupByName("content").String(), "/"))
		}
	}

	return
}

func GetCommand(ctx context.Context, content string) (commands []string) {
	// 校验合法性
	matched, err := commandFullRepattern.MatchString(content)
	if err != nil {
		log.Zlog.Error("GetCommand", zaplog.Error(err))
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
			log.Zlog.Error("GetCommand", zaplog.Error(err))
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
		log.Zlog.Error("GetCommand", zaplog.Error(err))
		return matched
	}
	return matched
}

func AddReaction2DB(ctx context.Context, msgID string) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	log.Zlog.Info("AddTraceLog2DB", zaplog.String("msgID", msgID), zaplog.String("traceID", span.SpanContext().TraceID().String()))
	if result := database.GetDbConnection().Create(&database.MsgTraceLog{
		MsgID:   msgID,
		TraceID: span.SpanContext().TraceID().String(),
	}); result.Error != nil {
		log.Zlog.Error("AddTraceLog2DB", zaplog.Error(result.Error))
	}
}

func ReplyMsgRawContentType(ctx context.Context, msgID, msgType, content, suffix string, replyInThread bool) (resp *larkim.ReplyMessageResp, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("msgType").String(msgType), attribute.Key("content").String(content))
	defer span.End()
	uuid := (msgID + suffix)
	if len(uuid) > 50 {
		uuid = uuid[:50]
	}

	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(NewTextMsgBuilder().Text(content).Build()).
			ReplyInThread(replyInThread).
			Uuid(GenUUIDStr(uuid, 50)).Build(),
	).MessageId(msgID).Build()

	resp, err = LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		log.Zlog.Error("ReplyMessage", zaplog.Error(err))
		return nil, err
	}
	if !resp.Success() {
		log.Zlog.Error("ReplyMessage", zaplog.String("Error", larkcore.Prettify(resp.CodeError.Err)))
		return nil, errors.New(resp.Error())
	}
	RecordReplyMessage2Opensearch(ctx, resp, content)
	return
}

// ReplyMsgText ReplyMsgText 注意：不要传入已经Build过的文本
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyMsgText(ctx context.Context, text, msgID, suffix string, replyInThread bool) (resp *larkim.ReplyMessageResp, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(text))
	defer span.End()
	return ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeText, text, suffix, replyInThread)
}

func RecordMessage2Opensearch(ctx context.Context, resp *larkim.CreateMessageResp, contents ...string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	var content string
	if len(contents) > 0 {
		content = strings.Join(contents, "\n")
	} else {
		content = getContentFromTextMsg(utility.AddressORNil(resp.Data.Body.Content))
	}
	msgLog := &handlertypes.MessageLog{
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
		log.Zlog.Error("EmbeddingText error", zaplog.Error(err))
	}
	// jieba := gojieba.NewJieba()
	// defer jieba.Free()
	// ws := jieba.Cut(content, true)
	err = opensearchdal.InsertData(ctx, consts.LarkMsgIndex,
		utility.AddressORNil(resp.Data.MessageId),
		&handlertypes.MessageIndex{
			MessageLog: msgLog,
			ChatName:   GetChatName(ctx, utility.AddressORNil(resp.Data.ChatId)),
			RawMessage: content,
			// RawMessageJieba: strings.Join(ws, " "),
			CreateTime: utility.EpoMil2DateStr(*resp.Data.CreateTime),
			Message:    embedded,
			UserID:     "你",
			UserName:   "你",
			TokenUsage: usage,
		},
	)
	if err != nil {
		log.Zlog.Error("InsertData", zaplog.Error(err))
		return
	}
}

func RecordCardAction2Opensearch(ctx context.Context, cardAction *callback.CardActionTriggerEvent) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	chatID := cardAction.Event.Context.OpenChatID
	userID := cardAction.Event.Operator.OpenID
	member, err := GetUserMemberFromChat(ctx, chatID, userID)
	if err != nil {
		return
	}
	idxData := &handlertypes.CardActionIndex{
		CardActionTriggerEvent: cardAction,
		ChatName:               GetChatName(ctx, userID),
		CreateTime:             utility.EpoMicro2DateStr(cardAction.EventV2Base.Header.CreateTime),
		UserID:                 cardAction.Event.Operator.OpenID,
		UserName:               utility.AddressORNil(member.Name),
		ActionValue:            cardAction.Event.Action.Value,
	}
	fmt.Println(sonic.MarshalString(idxData))
	err = opensearchdal.InsertData(ctx,
		consts.LarkCardActionIndex,
		cardAction.Event.Operator.OpenID,
		idxData,
	)
	if err != nil {
		log.Zlog.Error("InsertData", zaplog.Error(err))
		return
	}
}

func RecordReplyMessage2Opensearch(ctx context.Context, resp *larkim.ReplyMessageResp, contents ...string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	var content string
	if len(contents) > 0 {
		content = strings.Join(contents, "\n")
	} else {
		content = getContentFromTextMsg(utility.AddressORNil(resp.Data.Body.Content))
	}
	msgLog := &handlertypes.MessageLog{
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
		log.Zlog.Error("EmbeddingText error", zaplog.Error(err))
	}
	// jieba := gojieba.NewJieba()
	// defer jieba.Free()
	// ws := jieba.Cut(content, true)

	err = opensearchdal.InsertData(ctx, consts.LarkMsgIndex, utility.AddressORNil(resp.Data.MessageId),
		&handlertypes.MessageIndex{
			MessageLog: msgLog,
			ChatName:   GetChatName(ctx, utility.AddressORNil(resp.Data.ChatId)),
			RawMessage: content,
			// RawMessageJieba: strings.Join(ws, " "),
			CreateTime: utility.EpoMil2DateStr(*resp.Data.CreateTime),
			Message:    embedded,
			UserID:     "你",
			UserName:   "你",
			TokenUsage: usage,
		},
	)
	if err != nil {
		log.Zlog.Error("InsertData", zaplog.Error(err))
		return
	}
}

// CreateMsgText 不需要自行BuildText
func CreateMsgText(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(content))
	defer span.End()

	// TODO: Add id saving
	return CreateMsgTextRaw(ctx, NewTextMsgBuilder().Text(content).Build(), msgID, chatID)
}

// CreateMsgTextRaw 需要自行BuildText
func CreateMsgTextRaw(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
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
					Uuid(GenUUIDStr(uuid, 50)).
					MsgType(larkim.MsgTypeText).
					Build(),
			).
			Build(),
	)
	if err != nil {
		log.Zlog.Error("CreateMessage", zaplog.Error(err))
		return err
	}
	if !resp.Success() {
		log.Zlog.Error("CreateMessage", zaplog.String("Error", resp.Error()))
		return errors.New(resp.Error())
	}
	RecordMessage2Opensearch(ctx, resp)
	return
}

func AddReaction(ctx context.Context, reactionType, msgID string) (reactionID string, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()

	req := larkim.NewCreateMessageReactionReqBuilder().Body(larkim.NewCreateMessageReactionReqBodyBuilder().ReactionType(larkim.NewEmojiBuilder().EmojiType(reactionType).Build()).Build()).MessageId(msgID).Build()
	resp, err := LarkClient.Im.V1.MessageReaction.Create(ctx, req)
	if err != nil {
		log.Zlog.Error("AddReaction", zaplog.Error(err))
		return "", err
	}
	if !resp.Success() {
		log.Zlog.Error("AddReaction", zaplog.String("Error", resp.Error()))
		return "", errors.New(resp.Error())
	}
	AddReaction2DB(ctx, msgID)
	return *resp.Data.ReactionId, err
}

func AddReactionAsync(ctx context.Context, reactionType, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()

	req := larkim.NewCreateMessageReactionReqBuilder().Body(larkim.NewCreateMessageReactionReqBodyBuilder().ReactionType(larkim.NewEmojiBuilder().EmojiType(reactionType).Build()).Build()).MessageId(msgID).Build()
	go func() {
		resp, err := LarkClient.Im.V1.MessageReaction.Create(ctx, req)
		if err != nil {
			log.Zlog.Error("AddReaction", zaplog.Error(err))
			return
		}
		if !resp.Success() {
			log.Zlog.Error("AddReaction", zaplog.String("Error", resp.Error()))
			return
		}
		AddReaction2DB(ctx, msgID)
	}()
	return nil
}

func RemoveReaction(ctx context.Context, reactionID, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	req := larkim.NewDeleteMessageReactionReqBuilder().MessageId(msgID).ReactionId(reactionID).Build()
	resp, err := LarkClient.Im.V1.MessageReaction.Delete(ctx, req)
	if err != nil {
		log.Zlog.Error("RemoveReaction", zaplog.Error(err))
		return err
	}
	if !resp.Success() {
		log.Zlog.Error("RemoveReaction", zaplog.String("Error", resp.Error()))
		return errors.New(resp.Error())
	}
	AddReaction2DB(ctx, msgID)
	return
}
