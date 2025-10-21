package history

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	commonutils "github.com/BetaGoRobot/go_utils/common_utils"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"go.opentelemetry.io/otel/attribute"
)

// Helper  to be filled
//
//	@author kevinmatthe
//	@update 2025-04-30 13:10:06
type Helper struct {
	context.Context

	req    *osquery.SearchRequest
	index  string
	query  osquery.Mappable
	source []string
	size   uint64
}

// New to be filled
//
//	@return *HistoryHelper
//	@author kevinmatthe
//	@update 2025-04-30 13:10:02
func New(ctx context.Context) *Helper {
	return &Helper{
		req: osquery.Search(),
	}
}

// Index to be filled
//
//	@receiver h *Helper
//	@param index string
//	@return *Helper
//	@author kevinmatthe
//	@update 2025-04-30 13:11:28
func (h *Helper) Index(index string) *Helper {
	h.index = index
	return h
}

// Size to be filled
//
//	@receiver h *Helper
//	@param size uint64
//	@return *Helper
//	@author kevinmatthe
//	@update 2025-04-30 13:16:42
func (h *Helper) Size(size uint64) *Helper {
	h.req.Size(size)
	return h
}

// Query to be filled
//
//	@receiver h *HistoryHelper
//	@param query osquery.Mappable
//	@return *HistoryHelper
//	@author kevinmatthe
//	@update 2025-04-30 13:09:55
func (h *Helper) Query(query osquery.Mappable) *Helper {
	h.req.Query(query)
	return h
}

// Aggs to be filled
//
//	@receiver h *HistoryHelper
//	@param query osquery.Mappable
//	@return *HistoryHelper
//	@author kevinmatthe
//	@update 2025-04-30 13:09:55
func (h *Helper) Aggs(aggs ...osquery.Aggregation) *Helper {
	h.req.Aggs(aggs...)
	return h
}

// Source to be filled
//
//	@receiver h *HistoryHelper
//	@param source []string
//	@return *HistoryHelper
//	@author kevinmatthe
//	@update 2025-04-30 13:10:00
func (h *Helper) Source(source ...string) *Helper {
	h.req.SourceIncludes(source...)
	return h
}

// Sort to be filled
//
//	@receiver h *Helper
//	@param name string
//	@param order osquery.Order
//	@return *Helper
//	@author kevinmatthe
//	@update 2025-04-30 13:14:55
func (h *Helper) Sort(name string, order osquery.Order) *Helper {
	h.req.Sort(name, order)
	return h
}

