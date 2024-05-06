// Package meeting_room code generated by oapi sdk gen
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

package larkmeeting_room

import (
	"github.com/larksuite/oapi-sdk-go/v3/event"
)

type DepartmentId struct {
	DepartmentId     *string `json:"department_id,omitempty"`      //
	OpenDepartmentId *string `json:"open_department_id,omitempty"` //
}

type DepartmentIdBuilder struct {
	departmentId         string //
	departmentIdFlag     bool
	openDepartmentId     string //
	openDepartmentIdFlag bool
}

func NewDepartmentIdBuilder() *DepartmentIdBuilder {
	builder := &DepartmentIdBuilder{}
	return builder
}

// 示例值：
func (builder *DepartmentIdBuilder) DepartmentId(departmentId string) *DepartmentIdBuilder {
	builder.departmentId = departmentId
	builder.departmentIdFlag = true
	return builder
}

// 示例值：
func (builder *DepartmentIdBuilder) OpenDepartmentId(openDepartmentId string) *DepartmentIdBuilder {
	builder.openDepartmentId = openDepartmentId
	builder.openDepartmentIdFlag = true
	return builder
}

func (builder *DepartmentIdBuilder) Build() *DepartmentId {
	req := &DepartmentId{}
	if builder.departmentIdFlag {
		req.DepartmentId = &builder.departmentId

	}
	if builder.openDepartmentIdFlag {
		req.OpenDepartmentId = &builder.openDepartmentId

	}
	return req
}

type EventTime struct {
	TimeStamp *int `json:"time_stamp,omitempty"` //
}

type EventTimeBuilder struct {
	timeStamp     int //
	timeStampFlag bool
}

func NewEventTimeBuilder() *EventTimeBuilder {
	builder := &EventTimeBuilder{}
	return builder
}

// 示例值：
func (builder *EventTimeBuilder) TimeStamp(timeStamp int) *EventTimeBuilder {
	builder.timeStamp = timeStamp
	builder.timeStampFlag = true
	return builder
}

func (builder *EventTimeBuilder) Build() *EventTime {
	req := &EventTime{}
	if builder.timeStampFlag {
		req.TimeStamp = &builder.timeStamp

	}
	return req
}

type MeetingRoom struct {
	RoomId *int `json:"room_id,omitempty"` // your description here
}

type MeetingRoomBuilder struct {
	roomId     int // your description here
	roomIdFlag bool
}

func NewMeetingRoomBuilder() *MeetingRoomBuilder {
	builder := &MeetingRoomBuilder{}
	return builder
}

// your description here
//
// 示例值：
func (builder *MeetingRoomBuilder) RoomId(roomId int) *MeetingRoomBuilder {
	builder.roomId = roomId
	builder.roomIdFlag = true
	return builder
}

func (builder *MeetingRoomBuilder) Build() *MeetingRoom {
	req := &MeetingRoom{}
	if builder.roomIdFlag {
		req.RoomId = &builder.roomId

	}
	return req
}

type UserInfo struct {
	OpenId *string `json:"open_id,omitempty"` //
	UserId *string `json:"user_id,omitempty"` // 用户在ISV下的唯一标识，申请了"获取用户 user ID"权限后才会返回
}

type UserInfoBuilder struct {
	openId     string //
	openIdFlag bool
	userId     string // 用户在ISV下的唯一标识，申请了"获取用户 user ID"权限后才会返回
	userIdFlag bool
}

func NewUserInfoBuilder() *UserInfoBuilder {
	builder := &UserInfoBuilder{}
	return builder
}

// 示例值：
func (builder *UserInfoBuilder) OpenId(openId string) *UserInfoBuilder {
	builder.openId = openId
	builder.openIdFlag = true
	return builder
}

// 用户在ISV下的唯一标识，申请了"获取用户 user ID"权限后才会返回
//
// 示例值：
func (builder *UserInfoBuilder) UserId(userId string) *UserInfoBuilder {
	builder.userId = userId
	builder.userIdFlag = true
	return builder
}

func (builder *UserInfoBuilder) Build() *UserInfo {
	req := &UserInfo{}
	if builder.openIdFlag {
		req.OpenId = &builder.openId

	}
	if builder.userIdFlag {
		req.UserId = &builder.userId

	}
	return req
}

type P2MeetingRoomCreatedV1Data struct {
	RoomName *string `json:"room_name,omitempty"` //
	RoomId   *string `json:"room_id,omitempty"`   //
}

type P2MeetingRoomCreatedV1 struct {
	*larkevent.EventV2Base                             // 事件基础数据
	*larkevent.EventReq                                // 请求原生数据
	Event                  *P2MeetingRoomCreatedV1Data `json:"event"` // 事件内容
}

func (m *P2MeetingRoomCreatedV1) RawReq(req *larkevent.EventReq) {
	m.EventReq = req
}

type P2MeetingRoomDeletedV1Data struct {
	RoomName *string `json:"room_name,omitempty"` //
	RoomId   *string `json:"room_id,omitempty"`   //
}

type P2MeetingRoomDeletedV1 struct {
	*larkevent.EventV2Base                             // 事件基础数据
	*larkevent.EventReq                                // 请求原生数据
	Event                  *P2MeetingRoomDeletedV1Data `json:"event"` // 事件内容
}

func (m *P2MeetingRoomDeletedV1) RawReq(req *larkevent.EventReq) {
	m.EventReq = req
}

type P2MeetingRoomStatusChangedV1Data struct {
	RoomName *string `json:"room_name,omitempty"` // 会议室名称
	RoomId   *string `json:"room_id,omitempty"`   // 会议室 ID
}

type P2MeetingRoomStatusChangedV1 struct {
	*larkevent.EventV2Base                                   // 事件基础数据
	*larkevent.EventReq                                      // 请求原生数据
	Event                  *P2MeetingRoomStatusChangedV1Data `json:"event"` // 事件内容
}

func (m *P2MeetingRoomStatusChangedV1) RawReq(req *larkevent.EventReq) {
	m.EventReq = req
}

type P2MeetingRoomUpdatedV1Data struct {
	RoomName *string `json:"room_name,omitempty"` //
	RoomId   *string `json:"room_id,omitempty"`   //
}

type P2MeetingRoomUpdatedV1 struct {
	*larkevent.EventV2Base                             // 事件基础数据
	*larkevent.EventReq                                // 请求原生数据
	Event                  *P2MeetingRoomUpdatedV1Data `json:"event"` // 事件内容
}

func (m *P2MeetingRoomUpdatedV1) RawReq(req *larkevent.EventReq) {
	m.EventReq = req
}