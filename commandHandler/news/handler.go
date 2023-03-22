package news

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/httptool"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel/attribute"
)

var (
	zapLogger   = utility.ZapLogger
	sugerLogger = utility.SugerLogger
)

var apiKey = os.Getenv("NEWS_API_KEY")

var apiBaseURL = "https://api.itapi.cn/api/hotnews/all"

// NewsData a
type NewsData struct {
	Rank    int    `json:"rank"`
	Name    string `json:"name"`
	ViewNum string `json:"viewnum"`
	URL     string `json:"url"`
}

// NewsDataRaw 原始
type NewsDataRaw struct {
	Data []NewsData `json:"data"`
}

var newsCache = cache.New(5*time.Hour, time.Hour)

// Handler asd
//
//	@param targetID
//	@param quoteID
//	@param authorID
//	@param args
//	@return err
func Handler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.End()

	newsType := "weibo"
	if len(args) > 0 {
		newsType = args[0]
	}
	var res NewsDataRaw
	resCache, found := newsCache.Get(newsType)
	if found {
		res = resCache.(NewsDataRaw)
	} else {
		resp, err := httptool.GetWithParams(httptool.RequestInfo{
			URL:     apiBaseURL,
			Cookies: []*http.Cookie{},
			Params: map[string][]string{
				"type": {newsType},
				"key":  {apiKey},
			},
		})
		if err != nil || resp.StatusCode != http.StatusOK {
			zapLogger.Error("获取新闻失败...状态码："+strconv.Itoa(resp.StatusCode), zaplog.Error(err))
			return err
		}

		resRaw, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			zapLogger.Error("Read Body err", zaplog.Error(err))
			return err
		}

		res = NewsDataRaw{
			Data: make([]NewsData, 0),
		}
		err = json.Unmarshal(resRaw, &res)
		if err != nil {
			zapLogger.Error("Unmarshal err", zaplog.Error(err))
			return err
		}
		newsCache.Set(newsType, res, 0)
	}

	title := fmt.Sprintf("每日%s热榜", newsType)
	if len(res.Data) != 0 {
		modules := make([]interface{}, 0)
		modules = append(modules, betagovar.CardMessageTextModule{
			Type: "header",
			Text: struct {
				Type    string "json:\"type\""
				Content string "json:\"content\""
			}{"plain-text", title + emoji.Newspaper.String()},
		})
		for _, data := range res.Data[:10] {
			modules = append(modules,
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: fmt.Sprintf("%d. [%s](%s) ***%s***", data.Rank, data.Name, data.URL, data.ViewNum),
					},
				},
			)
		}
		cardMessageStr, err := kook.CardMessage{
			&kook.CardMessageCard{
				Theme: "secondary",
				Size:  "lg",
				Modules: append(
					modules,
					&kook.CardMessageSection{
						Mode: kook.CardMessageSectionModeLeft,
						Text: &kook.CardMessageElementKMarkdown{
							Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
						},
					},
				),
			},
		}.BuildMessage()
		if err != nil {
			return err
		}
		betagovar.GlobalSession.MessageCreate(
			&kook.MessageCreate{
				MessageCreateBase: kook.MessageCreateBase{
					Type:     kook.MessageTypeCard,
					TargetID: targetID,
					Content:  cardMessageStr,
					Quote:    quoteID,
				},
			},
		)
	}
	return
}
