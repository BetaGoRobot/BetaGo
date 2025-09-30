package chunking

import (
	"context"
	"encoding/json"
	"errors"
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

// Constants for chunking behavior
const (
	// INACTIVITY_TIMEOUT 定义会话非活跃超时时间
	// INACTIVITY_TIMEOUT = 30 * time.Second
	INACTIVITY_TIMEOUT = 5 * time.Minute
	// MAX_CHUNK_SIZE 定义在强制合并前一个块中的最大消息数
	MAX_CHUNK_SIZE = 50
)

const (
	// redisSessionKeyPrefix is the prefix for the session buffer hash
	redisSessionKeyPrefix = "chat:session:"
	// redisActiveSessionsKey is the key for the sorted set indexing active sessions
	redisActiveSessionsKey = "active_sessions"
)

// SessionBuffer 代表在Redis中存储的会话缓冲区
type SessionBuffer struct {
	Messages     []StandardMsg `json:"messages"`
	LastActiveTs int64         `json:"last_active_ts"`
}

type StandardMsg struct {
	GroupID_   string `json:"group_id"`
	MsgID_     string `json:"msg_id"`
	TimeStamp_ int64  `json:"timestamp"`
	BuildLine_ string `json:"line"`
}

func (m *StandardMsg) BuildLine() string {
	return m.BuildLine_
}

func (m *StandardMsg) TimeStamp() int64 {
	return m.TimeStamp_
}

func (m *StandardMsg) GroupID() string {
	return m.GroupID_
}

func (m *StandardMsg) MsgID() string {
	return m.MsgID_
}

func BuildStdMsg(msg GenericMsg) StandardMsg {
	return StandardMsg{
		GroupID_:   msg.GroupID(),
		MsgID_:     msg.MsgID(),
		TimeStamp_: msg.TimeStamp(),
		BuildLine_: msg.BuildLine(),
	}
}

type Chunk struct {
	GroupID  string
	Messages []StandardMsg
}

// Management is the main struct for managing message chunking.
type Management struct {
	redisClient     *redis.Client
	processingQueue chan *Chunk
}

type GenericMsg interface {
	GroupID() string
	MsgID() string
	TimeStamp() int64
	BuildLine() string
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
func NewManagement() *Management {
	return &Management{
		redisClient:     redis_client.GetRedisClient(),
		processingQueue: make(chan *Chunk, 100), // Buffered channel for processing chunks
	}
}

// SubmitMessage 处理新的传入消息。它将消息添加到Redis中相应的会话缓冲区。
// 如果缓冲区达到MAX_CHUNK_SIZE，它会触发立即合并。否则，它会更新会话的
// 最后活动时间戳，以用于基于超时的机制。
func (m *Management) SubmitMessage(ctx context.Context, msg GenericMsg) (err error) {
	groupID := msg.GroupID()
	newTimestamp := msg.TimeStamp()
	if groupID == "" {
		return fmt.Errorf("group ID is empty, skipping message")
	}

	sessionKey := redisSessionKeyPrefix + groupID

	// 1. 从Redis获取当前会话
	val, err := m.redisClient.Get(ctx, sessionKey).Result()
	if err != nil && err != redis.Nil {
		log.Zlog.Error("Failed to get session from Redis", zaplog.String("groupID", groupID), zaplog.Error(err))
		return err
	}

	var buffer SessionBuffer
	// 如果会话存在，则反序列化它。否则，将使用一个新的空缓冲区。
	if err == nil || errors.Is(err, redis.Nil) {
		if err := json.Unmarshal([]byte(val), &buffer); err != nil {
			log.Zlog.Warn("Failed to unmarshal session buffer, starting a new one", zaplog.String("data", val), zaplog.String("groupID", groupID), zaplog.Error(err))
			// 数据可能已损坏，从一个新缓冲区开始
			m.redisClient.Del(ctx, sessionKey)
			buffer = SessionBuffer{}
		}
	}
	// 2. 附加新消息并更新时间戳
	buffer.Messages = append(buffer.Messages, BuildStdMsg(msg))
	buffer.LastActiveTs = newTimestamp

	// 3. 检查缓冲区大小是否超过限制
	if len(buffer.Messages) >= MAX_CHUNK_SIZE {
		log.Zlog.Info("Chunk reached max size, triggering immediate merge",
			zaplog.String("groupID", groupID),
			zaplog.Int("size", len(buffer.Messages)),
		)

		// 将完整的块发送到处理队列
		m.processingQueue <- &Chunk{
			GroupID:  groupID,
			Messages: buffer.Messages,
		}

		// 通过删除会话键并将其从活动集合中移除来清理Redis
		pipe := m.redisClient.Pipeline()
		pipe.Del(ctx, sessionKey)
		pipe.ZRem(ctx, redisActiveSessionsKey, groupID)
		_, err = pipe.Exec(ctx)
		if err != nil {
			// 记录错误但继续，因为块已排队等待处理。
			// 超时机制后续可能会尝试处理一个不存在的键，这是无害的。
			log.Zlog.Error("Failed to execute Redis cleanup pipeline after max size merge", zaplog.String("groupID", groupID), zaplog.Error(err))
		}
		return nil // 触发合并，操作完成
	}

	// 4. 如果未达到大小限制，则在Redis中更新会话
	bufferJSON, err := json.Marshal(buffer)
	if err != nil {
		log.Zlog.Error("Failed to marshal session buffer", zaplog.String("groupID", groupID), zaplog.Error(err))
		return err
	}

	pipe := m.redisClient.Pipeline()
	// 持久化会话数据。后台任务将在超时时清理它。
	pipe.Set(ctx, sessionKey, bufferJSON, 0)
	// 更新有序集合中的分数以反映新的活动时间。
	pipe.ZAdd(ctx, redisActiveSessionsKey, redis.Z{Score: float64(newTimestamp), Member: groupID})
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Zlog.Error("Failed to execute Redis update pipeline", zaplog.String("groupID", groupID), zaplog.Error(err))
		return err
	}

	log.Zlog.Debug("Message submitted and session updated", zaplog.String("groupID", groupID), zaplog.Int("buffer_size", len(buffer.Messages)))
	return nil
}

// OnMerge is called when a chunk is ready to be processed.
// It builds a single string from the chunk and sends it to an LLM.
func (m *Management) OnMerge(ctx context.Context, chunk *Chunk) (err error) {
	if chunk == nil || len(chunk.Messages) == 0 {
		return nil
	}
	// 写入大模型
	chunkLines := make([]string, len(chunk.Messages))
	msgIDs := make([]string, len(chunk.Messages))
	for idx, c := range chunk.Messages {
		msgLine := c.BuildLine()
		chunkLines[idx] = msgLine
		msgIDs[idx] = c.MsgID()
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
	err = tp.Execute(sysPrompt, map[string]string{"CurrentTimeStamp": time.Now().In(utility.UTCPlus8Loc()).Format(time.DateTime)})
	if err != nil {
		return
	}
	chunkStr := strings.Join(chunkLines, "\n")
	res, err := doubao.SingleChat(ctx, sysPrompt.String(), chunkStr)
	if err != nil {
		return
	}
	res = strings.Trim(res, "```")
	res = strings.TrimLeft(res, "json")
	log.SLog.Infof(
		"OnMerge chunk processed by LLM:\n records: %s\nres: %s\n", chunkStr, res,
	)

	chunkLog := &handlertypes.MessageChunkLogV3{
		ID:        uuid.NewV1().String(),
		Timestamp: utility.UTCPlus8Time().Format(time.DateTime),
		GroupID:   chunk.GroupID,
		MsgIDs:    msgIDs,
		MsgList:   chunkLines,
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
func (m *Management) StartBackgroundCleaner(ctx context.Context) {
	log.Zlog.Info("Starting background cleaner for timed-out sessions...")
	// Start the consumer goroutine
	go func() {
		for chunk := range m.processingQueue {
			if chunk == nil {
				continue
			}
			// Each chunk is processed in its own goroutine to avoid blocking the queue consumer
			go func(c *Chunk) {
				log.Zlog.Info("Processing a merged chunk", zaplog.Int("message_count", len(c.Messages)))
				if err := m.OnMerge(ctx, c); err != nil {
					log.Zlog.Error("Error during OnMerge", zaplog.Error(err))
				}
			}(chunk)
		}
	}()

	// Start the ticker for scanning Redis
	go func() {
		ticker := time.NewTicker(INACTIVITY_TIMEOUT / 10)
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
func (m *Management) scanAndProcessTimeouts(ctx context.Context) {
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
		log.Zlog.Info("not session is timed out, will do nothing...")
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

		var buffer SessionBuffer
		if err := json.Unmarshal([]byte(val), &buffer); err != nil {
			log.Zlog.Error("Failed to unmarshal timed-out session buffer", zaplog.String("groupID", groupID), zaplog.Error(err))
			continue
		}

		// Send the collected messages to the processing queue
		if len(buffer.Messages) > 0 {
			m.processingQueue <- &Chunk{
				GroupID:  groupID,
				Messages: buffer.Messages,
			}
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

// BuildEmbeddingInput 函数接收一个更新后的对话文档，然后构建一个高质量的字符串用于生成embedding。
func BuildEmbeddingInput(doc *handlertypes.MessageChunkLogV3) string {
	// 使用 strings.Builder 来高效地拼接字符串
	var builder strings.Builder

	// 1. 核心摘要和主要意图：这是对话最高级别的概括。
	if doc.Summary != "" {
		builder.WriteString("核心摘要: ")
		builder.WriteString(doc.Summary)
		builder.WriteString("\n")
	}
	if doc.Intent != "" {
		builder.WriteString("主要意图: ")
		builder.WriteString(doc.Intent)
		builder.WriteString("\n")
	}

	// 2. 实体 - 对话的具体内容和主体。
	// 将所有关键实体信息组合在一起，形成对“聊了什么”的全面描述。
	if len(doc.Entities.MainTopicsOrActivities) > 0 {
		builder.WriteString("核心议题与活动: ")
		builder.WriteString(strings.Join(doc.Entities.MainTopicsOrActivities, ", "))
		builder.WriteString("\n")
	}
	if len(doc.Entities.KeyConceptsAndNouns) > 0 {
		builder.WriteString("关键概念: ")
		builder.WriteString(strings.Join(doc.Entities.KeyConceptsAndNouns, ", "))
		builder.WriteString("\n")
	}
	if len(doc.Entities.MentionedPeople) > 0 {
		builder.WriteString("提及人物: ")
		builder.WriteString(strings.Join(doc.Entities.MentionedPeople, ", "))
		builder.WriteString("\n")
	}
	if len(doc.Entities.LocationsAndVenues) > 0 {
		builder.WriteString("涉及地点: ")
		builder.WriteString(strings.Join(doc.Entities.LocationsAndVenues, ", "))
		builder.WriteString("\n")
	}
	if len(doc.Entities.MediaAndWorks) > 0 {
		var works []string
		for _, w := range doc.Entities.MediaAndWorks {
			works = append(works, fmt.Sprintf("%s (%s)", w.Title, w.Type))
		}
		builder.WriteString("提及作品: ")
		builder.WriteString(strings.Join(works, ", "))
		builder.WriteString("\n")
	}

	// 3. 结果 - 对话产生了什么结论和计划。
	if len(doc.Outcomes.ConclusionsOrAgreements) > 0 {
		builder.WriteString("共识与结论: ")
		builder.WriteString(strings.Join(doc.Outcomes.ConclusionsOrAgreements, "; "))
		builder.WriteString("\n")
	}
	if len(doc.Outcomes.PlansAndSuggestions) > 0 {
		var plans []string
		for _, p := range doc.Outcomes.PlansAndSuggestions {
			plans = append(plans, p.ActivityOrSuggestion)
		}
		builder.WriteString("计划与提议: ")
		builder.WriteString(strings.Join(plans, "; "))
		builder.WriteString("\n")
	}
	if len(doc.Outcomes.OpenThreadsOrPendingPoints) > 0 {
		builder.WriteString("待定事项: ")
		builder.WriteString(strings.Join(doc.Outcomes.OpenThreadsOrPendingPoints, "; "))
		builder.WriteString("\n")
	}

	// 4. 情感与氛围：为对话添加情感色彩的上下文。
	if doc.SentimentAndTone.Sentiment != "" {
		builder.WriteString("整体情绪: ")
		builder.WriteString(doc.SentimentAndTone.Sentiment)
		if len(doc.SentimentAndTone.Tones) > 0 {
			builder.WriteString(fmt.Sprintf(" (主要语气: %s)", strings.Join(doc.SentimentAndTone.Tones, ", ")))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
