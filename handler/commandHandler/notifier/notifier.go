package notifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/handler/scheduletask"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
)

// StartAutoService  启动自动服务
func StartAutoService() {
	StartUpMessage(consts.GlobalSession)
	go scheduletask.DailyTask()
	go scheduletask.OnlineTest()
}

// StartUpMessage  启动时的消息
//
//	@param session
//	@return err
func StartUpMessage(session *kook.Session) (err error) {
	RestartMsgID, _ := redis.GetRedisClient().GetDel(context.Background(), "RestartMsgID").Result()
	RestartTargetID, _ := redis.GetRedisClient().GetDel(context.Background(), "RestartTargetID").Result()
	RestartAuthorID, _ := redis.GetRedisClient().GetDel(context.Background(), "RestartAuthorID").Result()
	utility.SendMessageTempAndDelete(RestartTargetID, RestartMsgID, RestartAuthorID, "重启成功。")
	// StartUp for debug:
	currentIP, err := utility.GetPubIP()
	if err != nil {
		return
	}
	go func() {
		title := emoji.DesertIsland.String() + "Online Notifacation" + emoji.Information.String()

		content := strings.Join(
			[]string{
				"Name: \t**", consts.RobotName, "**\n",
				"Time: \t**", time.Now().Add(time.Hour * 8).Local().Format("2006-01-02 15:04:05"), "**\n",
				"IP: \t**", currentIP, "**\n",
				"Message: \t**" + env.GitCommitMessage + "**\n",
				"Commit-Page: \t[CommitPage](", fmt.Sprintf("https://github.com/BetaGoRobot/BetaGo/commit/%s", env.GithubSha), ")\n",
				"LeaveYourCommentHere: \t[CommentPage](", consts.CommentsURL, ")\n",
			},
			"")
		cardMessage, _ := kook.CardMessage{
			&kook.CardMessageCard{
				Theme: kook.CardThemeInfo,
				Size:  kook.CardSizeLg,
				Modules: []interface{}{
					kook.CardMessageHeader{
						Text: kook.CardMessageElementText{
							Content: title,
							Emoji:   false,
						},
					},
					kook.CardMessageSection{
						Text: kook.CardMessageElementKMarkdown{
							Content: content,
						},
					},
				},
			},
		}.BuildMessage()
		gotify.SendMessage(context.Background(), title, content, 5)
		session.MessageCreate(
			&kook.MessageCreate{
				MessageCreateBase: kook.MessageCreateBase{
					Type:     kook.MessageTypeCard,
					TargetID: consts.TestChanID,
					Content:  cardMessage,
				},
			},
		)
	}()
	if consts.BetaGoTest {
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
									"更新内容: \t**" + consts.CommitMessage + "**\n",
									"Commit-Page: \t[CommitPage](", consts.HTMLURL, ")\n",
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
					TargetID: consts.BetaGoUpdateChanID,
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
	currentIP, err := utility.GetPubIP()
	if err != nil {
		return
	}
	title := emoji.DesertIsland.String() + "Offline Notifacation" + emoji.Information.String()
	content := strings.Join([]string{
		"Name: \t**", consts.RobotName, "**\n",
		"Time: \t**", time.Now().Add(time.Hour * 8).Format("2006-01-02 15:04:05"), "**\n",
		"IP: \t**", currentIP, "**\n",
		"Message: \t**", consts.CommitMessage, "**\n",
		"Commit-Page: \t[CommitPage](", consts.HTMLURL, ")\n",
		"LeaveYourCommentHere: \t[CommentPage](", consts.CommentsURL, ")\n",
	},
		"")
	cardMessage, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: title,
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: content,
					},
				},
			},
		},
	}.BuildMessage()
	gotify.SendMessage(context.Background(), title, content, 5)
	session.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: consts.TestChanID,
				Content:  cardMessage,
			},
		},
	)
	return
}
