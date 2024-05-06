package gotify

import (
	"context"
	"testing"
)

func TestSendMessage(t *testing.T) {
	SendMessage(context.Background(), "", "SourceChannelID: `7485159615915618596`", 5)
}
