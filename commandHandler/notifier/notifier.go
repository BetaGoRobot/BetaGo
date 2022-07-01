package notifier

import (
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// StartUpMessage  启动时的消息
//  @param session
//  @return err
func StartUpMessage(session *khl.Session) (err error) {
	// StartUp for debug:
	currentIP, err := utility.GetOutBoundIP()
	if err != nil {
		return
	}
	go func() {
		cardMessage, _ := khl.CardMessage{
			&khl.CardMessageCard{
				Theme: khl.CardThemeInfo,
				Size:  khl.CardSizeLg,
				Modules: []interface{}{
					khl.CardMessageHeader{
						Text: khl.CardMessageElementText{
							Content: emoji.DesertIsland.String() + "Online Notifacation" + emoji.Information.String(),
							Emoji:   false,
						},
					},
					khl.CardMessageSection{
						Text: khl.CardMessageElementKMarkdown{
							Content: strings.Join(
								[]string{
									"Name: \t**", betagovar.RobotName, "**\n",
									"Time: \t**", time.Now().Local().Format("2006-01-02 15:04:05"), "**\n",
									"IP: \t**", currentIP, "**\n",
									"Message: \t**" + betagovar.CommitMessage + "**\n",
									"Commit-Page: \t[CommitPage](", betagovar.HTMLURL, ")\n",
									"LeaveYourCommentHere: \t[CommentPage](", betagovar.CommentsURL, ")\n",
								},
								""),
						},
					},
				},
			},
		}.BuildMessage()
		session.MessageCreate(
			&khl.MessageCreate{
				MessageCreateBase: khl.MessageCreateBase{
					Type:     khl.MessageTypeCard,
					TargetID: betagovar.TestChanID,
					Content:  cardMessage,
				},
			},
		)
	}()
	if betagovar.BetaGoTest {
		return
	}
	go func() { // StartUp for info:
		cardMessage, _ := khl.CardMessage{
			&khl.CardMessageCard{
				Theme: khl.CardThemeInfo,
				Size:  khl.CardSizeLg,
				Modules: []interface{}{
					khl.CardMessageHeader{
						Text: khl.CardMessageElementText{
							Content: emoji.DesertIsland.String() + "BetaGo更新信息" + emoji.Information.String(),
							Emoji:   false,
						},
					},
					khl.CardMessageSection{
						Text: khl.CardMessageElementKMarkdown{
							Content: strings.Join(
								[]string{
									"Time: \t**", time.Now().Format("2006-01-02 15:04:05"), "**\n",
									"更新内容: \t**" + betagovar.CommitMessage + "**\n",
									"Commit-Page: \t[CommitPage](", betagovar.HTMLURL, ")\n",
								},
								""),
						},
					},
				},
			},
		}.BuildMessage()
		session.MessageCreate(
			&khl.MessageCreate{
				MessageCreateBase: khl.MessageCreateBase{
					Type:     khl.MessageTypeCard,
					TargetID: betagovar.BetaGoUpdateChanID,
					Content:  cardMessage,
				},
			},
		)
	}()
	return
}

// OfflineMessage 离线时的消息
//  @param session
//  @return err
func OfflineMessage(session *khl.Session) (err error) {
	currentIP, err := utility.GetOutBoundIP()
	if err != nil {
		return
	}
	cardMessage, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: emoji.DesertIsland.String() + "Offline Notifacation" + emoji.Information.String(),
						Emoji:   false,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementKMarkdown{
						Content: strings.Join([]string{
							"Name: \t**", betagovar.RobotName, "**\n",
							"Time: \t**", time.Now().Format("2006-01-02 15:04:05"), "**\n",
							"IP: \t**", currentIP, "**\n",
							"Message: \t**", betagovar.CommitMessage, "**\n",
							"Commit-Page: \t[CommitPage](", betagovar.HTMLURL, ")\n",
							"LeaveYourCommentHere: \t[CommentPage](", betagovar.CommentsURL, ")\n"},
							""),
					},
				},
			},
		},
	}.BuildMessage()
	session.MessageCreate(
		&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     khl.MessageTypeCard,
				TargetID: betagovar.TestChanID,
				Content:  cardMessage,
			},
		},
	)
	return
}
