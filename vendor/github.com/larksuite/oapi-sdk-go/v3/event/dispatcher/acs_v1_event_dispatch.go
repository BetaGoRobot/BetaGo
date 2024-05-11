// Package dispatcher code generated by oapi sdk gen
/*
 * MIT License
 *
 * Copyright (c) 2022 Lark Technologies Pte. Ltd.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice, shall be included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package dispatcher

import (
	"context"
	"github.com/larksuite/oapi-sdk-go/v3/service/acs/v1"
)

// 新增门禁访问记录
//
// - 门禁设备识别用户成功后发送该事件给订阅应用。
//
// - 事件描述文档链接:https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/acs-v1/access_record/events/created
func (dispatcher *EventDispatcher) OnP2AccessRecordCreatedV1(handler func(ctx context.Context, event *larkacs.P2AccessRecordCreatedV1) error) *EventDispatcher {
	_, existed := dispatcher.eventType2EventHandler["acs.access_record.created_v1"]
	if existed {
		panic("event: multiple handler registrations for " + "acs.access_record.created_v1")
	}
	dispatcher.eventType2EventHandler["acs.access_record.created_v1"] = larkacs.NewP2AccessRecordCreatedV1Handler(handler)
	return dispatcher
}

// 用户信息变更
//
// - 智能门禁用户特征值变化时，发送此事件。
//
// - 事件描述文档链接:https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/acs-v1/user/events/updated
func (dispatcher *EventDispatcher) OnP2UserUpdatedV1(handler func(ctx context.Context, event *larkacs.P2UserUpdatedV1) error) *EventDispatcher {
	_, existed := dispatcher.eventType2EventHandler["acs.user.updated_v1"]
	if existed {
		panic("event: multiple handler registrations for " + "acs.user.updated_v1")
	}
	dispatcher.eventType2EventHandler["acs.user.updated_v1"] = larkacs.NewP2UserUpdatedV1Handler(handler)
	return dispatcher
}
