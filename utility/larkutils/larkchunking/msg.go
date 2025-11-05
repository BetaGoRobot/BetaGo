package larkchunking

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/grouputil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkconsts"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/bytedance/sonic"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type LarkMessageEvent struct {
	*larkim.P2MessageReceiveV1
}

func (m *LarkMessageEvent) GroupID() (res string) {
	return *m.Event.Message.ChatId
}

func (m *LarkMessageEvent) MsgID() (res string) {
	return *m.Event.Message.MessageId
}

func (m *LarkMessageEvent) TimeStamp() (res int64) {
	t, err := strconv.ParseInt(*m.Event.Message.CreateTime, 10, 64)
	if err != nil {
		logs.L.Error().Err(err).Msg("getTimestampFunc error")
		return time.Now().UnixMilli()
	}
	return t
}

func (m *LarkMessageEvent) BuildLine() (line string) {
	mentions := m.Event.Message.Mentions

	tmpList := make([]string, 0)
	for msgItem := range larkmsgutils.
		GetContentItemsSeq(
			&larkim.EventMessage{
				Content:     m.Event.Message.Content,
				MessageType: m.Event.Message.MessageType,
			},
		) {
		switch msgItem.Tag {
		case "at", "text":
			if msgItem.Tag == "text" {
				m := map[string]string{}
				if err := sonic.UnmarshalString(msgItem.Content, &m); err == nil {
					msgItem.Content = m["text"]
				}
			}
			if len(mentions) > 0 {
				for _, mention := range mentions {
					if mention.Key != nil {
						if strings.HasPrefix(*mention.Name, "不太正经的网易云音乐机器人") {
							*mention.Name = "你"
						}
						msgItem.Content = strings.ReplaceAll(msgItem.Content, *mention.Key, fmt.Sprintf("@%s", *mention.Name))
					}
				}
			}
			fallthrough
		default:
			content := strings.ReplaceAll(msgItem.Content, "\n", "<换行>")
			if strings.TrimSpace(content) != "" {
				tmpList = append(tmpList, content)
			}
		}
	}
	userName := ""
	if *m.Event.Sender.SenderId.OpenId == larkconsts.BotAppID {
		userName = "机器人"
	} else {
		member, err := grouputil.GetUserMemberFromChat(context.Background(), *m.Event.Message.ChatId, *m.Event.Sender.SenderId.OpenId)
		if err != nil {
			logs.L.Error().Ctx(context.Background()).Err(err).Msg("got error openID")
		}
		if member == nil {
			userName = "NULL"
		} else {
			userName = *member.Name
		}
	}

	createTime := time.UnixMilli(m.TimeStamp()).In(utility.UTCPlus8Loc()).Format(time.DateTime)
	return fmt.Sprintf("[%s](%s) <%s>: %s", createTime, *m.Event.Sender.SenderId.OpenId, userName, strings.Join(tmpList, ";"))
}
