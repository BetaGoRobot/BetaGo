package helper

import (
	"context"
	"testing"
)

func TestAdminCommandHelperHandler(t *testing.T) {
	AdminCommandHelperHandler(context.Background(), "7419593543056418", "", "")
}

func TestUserCommandHelperHandler(t *testing.T) {
	UserCommandHelperHandler(context.Background(), "7419593543056418", "", "")
}
