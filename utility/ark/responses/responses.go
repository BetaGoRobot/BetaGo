package responses

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/ark"
	"github.com/BetaGoRobot/BetaGo/utility/ark/tools"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	redisdal "github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/redis/go-redis/v9"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func ResponseWithCache(ctx context.Context, sysPrompt, userPrompt, modelID string) (res string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	span.SetAttributes(attribute.Key("user_prompt").String(userPrompt))
	defer span.End()
	defer func() { span.RecordError(err) }()
	key := fmt.Sprintf("ark:response:cache:chunking:%s:%s", modelID, userPrompt)

	respID, err := redisdal.GetRedisClient().Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		logs.L().Ctx(ctx).Error("get cache error", zap.Error(err))
		return
	}
	if respID == "" {
		exp := time.Now().Add(time.Hour).Unix()
		req := &responses.ResponsesRequest{
			Model: modelID,
			Input: &responses.ResponsesInput{
				Union: &responses.ResponsesInput_ListValue{
					ListValue: &responses.InputItemList{
						ListValue: []*responses.InputItem{
							{
								Union: &responses.InputItem_InputMessage{InputMessage: &responses.ItemInputMessage{
									Role: responses.MessageRole_system,
									Content: []*responses.ContentItem{
										{
											Union: &responses.ContentItem_Text{Text: &responses.ContentItemText{Type: responses.ContentItemType_input_text, Text: sysPrompt}},
										},
									},
								}},
							},
						},
					},
				},
			},
			Store: utility.Ptr(true),
			Caching: &responses.ResponsesCaching{
				Type: responses.CacheType_enabled.Enum(),
			},
			ExpireAt: utility.Ptr(exp),
		}
		// 先创建cache
		resp, err := ark.Cli().CreateResponses(ctx, req)
		if err != nil {
			logs.L().Ctx(ctx).Error("responses error", zap.Error(err))
			return "", err
		}
		if err := redisdal.GetRedisClient().Set(ctx, key, resp.Id, 0).Err(); err != nil && err != redis.Nil {
			logs.L().Ctx(ctx).Error("set cache error", zap.Error(err))
			return "", err
		}
		if err := redisdal.GetRedisClient().ExpireAt(ctx, key, time.Unix(exp, 0)).Err(); err != nil && err != redis.Nil {
			logs.L().Ctx(ctx).Error("expire cache error", zap.Error(err))
			return "", err
		}
		respID = resp.Id
	}

	previousResponseID := respID
	secondReq := &responses.ResponsesRequest{
		Model: modelID,
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{
					ListValue: []*responses.InputItem{
						{
							Union: &responses.InputItem_InputMessage{InputMessage: &responses.ItemInputMessage{
								Role: responses.MessageRole_user,
								Content: []*responses.ContentItem{
									{
										Union: &responses.ContentItem_Text{Text: &responses.ContentItemText{Type: responses.ContentItemType_input_text, Text: userPrompt}},
									},
								},
							}},
						},
					},
				},
			},
		},
		PreviousResponseId: &previousResponseID,
	}

	resp, err := ark.Cli().CreateResponses(ctx, secondReq)
	if err != nil {
		logs.L().Ctx(ctx).Error("responses error", zap.Error(err))
		return "", err
	}

	for _, output := range resp.GetOutput() {
		if msg := output.GetOutputMessage(); msg != nil {
			if content := msg.GetContent(); len(content) > 0 {
				return content[0].GetText().GetText(), nil
			}
		}
	}
	return "", errors.New("text is nil")
}

type ContentStruct struct {
	Decision             string `json:"decision"`
	Thought              string `json:"thought"`
	ReferenceFromWeb     string `json:"reference_from_web"`
	ReferenceFromHistory string `json:"reference_from_history"`
	Reply                string `json:"reply"`
}

func (s *ContentStruct) BuildOutput() string {
	output := strings.Builder{}
	if s.Decision != "" {
		output.WriteString(fmt.Sprintf("- 决策: %s\n", s.Decision))
	}
	if s.Thought != "" {
		output.WriteString(fmt.Sprintf("- 思考: %s\n", s.Thought))
	}
	if s.Reply != "" {
		output.WriteString(fmt.Sprintf("- 回复: %s\n", s.Reply))
	}
	if s.ReferenceFromWeb != "" {
		output.WriteString(fmt.Sprintf("- 参考网络: %s\n", s.ReferenceFromWeb))
	}
	if s.ReferenceFromHistory != "" {
		output.WriteString(fmt.Sprintf("- 参考历史: %s\n", s.ReferenceFromHistory))
	}
	return output.String()
}

