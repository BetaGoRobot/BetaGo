package larkchunking

import (
	"context"
	"fmt"
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

// LarkMessage defines a common interface for Lark message responses.
type LarkMessage interface {
	GroupID() string
	MsgID() string
	TimeStamp() int64
	BuildLine() string
}

// buildLineCommon encapsulates the common logic for building a message line.
func buildLineCommon(
	ctx context.Context,
	content *string,
	messageType *string,
	mentions []*larkim.Mention,
	senderID *string,
	chatID *string,
	timestamp int64,
) (line string) {
	tmpList := make([]string, 0)
	for msgItem := range larkmsgutils.
		GetContentItemsSeq(
			&larkim.EventMessage{
				Content:     content,
				MessageType: messageType,
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
	if *senderID == larkconsts.BotAppID {
		userName = "机器人"
	} else {
		member, err := grouputil.GetUserMemberFromChat(ctx, *chatID, *senderID)
		if err != nil {
			log.Zlog.Error("got error openID", zaplog.String("openID", *senderID))
		}
		if member == nil {
			userName = "NULL"
		} else {
			userName = *member.Name
		}
	}

	createTime := time.UnixMilli(timestamp).In(utility.UTCPlus8Loc()).Format(time.DateTime)
	return fmt.Sprintf("[%s](%s) <%s>: %s", createTime, *senderID, userName, strings.Join(tmpList, ";"))
}
