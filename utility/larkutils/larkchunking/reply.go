package larkchunking

import (
	"context"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type LarkMessageRespReply struct {
	*larkim.ReplyMessageResp
}

func (m *LarkMessageRespReply) GroupID() (res string) {
	return *m.Data.ChatId
}

func (m *LarkMessageRespReply) MsgID() (res string) {
	return *m.Data.MessageId
}

func (m *LarkMessageRespReply) TimeStamp() (res int64) {
	t, err := strconv.ParseInt(*m.Data.CreateTime, 10, 64)
	if err != nil {
		logs.L.Error().Err(err).Msg("getTimestampFunc error")
		return time.Now().UnixMilli()
	}
	return t
}

func (m *LarkMessageRespReply) BuildLine() (line string) {
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
