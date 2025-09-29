package lark

import (
	"github.com/BetaGoRobot/BetaGo/consts/env"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

var LarkClient *lark.Client = lark.NewClient(env.LarkAppID, env.LarkAppSecret)
