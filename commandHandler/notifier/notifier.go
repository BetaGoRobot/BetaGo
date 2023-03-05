package notifier

import (
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/httptool"
	"github.com/BetaGoRobot/BetaGo/scheduletask"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
)

// StartAutoService  启动自动服务
func StartAutoService() {
	StartUpMessage(betagovar.GlobalSession)
	go scheduletask.DailyRecommand()
	go scheduletask.DailyRate()
	go scheduletask.DailyNews()
}

// StartUpMessage  启动时的消息
//
//	@param session
//	@return err
func StartUpMessage(session *kook.Session) (err error) {
	// StartUp for debug:
	currentIP, err := httptool.GetPubIP()
	if err != nil {
		return
	}
	go func() {
		cardMessage, _ := kook.CardMessage{
			&kook.CardMessageCard{
				Theme: kook.CardThemeInfo,
				Size:  kook.CardSizeLg,
				Modules: []interface{}{
					kook.CardMessageHeader{
						Text: kook.CardMessageElementText{
							Content: emoji.DesertIsland.String() + "Online Notifacation" + emoji.Information.String(),
							Emoji:   false,
						},
					},
					kook.CardMessageSection{
						Text: kook.CardMessageElementKMarkdown{
							Content: strings.Join(
								[]string{
									"Name: \t**", betagovar.RobotName, "**\n",
									"Time: \t**", time.Now().Add(time.Hour * 8).Local().Format("2006-01-02 15:04:05"), "**\n",
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
			&kook.MessageCreate{
				MessageCreateBase: kook.MessageCreateBase{
					Type:     kook.MessageTypeCard,
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
		cardMessage, _ := kook.CardMessage{
			&kook.CardMessageCard{
				Theme: kook.CardThemeInfo,
				Size:  kook.CardSizeLg,
				Modules: []interface{}{
					kook.CardMessageHeader{
						Text: kook.CardMessageElementText{
							Content: emoji.DesertIsland.String() + "BetaGo更新信息" + emoji.Information.String(),
							Emoji:   false,
						},
					},
					kook.CardMessageSection{
						Text: kook.CardMessageElementKMarkdown{
							Content: strings.Join(
								[]string{
									"Time: \t**", time.Now().Add(time.Hour * 8).Format("2006-01-02 15:04:05"), "**\n",
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
			&kook.MessageCreate{
				MessageCreateBase: kook.MessageCreateBase{
					Type:     kook.MessageTypeCard,
					TargetID: betagovar.BetaGoUpdateChanID,
					Content:  cardMessage,
				},
			},
		)
	}()
	return
}

// OfflineMessage 离线时的消息
//
//	@param session
//	@return err
func OfflineMessage(session *kook.Session) (err error) {
	currentIP, err := httptool.GetPubIP()
	if err != nil {
		return
	}
	cardMessage, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.DesertIsland.String() + "Offline Notifacation" + emoji.Information.String(),
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: strings.Join([]string{
							"Name: \t**", betagovar.RobotName, "**\n",
							"Time: \t**", time.Now().Add(time.Hour * 8).Format("2006-01-02 15:04:05"), "**\n",
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
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: betagovar.TestChanID,
				Content:  cardMessage,
			},
		},
	)
	return
}
