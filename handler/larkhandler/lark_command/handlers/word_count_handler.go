package handlers

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
	commonutils "github.com/BetaGoRobot/go_utils/common_utils"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	. "github.com/olivere/elastic/v7"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"go.opentelemetry.io/otel/attribute"
)

func WordCloudHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	var (
		days     = 7
		interval = "1d"
		st, et   time.Time
	)
	chatID := *data.Event.Message.ChatId
	top := 10
	argMap, _ := parseArgs(args...)
	if inputInterval, ok := argMap["interval"]; ok {
		interval = inputInterval
	}
	if daysStr, ok := argMap["days"]; ok {
		newDays, err := strconv.Atoi(daysStr)
		if err == nil && newDays > 0 {
			days = newDays
		}
	}
	if chatIDInput, ok := argMap["chat_id"]; ok {
		chatID = chatIDInput
	}
	if topStr, ok := argMap["top"]; ok {
		top, err = strconv.Atoi(topStr)
		if err != nil || top <= 0 {
			top = 15
		}
	}

	st, et = GetBackDays(days)
	// 如果有st，et的配置，用st，et的配置来覆盖
	if stStr, ok := argMap["st"]; ok {
		if etStr, ok := argMap["et"]; ok {
			st, err = time.ParseInLocation(time.DateTime, stStr, utility.UTCPlus8Loc())
			if err != nil {
				return err
			}
			et, err = time.ParseInLocation(time.DateTime, etStr, utility.UTCPlus8Loc())
			if err != nil {
				return err
			}
		}
	}

	helper := &trendInternalHelper{
		days: days, st: st, et: et, msgID: *data.Event.Message.MessageId, chatID: chatID, interval: interval,
	}

	userList, err := genHotRate(ctx, helper, top)
	if err != nil {
		return
	}

	wc, err := genWordCount(ctx, chatID, st, et)
	if err != nil {
		return
	}

	chunks, err := getChunks(ctx, chatID, st, et)
	if err != nil {
		return
	}
	wordCloud := vadvisor.NewWordCloudChartsGraphWithPlayer[string, int]()
	for _, bucket := range wc.Dimension.Dimension.Dimension.Buckets {
		wordCloud.AddData("user_name",
			&vadvisor.ValueUnit[string, int]{
				XField:      bucket.Key,
				YField:      bucket.DocCount,
				SeriesField: strconv.Itoa(bucket.DocCount),
			})
	}
	wordCloud.Build(ctx)

	tpl := templates.GetTemplateV2(templates.WordCountTemplate)
	cardVar := &templates.WordCountCardVars{
		UserList:  userList,
		WordCloud: wordCloud,
		Chunks:    chunks,
		StartTime: st.Format("2006-01-02 15:04"),
		EndTime:   et.Format("2006-01-02 15:04"),
	}
	tpl.WithData(cardVar)
	cardContent := templates.NewCardContentV2(ctx, tpl)

	if metaData != nil && metaData.Refresh {
		err = larkutils.PatchCard(ctx,
			cardContent,
			*data.Event.Message.MessageId)
	} else {
		err = larkutils.ReplyCard(ctx,
			cardContent,
			*data.Event.Message.MessageId, "", true)
	}
	return
}

type WordCountType struct {
	Dimension struct {
		DocCount  int `json:"doc_count"`
		Dimension struct {
			DocCount  int `json:"doc_count"`
			Dimension struct {
				DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
				SumOtherDocCount        int `json:"sum_other_doc_count"`
				Buckets                 []struct {
					Key      string `json:"key"`
					DocCount int    `json:"doc_count"`
				} `json:"buckets"`
			} `json:"dimension"`
		} `json:"dimension"`
	} `json:"dimension"`
}

// Style 定义了每个意图的展示风格，包括短语和颜色。
type Style struct {
	Phrase string
	Color  string
}

// IntentStyleMap 存储了意图与其对应风格的映射。
var IntentStyleMap = map[string]*Style{
	"SOCIAL_COORDINATION": {
		Phrase: "共商议事",
		Color:  "blue",
	},
	"INFORMATION_SHARING": {
		Phrase: "见闻共飨",
		Color:  "neutral",
	},
	"SEEKING_HELP_OR_ADVICE": {
		Phrase: "求教问策",
		Color:  "green",
	},
	"DEBATE_OR_DISCUSSION": {
		Phrase: "明辨事理",
		Color:  "indigo",
	},
	"EMOTIONAL_SHARING_OR_SUPPORT": {
		Phrase: "悲欢与共",
		Color:  "violet",
	},
	"REQUESTING_RECOMMENDATION": {
		Phrase: "求珍问宝",
		Color:  "orange",
	},
	"CASUAL_CHITCHAT": {
		Phrase: "谈天说地",
		Color:  "yellow",
	},
}

