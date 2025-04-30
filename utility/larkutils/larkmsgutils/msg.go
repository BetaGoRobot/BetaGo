package larkmsgutils

import (
	"iter"

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
		return trans2Item(*m.MessageType, *m.Content)
	case *larkim.Message:
		// 处理普通消息
		return trans2Item(*m.MsgType, *m.Body.Content)
	}
	return nil
}

// trans2Item to be filled
//
//	@param msgType string
//	@param content string
//	@return itemList []*Item
//	@author kevinmatthe
//	@update 2025-04-30 14:04:48
func trans2Item(msgType, content string) (itemList iter.Seq[*Item]) {
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
					if ele2.Tag == "at" {
						if !yield(&Item{Tag: "at", Content: ele2.UserID}) {
							return
						}
					} else if ele2.Tag == "text" {
						if !yield(&Item{Tag: "text", Content: ele2.Text}) {
							return
						}
					} else if ele2.Tag == "image" {
						if !yield(&Item{Tag: "image", Content: ele2.ImageKey}) {
							return
						}
					} else if ele2.Tag == "sticker" {
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
