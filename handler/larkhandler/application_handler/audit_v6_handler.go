package applicationhandler

import (
	"context"

	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
)

func AuditV6Handler(ctx context.Context, event *larkapplication.P2ApplicationAppVersionAuditV6) error {
	return nil
}
