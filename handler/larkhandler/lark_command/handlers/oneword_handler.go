package handlers

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/hitokoto"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/enescakir/emoji"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const (
	yiyanURL     = "https://api.fanlisky.cn/niuren/getSen"
	yiyanPoemURL = "https://v1.jinrishici.com/all.json"
)

// RespBody  一言返回体
type RespBody struct {
	ID         int         `json:"id"`
	UUID       string      `json:"uuid"`
	Hitokoto   string      `json:"hitokoto"`
	Type       string      `json:"type"`
	From       string      `json:"from"`
	FromWho    interface{} `json:"from_who"`
	Creator    string      `json:"creator"`
	CreatorUID int         `json:"creator_uid"`
	Reviewer   int         `json:"reviewer"`
	CommitFrom string      `json:"commit_from"`
	CreatedAt  string      `json:"created_at"`
	Length     int         `json:"length"`
}

func OneWordHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	oneWordArgs := []string{}

	argsMap, _ := parseArgs(args...)
	wordType, ok := argsMap["type"]
	if ok {
		switch wordType {
		case "二次元":
			oneWordArgs = append(oneWordArgs, []string{"a", "b"}...)
		case "游戏":
			oneWordArgs = append(oneWordArgs, "c")
		case "文学":
			oneWordArgs = append(oneWordArgs, "d")
		case "原创":
			oneWordArgs = append(oneWordArgs, "e")
		case "网络":
			oneWordArgs = append(oneWordArgs, "f")
		case "其他":
			oneWordArgs = append(oneWordArgs, "g")
		case "影视":
			oneWordArgs = append(oneWordArgs, "h")
		case "诗词":
			oneWordArgs = append(oneWordArgs, "i")
		case "网易云":
			oneWordArgs = append(oneWordArgs, "j")
		case "哲学":
			oneWordArgs = append(oneWordArgs, "k")
		case "抖机灵":
			oneWordArgs = append(oneWordArgs, "l")
		}
	}

	hitokotoRes, err := hitokoto.GetHitokoto(oneWordArgs...)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("%s 很喜欢《%s》中的一句话\n%s", emoji.Mountain.String(), hitokotoRes.From, hitokotoRes.Hitokoto)
	err = larkutils.ReplyMsgText(ctx, msg, *data.Event.Message.MessageId, "_oneWord", false)
	return
}
