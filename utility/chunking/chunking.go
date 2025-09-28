package chunking

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/bytedance/sonic"

	redis_client "github.com/BetaGoRobot/BetaGo/utility/redis" // Renamed to avoid conflict with package name
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
)

// INACTIVITY_TIMEOUT 定义会话非活跃超时时间
const INACTIVITY_TIMEOUT = 3 * time.Minute

const (
	// redisSessionKeyPrefix is the prefix for the session buffer hash
	redisSessionKeyPrefix = "chat:session:"
	// redisActiveSessionsKey is the key for the sorted set indexing active sessions
	redisActiveSessionsKey = "active_sessions"
)

// SessionBuffer 代表在Redis中存储的会话缓冲区
type SessionBuffer[M any] struct {
	Messages     []M   `json:"messages"`
	LastActiveTs int64 `json:"last_active_ts"`
}

// Management is the main struct for managing message chunking.
type Management[M any] struct {
	redisClient      *redis.Client
	processingQueue  chan []M
	getGroupIDFunc   func(M) string
	getTimestampFunc func(M) int64
}

type (
	// ChunkMessage is an alias for the Lark message event type.
	ChunkMessage larkim.P2MessageReceiveV1
	// ChunkMessageLark is a slice of Lark message events.
	ChunkMessageLark []*larkim.P2MessageReceiveV1
)

// NewManagement creates a new Management instance.
// getGroupIDFunc: A function to extract the group/chat ID from a message.
// getTimestampFunc: A function to extract the Unix timestamp from a message.
func NewManagement[M any](getGroupIDFunc func(M) string, getTimestampFunc func(M) int64) *Management[M] {
	return &Management[M]{
		redisClient:      redis_client.GetRedisClient(),
		processingQueue:  make(chan []M, 100), // Buffered channel for processing chunks
		getGroupIDFunc:   getGroupIDFunc,
		getTimestampFunc: getTimestampFunc,
	}
}

// SubmitMessage handles a new incoming message, adding it to the appropriate chunk buffer in Redis
// and updating its last active timestamp.
func (m *Management[M]) SubmitMessage(ctx context.Context, msg M) (err error) {
	groupID := m.getGroupIDFunc(msg)
	newTimestamp := m.getTimestampFunc(msg)
	if groupID == "" {
		return fmt.Errorf("group ID is empty, skipping message")
	}

	sessionKey := redisSessionKeyPrefix + groupID

	// 1. Get current session from Redis
	val, err := m.redisClient.Get(ctx, sessionKey).Result()
	if err != nil && err != redis.Nil {
		log.Zlog.Error("Failed to get session from Redis", zaplog.String("groupID", groupID), zaplog.Error(err))
		return err
	}

	var buffer SessionBuffer[M]
	// If a session exists, unmarshal it. Otherwise, a new empty buffer will be used.
	if err == nil {
		if err := json.Unmarshal([]byte(val), &buffer); err != nil {
			log.Zlog.Warn("Failed to unmarshal session buffer, starting a new one", zaplog.String("groupID", groupID), zaplog.Error(err))
			// Data might be corrupted, start with a fresh buffer
			buffer = SessionBuffer[M]{}
		}
	}

	// 2. Append the new message and ALWAYS update the timestamp.
	// The core logic is here: we simply add the message and refresh the last active time.
	// Timeout detection is now handled exclusively by the background cleaner.
	buffer.Messages = append(buffer.Messages, msg)
	buffer.LastActiveTs = newTimestamp

	// 3. Write back to Redis using a pipeline for atomicity
	bufferJSON, err := json.Marshal(buffer)
	if err != nil {
		log.Zlog.Error("Failed to marshal session buffer", zaplog.String("groupID", groupID), zaplog.Error(err))
		return err
	}

	pipe := m.redisClient.Pipeline()
	// Persist the session data. The background task will clean it up on timeout.
	pipe.Set(ctx, sessionKey, bufferJSON, 0)
	// Update the score in the sorted set to reflect the new activity time.
	// This ensures the background cleaner knows this session is active.
	pipe.ZAdd(ctx, redisActiveSessionsKey, redis.Z{Score: float64(newTimestamp), Member: groupID})
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Zlog.Error("Failed to execute Redis pipeline", zaplog.String("groupID", groupID), zaplog.Error(err))
		return err
	}

	log.Zlog.Debug("Message submitted and session updated", zaplog.String("groupID", groupID), zaplog.Int("buffer_size", len(buffer.Messages)))
	return nil
}

