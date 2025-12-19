package larkutils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/lark"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/ark/embedding"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkchunking"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/userutil"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/retriver"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/tmc/langchaingo/schema"
	"github.com/yanyiwu/gojieba"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
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
	rawContent := getContentFromTextMsg(*event.Event.Message.Content)
	if len(event.Event.Message.Mentions) > 0 {
		for _, mention := range event.Event.Message.Mentions {
			rawContent = strings.ReplaceAll(rawContent, *mention.Key, fmt.Sprintf("@%s", *mention.Name))
		}
	}
	return rawContent
}

func getContentFromTextMsg(s string) string {
	msgMap := make(map[string]interface{})
	err := sonic.UnmarshalString(s, &msgMap)
	if err != nil {
		logs.L().Error("repeatMessage", zap.Error(err))
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
	resp, err := lark.LarkClient.Im.V1.Message.Get(ctx, larkim.NewGetMessageReqBuilder().MessageId(msgID).Build())
	if err != nil {
		logs.L().Ctx(ctx).Error("GetMsgByID", zap.Error(err))
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("GetMsgByID", zap.String("error", resp.Error()))
	}
	return *resp.Data.Items[0].Body.Content
}

func GetMsgFullByID(ctx context.Context, msgID string) *larkim.GetMessageResp {
	resp, err := lark.LarkClient.Im.V1.Message.Get(ctx, larkim.NewGetMessageReqBuilder().MessageId(msgID).Build())
	if err != nil {
		logs.L().Ctx(ctx).Error("GetMsgByID", zap.Error(err))
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("GetMsgByID", zap.String("error", resp.Error()))
	}
	return resp
}

func GetCommandWithMatched(ctx context.Context, content string) (commands []string, isCommand bool) {
	if IsCommand(ctx, content) {
		isCommand = true
		match, err := commandMsgRepattern.FindStringMatch(content)
		if err != nil {
			logs.L().Ctx(ctx).Error("GetCommand", zap.Error(err))
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
		logs.L().Ctx(ctx).Error("GetCommand", zap.Error(err))
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
			logs.L().Ctx(ctx).Error("GetCommand", zap.Error(err))
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
		logs.L().Ctx(ctx).Error("GetCommand", zap.Error(err))
		return matched
	}
	return matched
}

func AddTrace2DB(ctx context.Context, msgID string) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	logs.L().Ctx(ctx).Info("AddTraceLog2DB",
		zap.String("msgID", msgID),
		zap.String("traceID", span.SpanContext().TraceID().String()),
	)
	if result := database.GetDbConnection().Create(&database.MsgTraceLog{
		MsgID:   msgID,
		TraceID: span.SpanContext().TraceID().String(),
	}); result.Error != nil {
		logs.L().Ctx(ctx).Error("AddTraceLog2DB", zap.Error(result.Error))
	}
}

func ReplyMsgRawAsText(ctx context.Context, msgID, msgType, content, suffix string, replyInThread bool) (resp *larkim.ReplyMessageResp, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("msgType").String(msgType), attribute.Key("content").String(content))
	defer span.End()
	defer func() { span.RecordError(err) }()
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

	resp, err = lark.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		logs.L().Ctx(ctx).Error("ReplyMessage", zap.Error(err))
		return nil, err
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("ReplyMessage", zap.String("Error", larkcore.Prettify(resp.CodeError.Err)))
		return nil, errors.New(resp.Error())
	}
	go RecordReplyMessage2Opensearch(ctx, resp, content)
	return
}

