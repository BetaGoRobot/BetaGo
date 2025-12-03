package tools

import larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

type FunctionCallMeta struct {
	ChatID string
	UserID string

	LarkData *larkim.P2MessageReceiveV1
}

const (
	ToolSearchHistory = "search_history"
	ToolSearchMusic   = "search_music"
)
