// Code generated by Lark OpenAPI.

package larkhuman_authentication

import (
	"context"
	"github.com/larksuite/oapi-sdk-go/v3/core"
	"net/http"
)

type V1 struct {
	Identity *identity // 实名认证
}

func New(config *larkcore.Config) *V1 {
	return &V1{
		Identity: &identity{config: config},
	}
}

type identity struct {
	config *larkcore.Config
}

// Create 录入身份信息
//
// - 该接口用于录入实名认证的身份信息，在唤起有源活体认证前，需要使用该接口进行实名认证。
//
// - 实名认证接口会有计费管理，接入前请联系飞书开放平台工作人员，邮箱：openplatform@bytedance.com。;;仅通过计费申请的应用，才能在[开发者后台](https://open.feishu.cn/app)查找并申请该接口的权限。
//
// - 官网API文档链接:https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/human_authentication-v1/identity/create
//
// - 使用Demo链接:https://github.com/larksuite/oapi-sdk-go/tree/v3_main/sample/apiall/human_authenticationv1/create_identity.go
func (i *identity) Create(ctx context.Context, req *CreateIdentityReq, options ...larkcore.RequestOptionFunc) (*CreateIdentityResp, error) {
	// 发起请求
	apiReq := req.apiReq
	apiReq.ApiPath = "/open-apis/human_authentication/v1/identities"
	apiReq.HttpMethod = http.MethodPost
	apiReq.SupportedAccessTokenTypes = []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant}
	apiResp, err := larkcore.Request(ctx, apiReq, i.config, options...)
	if err != nil {
		return nil, err
	}
	// 反序列响应结果
	resp := &CreateIdentityResp{ApiResp: apiResp}
	err = apiResp.JSONUnmarshalBody(resp, i.config)
	if err != nil {
		return nil, err
	}
	return resp, err
}
