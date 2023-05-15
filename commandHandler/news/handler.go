package news

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
	"github.com/patrickmn/go-cache"
	"github.com/spyzhov/ajson"
	"go.opentelemetry.io/otel/attribute"
)

var (
	zapLogger   = utility.ZapLogger
	sugerLogger = utility.SugerLogger
)

var apiKey = os.Getenv("NEWS_API_KEY")

var apiBaseURL = "https://api.itapi.cn/api/hotnews/all"

var apiDailyMorningReport = "https://v2.alapi.cn/api/zaobao"

// NewsData a
type NewsData struct {
	Rank    int         `json:"rank"`
	Name    string      `json:"name"`
	ViewNum interface{} `json:"viewnum"`
	URL     string      `json:"url"`
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
	defer span.RecordError(err)
	defer span.End()

	newsType := "weibo"
	if len(args) > 0 {
		newsType = args[0]
	}
	if newsType == "morning" {
		return MorningHandler(ctx, targetID, quoteID, authorID)
	}
	var res NewsDataRaw
	if resCache, found := newsCache.Get(newsType); found {
		res = resCache.(NewsDataRaw)
	} else {
		resp, err := betagovar.HttpClient.R().
			SetQueryParam("key", apiKey).
			SetQueryParam("type", newsType).
			Get(apiBaseURL)
		if err != nil {
			zapLogger.Error("获取新闻失败...", zaplog.Error(err))
			return err
		}

		res = NewsDataRaw{
			Data: make([]NewsData, 0),
		}
		err = json.Unmarshal(resp.Body(), &res)
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
			if !strings.HasPrefix(data.URL, "http") {
				data.URL = "http:" + data.URL
			}
			modules = append(modules,
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: fmt.Sprintf("%d. %s [%s](%s) **%s**", data.Rank, data.Name, data.URL, data.URL, data.ViewNum),
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
					&kook.CardMessageDivider{},
					utility.GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
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

// MorningHandler  每日早报
//
//	@param ctx
//	@param targetID
//	@param quoteID
//	@param authorID
//	@param args
//	@return err
func MorningHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	var (
		newsList = make([]string, 0)
		imageURL string
	)
	cacheRes, found := newsCache.Get("Morning")
	if found {
		imageURL = cacheRes.([]string)[0]
		newsList = newsList[1:]
	} else {
		resp, err := betagovar.HttpClient.R().
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetBody(fmt.Sprintf("token=%s&format=json", apiKey)).
			Post(apiDailyMorningReport)
		if err != nil || resp.StatusCode() != 200 {
			zapLogger.Error("获取新闻失败...", zaplog.Error(err))
			return fmt.Errorf("StatusCode: %d, err is %v", resp.StatusCode(), err)
		}
		fmt.Println(resp)
		newsNode, err := ajson.JSONPath(resp.Body(), "$.data.news")
		if err != nil {
			return err
		}
		recordList, _ := newsNode[0].GetArray()
		imageNode, err := ajson.JSONPath(resp.Body(), "$.data.head_image")
		if err != nil {
			return err
		}
		imageURL, _ = imageNode[0].GetString()

		for _, r := range recordList {
			newsString, _ := r.GetString()
			newsList = append(newsList, newsString)
		}
		newsCache.Set("morning", append([]string{imageURL}, newsList...), cache.DefaultExpiration)
	}

	modules := make([]interface{}, 0)
	modules = append(modules,
		betagovar.CardMessageTextModule{
			Type: "header",
			Text: struct {
				Type    string "json:\"type\""
				Content string "json:\"content\""
			}{"plain-text", "每日早报" + emoji.Newspaper.String()},
		},
		kook.CardMessageContainer{
			kook.CardMessageElementImage{
				Src:  imageURL,
				Size: string(kook.CardSizeLg),
			},
		},
		kook.CardMessageSection{
			Text: kook.CardMessageElementKMarkdown{
				Content: strings.Join(newsList, "\n"),
			},
		},
	)

	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: kook.CardThemeSecondary,
			Size:  kook.CardSizeLg,
			Modules: append(
				modules,
				&kook.CardMessageDivider{},
				utility.GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
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
	return
}