// GetIntentPhraseWithFallback 是一个更简洁的转换函数。
// 它接受一个意图 key，如果找到则返回对应的中文短语。
// 如果未找到，它会返回原始的 key 作为备用值，这样调用方总能获得一个可显示的字符串。
func GetIntentPhraseWithFallback(intentKey string) (phrase string, color string) {
	if phrase, ok := IntentStyleMap[intentKey]; ok {
		return phrase.Phrase, phrase.Color
	}
	// 返回原始 key 作为备用
	return intentKey, "neutral"
}

// ToneStyleMap 存储了语气与其对应风格的映射。
var ToneStyleMap = map[string]*Style{
	"HUMOROUS":      {Phrase: "妙语连珠", Color: "lime"},
	"SUPPORTIVE":    {Phrase: "暖心慰藉", Color: "turquoise"},
	"CURIOUS":       {Phrase: "寻根究底", Color: "purple"},
	"EXCITED":       {Phrase: "兴高采烈", Color: "carmine"},
	"URGENT":        {Phrase: "迫在眉睫", Color: "red"},
	"FORMAL":        {Phrase: "严谨庄重", Color: "indigo"},
	"INFORMAL":      {Phrase: "随心而谈", Color: "wathet"},
	"SARCASTIC":     {Phrase: "反语相讥", Color: "yellow"},
	"ARGUMENTATIVE": {Phrase: "唇枪舌剑", Color: "orange"},
	"NOSTALGIC":     {Phrase: "追忆往昔", Color: "violet"},
}

// GetToneStyle 函数用于安全地获取语气的风格。
func GetToneStyle(key string) (phrase string, color string) {
	if phrase, ok := ToneStyleMap[key]; ok {
		return phrase.Phrase, phrase.Color
	}
	// 返回原始 key 作为备用
	return key, "neutral"
}

func getChunks(ctx context.Context, chatID string, st, et time.Time) (chunks []*templates.ChunkData, err error) {
	chunks = make([]*templates.ChunkData, 0)
	queryReq := NewSearchRequest().
		Query(NewBoolQuery().Must(
			NewTermQuery("group_id", chatID),
			NewRangeQuery("timestamp").Gte(st.UTC().Format(time.DateTime)).Lte(et.UTC().Format(time.DateTime)),
		)).
		FetchSourceIncludeExclude(
			[]string{}, []string{"conversation_embedding", "msg_ids", "msg_list"},
		).
		SortBy(
			NewScriptSort(
				NewScript("script").Script("doc['msg_ids'].size()").Lang("painless"), "number",
			).Order(false)).
		// SortBy(
		// 	NewFieldSort("timestamp").Desc(),
		// ).
		Size(5)

	data, err := queryReq.Body()
	if err != nil {
		return
	}
	resp, err := opensearchdal.SearchDataStr(ctx, consts.LarkChunkIndex, data)
	if err != nil {
		return
	}

	return commonutils.TransSlice(resp.Hits.Hits, func(hit opensearchapi.SearchHit) (target *templates.ChunkData) {
		chunkLog := &handlertypes.MessageChunkLogV3{}
		sonic.Unmarshal(hit.Source, chunkLog)
		chunkData := &templates.ChunkData{
			MessageChunkLogV3: chunkLog,
		}
		chunkData.Intent = larkmsgutils.TagText(GetIntentPhraseWithFallback(chunkLog.Intent))
		chunkData.Sentiment = larkmsgutils.TagText(SentimentColor(chunkData.SentimentAndTone.Sentiment))
		chunkData.Tones = strings.Join(commonutils.TransSlice(chunkData.SentimentAndTone.Tones, func(s string) string { return larkmsgutils.TagText(GetToneStyle(s)) }), "")
		chunkData.SentimentAndTone = nil

		chunkData.UserIDs4Lark = commonutils.TransSlice(chunkLog.InteractionAnalysis.Participants, func(p *handlertypes.Participant) *templates.UserUnit { return &templates.UserUnit{ID: p.UserID} })
		chunkData.UserIDs = nil

		chunkData.UnresolvedQuestions = strings.Join(commonutils.TransSlice(chunkLog.InteractionAnalysis.UnresolvedQuestions, func(q string) string { return larkmsgutils.TagText(q, "red") }), "")
		return chunkData
	}), err
}