// RetriveChunk is not implemented as it doesn't fit the event-driven, timeout-based merging strategy.
// Chunks are automatically processed when they are considered complete.
func (m *Management[M]) RetriveChunk(cnt int) ([][]*M, error) {
	return nil, fmt.Errorf("RetriveChunk is not applicable in this automated, timeout-based design")
}

// OnMerge is called when a chunk is ready to be processed.
// It builds a single string from the chunk and sends it to an LLM.
func (m *Management[M]) OnMerge(ctx context.Context, chunk []M, buildLine func(M) string) (err error) {
	if len(chunk) == 0 {
		return nil
	}
	// 写入大模型
	chunkLines := make([]string, len(chunk))
	for idx, c := range chunk {
		msgLine := buildLine(c)
		chunkLines[idx] = msgLine
	}

	// Note: It's better to fetch templates once, not on every merge. This is kept as per the original code.
	templateRows, _ := database.FindByCacheFunc(database.PromptTemplateArgs{PromptID: 3}, func(d database.PromptTemplateArgs) string { return strconv.Itoa(d.PromptID) })
	if len(templateRows) == 0 {
		return fmt.Errorf("prompt template with ID 3 not found")
	}

	promptTemplateStr := templateRows[0].TemplateStr
	tp, err := template.New("prompt").Parse(promptTemplateStr)
	if err != nil {
		return
	}
	sysPrompt := &strings.Builder{}
	err = tp.Execute(sysPrompt, map[string]string{"CurrentTimeStamp": time.Now().Format(time.DateTime)})
	if err != nil {
		return
	}
	chunkStr := strings.Join(chunkLines, "\n")
	res, err := doubao.SingleChat(ctx, sysPrompt.String(), chunkStr)
	if err != nil {
		return
	}
	log.SLog.Infof(
		"OnMerge chunk processed by LLM:\n records: %s\nres: %s\n", chunkStr, res,
	)

	chunkLog := &handlertypes.MessageChunkLog{
		Timestamp: utility.UTCPlus8Time().Format(time.DateTime),
	}
	err = sonic.UnmarshalString(res, &chunkLog)
	if err != nil {
		return
	}
	embedding, _, err := doubao.EmbeddingText(ctx, BuildEmbeddingInput(chunkLog))
	if err != nil {
		log.Zlog.Info("embedding error")
		return
	}
	chunkLog.ConversationEmbedding = Normalize(embedding)
	err = opensearchdal.InsertData(
		ctx, consts.LarkChunkIndex, uuid.NewV4().String(),
		chunkLog,
	)
	return
}

// StartBackgroundCleaner starts a goroutine to periodically scan for and process timed-out sessions.
func (m *Management[M]) StartBackgroundCleaner(ctx context.Context, buildLine func(M) string) {
	log.Zlog.Info("Starting background cleaner for timed-out sessions...")
	// Start the consumer goroutine
	go func() {
		for chunk := range m.processingQueue {
			// Each chunk is processed in its own goroutine to avoid blocking the queue consumer
			go func(c []M) {
				log.Zlog.Info("Processing a merged chunk", zaplog.Int("message_count", len(c)))
				if err := m.OnMerge(ctx, c, buildLine); err != nil {
					log.Zlog.Error("Error during OnMerge", zaplog.Error(err))
				}
			}(chunk)
		}
	}()

	// Start the ticker for scanning Redis
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Zlog.Info("Stopping background cleaner.")
				return
			case <-ticker.C:
				m.scanAndProcessTimeouts(ctx)
			}
		}
	}()
}