func ReplyMsgRawContentType(ctx context.Context, msgID, msgType, content, suffix string, replyInThread bool) (resp *larkim.ReplyMessageResp, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("msgType").String(msgType), attribute.Key("content").String(content))
	defer span.End()
	defer func() { span.RecordError(err) }()
	uuid := (msgID + suffix)
	if len(uuid) > 50 {
		uuid = uuid[:50]
	}

	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			ReplyInThread(replyInThread).
			Uuid(GenUUIDStr(uuid, 50)).Build(),
	).MessageId(msgID).Build()

	resp, err = lark.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		logs.L().Ctx(ctx).Error("ReplyMessage", zap.Error(err))
		return nil, err
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("ReplyMessage", zap.String("Error", larkcore.Prettify(resp.CodeError.Err)))
		return nil, errors.New(resp.Error())
	}
	go RecordReplyMessage2Opensearch(ctx, resp, content)
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
	defer func() { span.RecordError(err) }()
	return ReplyMsgRawAsText(ctx, msgID, larkim.MsgTypeText, text, suffix, replyInThread)
}

func RecordMessage2Opensearch(ctx context.Context, resp *larkim.CreateMessageResp, contents ...string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	defer larkchunking.M.SubmitMessage(ctx, &larkchunking.LarkMessageRespCreate{resp})

	var content string
	if len(contents) > 0 {
		content = strings.Join(contents, "\n")
	} else {
		content = getContentFromTextMsg(utility.AddressORNil(resp.Data.Body.Content))
	}

	config, _ := database.FindByCacheFunc(
		database.PrivateMode{ChatID: utility.AddressORNil(resp.Data.ChatId)},
		func(d database.PrivateMode) string { return d.ChatID },
	)
	if len(config) > 0 && config[0].Enable {
		// 隐私模式，不存了
		logs.L().Ctx(ctx).Info("ChatID hit private config, will not record data...",
			zap.String("chat_id", utility.AddressORNil(resp.Data.ChatId)),
		)
		return
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
	embedded, usage, err := embedding.EmbeddingText(ctx, utility.AddressORNil(resp.Data.Body.Content))
	if err != nil {
		logs.L().Ctx(ctx).Error("EmbeddingText error", zap.Error(err))
	}
	jieba := gojieba.NewJieba()
	defer jieba.Free()
	ws := jieba.Cut(content, true)

	err = opensearchdal.InsertData(ctx, consts.LarkMsgIndex,
		utility.AddressORNil(resp.Data.MessageId),
		&handlertypes.MessageIndex{
			MessageLog:      msgLog,
			ChatName:        GetChatName(ctx, utility.AddressORNil(resp.Data.ChatId)),
			RawMessage:      content,
			RawMessageJieba: strings.Join(ws, " "),
			CreateTime:      utility.Epo2DateZoneMil(utility.MustInt(*resp.Data.CreateTime), time.UTC, time.DateTime),
			CreateTimeV2:    utility.Epo2DateZoneMil(utility.MustInt(*resp.Data.CreateTime), utility.UTCPlus8Loc(), time.RFC3339),
			Message:         embedded,
			UserID:          "你",
			UserName:        "你",
			TokenUsage:      usage,
		},
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("InsertData", zap.Error(err))
		return
	}
	err = retriver.Cli().AddDocuments(ctx, utility.AddressORNil(resp.Data.ChatId),
		[]schema.Document{{
			PageContent: content,
			Metadata: map[string]any{
				"chat_id":     utility.AddressORNil(resp.Data.ChatId),
				"user_id":     utility.AddressORNil(resp.Data.Sender.Id),
				"msg_id":      utility.AddressORNil(resp.Data.MessageId),
				"create_time": utility.EpoMil2DateStr(*resp.Data.CreateTime),
				"user_name":   "你",
			},
		}},
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("AddDocuments error", zap.Error(err))
	}
}

func RecordCardAction2Opensearch(ctx context.Context, cardAction *callback.CardActionTriggerEvent) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	chatID := cardAction.Event.Context.OpenChatID
	userID := cardAction.Event.Operator.OpenID
	userInfo, err := userutil.GetUserInfoCache(ctx, cardAction.Event.Context.OpenChatID, userID)
	if err != nil {
		logs.L().Ctx(ctx).Error("GetUserInfo error", zap.Error(err))
		return
	}
	idxData := &handlertypes.CardActionIndex{
		CardActionTriggerEvent: cardAction,
		ChatName:               GetChatName(ctx, chatID),
		CreateTime:             utility.EpoMicro2DateStr(cardAction.EventV2Base.Header.CreateTime),
		UserID:                 userID,
		UserName:               utility.AddressORNil(userInfo.Name),
		ActionValue:            cardAction.Event.Action.Value,
	}
	err = opensearchdal.InsertData(ctx,
		consts.LarkCardActionIndex,
		cardAction.Event.Operator.OpenID,
		idxData,
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("InsertData", zap.Error(err))
		return
	}
}

