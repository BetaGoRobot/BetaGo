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

func (h *Helper) GetMsg() (messageList []string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(h.Context, reflecting.GetCurrentFunc())
	defer span.End()
	span.SetAttributes(
		attribute.Key("index").String(h.index),
		attribute.Key("query").String(utility.MustMashal(h.query)),
		attribute.Key("source").StringSlice(h.source),
		attribute.Key("size").Int64(int64(h.size)),
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
	messageList = FilterMessage(resp.Hits.Hits)
	return
}

func FilterMessage(hits []opensearchapi.SearchHit) (msgList []string) {
	msgList = make([]string, 0)
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
							if *mention.Name == "不太正经的网易云音乐机器人" {
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
		if r := fmt.Sprintf("[%s] <%s>: %s", res.CreateTime, res.UserName, strings.Join(tmpList, ";")); r != "" {
			msgList = append(msgList, r)
		}
	}
	slices.Reverse(msgList)
	return msgList
}
