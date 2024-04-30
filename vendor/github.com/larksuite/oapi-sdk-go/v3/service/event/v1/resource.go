// Code generated by Lark OpenAPI.

package larkevent

import (
	"context"
	"github.com/larksuite/oapi-sdk-go/v3/core"
	"net/http"
)

type V1 struct {
	OutboundIp *outboundIp // 事件订阅
}

func New(config *larkcore.Config) *V1 {
	return &V1{
		OutboundIp: &outboundIp{config: config},
	}
}

type outboundIp struct {
	config *larkcore.Config
}

// List 获取事件出口 IP
//
// - 飞书开放平台向应用配置的回调地址推送事件时，是通过特定的 IP 发送出去的，应用可以通过本接口获取所有相关的 IP 地址。
//
// - IP 地址有变更可能，建议应用每隔 6 小时定时拉取最新的 IP 地址，以免由于企业防火墙设置，导致应用无法及时接收到飞书开放平台推送的事件。
//
// - 官网API文档链接:https://open.feishu.cn/document/ukTMukTMukTM/uYDNxYjL2QTM24iN0EjN/event-v1/outbound_ip/list
//
// - 使用Demo链接:https://github.com/larksuite/oapi-sdk-go/tree/v3_main/sample/apiall/eventv1/list_outboundIp.go
func (o *outboundIp) List(ctx context.Context, req *ListOutboundIpReq, options ...larkcore.RequestOptionFunc) (*ListOutboundIpResp, error) {
	// 发起请求
	apiReq := req.apiReq
	apiReq.ApiPath = "/open-apis/event/v1/outbound_ip"
	apiReq.HttpMethod = http.MethodGet
	apiReq.SupportedAccessTokenTypes = []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant}
	apiResp, err := larkcore.Request(ctx, apiReq, o.config, options...)
	if err != nil {
		return nil, err
	}
	// 反序列响应结果
	resp := &ListOutboundIpResp{ApiResp: apiResp}
	err = apiResp.JSONUnmarshalBody(resp, o.config)
	if err != nil {
		return nil, err
	}
	return resp, err
}
func (o *outboundIp) ListByIterator(ctx context.Context, req *ListOutboundIpReq, options ...larkcore.RequestOptionFunc) (*ListOutboundIpIterator, error) {
	return &ListOutboundIpIterator{
		ctx:      ctx,
		req:      req,
		listFunc: o.List,
		options:  options,
		limit:    req.Limit}, nil
}