func RecordReplyMessage2Opensearch(ctx context.Context, resp *larkim.ReplyMessageResp, contents ...string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	defer larkchunking.M.SubmitMessage(ctx, &larkchunking.LarkMessageRespReply{resp})
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

	embedded, usage, err := embedding.EmbeddingText(ctx, utility.AddressORNil(resp.Data.Body.Content))
	if err != nil {
		logs.L().Ctx(ctx).Error("EmbeddingText error", zap.Error(err))
	}
	jieba := gojieba.NewJieba()
	defer jieba.Free()
	ws := jieba.Cut(content, true)

	err = opensearchdal.InsertData(ctx, consts.LarkMsgIndex, utility.AddressORNil(resp.Data.MessageId),
		&handlertypes.MessageIndex{
			MessageLog:      msgLog,
			ChatName:        GetChatName(ctx, utility.AddressORNil(resp.Data.ChatId)),
			RawMessage:      content,
			RawMessageJieba: strings.Join(ws, " "),
			CreateTime:      utility.Epo2DateZoneMil(utility.MustInt(*resp.Data.CreateTime), time.UTC, time.DateTime),
			CreateTimeV2:    utility.Epo2DateZoneMil(utility.MustInt(*resp.Data.CreateTime), utility.UTCPlus8Loc(), time.RFC3339),
			Message:         embedded,
			UserID:          "你",
			UserName:        "你",
			TokenUsage:      usage,
		},
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("InsertData", zap.Error(err))
		return
	}
	err = retriver.Cli().AddDocuments(ctx, utility.AddressORNil(resp.Data.ChatId),
		[]schema.Document{{
			PageContent: content,
			Metadata: map[string]any{
				"chat_id":     utility.AddressORNil(resp.Data.ChatId),
				"user_id":     utility.AddressORNil(resp.Data.Sender.Id),
				"msg_id":      utility.AddressORNil(resp.Data.MessageId),
				"create_time": utility.EpoMil2DateStr(*resp.Data.CreateTime),
				"user_name":   "你",
			},
		}},
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("AddDocuments error", zap.Error(err))
	}
}

// CreateMsgText 不需要自行BuildText
func CreateMsgText(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(content))
	defer span.End()
	defer func() { span.RecordError(err) }()

	// TODO: Add id saving
	return CreateMsgTextRaw(ctx, NewTextMsgBuilder().Text(content).Build(), msgID, chatID)
}

// CreateMsgTextRaw 需要自行BuildText
func CreateMsgTextRaw(ctx context.Context, content, msgID, chatID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("content").String(content))
	defer span.End()
	defer func() { span.RecordError(err) }()
	// TODO: Add id saving
	uuid := (msgID + "_create")
	if len(uuid) > 50 {
		uuid = uuid[:50]
	}
	resp, err := lark.LarkClient.Im.Message.Create(ctx,
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
		logs.L().Ctx(ctx).Error("CreateMessage", zap.Error(err))
		return err
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("CreateMessage", zap.String("respError", resp.Error()))
		return errors.New(resp.Error())
	}
	go RecordMessage2Opensearch(ctx, resp)
	return
}

