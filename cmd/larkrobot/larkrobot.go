package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/BetaGoRobot/BetaGo/larkcards"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/bytedance/sonic"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
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
	}
	return *resp.Data.ImageKey, err
}

func SendMusicCard(songID string) {
}

func main() {
	// 注册事件回调
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			if !larkcards.IsMentioned(event.Event.Message.Mentions) {
				return nil
			}
			msgMap := make(map[string]interface{})
			msg := *event.Event.Message.Content
			err := sonic.UnmarshalString(msg, &msgMap)
			if err != nil {
				return err
			}
			if text, ok := msgMap["text"]; ok {
				msg = text.(string)
			}
			res, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, msg)
			if err != nil {
				return err
			}
			imageKey, err := uploadPic(ctx, res[0].PicURL)
			if err != nil {
				return err
			}
			lyrics := neteaseapi.NetEaseGCtx.GetLyrics(ctx, res[0].ID)
			lyrics = larkcards.TrimLyrics(lyrics)
			lyrics = strings.ReplaceAll(lyrics, "\n", "\n\n")
			cardStr := larkcards.GenerateMusicCardByStruct(imageKey, res[0].Name, res[0].ArtistName, res[0].SongURL, lyrics)
			fmt.Println(cardStr)
			req := larkim.NewReplyMessageReqBuilder().Body(
				larkim.NewReplyMessageReqBodyBuilder().Content(cardStr).MsgType(larkim.MsgTypeInteractive).ReplyInThread(false).Uuid(*event.Event.Message.MessageId).Build(),
			).MessageId(*event.Event.Message.MessageId).Build()
			resp, err := larkClient.Im.V1.Message.Reply(ctx, req)
			if err != nil {
				return err
			}
			fmt.Println(resp)
			return nil
		}).
		OnCustomizedEvent("message", func(ctx context.Context, event *larkevent.EventReq) error {
			fmt.Printf("[ OnCustomizedEvent access ], type: message, data: %s\n", string(event.Body))
			return nil
		})
	// 创建Client
	cli := larkws.NewClient(os.Getenv("LARK_CLIENT_ID"), os.Getenv("LARK_SECRET"),
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelDebug),
	)
	// 启动客户端
	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
