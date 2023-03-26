package helper

import (
	"context"
	"testing"
)

func TestAdminCommandHelperHandler(t *testing.T) {
	AdminCommandHelperHandler(context.Background(), "4988093461275944", "", "")
}

func TestUserCommandHelperHandler(t *testing.T) {
	UserCommandHelperHandler(context.Background(), "4988093461275944", "", "")
}
