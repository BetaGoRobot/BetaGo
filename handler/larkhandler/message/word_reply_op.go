package message

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ Op = &WordReplyMsgOperator{}

// WordReplyMsgOperator  Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:11
type WordReplyMsgOperator struct {
	OpBase
}

// PreRun Repeat
//
//	@receiver r *WordReplyMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:17
func (r *WordReplyMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)

	// 先判断群聊的功能启用情况
	if !larkutils.CheckFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionWordReply) {
		return errors.Wrap(consts.ErrStageSkip, "WordReplyMsgOperator: Not enabled")
	}

	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "WordReplyMsgOperator: Is Command")
	}
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *WordReplyMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)

	msg := larkutils.PreGetTextMsg(ctx, event)
	var replyItem *database.ReplyNType
	// 检查定制化逻辑, Key为GuildID, 拿到GUI了dID下的所有SubStr配置
	customConfig, hitCache := database.FindByCacheFunc(database.QuoteReplyMsgCustom{GuildID: *event.Event.Message.ChatId},
		func(d database.QuoteReplyMsgCustom) string {
			return d.GuildID
		},
	)
	replyList := make([]*database.ReplyNType, 0)
	span.SetAttributes(attribute.Bool("QuoteReplyMsgCustom hitCache", hitCache))
	for _, data := range customConfig {
		if CheckQuoteKeywordMatch(msg, data.Keyword, data.MatchType) {
			replyList = append(replyList, &database.ReplyNType{Reply: data.Reply, ReplyType: data.ReplyType})
		}
	}

	if len(replyList) == 0 {
		// 无定制化逻辑，走通用判断
		data, hitCache := database.FindByCacheFunc(
			database.QuoteReplyMsg{},
			func(d database.QuoteReplyMsg) string {
				return d.Keyword
			},
		)
		span.SetAttributes(attribute.Bool("QuoteReplyMsg hitCache", hitCache))
		for _, d := range data {
			if CheckQuoteKeywordMatch(msg, d.Keyword, d.MatchType) {
				replyList = append(replyList, &database.ReplyNType{Reply: d.Reply, ReplyType: d.ReplyType})
			}
		}
	}
	if len(replyList) > 0 {
		replyItem = utility.SampleSlice(replyList)
		_, subSpan := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		defer subSpan.End()
		if replyItem.ReplyType == consts.ReplyTypeText {
			_, err := larkutils.ReplyMsgText(ctx, replyItem.Reply, *event.Event.Message.MessageId, "_wordReply", false)
			if err != nil {
				log.Zlog.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
				return err
			}
		} else if replyItem.ReplyType == consts.ReplyTypeImg {
			var msgType, content string
			if strings.HasPrefix(replyItem.Reply, "img") {
				msgType = larkim.MsgTypeImage
				content, _ = sonic.MarshalString(map[string]string{
					"image_key": replyItem.Reply,
				})
			} else {
				msgType = larkim.MsgTypeSticker
				content, _ = sonic.MarshalString(map[string]string{
					"file_key": replyItem.Reply,
				})
			}
			_, err := larkutils.ReplyMsgRawContentType(ctx, *event.Event.Message.MessageId, msgType, content, "_wordReply", false)
			if err != nil {
				log.Zlog.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
				return err
			}
		} else {
			return errors.New("unknown reply type")
		}

	}
	return
}

func CheckQuoteKeywordMatch(msg string, keyword string, matchType consts.WordMatchType) bool {
	switch matchType {
	case consts.MatchTypeFull:
		return msg == keyword
	case consts.MatchTypeSubStr:
		return strings.Contains(msg, keyword)
	case consts.MatchTypeRegex:
		return utility.RegexpMatch(msg, keyword)
	default:
		panic("unknown match type" + matchType)
	}
}
