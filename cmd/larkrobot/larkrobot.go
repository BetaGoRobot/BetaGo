package main

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

func longConn() { // 注册事件回调
	eventHandler := dispatcher.
		NewEventDispatcher("", "").
		OnP2MessageReactionCreatedV1(larkhandler.MessageReactionHandler).
		OnP2MessageReceiveV1(larkhandler.MessageV2Handler).
		OnP2ApplicationAppVersionAuditV6(larkhandler.AuditV6Handler).
		OnP2CardActionTrigger(larkhandler.WebHookHandler)
	// 创建Client
	cli := larkws.NewClient(env.LarkAppID, env.LarkAppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)
	// 启动客户端
	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

func main() {
	go longConn()
	select {}
}
