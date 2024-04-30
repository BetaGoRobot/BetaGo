package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BetaGoRobot/BetaGo/larkcards"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/bytedance/sonic"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"go.opentelemetry.io/otel/attribute"
)

var larkClient *lark.Client = lark.NewClient(os.Getenv("LARK_CLIENT_ID"), os.Getenv("LARK_SECRET"))

func uploadPic(ctx context.Context, imageURL string) (key string, err error) {
	picResp, err := http.Get(imageURL)
	req := larkim.NewCreateImageReqBuilder().
		Body(
			larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkim.ImageTypeMessage).
				Image(picResp.Body).
				Build(),
		).
		Build()
	resp, err := larkClient.Im.Image.Create(ctx, req)
	if err != nil {
		log.Println(err)
		return
	}
	return *resp.Data.ImageKey, err
}

func getMusicAndSend(ctx context.Context, event *larkim.P2MessageReceiveV1, msg string) (err error) {
	ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()

	msg = strings.Split(msg, " ")[1]
	res, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, msg)
	if err != nil {
		return err
	}
	listMsg := larkcards.NewSearchListCard()
	var (
		tmpChan = make(chan *struct {
			Index    int
			imageKey string
			Name     string
			Artist   string
			ID       string
		})
		wg = &sync.WaitGroup{}
	)
	go func() {
		for i, r := range res {
			wg.Add(1)
			go func(index int, res neteaseapi.SearchMusicRes) {
				imageKey, err := uploadPic(ctx, res.PicURL)
				if err != nil {
					return
				}
				tmpChan <- &struct {
					Index    int
					imageKey string
					Name     string
					Artist   string
					ID       string
				}{
					Index:    index,
					imageKey: imageKey,
					Name:     res.Name,
					Artist:   res.ArtistName,
					ID:       res.ID,
				}
				wg.Done()
			}(i, *r)

		}
		wg.Wait()
		close(tmpChan)
	}()
	tmpList := make([]*struct {
		Index    int
		imageKey string
		Name     string
		Artist   string
		ID       string
	}, 0)
	for item := range tmpChan {
		tmpList = append(tmpList, item)
	}
	sort.Slice(tmpList, func(i, j int) bool {
		return tmpList[i].Index < tmpList[j].Index
	})
	for _, item := range tmpList {
		listMsg.AddColumn(ctx, item.imageKey, item.Name, item.Artist, item.ID)
	}
	listMsg.AddJaegerTracer(ctx, span)
	cardStr, err := sonic.MarshalString(listMsg)
	if err != nil {
		return err
	}

	fmt.Println(cardStr)
	req := larkim.NewReplyMessageReqBuilder().
		Body(
			larkim.NewReplyMessageReqBodyBuilder().
				Content(cardStr).
				MsgType(larkim.MsgTypeInteractive).
				ReplyInThread(false).
				Uuid(*event.Event.Message.MessageId).
				Build(),
		).MessageId(*event.Event.Message.MessageId).
		Build()

	resp, err := larkClient.Im.V1.Message.Reply(ctx, req)
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
				ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
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
	cli := larkws.NewClient(os.Getenv("LARK_CLIENT_ID"), os.Getenv("LARK_SECRET"),
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
	ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	const (
		maxSingleLineLen = 48
		maxPageSize      = 9
	)
	musicURL, err := neteaseapi.NetEaseGCtx.GetMusicURL(musicID)
	if err != nil {
		log.Println(err)
		return ""
	}

	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]
	picURL := songDetail.Al.PicURL
	imageKey, err := uploadPic(ctx, picURL)
	if err != nil {
		log.Println(err)
		return ""
	}
	lyric := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyric = larkcards.TrimLyrics(lyric)
	// eg: page = 1
	quotaRemain := maxPageSize
	lyricList := strings.Split(lyric, "\n")
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

	lyric = strings.Join(newList, "\n")

	lyric = strings.ReplaceAll(lyric, "\n", "\n\n")
	cardStr := larkcards.GenerateMusicCardByStruct(ctx, imageKey, songDetail.Name, songDetail.Ar[0].Name, musicURL, lyric, musicID)
	return cardStr
}

func SendMusicCard(ctx context.Context, musicID string, msgID string, page int) {
	ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	cardStr := GetCardMusicByPage(ctx, musicID, page)
	fmt.Println(cardStr)
	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(cardStr).MsgType(larkim.MsgTypeInteractive).ReplyInThread(false).Uuid(msgID).Build(),
	).MessageId(msgID).Build()
	resp, err := larkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return
	}

	fmt.Println(resp)
}

func HandleFullLyrics(ctx context.Context, musicID, msgID string) {
	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]

	lyric := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyric = larkcards.TrimLyrics(lyric)
	sp := strings.Split(lyric, "\n")
	left := strings.Join(sp[:len(sp)/2], "\n\n")
	right := strings.Join(sp[len(sp)/2+1:], "\n\n")
	cardStr := larkcards.GenFullLyricsCard(ctx, songDetail.Name, songDetail.Ar[0].Name, left, right)

	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(cardStr).MsgType(larkim.MsgTypeInteractive).ReplyInThread(false).Uuid(msgID).Build(),
	).MessageId(msgID).Build()
	resp, err := larkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return
	}
	fmt.Println(resp.CodeError.Msg)
}

func webHook() {
	// 创建 card 处理器
	cardHandler := larkcard.NewCardActionHandler(os.Getenv("LARK_VERIFICATION"), os.Getenv("LARK_ENCRYPTION"), func(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
		ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
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
