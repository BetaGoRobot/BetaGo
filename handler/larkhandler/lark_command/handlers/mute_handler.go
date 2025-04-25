package handlers

import (
	"context"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
)

const (
	MuteRedisKeyPrefix = "mute:"
)

func MuteHandler(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	var (
		res              string
		muteTimeDuration time.Duration
	)
	argMap, _ := parseArgs(args...)
	if argMap["cancel"] != "" {
		// 取消禁言
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
