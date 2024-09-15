package larkhandler

import (
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type MessageIndex struct {
	*database.MessageLog
	ChatName   string      `json:"chat_name"`
	CreateTime string      `json:"create_time"`
	Message    []float32   `json:"message"`
	UserID     string      `json:"user_id"`
	UserName   string      `json:"user_name"`
	RawMessage string      `json:"raw_message"`
	TokenUsage model.Usage `json:"token_usage"`
}
