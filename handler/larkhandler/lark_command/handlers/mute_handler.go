package handlers

import (
	"context"
	"time"

	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/ark/tools"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/gg/goption"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
)

const (
	MuteRedisKeyPrefix = "mute:"
)

func MuteHandler(ctx context.Context, event *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	var (
		res              string
		muteTimeDuration time.Duration
	)
	defer func() { metaData.SetExtra("mute_result", res) }()
	argMap, _ := parseArgs(args...)
	if _, ok := argMap["cancel"]; ok {
		// 取消禁言
		// 先检查是否已经取消禁言
		if ext, err := redis.GetRedisClient().
			Exists(ctx, MuteRedisKeyPrefix+*event.Event.Message.ChatId).Result(); err != nil {
			return err
		} else if ext == 0 {
			res = "没有禁言，不需要取消, 直接发言即可"
			return nil // Do nothing
		}
		if err := redis.GetRedisClient().Del(ctx, MuteRedisKeyPrefix+*event.Event.Message.ChatId).Err(); err != nil {
			return err
		}
		res = "禁言已取消"
	} else if muteTime, ok := argMap["t"]; ok {
		muteTimeDuration, err = time.ParseDuration(muteTime)
		if err != nil {
			return errors.Wrap(err, "parse time error")
		}
	} else {
		muteTimeDuration = time.Minute * 3 // 默认三分钟
	}
	if muteTimeDuration > 0 {
		if err := redis.GetRedisClient().
			Set(ctx, MuteRedisKeyPrefix+*event.Event.Message.ChatId, 1, muteTimeDuration).
			Err(); err != nil {
			return err
		}
		res = "已启用" + muteTimeDuration.String() + "禁言"
	}
	err = larkutils.ReplyCardText(ctx, res, *event.Event.Message.MessageId, "_mute", true)
	if err != nil {
		return err
	}
	return
}

func init() {
	params := tools.NewParameters("object").
		AddProperty("time", &tools.Property{
			Type:        "string",
			Description: "禁言的时间 duration 格式, 例如 3m 表示禁言三分钟",
		}).AddProperty("cancel", &tools.Property{
		Type:        "boolean",
		Description: "是否取消禁言, 默认为 false",
	})
	fcu := tools.NewFunctionCallUnit().
		Name("mute_robot").Desc("为机器人设置或解除禁言.当用户要求机器人说话时，可以先尝试调用此函数取消禁言。当用户要求机器人闭嘴或者不要说话时，需要调用此函数设置禁言").Params(params).Func(muteWrap)
	tools.M().Add(fcu)
}

func muteWrap(ctx context.Context, meta *tools.FunctionCallMeta, args string) (any, error) {
	s := struct {
		Time   string `json:"time"`
		Cancel bool   `json:"cancel"`
	}{}
	err := utility.UnmarshallStringPre(args, &s)
	if err != nil {
		return nil, err
	}
	argsSlice := make([]string, 0)
	if s.Cancel {
		argsSlice = append(argsSlice, "--cancel")
	}
	if s.Time != "" {
		argsSlice = append(argsSlice, "--t="+s.Time)
	}
	metaData := handlerbase.NewBaseMetaDataWithChatIDUID(ctx, meta.ChatID, meta.UserID)
	if err := MuteHandler(ctx, meta.LarkData, metaData, argsSlice...); err != nil {
		return nil, err
	}
	return goption.Of(metaData.GetExtra("mute_result")).ValueOr("执行完成但没有结果"), nil
}
