package larkcards

import (
	"context"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/patrickmn/go-cache"
)

var (
	repeatConfigCache   = cache.New(time.Second*30, time.Second*30)
	repeatWordRateCache = cache.New(time.Second*30, time.Second*30)
)

func RepeatMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	msg := PreGetTextMsg(ctx, event)
	// 先判断群聊的功能启用情况
	if enabled, exists := repeatConfigCache.Get(*event.Event.Message.ChatId); exists {
		// 缓存中已存在，直接取值
		if !enabled.(bool) {
			return
		}
	} else {
		// 缓存中不存在，从数据库中取值
		var count int64
		database.GetDbConnection().Find(&database.RepeatWhitelist{GuildID: *event.Event.Message.ChatId}).Count(&count)
		if count == 0 {
			repeatConfigCache.Set(*event.Event.Message.ChatId, false, cache.DefaultExpiration)
			return
		}
		repeatConfigCache.Set(*event.Event.Message.ChatId, true, cache.DefaultExpiration)
	}

	// 开始摇骰子, 默认概率10%
	realRate := utility.MustAtoI(utility.GetEnvWithDefault("REPEAT_DEFAULT_RATE", "10"))
	if rate, exists := repeatWordRateCache.Get(msg); exists {
		if r := rate.(int); r != -1 {
			repeatWordRateCache.Set(msg, rate.(int), cache.DefaultExpiration)
			realRate = r
		}
	} else {
		wordRate := database.RepeatWordsRate{
			Word: msg,
		}
		database.GetDbConnection().Find(&wordRate)
		if wordRate.Rate != 0 && wordRate.Rate != -1 {
			realRate = wordRate.Rate
		}
		repeatWordRateCache.Set(msg, realRate, cache.DefaultExpiration)
	}
	if utility.Probability(float64(realRate) / 100) {
		// sendMsg
		textMsg := larkim.NewTextMsgBuilder().Text(msg).Build()
		req := larkim.NewCreateMessageReqBuilder().ReceiveIdType(larkim.ReceiveIdTypeChatId).Body(
			larkim.NewCreateMessageReqBodyBuilder().
				ReceiveId(*event.Event.Message.ChatId).
				Content(textMsg).
				MsgType(larkim.MsgTypeText).
				Uuid(*event.Event.Message.MessageId).
				Build(),
		).Build()
		resp, err := larkutils.LarkClient.Im.V1.Message.Create(ctx, req)
		if err != nil {
			log.ZapLogger.Error("repeatMessage", zaplog.Error(err))
		}
		log.ZapLogger.Info("repeatMessage", zaplog.Any("resp", resp))
	}
}

func ReactionMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	// 先判断群聊的功能启用情况
	if enabled, exists := repeatConfigCache.Get(*event.Event.Message.ChatId); exists {
		// 缓存中已存在，直接取值
		if !enabled.(bool) {
			return
		}
	} else {
		// 缓存中不存在，从数据库中取值
		var count int64
		database.GetDbConnection().Find(&database.ReactionWhitelist{GuildID: *event.Event.Message.ChatId}).Count(&count)
		if count == 0 {
			repeatConfigCache.Set(*event.Event.Message.ChatId, false, cache.DefaultExpiration)
			return
		}
		repeatConfigCache.Set(*event.Event.Message.ChatId, true, cache.DefaultExpiration)
	}

	// 开始摇骰子, 默认概率10%
	realRate := utility.MustAtoI(utility.GetEnvWithDefault("REACTION_DEFAULT_RATE", "10"))
	if utility.Probability(float64(realRate) / 100) {
		// sendMsg
		req := larkim.NewCreateMessageReactionReqBuilder().
			MessageId(*event.Event.Message.MessageId).
			Body(
				larkim.NewCreateMessageReactionReqBodyBuilder().
					ReactionType(
						larkim.NewEmojiBuilder().
							EmojiType("THUMBSUP").
							Build(),
					).
					Build(),
			).
			Build()
		resp, err := larkutils.LarkClient.Im.V1.MessageReaction.Create(ctx, req)
		if err != nil {
			log.ZapLogger.Error("reactMessage", zaplog.Error(err))
		}
		log.ZapLogger.Info("reactMessage", zaplog.Any("resp", resp))
	}
}
