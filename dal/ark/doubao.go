package ark

import "github.com/volcengine/volcengine-go-sdk/service/arkruntime/utils"

var _ ArkLLMInteraction = &DoubaoChat{}

type DoubaoChat struct{}

func (c *DoubaoChat) PrePrompt(input string) {
}

func (c *DoubaoChat) Context() {
}

func (c *DoubaoChat) Input(input string) {
}

func (c *DoubaoChat) Receive() string {
	return ""
}

func (c *DoubaoChat) ReceiveStream() *utils.ChatCompletionStreamReader {
	return nil
}