func genHotRate(ctx context.Context, helper *trendInternalHelper, top int) (userList []*templates.UserListItem, err error) {
	// 统计用户发送的消息数量
	trendMap := make(map[string]*templates.UserListItem)
	msgTrend, err := helper.TrendRate(ctx, consts.LarkMsgIndex, "user_id")
	for _, bucket := range msgTrend.Dimension.Buckets {
		trendMap[bucket.Key] = &templates.UserListItem{Number: -1, User: []*templates.UserUnit{{ID: bucket.Key}}, MsgCnt: bucket.DocCount}
	}
	type UserCountResult struct {
		OpenID string `gorm:"column:open_id"`
		Total  int64  `gorm:"column:total"`
	}
	actionRes := []*UserCountResult{}
	ins := database.GetDbConnection().Model(&database.InteractionStats{}).
		Select("open_id, count(*) as total").
		Where("guild_id = ? and created_at > ? and created_at < ?", helper.chatID, helper.st, helper.et).
		Group("open_id")
	if err = ins.Find(&actionRes).Error; err != nil {
		return
	}

	for _, res := range actionRes {
		if item, ok := trendMap[res.OpenID]; ok {
			item.ActionCnt = int(res.Total)
		} else {
			trendMap[res.OpenID] = &templates.UserListItem{Number: -1, User: []*templates.UserUnit{{ID: res.OpenID}}, ActionCnt: int(res.Total)}
		}
	}

	userList = make([]*templates.UserListItem, 0, len(trendMap))
	for _, item := range trendMap {
		userList = append(userList, item)
	}

	sort.Slice(userList, func(i, j int) bool {
		return userList[i].MsgCnt*10+userList[i].ActionCnt > userList[j].MsgCnt*10+userList[j].ActionCnt
	})
	for idx, item := range userList {
		item.Number = idx + 1
	}
	if len(userList) > top {
		userList = userList[:top]
	}
	return
}

func genWordCount(ctx context.Context, chatID string, st, et time.Time) (wc WordCountType, err error) {
	// 统计用户发送的
	tagsToInclude := []interface{}{
		"n", "nr", "ns", "nt", "nz",
		"v", "vd", "vn",
		"a", "ad", "an",
		"i", "l",
	}
	// 1. 构建最内层的聚合：统计词频 (word_counts)
	// 这是一个 terms aggregation
	wordCountsAgg := osquery.TermsAgg("dimension", "raw_message_jieba_tag.word").Size(100) // 返回前 50 个

	// 2. 构建中间层的聚合：根据词性进行过滤 (filtered_tags)
	// 这是一个 filter aggregation
	filteredTagsAgg := osquery.FilterAgg(
		"dimension",
		osquery.Bool().Must(
			osquery.Terms("raw_message_jieba_tag.tag", tagsToInclude...),
			osquery.CustomAgg("script", map[string]any{
				"script": map[string]any{
					"script": map[string]any{
						"source": "doc['raw_message_jieba_tag.word'].value.length() > 1",
						"lang":   "painless",
					},
				},
			}),
		),
	).Aggs(wordCountsAgg)

	// 3. 构建最外层的聚合：处理嵌套字段 (word_cloud_tags)
	// 这是一个 nested aggregation
	wordCloudTagsAgg := osquery.NestedAgg(
		"dimension",
		"raw_message_jieba_tag",
	).Aggs(filteredTagsAgg)

	// 4. 构建最终的查询对象
	query := osquery.Query(osquery.Bool().
		Must(
			osquery.Term("chat_id", chatID),
			osquery.Range("create_time").
				Gte(st.In(utility.UTCPlus8Loc()).Format(time.DateTime)).
				Lte(et.In(utility.UTCPlus8Loc()).Format(time.DateTime)),
		)).
		Size(0). // 设置 size 为 0，表示不返回任何文档，只关心聚合结果
		Aggs(wordCloudTagsAgg)

	// 统计一下词频
	resp, err := opensearchdal.SearchData(ctx, consts.LarkMsgIndex, query)
	if err != nil {
		return
	}

	wc = WordCountType{}
	err = sonic.Unmarshal(resp.Aggregations, &wc)
	if err != nil {
		return
	}
	return
}

func SentimentColor(sentiment string) (string, string) {
	// `POSITIVE`, `NEGATIVE`, `NEUTRAL`, `MIXED`
	switch sentiment {
	case "POSITIVE":
		return "正向", "green"
	case "NEGATIVE":
		return "负向", "red"
	case "NEUTRAL":
		return "中性", "blue"
	case "MIXED":
		return "混合", "yellow"
	default:
		return sentiment, "lime"
	}
}
