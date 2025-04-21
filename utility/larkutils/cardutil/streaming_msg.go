package cardutil

import "github.com/bytedance/sonic"

type CardEntitySendContent struct {
	Type string              `json:"type"`
	Data *CardEntitySendData `json:"data"`
}

type CardEntitySendData struct {
	CardID string `json:"card_id"`
}

func NewCardEntityContent(cardID string) *CardEntitySendContent {
	return &CardEntitySendContent{
		"card",
		&CardEntitySendData{
			cardID,
		},
	}
}

func (e *CardEntitySendContent) String() string {
	s, _ := sonic.MarshalString(e)
	return s
}
