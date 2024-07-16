package ark

import (
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/utils"
)

type ArkLLMInteraction interface {
	PrePrompt(string)
	Context()
	Input(string)
	Receive() string
	ReceiveStream() *utils.ChatCompletionStreamReader
}
