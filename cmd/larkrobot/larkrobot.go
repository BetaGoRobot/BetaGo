package main

import (
	"context"
	"net/http"
	"os"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

func longConn() { // 注册事件回调
	eventHandler := dispatcher.
		NewEventDispatcher("", "").
		OnP2MessageReactionCreatedV1(larkhandler.MessageReactionHandler).
		OnP2MessageReceiveV1(larkhandler.MessageV2Handler).
		OnP2ApplicationAppVersionAuditV6(larkhandler.AuditV6Handler)
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

func webHook() {
	// 创建 card 处理器
	cardHandler := larkcard.
		NewCardActionHandler(
			os.Getenv("LARK_VERIFICATION"),
			os.Getenv("LARK_ENCRYPTION"),
			larkhandler.WebHookHandler,
		)

	// 注册处理器
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/webhook/card", httpserverext.NewCardActionHandlerFunc(cardHandler, larkevent.WithLogLevel(larkcore.LogLevelDebug)))

	// 启动 http 服务
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	go longConn()
	go webHook()
	select {}
}
