package handlers

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
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
	top := 15
	argMap, _ := parseArgs(args...)
	if inputInterval, ok := argMap["interval"]; ok {
		interval = inputInterval
	}
	if daysStr, ok := argMap["days"]; ok {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
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
		days:     days,
		st:       st,
		et:       et,
		msgID:    *data.Event.Message.MessageId,
		chatID:   chatID,
		interval: interval,
	}

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
	if err = database.GetDbConnection().Model(&database.InteractionStats{}).
		Select("open_id, count(*) as total").
		Where("guild_id = ? ", chatID).
		Group("open_id").Find(&actionRes).Error; err != nil {
		return
	}

	for _, res := range actionRes {
		if item, ok := trendMap[res.OpenID]; ok {
			item.ActionCnt = int(res.Total)
		} else {
			trendMap[res.OpenID] = &templates.UserListItem{Number: -1, User: []*templates.UserUnit{{ID: res.OpenID}}, ActionCnt: int(res.Total)}
		}
	}

	userList := make([]*templates.UserListItem, 0, len(trendMap))
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

	// 统计用户发送的
	tagsToInclude := []interface{}{
		"n", "nr", "ns", "nt", "nz",
		"v", "vd", "vn",
		"a", "ad", "an",
		"i", "l",
	}
	// 1. 构建最内层的聚合：统计词频 (word_counts)
	// 这是一个 terms aggregation
	wordCountsAgg := osquery.TermsAgg("dimension", "raw_message_jieba_tag.word").
		Size(50) // 返回前 100 个

	// 2. 构建中间层的聚合：根据词性进行过滤 (filtered_tags)
	// 这是一个 filter aggregation
	filteredTagsAgg := osquery.FilterAgg(
		"dimension",
		osquery.Bool().Must(
			osquery.Terms("raw_message_jieba_tag.tag", tagsToInclude...)),
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

	wc := WordCountType{}
	err = sonic.Unmarshal(resp.Aggregations, &wc)
	if err != nil {
		return err
	}

	wordCloud := vadvisor.NewWordCloudChartsGraphWithPlayer[string, int]()
	for idx, bucket := range wc.Dimension.Dimension.Dimension.Buckets {
		wordCloud.AddData("user_name",
			&vadvisor.ValueUnit[string, int]{
				XField:      bucket.Key,
				YField:      bucket.DocCount,
				SeriesField: strconv.Itoa(idx),
			})
	}

	wordCloud.Build(ctx)
	tpl := templates.GetTemplateV2(templates.WordCountTemplate)
	cardVar := &templates.WordCountCardVars{
		UserList:  userList,
		WordCloud: wordCloud,
	}
	tpl.WithData(cardVar)
	cardContent := templates.NewCardContentV2(ctx, tpl)
	err = larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "_replyGet", false)
	if err != nil {
		return err
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