func AddReaction(ctx context.Context, reactionType, msgID string) (reactionID string, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	defer func() { span.RecordError(err) }()

	req := larkim.NewCreateMessageReactionReqBuilder().Body(larkim.NewCreateMessageReactionReqBodyBuilder().ReactionType(larkim.NewEmojiBuilder().EmojiType(reactionType).Build()).Build()).MessageId(msgID).Build()
	resp, err := lark.LarkClient.Im.V1.MessageReaction.Create(ctx, req)
	if err != nil {
		logs.L().Ctx(ctx).Error("AddReaction", zap.Error(err))
		return "", err
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("AddReaction", zap.String("respError", resp.Error()))
		return "", errors.New(resp.Error())
	}
	AddTrace2DB(ctx, msgID)
	return *resp.Data.ReactionId, err
}

func AddReactionAsync(ctx context.Context, reactionType, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	defer func() { span.RecordError(err) }()

	req := larkim.NewCreateMessageReactionReqBuilder().Body(larkim.NewCreateMessageReactionReqBodyBuilder().ReactionType(larkim.NewEmojiBuilder().EmojiType(reactionType).Build()).Build()).MessageId(msgID).Build()
	go func() {
		resp, err := lark.LarkClient.Im.V1.MessageReaction.Create(ctx, req)
		if err != nil {
			logs.L().Ctx(ctx).Error("AddReaction", zap.Error(err))
			return
		}
		if !resp.Success() {
			logs.L().Ctx(ctx).Error("AddReaction", zap.String("respError", resp.Error()))
			return
		}
		AddTrace2DB(ctx, msgID)
	}()
	return nil
}

func RemoveReaction(ctx context.Context, reactionID, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	defer func() { span.RecordError(err) }()
	req := larkim.NewDeleteMessageReactionReqBuilder().MessageId(msgID).ReactionId(reactionID).Build()
	resp, err := lark.LarkClient.Im.V1.MessageReaction.Delete(ctx, req)
	if err != nil {
		logs.L().Ctx(ctx).Error("RemoveReaction", zap.Error(err))
		return err
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("RemoveReaction", zap.String("respError", resp.Error()))
		return errors.New(resp.Error())
	}
	AddTrace2DB(ctx, msgID)
	return
}

func RemoveReactionAsync(ctx context.Context, reactionID, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	defer func() { span.RecordError(err) }()
	req := larkim.NewDeleteMessageReactionReqBuilder().MessageId(msgID).ReactionId(reactionID).Build()
	go func() {
		resp, err := lark.LarkClient.Im.V1.MessageReaction.Delete(ctx, req)
		if err != nil {
			logs.L().Ctx(ctx).Error("RemoveReaction", zap.Error(err))
			return
		}
		if !resp.Success() {
			logs.L().Ctx(ctx).Error("RemoveReaction", zap.String("respError", resp.Error()))
			err = errors.New(resp.Error())
			return
		}
		AddTrace2DB(ctx, msgID)
	}()
	return
}

// UpdateMessageTextRaw textMsg必须是序列化后的JSON
//
//	@param ctx context.Context
//	@param msgID string
//	@param textMsg string
//	@return err error
//	@author kevinmatthe
//	@update 2025-06-05 17:06:39
func UpdateMessageTextRaw(ctx context.Context, msgID, textMsg string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := lark.LarkClient.Im.V1.Message.Update(
		ctx,
		larkim.NewUpdateMessageReqBuilder().MessageId(msgID).
			Body(
				larkim.NewUpdateMessageReqBodyBuilder().MsgType(larkim.MsgTypeText).Content(textMsg).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}
	return
}

// UpdateMessageText 1
//
//	@param ctx context.Context
//	@param msgID string
//	@param textMsg string
//	@return err error
//	@author kevinmatthe
//	@update 2025-06-05 17:06:39
func UpdateMessageText(ctx context.Context, msgID, textMsg string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := lark.LarkClient.Im.V1.Message.Update(
		ctx,
		larkim.NewUpdateMessageReqBuilder().MessageId(msgID).
			Body(
				larkim.NewUpdateMessageReqBodyBuilder().MsgType(larkim.MsgTypeText).Content(NewTextMsgBuilder().Text(textMsg).Build()).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}
	return
}
