package message

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/lark"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var _ Op = &RepeatMsgOperator{}

// RepeatMsgOperator  RepeatMsg Op
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:51
type RepeatMsgOperator struct {
	OpBase
}

// PreRun Repeat
//
//	@receiver r *RepeatMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:35
func (r *RepeatMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	// 先判断群聊的功能启用情况
	if !larkutils.CheckFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomRepeat) {
		return errors.Wrap(consts.ErrStageSkip, "RepeatMsgOperator: Not enabled")
	}
	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "RepeatMsgOperator: Is Command")
	}
	return
}

// Run Repeat
//
//	@receiver r *RepeatMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:41
func (r *RepeatMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer func() { span.RecordError(err) }()

	// Repeat
	msg := larkutils.PreGetTextMsg(ctx, event)

	// 开始摇骰子, 默认概率10%
	realRate := utility.MustAtoI(utility.GetEnvWithDefault("REPEAT_DEFAULT_RATE", "10"))
	// 群聊定制化
	config, hitCache := database.FindByCacheFunc(
		database.RepeatWordsRateCustom{
			GuildID: *event.Event.Message.ChatId,
			Word:    msg,
		},
		func(d database.RepeatWordsRateCustom) string {
			return d.GuildID + d.Word
		},
	)
	span.SetAttributes(attribute.Bool("RepeatWordsRateCustom hitCache", hitCache))

	if len(config) != 0 {
		realRate = config[0].Rate
	} else {
		config, hitCache := database.FindByCacheFunc(
			database.RepeatWordsRate{
				Word: msg,
			},
			func(d database.RepeatWordsRate) string {
				return d.Word
			},
		)
		span.SetAttributes(attribute.Bool("RepeatWordsRate hitCache", hitCache))
		if len(config) != 0 {
			realRate = config[0].Rate
		}
	}

	if utility.Probability(float64(realRate) / 100) {
		msgType := strings.ToLower(*event.Event.Message.MessageType)
		if msgType == "text" {
			m, err := utility.JSON2Map(*event.Event.Message.Content)
			if err != nil {
				return err
			}
			for _, mention := range event.Event.Message.Mentions {
				m["text"] = strings.ReplaceAll(m["text"].(string), *mention.Key, larkmsgutils.AtUser(*mention.Id.OpenId, *mention.Name))
			}
			err = larkutils.CreateMsgTextRaw(
				ctx,
				utility.MustMashal(m),
				*event.Event.Message.MessageId,
				*event.Event.Message.ChatId,
			)
			if err != nil {
				logs.L.Error().Ctx(ctx).Err(err).Str("TraceID", span.SpanContext().TraceID().String()).Msg("repeatMessage error")
			}
		} else {
			repeatReq := larkim.NewCreateMessageReqBuilder().
				Body(
					larkim.NewCreateMessageReqBodyBuilder().
						Content(*event.Event.Message.Content).
						ReceiveId(*event.Event.Message.ChatId).
						MsgType(*event.Event.Message.MessageType).
						Build(),
				).
				ReceiveIdType(larkim.ReceiveIdTypeChatId).
				Build()
			resp, err := lark.LarkClient.Im.V1.Message.Create(ctx, repeatReq)
			if err != nil {
				return err
			}
			if !resp.Success() {
				if strings.Contains(resp.Error(), "invalid image_key") {
					logs.L.Error().Ctx(ctx).Err(err).Str("TraceID", span.SpanContext().TraceID().String()).Msg("repeatMessage error")
					return nil
				}
				return errors.New(resp.Error())
			}
			larkutils.RecordMessage2Opensearch(ctx, resp)
		}
	}
	return
}

func RebuildAtMsg(input string, substrings []string) []string {
	result := []string{}
	start := 0

	// Keep track of the positions to split
	splitPositions := []int{}

	// Iterate through the input to find all occurrences of substrings
	for _, sub := range substrings {
		start = 0
		for {
			pos := strings.Index(input[start:], sub)
			if pos == -1 {
				break
			}
			actualPos := start + pos
			splitPositions = append(splitPositions, actualPos, actualPos+len(sub))
			start = actualPos + len(sub)
		}
	}

	// Sort the positions to split
	sort.Slice(splitPositions, func(i, j int) bool { return splitPositions[i] < splitPositions[j] })

	if len(splitPositions) > 0 {
		// Remove duplicate positions
		uniquePositions := []int{}
		for i, pos := range splitPositions {
			if i == 0 || pos != splitPositions[i-1] {
				uniquePositions = append(uniquePositions, pos)
			}
		}

		// Add start and end of the string to the positions if not already present
		if uniquePositions[0] != 0 {
			uniquePositions = append([]int{0}, uniquePositions...)
		}
		if uniquePositions[len(uniquePositions)-1] != len(input) {
			uniquePositions = append(uniquePositions, len(input))
		}

		// Extract substrings based on split positions
		for i := 0; i < len(uniquePositions)-1; i++ {
			result = append(result, input[uniquePositions[i]:uniquePositions[i+1]])
		}
	} else {
		result = append(result, input)
	}
	return result
}
