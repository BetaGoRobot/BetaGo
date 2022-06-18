package roll

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/yiyan"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// OneWordHandler  获取一言
//  @param targetID
//  @param quoteID
//  @param authorID
//  @return err
func OneWordHandler(targetID, quoteID, authorID string, args ...string) (err error) {
	// 构建CardMessage
	poemMap := yiyan.GetPoem()
	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: emoji.Mountain.String() + "来自一言的推荐:",
						Emoji:   true,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementText{
						Content: poemMap["content"].(string) + "\n" + poemMap["author"].(string) + " --《" + poemMap["origin"].(string) + "》",
						Emoji:   true,
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		return
	}
	_, err = betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    quoteID,
		},
	})
	return
}
