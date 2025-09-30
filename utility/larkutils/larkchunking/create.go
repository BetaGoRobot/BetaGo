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
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/bytedance/sonic"
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
	mentions := m.Data.Mentions

	tmpList := make([]string, 0)
	for msgItem := range larkmsgutils.
		GetContentItemsSeq(
			&larkim.EventMessage{
				Content:     m.Data.Body.Content,
				MessageType: m.Data.MsgType,
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
						if *mention.Name == "不太正经的网易云音乐机器人" {
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
	if *m.Data.Sender.Id != larkconsts.BotAppID {
		userName = "机器人"
	} else {
		member, err := grouputil.GetUserMemberFromChat(context.Background(), *m.Data.ChatId, *m.Data.Sender.Id)
		if err != nil {
			log.Zlog.Error("got error openID", zaplog.String("openID", *m.Data.Sender.Id))
		}
		if member == nil {
			userName = "NULL"
		} else {
			userName = *member.Name
		}
	}

	createTime := time.UnixMilli(m.TimeStamp()).In(utility.UTCPlus8Loc()).Format(time.DateTime)
	return fmt.Sprintf("[%s](%s) <%s>: %s", createTime, *m.Data.Sender.Id, userName, strings.Join(tmpList, ";"))
}
