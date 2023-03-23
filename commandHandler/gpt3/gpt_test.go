package gpt3

import (
	"context"
	"testing"
)

func TestCreateChatCompletion(t *testing.T) {
	CreateChatCompletion(context.Background(), "zsh是什么", "")
}

func TestModerationCheck(t *testing.T) {
	ModerationCheck(context.Background(), "I want to kill you")
}
