package larkcards

import (
	"context"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/patrickmn/go-cache"
)

var (
	repeatConfigCache   = cache.New(time.Second*30, time.Second*30)
	repeatWordRateCache = cache.New(time.Second*30, time.Second*30)
)

func RepeatMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	msgMap := make(map[string]interface{})
	msg := *event.Event.Message.Content
	err := sonic.UnmarshalString(msg, &msgMap)
	if err != nil {
		zaplog.Logger.Error("repeatMessage", zaplog.Error(err))
		return
	}
	if text, ok := msgMap["text"]; ok {
		msg = text.(string)
	}
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
		if r := rate.(int); r > 0 {
			repeatWordRateCache.Set(msg, rate.(int)-1, cache.DefaultExpiration)
			realRate = r
		}
	} else {
		wordRate := database.RepeatWordsRate{
			Word: msg,
		}
		database.GetDbConnection().Find(&wordRate)
		if wordRate.Rate == 0 {
			repeatWordRateCache.Set(msg, 1, cache.DefaultExpiration)
			return
		}
		repeatWordRateCache.Set(msg, wordRate.Rate, cache.DefaultExpiration)
		realRate = wordRate.Rate
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
			log.SugerLogger.Error("repeatMessage", zaplog.Error(err))
		}
		log.SugerLogger.Info("repeatMessage", zaplog.Any("resp", resp))
	}
}
