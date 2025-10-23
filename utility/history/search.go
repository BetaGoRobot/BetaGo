package history

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	"github.com/BetaGoRobot/BetaGo/utility/logging"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yanyiwu/gojieba"
	"go.opentelemetry.io/otel/attribute"
)

// SearchResult 是我们最终返回给 LLM 的标准结果格式
type SearchResult struct {
	MessageID  string  `json:"message_id"`
	RawMessage string  `json:"raw_message"`
	UserName   string  `json:"user_name"`
	ChatName   string  `json:"chat_name"`
	CreateTime string  `json:"create_time"`
	Mentions   string  `json:"mentions"`
	Score      float64 `json:"score"`
}

// HybridSearchRequest 定义了搜索的输入参数
type HybridSearchRequest struct {
	QueryText []string `json:"query"`
	TopK      int      `json:"top_k"`
	UserID    string   `json:"user_id,omitempty"`
	ChatID    string   `json:"chat_id,omitempty"`
}

type EmbeddingFunc func(ctx context.Context, text string) (vector []float32, tokenUsage model.Usage, err error)

// HybridSearch 执行混合搜索
func HybridSearch(ctx context.Context, req HybridSearchRequest, embeddingFunc EmbeddingFunc) (searchResults []*SearchResult, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	logging.Logger.Info().Ctx(ctx).Str("query_text", strings.Join(req.QueryText, " ")).Msg("开始混合搜索")
	if req.TopK <= 0 {
		req.TopK = 5
	}

	// --- A. 构建过滤 (Filter) 子句 ---
	// 'filter' 用于精确匹配，不影响评分，例如过滤 user_id
	filters := []map[string]interface{}{}
	if req.UserID != "" {
		filters = append(filters, map[string]interface{}{"term": map[string]interface{}{"user_id": req.UserID}})
	}
	if req.ChatID != "" {
		filters = append(filters, map[string]interface{}{"term": map[string]interface{}{"chat_id": req.ChatID}})
	}

	queryTerms := make([]string, 0)
	jieba := gojieba.NewJieba()
	defer jieba.Free()
	for _, query := range req.QueryText {
		queryTerms = append(queryTerms, jieba.Cut(query, true)...)
	}

	queryVecList := make([]map[string]any, 0, len(req.QueryText))
	for _, query := range req.QueryText {
		var queryVec []float32
		queryVec, _, err = embeddingFunc(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("获取向量失败: %w", err)
		}

		queryVecList = append(queryVecList, map[string]any{
			"knn": map[string]interface{}{
				"message": map[string]interface{}{
					"vector": queryVec,
					"k":      req.TopK, // KNN 召回 K 个最近邻
					"boost":  2.0,      // 向量分数权重 (示例值，请调优)
				},
			},
		})
	}

	shouldClauses := []map[string]interface{}{
		{
			// 1. 关键词 (BM25) 查询
			// 我们查询 'message_str' 字段
			"terms": map[string]interface{}{"raw_message_jieba_array": queryTerms},
		},
		{
			"bool": map[string]any{"should": queryVecList},
		},
	}
	query := map[string]interface{}{
		"size": req.TopK,
		"_source": []string{ // 只拉取我们需要的字段
			"message_id",
			"raw_message",
			"user_name",
			"chat_name",
			"create_time",
			"mentions",
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":                 filters,       // 必须满足的过滤条件
				"should":               shouldClauses, // 应该满足的召回条件
				"minimum_should_match": 1,             // 至少匹配 'should' 中的一个
			},
		},
	}
	span.SetAttributes(attribute.Key("query").String(utility.MustMashal(query)))
	res, err := opensearchdal.SearchData(ctx, consts.LarkMsgIndex, query)
	if err != nil {
		return nil, fmt.Errorf("搜索请求失败: %w", err)
	}

	resultList := make([]*SearchResult, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		result := &SearchResult{}
		if err = sonic.Unmarshal(hit.Source, &result); err != nil {
			return
		}
		mentions := make([]*larkmsgutils.Mention, 0)
		if err = sonic.UnmarshalString(result.Mentions, &mentions); err != nil {
			return
		}
		result.RawMessage = larkmsgutils.ReplaceMentionToName(result.RawMessage, mentions)
		resultList = append(resultList, result)
	}

	return resultList, nil
}
