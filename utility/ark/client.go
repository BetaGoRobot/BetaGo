package ark

import (
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
)

var client = arkruntime.NewClientWithApiKey(env.DOUBAO_API_KEY)

func Cli() *arkruntime.Client {
	return client
}
