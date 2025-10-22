package larkchunking

import (
	"context"
	"strconv"
	"time"

	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type LarkMessageRespCreate struct {
	*larkim.CreateMessageResp
}

func (m *LarkMessageRespCreate) GroupID() (res string) {
	return *m.Data.ChatId
}

func (m *LarkMessageRespCreate) MsgID() (res string) {
	return *m.Data.MessageId
}

func (m *LarkMessageRespCreate) TimeStamp() (res int64) {
	t, err := strconv.ParseInt(*m.Data.CreateTime, 10, 64)
	if err != nil {
		zaplog.Logger.Error("getTimestampFunc error", zaplog.Error(err))
		return time.Now().UnixMilli()
	}
	return t
}

func (m *LarkMessageRespCreate) BuildLine() (line string) {
	return buildLineCommon(
		context.Background(),
		m.Data.Body.Content,
		m.Data.MsgType,
		m.Data.Mentions,
		m.Data.Sender.Id,
		m.Data.ChatId,
		m.TimeStamp(),
	)
}
