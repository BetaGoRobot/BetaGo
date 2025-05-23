package hitokoto

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/enescakir/emoji"
	jsoniter "github.com/json-iterator/go"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	hitokotoURL = "https://v1.hitokoto.cn"
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

// GetHitokotoHandler 获取一言
//
//	@param targetID
//	@param msgID
//	@param authorID
//	@param args
//	@return err
func GetHitokotoHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	params := make([]string, 0)
	for index := range args {
		switch args[index] {
		case "二次元":
			params = append(params, []string{"a", "b"}...)
		case "游戏":
			params = append(params, "c")
		case "文学":
			params = append(params, "d")
		case "原创":
			params = append(params, "e")
		case "网络":
			params = append(params, "f")
		case "其他":
			params = append(params, "g")
		case "影视":
			params = append(params, "h")
		case "诗词":
			params = append(params, "i")
		case "网易云":
			params = append(params, "j")
		case "哲学":
			params = append(params, "k")
		case "抖机灵":
			params = append(params, "l")
		}
	}
	hitokotoRes, err := GetHitokoto(params...)
	if err != nil {
		return
	}
	cardMessageStr, err := utility.BuildCardMessage(
		"info",
		"lg",
		"",
		quoteID,
		span,
		kook.CardMessageHeader{
			Text: kook.CardMessageElementText{
				Content: fmt.Sprintf("%s 很喜欢《%s》中的一句话", emoji.Mountain.String(), hitokotoRes.From),
				Emoji:   true,
			},
		},
		kook.CardMessageSection{
			Text: kook.CardMessageElementText{
				Content: hitokotoRes.Hitokoto + "\n",
				Emoji:   true,
			},
		},
	)
	if err != nil {
		return
	}
	_, err = consts.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    quoteID,
		},
	})
	fmt.Println(cardMessageStr)
	return
}

// GetHitokoto 获取一言
//
//	@param parameters
func GetHitokoto(field ...string) (hitokotoRes RespBody, err error) {
	resp, err := requests.Req().SetQueryParamsFromValues(map[string][]string{"c": field}).Get(hitokotoURL)
	if err != nil {
		log.Zlog.Error("获取一言失败", zaplog.Error(err))
		return
	}
	if err = json.Unmarshal(resp.Body(), &hitokotoRes); err != nil {
		log.Zlog.Error("获取一言失败", zaplog.Error(err))
		return
	}
	return
}
