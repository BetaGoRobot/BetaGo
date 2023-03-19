package gotify

import "testing"

func TestSendMessage(t *testing.T) {
	SendMessage("", "SourceChannelID: `7485159615915618596`", 5)
}