// scanAndProcessTimeouts is the internal logic for the background cleaner.
func (m *Management[M]) scanAndProcessTimeouts(ctx context.Context) {
	log.Zlog.Debug("Scanning for timed-out sessions...")
	// Calculate the timestamp threshold for timeout. Sessions older than this will be processed.
	timeoutThreshold := time.Now().Add(-INACTIVITY_TIMEOUT).UnixMilli()

	// 1. Find all group IDs that have timed out using ZRangeByScore
	timedOutGroupIDs, err := m.redisClient.ZRangeByScore(ctx, redisActiveSessionsKey, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(timeoutThreshold, 10),
	}).Result()
	if err != nil {
		log.Zlog.Error("Failed to get timed-out sessions from Redis", zaplog.Error(err))
		return
	}

	if len(timedOutGroupIDs) == 0 {
		return // Nothing to do
	}
	log.Zlog.Info("Found timed-out sessions", zaplog.Int("count", len(timedOutGroupIDs)), zaplog.Strings("group_ids", timedOutGroupIDs))

	// 2. Process and clean up each timed-out session
	for _, groupID := range timedOutGroupIDs {
		sessionKey := redisSessionKeyPrefix + groupID
		// Atomically get the session data and delete the key.
		// This prevents a race condition where a new message arrives while we are processing the timeout.
		val, err := m.redisClient.GetDel(ctx, sessionKey).Result()
		if err == redis.Nil {
			// Session was already processed or removed. Clean up the sorted set entry just in case.
			m.redisClient.ZRem(ctx, redisActiveSessionsKey, groupID)
			continue
		}
		if err != nil {
			log.Zlog.Error("Failed to GetDel session from Redis", zaplog.String("groupID", groupID), zaplog.Error(err))
			continue
		}

		// At this point, the session key is deleted from Redis. Now we clean up the sorted set.
		m.redisClient.ZRem(ctx, redisActiveSessionsKey, groupID)

		var buffer SessionBuffer[M]
		if err := json.Unmarshal([]byte(val), &buffer); err != nil {
			log.Zlog.Error("Failed to unmarshal timed-out session buffer", zaplog.String("groupID", groupID), zaplog.Error(err))
			continue
		}

		// Send the collected messages to the processing queue
		if len(buffer.Messages) > 0 {
			m.processingQueue <- buffer.Messages
		}
	}
}

// Normalize 函数接收一个 float32 类型的向量（切片），返回其归一化后的新向量。
// L2 归一化步骤：
// 1. 计算向量所有元素平方和的平方根（即 L2 范数或称“长度”）。
// 2. 向量中的每个元素都除以这个长度。
func Normalize(vec []float32) []float32 {
	// 1. 计算所有元素的平方和
	// 使用 float64 来进行中间计算，可以防止因累加大量数值而导致的精度损失。
	var sumOfSquares float64
	for _, val := range vec {
		sumOfSquares += float64(val) * float64(val)
	}

	// 2. 计算向量的长度（L2 范数）
	magnitude := math.Sqrt(sumOfSquares)

	// 处理零向量的特殊情况：如果向量长度为0，无法进行归一化（会导致除以零）。
	// 在这种情况下，直接返回一个与原向量等长的零向量。
	if magnitude == 0 {
		return make([]float32, len(vec))
	}

	// 3. 创建一个新的切片来存储归一化后的结果
	normalizedVec := make([]float32, len(vec))

	// 4. 将原向量的每个元素除以长度
	for i, val := range vec {
		normalizedVec[i] = val / float32(magnitude)
	}

	return normalizedVec
}

// BuildEmbeddingInput 函数接收一个对话文档，然后构建一个高质量的字符串用于生成embedding。
func BuildEmbeddingInput(doc *handlertypes.MessageChunkLog) string {
	// 使用 strings.Builder 来高效地拼接字符串，性能远优于简单的 '+' 拼接。
	var builder strings.Builder

	// 1. 添加核心摘要
	// 我们在每个部分前添加一个简单的标签（如“摘要：”），可以帮助模型更好地理解不同部分的上下文。
	if doc.Summary != "" {
		builder.WriteString("核心摘要: ")
		builder.WriteString(doc.Summary)
		builder.WriteString("\n") // 使用换行符分隔不同部分
	}

	// 2. 添加涉及的项目和议题
	if len(doc.Entities.ProjectsAndTopics) > 0 {
		builder.WriteString("涉及项目: ")
		builder.WriteString(strings.Join(doc.Entities.ProjectsAndTopics, ", "))
		builder.WriteString("\n")
	}

	// 3. 添加技术关键词
	if len(doc.Entities.TechnicalKeywords) > 0 {
		builder.WriteString("技术关键词: ")
		builder.WriteString(strings.Join(doc.Entities.TechnicalKeywords, ", "))
		builder.WriteString("\n")
	}

	// 4. 添加关键决策
	if len(doc.Outcomes.DecisionsMade) > 0 {
		builder.WriteString("关键决策: ")
		builder.WriteString(strings.Join(doc.Outcomes.DecisionsMade, "; "))
		builder.WriteString("\n")
	}

	return builder.String()
}
