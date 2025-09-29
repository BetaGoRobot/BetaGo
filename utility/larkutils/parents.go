package larkutils

import (
	"context"
	"errors"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func GetAllParentMsg(ctx context.Context, data *larkim.P2MessageReceiveV1) (msgList []*larkim.Message, err error) {
	msgList = []*larkim.Message{}
	if data.Event.Message.ThreadId != nil { // 话题模式，找到所有的ID
		resp, err := LarkClient.Im.Message.List(ctx, larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return nil, err
		}
		for _, msg := range resp.Data.Items {
			msgList = append(msgList, msg)
		}
	} else if data.Event.Message.ParentId != nil {
		respMsg := GetMsgFullByID(ctx, *data.Event.Message.ParentId)
		msg := respMsg.Data.Items[0]
		if msg == nil {
			return nil, errors.New("No parent message found")
		}
		msgList = append(msgList, msg)
	}
	return
}
