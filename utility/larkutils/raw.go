package larkutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkauth "github.com/larksuite/oapi-sdk-go/v3/service/auth/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type EphemeralCardReqBody struct {
	ChatID  string      `json:"chat_id"`
	OpenID  string      `json:"open_id"`
	MsgType string      `json:"msg_type"`
	Card    interface{} `json:"card"`
}

func GetTenantAccessToken(ctx context.Context) string {
	resp, err := LarkClient.Auth.V3.TenantAccessToken.Internal(ctx,
		larkauth.NewInternalTenantAccessTokenReqBuilder().
			Body(
				larkauth.NewInternalTenantAccessTokenReqBodyBuilder().
					AppId(env.LarkAppID).
					AppSecret(env.LarkAppSecret).
					Build(),
			).
			Build(),
	)
	if err != nil {
		log.Zlog.Error("get tenant access token error", zaplog.Error(err))
		return ""
	}
	var resMap map[string]interface{}
	err = resp.JSONUnmarshalBody(&resMap, nil)
	if err != nil {
		log.Zlog.Error("json unmarshal error", zaplog.Error(err))
		return ""
	}
	return resMap["tenant_access_token"].(string)
}

func SendEphemeral(ctx context.Context, chatID, openID, card string) {
	var tmpCard interface{}
	err := sonic.UnmarshalString(card, &tmpCard)
	if err != nil {
		log.Zlog.Error("json unmarshal error", zaplog.Error(err))
		return
	}
	tmpCard = map[string]interface{}{"elements": tmpCard.(map[string]interface{})["i18n_elements"].(map[string]interface{})["zh_cn"]}
	// token := GetTenantAccessToken(ctx)
	body := &EphemeralCardReqBody{
		ChatID:  chatID,
		OpenID:  openID,
		MsgType: larkim.MsgTypeInteractive,
		Card:    tmpCard,
	}
	bodyJSON, err := sonic.MarshalString(body)
	fmt.Println(bodyJSON)
	if err != nil {
		log.Zlog.Error("json marshal error", zaplog.Error(err))
		return
	}
	resp, err := LarkClient.Do(
		context.Background(),
		&larkcore.ApiReq{
			HttpMethod:                http.MethodPost,
			ApiPath:                   "https://open.larkoffice.com/open-apis/ephemeral/v1/send",
			Body:                      bodyJSON,
			SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
		},
	)
	if err != nil {
		log.Zlog.Error("send ephemeral error", zaplog.Error(err))
		return
	}
	fmt.Println(string(resp.RawBody))
}