type ReplyUnit struct {
	ID      string
	Content string
}

type ModelStreamRespReasoning struct {
	ReasoningContent string
	Content          string
	ContentStruct    ContentStruct
	Reply2Show       *ReplyUnit
}

func SingleChatStreamingPrompt(ctx context.Context, sysPrompt, modelID string, files ...string) (seq iter.Seq[*ModelStreamRespReasoning], err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	span.SetAttributes(attribute.Key("model_id").String(modelID))
	span.SetAttributes(attribute.Key("files").String(strings.Join(files, "\n")))

	var req model.CreateChatCompletionRequest
	if len(files) > 0 {
		span.SetAttributes(attribute.Key("files").String(strings.Join(files, ",")))
		modelID = env.ARK_VISION_EPID
		listValue := make([]*model.ChatCompletionMessageContentPart, len(files)+1)
		listValue[0] = &model.ChatCompletionMessageContentPart{
			Type: model.ChatCompletionMessageContentPartTypeText,
			Text: sysPrompt,
		}
		for i, f := range files {
			listValue[i+1] = &model.ChatCompletionMessageContentPart{
				Type: model.ChatCompletionMessageContentPartTypeImageURL,
				ImageURL: &model.ChatMessageImageURL{
					URL:    f,
					Detail: model.ImageURLDetailAuto,
				},
			}
		}
		req = model.CreateChatCompletionRequest{
			Model: modelID,
			Messages: []*model.ChatCompletionMessage{
				{
					Role: "system",
					Content: &model.ChatCompletionMessageContent{
						ListValue: listValue,
					},
				},
			},
			Thinking: &model.Thinking{model.ThinkingTypeAuto},
		}
	} else {
		req = model.CreateChatCompletionRequest{
			Model: modelID,
			Messages: []*model.ChatCompletionMessage{
				{
					Role: "system",
					Content: &model.ChatCompletionMessageContent{
						StringValue: &sysPrompt,
					},
				},
			},
		}
	}

	r, err := ark.Cli().CreateChatCompletionStream(ctx, req, arkruntime.WithCustomHeader("x-is-encrypted", "true"))
	if err != nil {
		logs.L().Ctx(ctx).Error("chat error", zap.Error(err))
		return nil, err
	}

	return func(yield func(*ModelStreamRespReasoning) bool) {
		_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
		defer span.End()
		defer func() { span.RecordError(err) }()
		content := &strings.Builder{}
		reasoningContent := &strings.Builder{}
		defer span.SetAttributes(attribute.
			Key("content").
			String(content.String()))
		defer span.SetAttributes(attribute.
			Key("reasoning_content").
			String(reasoningContent.String()))

		for {
			resp, err := r.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}
			if len(resp.Choices) > 0 {
				c := resp.Choices[0]
				if rc := c.Delta.ReasoningContent; rc != nil {
					reasoningContent.WriteString(*rc)
				}
				if c := c.Delta.Content; c != "" {
					content.WriteString(c)
				}
				if !yield(&ModelStreamRespReasoning{
					ReasoningContent: reasoningContent.String(),
					Content:          content.String(),
				}) {
					return
				}
			}
		}
	}, nil
}

