package larkmsgutils

import (
	"fmt"
	"iter"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// MsgConstraints  to be filled
//
//	@author kevinmatthe
//	@update 2025-04-30 13:37:42
type MsgConstraints interface {
	*larkim.EventMessage | *larkim.Message
}

// GetContentItemsSeq to be filled
//
//	@param msg T
//	@return msgType string
//	@return msgContent string
//	@author kevinmatthe
//	@update 2025-04-30 13:37:40
func GetContentItemsSeq[T MsgConstraints](msg T) iter.Seq[*Item] {
	switch m := any(msg).(type) {
	case *larkim.EventMessage:
		// 处理事件消息
		return Trans2Item(*m.MessageType, *m.Content)
	case *larkim.Message:
		// 处理普通消息
		return Trans2Item(*m.MsgType, *m.Body.Content)
	}
	return nil
}

// Trans2Item to be filled
//
//	@param msgType string
//	@param content string
//	@return itemList []*Item
//	@author kevinmatthe
//	@update 2025-04-30 14:04:48
func Trans2Item(msgType, content string) (itemList iter.Seq[*Item]) {
	return func(yield func(*Item) bool) {
		switch msgType {
		case "text": // text是处理过的，直接返回
			if !yield(&Item{Tag: "text", Content: content}) {
				return
			}
		case "post":
			res := utility.MustUnmarshallString[postMsg](content)
			for _, ele := range res.Content {
				for _, ele2 := range ele {
					switch ele2.Tag {
					case "at":
						if !yield(&Item{Tag: "at", Content: ele2.UserID}) {
							return
						}
					case "text":
						if !yield(&Item{Tag: "text", Content: ele2.Text}) {
							return
						}
					case "image":
						if !yield(&Item{Tag: "image", Content: ele2.ImageKey}) {
							return
						}
					case "sticker":
						if !yield(&Item{Tag: "sticker", Content: ele2.FileKey}) {
							return
						}
					}
				}
			}
		case "image":
			res := utility.MustUnmarshallString[imageMsg](content)
			if !yield(&Item{Tag: "image", Content: res.ImageKey}) {
				return
			}
		case "file":
			res := utility.MustUnmarshallString[fileMsg](content)
			if !yield(&Item{Tag: "file", Content: res.FileKey}) {
				return
			}
		}
	}
}

func TagText(text string, color string) string {
	return fmt.Sprintf("<text_tag color='%s'>%s</text_tag>", color, text)
}

func AtUser(userID, userName string) string {
	return fmt.Sprintf("<at user_id=\"%s\">%s</at>", userID, userName)
}

type Mention struct {
	Key string `json:"key"`
	ID  struct {
		UserID  string `json:"user_id"`
		OpenID  string `json:"open_id"`
		UnionID string `json:"union_id"`
	} `json:"id"`
	Name      string `json:"name"`
	TenantKey string `json:"tenant_key"`
}

// ReplaceMentionToName 将@user_1 替换成 name
func ReplaceMentionToName(input string, mentions []*Mention) string {
	if mentions != nil {
		for _, mention := range mentions {
			// input = strings.ReplaceAll(input, mention.Key, fmt.Sprintf("<at user_id=\\\"%s\\\">%s</at>", mention.ID.UserID, mention.Name))
			input = strings.ReplaceAll(input, mention.Key, "")
			if len(input) > 0 && string(input[0]) == "/" {
				if inputs := strings.Split(input, " "); len(inputs) > 0 {
					input = strings.Join(inputs[1:], " ")
				}
			}

		}
	}
	return input
}