func (h *Helper) GetMsg() (messageList OpensearchMsgLogList, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(h.Context, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := h.GetRaw()
	if err != nil {
		return
	}
	messageList = FilterMessage(resp.Hits.Hits)
	return
}

func (h *Helper) GetRaw() (resp *opensearchapi.SearchResp, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(h.Context, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	span.SetAttributes(
		attribute.Key("index").String(h.index),
		attribute.Key("query").String(utility.MustMashal(h.query)),
		attribute.Key("source").StringSlice(h.source),
		attribute.Key("size").Int64(int64(h.size)),
	)

	resp, err = opensearchdal.
		SearchData(
			ctx,
			consts.LarkMsgIndex,
			h.req,
		)
	return
}

func (h *Helper) GetAll() (messageList []*handlertypes.MessageIndex, err error) {
	_, span := otel.LarkRobotOtelTracer.Start(h.Context, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := h.GetRaw()
	return commonutils.TransSlice(resp.Hits.Hits, func(hit opensearchapi.SearchHit) *handlertypes.MessageIndex {
		messageIndex := &handlertypes.MessageIndex{}
		err := sonic.Unmarshal(hit.Source, messageIndex)
		if err != nil {
			return nil
		}
		return messageIndex
	}), nil
}

type (
	TrendSeries []*TrendItem
	TrendItem   struct {
		Time  string `json:"time"`  // x轴
		Value int64  `json:"value"` // y轴
		Key   string `json:"key"`   // 序列
	}
	TrendAggData struct {
		Agg1 struct {
			Buckets []struct {
				KeyAsString string `json:"key_as_string"`
				Key         int64  `json:"key"`
				DocCount    int    `json:"doc_count"`
				Agg2        struct {
					DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
					SumOtherDocCount        int `json:"sum_other_doc_count"`
					Buckets                 []struct {
						Key      string `json:"key"`
						DocCount int    `json:"doc_count"`
					} `json:"buckets"`
				} `json:"agg2"`
			} `json:"buckets"`
		} `json:"agg1"`
	}
	Bucket struct {
		Key      string `json:"key"`
		DocCount int    `json:"doc_count"`
	}

	SingleDimAggregate struct {
		Dimension struct {
			DocCountErrorUpperBound int       `json:"doc_count_error_upper_bound"`
			SumOtherDocCount        int       `json:"sum_other_doc_count"`
			Buckets                 []*Bucket `json:"buckets"`
		} `json:"dimension"`
	}
)

func (h *Helper) GetTrend(interval, termField string) (trendList TrendSeries, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(h.Context, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	span.SetAttributes(
		attribute.Key("index").String(h.index),
		attribute.Key("query").String(utility.MustMashal(h.query)),
		attribute.Key("source").StringSlice(h.source),
		attribute.Key("size").Int64(int64(h.size)),
	)

	aggKey1 := "agg1"
	aggKey2 := "agg2"
	h.req.Aggs(
		osquery.CustomAgg(
			aggKey1,
			map[string]any{
				"date_histogram": map[string]any{
					"field":    "create_time",
					"interval": interval,
				},
				"aggs": map[string]any{
					aggKey2: map[string]any{
						"terms": map[string]any{
							"field": termField,
						},
					},
				},
			},
		),
	)
	resp, err := opensearchdal.
		SearchData(
			ctx,
			consts.LarkMsgIndex,
			h.req,
		)
	if err != nil {
		return
	}

	trendList = make(TrendSeries, 0)

	jsonBytes, err := resp.Aggregations.MarshalJSON()
	if err != nil {
		return
	}
	aggData := &TrendAggData{}
	err = sonic.ConfigStd.Unmarshal(jsonBytes, aggData)
	if err != nil {
		return
	}

	for _, bucket := range aggData.Agg1.Buckets {
		for _, item := range bucket.Agg2.Buckets {
			trendList = append(trendList, &TrendItem{
				Time:  bucket.KeyAsString,
				Value: int64(item.DocCount),
				Key:   item.Key,
			})
		}
	}
	return
}

type OpensearchMsgLogList []*OpensearchMsgLog

func (o OpensearchMsgLogList) ToLines() (msgList []string) {
	msgList = make([]string, 0)
	for _, item := range o {
		msgList = append(msgList, item.ToLine())
	}
	return
}

type OpensearchMsgLog struct {
	CreateTime  string            `json:"create_time"`
	UserID      string            `json:"user_id"`
	UserName    string            `json:"user_name"`
	MsgList     []string          `json:"msg_list"`
	MentionList []*larkim.Mention `json:"mention_list"`
}

func (o *OpensearchMsgLog) ToLine() (msgList string) {
	return fmt.Sprintf("[%s](%s) <%s>: %s", o.CreateTime, o.UserID, o.UserName, strings.Join(o.MsgList, ";"))
}

func FilterMessage(hits []opensearchapi.SearchHit) (msgList []*OpensearchMsgLog) {
	msgList = make([]*OpensearchMsgLog, 0)
	for _, hit := range hits {
		res := &handlertypes.MessageIndex{}
		b, _ := hit.Source.MarshalJSON()
		err := sonic.ConfigStd.Unmarshal(b, res)
		if err != nil {
			continue
		}
		mentions := make([]*larkim.Mention, 0)
		utility.UnmarshallStringPre(res.Mentions, &mentions)

		tmpList := make([]string, 0)
		for msgItem := range larkmsgutils.
			GetContentItemsSeq(
				&larkim.EventMessage{
					Content:     &res.RawMessage,
					MessageType: &res.MessageType,
				},
			) {
			switch msgItem.Tag {
			case "at", "text":
				if len(mentions) > 0 {
					for _, mention := range mentions {
						if mention.Key != nil {
							if strings.HasPrefix(*mention.Name, "不太正经的网易云音乐机器人") {
								*mention.Name = "你"
							}
							msgItem.Content = strings.ReplaceAll(msgItem.Content, *mention.Key, fmt.Sprintf("@%s", *mention.Name))
						}
					}
				}
				fallthrough
			default:
				content := strings.ReplaceAll(msgItem.Content, "\n", "<换行>")
				if strings.TrimSpace(content) != "" {
					tmpList = append(tmpList, content)
				}
			}
		}
		if len(tmpList) == 0 {
			continue
		}
		l := &OpensearchMsgLog{
			CreateTime:  res.CreateTime,
			UserID:      res.UserID,
			UserName:    res.UserName,
			MsgList:     tmpList,
			MentionList: mentions,
		}
		if r := l.ToLine(); r != "" {
			msgList = append(msgList, l)
		}
	}
	slices.Reverse(msgList)
	return msgList
}