func ResponseStreaming(ctx context.Context, sysPrompt, modelID string, meta *tools.FunctionCallMeta, files ...string) (seq iter.Seq[*ModelStreamRespReasoning], err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	span.SetAttributes(
		attribute.Key("sys_prompt.len").Int(len(sysPrompt)),
		attribute.Key("sys_prompt.preview").String(sysPrompt), // 假设有一个截断辅助函数
		attribute.Key("model_id").String(modelID),
		attribute.Key("files.count").Int(len(files)),
	)

	var req *responses.ResponsesRequest

	logs.L().Ctx(ctx).Info("preparing llm request",
		zap.String("model_id", modelID),
		zap.Int("file_count", len(files)),
		zap.Bool("use_vision", len(files) > 0),
	)
	if len(files) > 0 {
		span.SetAttributes(attribute.Key("files").String(strings.Join(files, ",")))
		modelID = env.ARK_VISION_EPID
		listValue := make([]*model.ChatCompletionMessageContentPart, len(files)+1)
		listValue[0] = &model.ChatCompletionMessageContentPart{
			Type: model.ChatCompletionMessageContentPartTypeText,
			Text: sysPrompt,
		}
		inputItems := make([]*responses.ContentItem, 0)
		for _, f := range files {
			inputItems = append(inputItems, &responses.ContentItem{
				Union: &responses.ContentItem_Image{
					Image: &responses.ContentItemImage{
						Type:     responses.ContentItemType_input_image,
						ImageUrl: utility.Ptr(f),
					},
				},
			})
		}
		inputItems = append(inputItems, &responses.ContentItem{
			Union: &responses.ContentItem_Text{
				Text: &responses.ContentItemText{
					Type: responses.ContentItemType_input_text,
					Text: sysPrompt,
				},
			},
		})
		inputMessage := &responses.ItemInputMessage{
			Role:    responses.MessageRole_user,
			Content: inputItems,
		}
		req = &responses.ResponsesRequest{
			Model: modelID,
			Input: &responses.ResponsesInput{
				Union: &responses.ResponsesInput_ListValue{
					ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
						Union: &responses.InputItem_InputMessage{InputMessage: inputMessage},
					}}},
				},
			},
			Temperature: utility.Ptr(0.1),
			Tools:       tools.M().Tools(),
			Stream:      utility.Ptr(true),
		}
	} else {
		req = &responses.ResponsesRequest{
			Model:       modelID,
			Input:       &responses.ResponsesInput{Union: &responses.ResponsesInput_StringValue{StringValue: sysPrompt}},
			Store:       utility.Ptr(true),
			Tools:       tools.M().Tools(),
			Temperature: utility.Ptr(0.1),
			Text: &responses.ResponsesText{
				Format: &responses.TextFormat{
					Type: responses.TextType_json_object,
				},
			},
			Stream: utility.Ptr(true),
		}
	}

	resp, err := ark.Cli().CreateResponsesStream(ctx, req)
	if err != nil {
		logs.L().Ctx(ctx).Error("failed to create responses stream", zap.Error(err))
		return nil, err
	}

	return func(yield func(s *ModelStreamRespReasoning) bool) {
		lastRespID := ""
		// 开启一个新的Span用于追踪流式接收过程
		subCtx, subSpan := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc()+".StreamIter")
		defer subSpan.End()
		defer func() { subSpan.RecordError(err) }() // 这里的err需要捕获闭包内的错误

		// 4. 定义流式过程中的统计变量
		streamStartTime := time.Now()
		var firstTokenTime *time.Time // 用于计算首字延迟 (TTFT)
		contentBuilder := &strings.Builder{}
		reasoningBuilder := &strings.Builder{}
		eventCount := 0

		// 使用 defer 统一打印流结束时的汇总日志
		defer func() {
			duration := time.Since(streamStartTime)
			logs.L().Ctx(subCtx).Info("stream response finished",
				zap.Duration("duration", duration),
				zap.String("last_resp_id", lastRespID),
				zap.Int("content_len", contentBuilder.Len()),
				zap.Int("reasoning_len", reasoningBuilder.Len()),
				zap.Int("event_count", eventCount),
			)
			// 将最终生成的完整内容写入 Span 属性（可视情况截断）
			subSpan.SetAttributes(
				attribute.Key("resp.content_len").Int(contentBuilder.Len()),
				attribute.Key("resp.duration_ms").Int64(duration.Milliseconds()),
			)
		}()

		functionName := ""

		for {
			event, err := resp.Recv()
			eventCount++

			if err == io.EOF {
				// EOF 不是错误，是正常结束，日志在 defer 中统一处理
				return
			}
			if err != nil {
				// 流式中断错误，属于高优日志
				logs.L().Ctx(subCtx).Error("stream receive error",
					zap.String("last_resp_id", lastRespID),
					zap.Error(err),
				)
				return
			}

			if id := event.GetResponse().GetResponse().GetId(); id != "" {
				lastRespID = id
			}

			// 5. 计算首字延迟 (TTFT)
			// 只要收到任何实质性的 delta (text 或 reasoning)，就算首字
			if firstTokenTime == nil {
				isContentDelta := event.GetEventType() == responses.EventType_response_output_text_delta.String() ||
					event.GetEventType() == responses.EventType_response_reasoning_summary_text_delta.String()

				if isContentDelta {
					now := time.Now()
					firstTokenTime = &now
					ttft := now.Sub(streamStartTime)
					logs.L().Ctx(subCtx).Info("ttft received", zap.Duration("ttft", ttft))
					subSpan.SetAttributes(attribute.Key("resp.ttft_ms").Int64(ttft.Milliseconds()))
				}
			}

			switch eventType := event.GetEventType(); eventType {

			case responses.EventType_response_output_item_added.String():
				item := event.GetItem()
				// 降噪：只在检测到函数调用时打印 Info，普通文本 Item 可以是 Debug
				if call := item.GetItem().GetFunctionToolCall(); call != nil {
					functionName = call.GetName()
					logs.L().Ctx(subCtx).Info("tool call detected",
						zap.String("function_name", functionName),
						zap.String("call_id", item.GetItem().GetFunctionToolCall().GetCallId()), // 假设有CallID
					)
					subSpan.AddEvent("tool_call_detected", trace.WithAttributes(attribute.String("func", functionName)))
				} else {
					logs.L().Ctx(subCtx).Debug("output item added", zap.String("type", "message"))
				}

			case responses.EventType_response_function_call_arguments_done.String():
				fa := event.GetFunctionCallArgumentsDone()
				args := fa.GetArguments()

				logs.L().Ctx(subCtx).Info("ready to execute function",
					zap.String("function_name", functionName),
					zap.String("arguments_preview", args),
				)

				fc, ok := tools.M().Get(functionName)
				if !ok {
					logs.L().Ctx(subCtx).Error("unknown function call", zap.String("function_name", functionName))
					continue
				}

				toolStart := time.Now()
				resp, err = tools.CallFunction(subCtx, fa, meta, modelID, lastRespID, fc)
				toolDuration := time.Since(toolStart)

				if err != nil {
					logs.L().Ctx(subCtx).Error("function execution failed",
						zap.String("function_name", functionName),
						zap.Duration("latency", toolDuration),
						zap.Error(err),
					)
					return
				}

				logs.L().Ctx(subCtx).Info("function execution completed",
					zap.String("function_name", functionName),
					zap.Duration("latency", toolDuration),
				)
				// 重置首字时间，因为 Function Call 后通常会重新开始流式输出
				firstTokenTime = nil
				continue

			case responses.EventType_response_reasoning_summary_text_delta.String():
				part := event.GetReasoningText()
				reasoningBuilder.WriteString(part.GetDelta())
				// 极其高频的日志，务必使用 Debug，否则会刷屏
				logs.L().Ctx(subCtx).Debug("recv reasoning delta", zap.Int("delta_len", len(part.GetDelta())))

			case responses.EventType_response_output_text_delta.String():
				part := event.GetText()
				contentBuilder.WriteString(part.GetDelta())
				logs.L().Ctx(subCtx).Debug("recv content delta", zap.Int("delta_len", len(part.GetDelta())))

			// Web Search 相关的状态流转，保留 Info 级别，这对于用户感知很重要
			case responses.EventType_response_web_search_call_searching.String():
				logs.L().Ctx(subCtx).Info("web search: searching")
			case responses.EventType_response_web_search_call_in_progress.String():
				logs.L().Ctx(subCtx).Info("web search: processing")
			case responses.EventType_response_web_search_call_completed.String():
				logs.L().Ctx(subCtx).Info("web search: completed")
			}

			// 统一处理 .done 事件，避免每个 case 里写一遍
			if strings.HasSuffix(event.GetEventType(), ".done") {
				logs.L().Ctx(subCtx).Debug("event lifecycle done",
					zap.String("event_type", event.GetEventType()),
				)
			}

			// 构造返回给上层的对象
			res := &ModelStreamRespReasoning{
				ReasoningContent: reasoningBuilder.String(),
				Content:          contentBuilder.String(),
			}

			if !yield(res) {
				logs.L().Ctx(subCtx).Warn("stream iterator stopped by caller")
				return
			}
		}
	}, nil
}
