// Code generated by Lark OpenAPI.

package mail

import (
	"github.com/larksuite/oapi-sdk-go/v3/core"
	v1 "github.com/larksuite/oapi-sdk-go/v3/service/mail/v1"
)

type Service struct {
	*v1.V1
}

func NewService(config *larkcore.Config) *Service {
	return &Service{
		V1: v1.New(config),
	}
}
