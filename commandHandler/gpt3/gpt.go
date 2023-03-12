package gpt3

import (
	"os"
)

var apiKey = os.Getenv("GPT_TOKEN")

// CreateChatCompletion
//
//	@param msg
//	@return message
//	@return err
func CreateChatCompletion(msg string) (message string, err error) {
	gptClient := &GPTClient{Model: "gpt-3.5-turbo"}
	gptClient.SetContent(msg)
	message, err = gptClient.Post()
	return
}
