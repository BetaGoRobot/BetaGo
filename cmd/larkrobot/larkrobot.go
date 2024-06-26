package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/dal/larkcards"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"

	"github.com/bytedance/sonic"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"go.opentelemetry.io/otel/attribute"
)

func getMusicAndSend(ctx context.Context, event *larkim.P2MessageReceiveV1, msg string) (err error) {
	defer larkutils.RecoverMsg(ctx, *event.Event.Message.MessageId)
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()

	keywords := strings.Split(msg, " ")[1:]
	if keyword := strings.ToLower(strings.Join(keywords, " ")); keyword == "try panic" {
		panic("try panic!")
	}
	res, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, keywords...)
	if err != nil {
		return err
	}
	listMsg := larkcards.NewSearchListCard()
	for _, item := range res {
		var invalid bool
		if item.SongURL == "" { // 无效歌曲
			invalid = true
		}
		listMsg.AddColumn(ctx, item.ImageKey, item.Name, item.ArtistName, item.ID, invalid)
	}
	listMsg.AddJaegerTracer(ctx, span)
	cardStr, err := sonic.MarshalString(listMsg)
	if err != nil {
		return err
	}

	fmt.Println(cardStr)
	// larkutils.SendEphemeral(ctx, *event.Event.Message.ChatId, *event.Event.Sender.SenderId.OpenId, cardStr)
	req := larkim.NewReplyMessageReqBuilder().
		Body(
			larkim.NewReplyMessageReqBodyBuilder().
				Content(cardStr).
				MsgType(larkim.MsgTypeInteractive).
				ReplyInThread(true).
				Uuid(*event.Event.Message.MessageId).
				Build(),
		).MessageId(*event.Event.Message.MessageId).
		Build()
	_, subSpan := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	subSpan.End()

	if err != nil {
		fmt.Println(resp)
		return err
	}
	fmt.Println(resp.CodeError.Msg)
	return
}

func longConn() { // 注册事件回调
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(
			func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
				ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
				defer larkutils.RecoverMsg(ctx, *event.Event.Message.MessageId)
				span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
				defer span.End()

				stamp, err := strconv.ParseInt(*event.Event.Message.CreateTime, 10, 64)
				if err != nil {
					panic(err)
				}
				if time.Now().Sub(time.Unix(stamp, 0)) > time.Second*10 {
					return nil
				}
				if !larkcards.IsMentioned(event.Event.Message.Mentions) || *event.Event.Sender.SenderId.OpenId == larkcards.BotOpenID {
					return nil
				}
				msgMap := make(map[string]interface{})
				msg := *event.Event.Message.Content
				err = sonic.UnmarshalString(msg, &msgMap)
				if err != nil {
					return err
				}
				if text, ok := msgMap["text"]; ok {
					msg = text.(string)
				}
				go getMusicAndSend(ctx, event, msg)
				fmt.Println(larkcore.Prettify(event))
				return nil
			},
		)
	// 创建Client
	cli := larkws.NewClient(env.LarkAppID, env.LarkAppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)
	// 启动客户端
	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

func GetCardMusicByPage(ctx context.Context, musicID string, page int) string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	const (
		maxSingleLineLen = 48
		maxPageSize      = 9
	)
	musicURL, err := neteaseapi.NetEaseGCtx.GetMusicURL(ctx, musicID)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return ""
	}

	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]
	picURL := songDetail.Al.PicURL
	imageKey, ossURL, err := larkutils.UploadPicAllinOne(ctx, picURL, musicID, true)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return ""
	}

	lyrics, lyricsURL := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyrics = larkcards.TrimLyrics(lyrics)

	artistNameLissst := make([]map[string]string, 0)
	for _, ar := range songDetail.Ar {
		artistNameLissst = append(artistNameLissst, map[string]string{"name": ar.Name})
	}
	artistJSON, err := sonic.MarshalString(artistNameLissst)
	if err != nil {
		log.ZapLogger.Error(err.Error())
	}
	lyricsURL = utility.BuildURL(lyricsURL, musicURL, ossURL, songDetail.Al.Name, songDetail.Name, artistJSON, songDetail.Dt)
	// eg: page = 1
	quotaRemain := maxPageSize
	lyricList := strings.Split(lyrics, "\n")
	newList := make([]string, 0)
	curPage := 1
	for _, l := range lyricList {
		quotaRemain--
		if len(l) > maxSingleLineLen {
			quotaRemain--
		}
		if quotaRemain <= 0 {
			curPage++
			quotaRemain = maxPageSize
			if curPage > page {
				break
			}
		}
		if curPage == page {
			newList = append(newList, l)
		}
	}

	lyrics = strings.Join(newList, "\n")

	lyrics = strings.ReplaceAll(lyrics, "\n", "\n\n")
	cardStr := larkcards.GenerateMusicCardByStruct(ctx, imageKey, songDetail.Name, songDetail.Ar[0].Name, lyricsURL, lyrics, musicID)
	return cardStr
}

func SendMusicCard(ctx context.Context, musicID string, msgID string, page int) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	cardStr := GetCardMusicByPage(ctx, musicID, page)
	fmt.Println(cardStr)
	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(cardStr).MsgType(larkim.MsgTypeInteractive).ReplyInThread(true).Uuid(msgID + musicID).Build(),
	).MessageId(msgID).Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return
	}

	fmt.Println(resp)
}

func HandleFullLyrics(ctx context.Context, musicID, msgID string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("musicID").String(musicID))
	defer span.End()

	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]

	lyric, _ := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyric = larkcards.TrimLyrics(lyric)
	sp := strings.Split(lyric, "\n")
	left := strings.Join(sp[:len(sp)/2], "\n\n")
	right := strings.Join(sp[len(sp)/2+1:], "\n\n")
	cardStr := larkcards.GenFullLyricsCard(ctx, songDetail.Name, songDetail.Ar[0].Name, left, right)

	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(cardStr).MsgType(larkim.MsgTypeInteractive).ReplyInThread(true).Uuid(msgID).Build(),
	).MessageId(msgID).Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return
	}
	fmt.Println(resp.CodeError.Msg)
}

func webHook() {
	// 创建 card 处理器
	cardHandler := larkcard.NewCardActionHandler(os.Getenv("LARK_VERIFICATION"), os.Getenv("LARK_ENCRYPTION"), func(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
		ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
		defer larkutils.RecoverMsg(ctx, cardAction.OpenMessageID)
		span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(cardAction)))
		defer span.End()

		// 处理 cardAction, 这里简单打印卡片内容
		if musicID, ok := cardAction.Action.Value["show_music"]; ok {
			go SendMusicCard(ctx, musicID.(string), cardAction.OpenMessageID, 1)
		}
		if musicID, ok := cardAction.Action.Value["music_id"]; ok {
			go HandleFullLyrics(ctx, musicID.(string), cardAction.OpenMessageID)
		}
		// 无返回值示例
		return nil, nil
	})

	// 注册处理器
	http.HandleFunc("/webhook/card", httpserverext.NewCardActionHandlerFunc(cardHandler, larkevent.WithLogLevel(larkcore.LogLevelDebug)))

	// 启动 http 服务
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	go longConn()
	go webHook()
	select {}
}
