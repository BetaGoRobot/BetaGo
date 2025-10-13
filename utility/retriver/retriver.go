package retriver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	opensearchgo "github.com/opensearch-project/opensearch-go"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

var Cli *RAGSystem

const (
	IndexNamePrefix = "langchaingo_default"
	vectorDimension = 2560
)

func init() {
	var err error
	ctx := context.Background()
	cfg := Config{
		OpenAIAPIKey:         doubao.DOUBAO_API_KEY,
		OpenAIBaseURL:        "https://ark.cn-beijing.volces.com/api/v3/",
		OpenAIModel:          doubao.ARK_NORMAL_EPID,
		OpenAIEmbeddingModel: doubao.DOUBAO_EMBEDDING_EPID,
		OpenAIEmbeddingDims:  vectorDimension, // 确保这个维度和你的模型匹配
		OpenSearchURL:        "https://" + os.Getenv("OPENSEARCH_DOMAIN") + ":9200",
		OpenSearchUsername:   os.Getenv("OPENSEARCH_USERNAME"),
		OpenSearchPassword:   os.Getenv("OPENSEARCH_PASSWORD"),
	}
	Cli, err = NewRAGSystem(ctx, cfg)
	if err != nil {
		log.Fatalf("初始化 RAG 系统失败: %v", err)
	}
	log.Println("RAG 系统初始化成功！")
}

// RAGSystem 结构体封装了 RAG 应用所需的所有核心组件
type RAGSystem struct {
	llm      *openai.LLM
	embedder *embeddings.EmbedderImpl
	store    opensearch.Store
}

// Config 结构体用于传递初始化 RAGSystem 所需的配置
type Config struct {
	OpenAIAPIKey         string
	OpenAIBaseURL        string
	OpenAIModel          string
	OpenAIEmbeddingModel string
	OpenAIEmbeddingDims  int
	OpenSearchURL        string
	OpenSearchUsername   string
	OpenSearchPassword   string
}

// NewRAGSystem 是 RAGSystem 的构造函数，负责初始化所有客户端和组件
// 这是我们的第一个“原子能力”：系统初始化
func NewRAGSystem(ctx context.Context, cfg Config) (*RAGSystem, error) {
	// 1. 创建 LLM 和 Embedding 模型
	llm, err := openai.New(
		openai.WithToken(cfg.OpenAIAPIKey),
		openai.WithBaseURL(cfg.OpenAIBaseURL),
		openai.WithModel(cfg.OpenAIModel),
		openai.WithEmbeddingModel(cfg.OpenAIEmbeddingModel),
		openai.WithEmbeddingDimensions(cfg.OpenAIEmbeddingDims),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 OpenAI 客户端失败: %w", err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("创建 Embedder 失败: %w", err)
	}

	// 2. 初始化 OpenSearch 客户端
	osClient, err := opensearchgo.NewClient(opensearchgo.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 仅用于开发，生产环境请使用正确的证书
		},
		Addresses: []string{cfg.OpenSearchURL},
		Username:  cfg.OpenSearchUsername,
		Password:  cfg.OpenSearchPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 OpenSearch 客户端失败: %w", err)
	}

	// 3. 创建向量存储实例
	store, err := opensearch.New(
		osClient,
		opensearch.WithEmbedder(embedder),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 OpenSearch 向量存储失败: %w", err)
	}

	return &RAGSystem{
		llm:      llm,
		embedder: embedder,
		store:    store,
	}, nil
}

// AddDocuments 是我们的第二个“原子能力”：插入文档
// 它负责创建索引（如果不存在）并将文档添加进去
func (rs *RAGSystem) AddDocuments(ctx context.Context, suffix string, docs []schema.Document) error {
	indexName := IndexNamePrefix + "_" + suffix
	log.Printf("正在为索引 '%s' 准备...", indexName)
	// 确保索引存在且维度正确
	// CreateIndex 是幂等的，如果索引已存在，会返回错误，我们可以检查并忽略特定错误，或者简单地尝试添加
	_, err := rs.store.CreateIndex(ctx, indexName, func(indexMap *map[string]interface{}) {
		// 动态设置向量维度
		(*indexMap)["mappings"].(map[string]interface{})["properties"].(map[string]interface{})["contentVector"].(map[string]interface{})["dimension"] = vectorDimension
	})
	if err != nil {
		// 如果索引已存在，通常会报错，这里可以根据实际错误类型进行更精细的判断
		log.Printf("创建索引 '%s' 时出现问题 (可能已存在): %v", indexName, err)
	}

	log.Printf("正在向索引 '%s' 添加 %d 个文档...", indexName, len(docs))
	_, err = rs.store.AddDocuments(ctx, docs, vectorstores.WithNameSpace(indexName))
	if err != nil {
		return fmt.Errorf("添加文档到 '%s' 失败: %w", indexName, err)
	}

	log.Printf("文档成功添加到 '%s'！", indexName)
	return nil
}

// RecallDocs 是我们的第三个“原子能力”：通过 query 召回文档
// 它根据查询字符串从指定索引中检索最相关的 k 个文档
func (rs *RAGSystem) RecallDocs(ctx context.Context, suffix string, query string, k int) ([]schema.Document, error) {
	indexName := IndexNamePrefix + "_" + suffix
	log.Printf("正在从索引 '%s' 中检索与 '%s' 相关的文档...", indexName, query)
	// 创建一个临时的检索器来执行查询
	retriever := vectorstores.ToRetriever(rs.store, k, vectorstores.WithNameSpace(indexName))

	retrievedDocs, err := retriever.GetRelevantDocuments(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("从检索器获取文档失败: %w", err)
	}

	return retrievedDocs, nil
}

// AnswerQuery 是我们的第四个“原子能力”：回答 query 的问题
// 这是一个完整的 RAG 流程，它先召回文档，然后让 LLM 基于这些文档回答问题
func (rs *RAGSystem) AnswerQuery(ctx context.Context, suffix string, query string, k int, chatID string) (string, []schema.Document, error) {
	indexName := IndexNamePrefix + "_" + suffix
	// 1. 使用召回能力获取上下文文档
	contextDocs, err := rs.RecallDocs(ctx, indexName, query, k)
	if err != nil {
		return "", nil, fmt.Errorf("RAG - 检索失败: %w", err)
	}

	if len(contextDocs) == 0 {
		log.Println("没有检索到相关文档，将直接调用 LLM 回答。")
	}

	// 2. 将检索到的文档内容格式化为上下文字符串
	var contextBuilder strings.Builder
	for _, doc := range contextDocs {
		contextBuilder.WriteString(doc.PageContent + "\n")
	}

	// 3. 创建 Prompt
	prompt := fmt.Sprintf(`请根据以下上下文回答问题。如果上下文没有提供相关信息，请直接回答你所知道的内容。

上下文:
%s

问题: %s`, contextBuilder.String(), query)

	// 4. 调用 LLM 生成最终答案
	log.Printf("正在调用 LLM 基于上下文生成回答...")
	answer, err := rs.llm.Call(ctx, prompt)
	if err != nil {
		return "", contextDocs, fmt.Errorf("RAG - LLM 调用失败: %w", err)
	}

	return answer, contextDocs, nil
}
