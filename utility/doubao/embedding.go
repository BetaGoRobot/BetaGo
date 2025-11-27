package doubao

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	redisdal "github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/redis/go-redis/v9"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var (
	DOUBAO_EMBEDDING_EPID = os.Getenv("DOUBAO_EMBEDDING_EPID")
	DOUBAO_API_KEY        = os.Getenv("DOUBAO_API_KEY")

	ARK_NORMAL_EPID = os.Getenv("ARK_NORMAL_EPID")
	ARK_REASON_EPID = os.Getenv("ARK_REASON_EPID")
	ARK_VISION_EPID = os.Getenv("ARK_VISION_EPID")
	ARK_CHUNK_EPID  = os.Getenv("ARK_CHUNK_EPID")

	NORMAL_MODEL_BOT_ID = os.Getenv("NORMAL_MODEL_BOT_ID")
	REASON_MODEL_BOT_ID = os.Getenv("REASON_MODEL_BOT_ID")
)

var client = arkruntime.NewClientWithApiKey(DOUBAO_API_KEY)

// EmbeddingText returns the embedding of the input text.
//
//	@param ctx
//	@param input
//	@return embedded
//	@return err
func EmbeddingText(ctx context.Context, input string) (embedded []float32, tokenUsage model.Usage, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("input").String(input))
	defer span.End()
	defer func() { span.RecordError(err) }()

	req := model.EmbeddingRequestStrings{
		Input: []string{input},
		Model: DOUBAO_EMBEDDING_EPID,
	}
	resp, err := client.CreateEmbeddings(
		ctx,
		req,
		arkruntime.WithCustomHeader("x-is-encrypted", "true"),
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("embeddings error", zap.Error(err))
		return
	}
	embedded = resp.Data[0].Embedding
	tokenUsage = resp.Usage
	return
}

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
		resp, err := client.CreateResponses(ctx, req)
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

	resp, err := client.CreateResponses(ctx, secondReq)
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
		modelID = ARK_VISION_EPID
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

	r, err := client.CreateChatCompletionStream(ctx, req, arkruntime.WithCustomHeader("x-is-encrypted", "true"))
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

func ResponseStreaming(ctx context.Context, sysPrompt, modelID string, meta *FunctionCallMeta, files ...string) (seq iter.Seq[*ModelStreamRespReasoning], err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	span.SetAttributes(attribute.Key("model_id").String(modelID))
	span.SetAttributes(attribute.Key("files").String(strings.Join(files, "\n")))

	var req *responses.ResponsesRequest
	if len(files) > 0 {
		span.SetAttributes(attribute.Key("files").String(strings.Join(files, ",")))
		modelID = ARK_VISION_EPID
		// responses.ResponsesInput_ListValue
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
			Tools:       Tools(),
			Stream:      utility.Ptr(true),
		}
	} else {
		req = &responses.ResponsesRequest{
			Model:       modelID,
			Input:       &responses.ResponsesInput{Union: &responses.ResponsesInput_StringValue{StringValue: sysPrompt}},
			Store:       utility.Ptr(true),
			Tools:       Tools(),
			Temperature: utility.Ptr(0.1),
			Text: &responses.ResponsesText{
				Format: &responses.TextFormat{
					Type: responses.TextType_json_object,
				},
			},
			Stream: utility.Ptr(true),
		}
	}
	resp, err := client.CreateResponsesStream(ctx, req)
	if err != nil {
		logs.L().Ctx(ctx).Error("responses error", zap.Error(err))
		return nil, err
	}

	return func(yield func(s *ModelStreamRespReasoning) bool) {
		lastRespID := ""
		subCtx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		defer span.End()
		defer func() { span.RecordError(err) }()
		span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetAttributes(attribute.Key("files").String(strings.Join(files, "\n")))
		// logFields := []zap.Field{
		// 	// zap.String("sys_prompt", sysPrompt),
		// 	zap.String("model_id", modelID),
		// 	zap.Strings("files", files),
		// }
		content := &strings.Builder{}
		reasoningContent := &strings.Builder{}
		defer logs.L().Ctx(subCtx).
			Info(
				"responsing done",
				zap.String("content", content.String()),
				zap.String("reasoning_content", reasoningContent.String()),
			)

		functionName := ""
		for {
			event, err := resp.Recv()
			if err == io.EOF {
				logs.L().Ctx(subCtx).Info("responses done",
					zap.String("content", content.String()),
					zap.String("reasoning_content", reasoningContent.String()),
				)
				return
			}
			if err != nil {
				logs.L().Ctx(subCtx).Error("responses error", zap.Error(err))
				return
			}
			if id := event.GetResponse().GetResponse().GetId(); id != "" {
				lastRespID = id
			}
			switch eventType := event.GetEventType(); eventType {
			case responses.EventType_response_output_item_added.String():
				item := event.GetItem()
				logs.L().Ctx(subCtx).Info("output item added", zap.Any("item", item))
				if call := item.GetItem().GetFunctionToolCall(); call != nil {
					functionName = call.GetName() // functionName写入
					logs.L().Ctx(subCtx).Info("decided to call function", zap.String("function_name", functionName))
				}
			case responses.EventType_response_function_call_arguments_done.String():
				fa := event.GetFunctionCallArgumentsDone()
				logs.L().Ctx(subCtx).Info("function call arguments", zap.String("arguments", fa.GetArguments()))
				fc, ok := CallFunctionMap[functionName]
				if !ok {
					logs.L().Ctx(subCtx).Error("function not found", zap.String("function_name", functionName))
					continue
				}
				resp, err = CallFunction(subCtx, fa, meta, modelID, lastRespID, fc)
				if err != nil {
					return
				}
				continue
			case responses.EventType_response_reasoning_summary_text_delta.String():
				part := event.GetReasoningText()
				reasoningContent.WriteString(part.GetDelta())
			case responses.EventType_response_output_text_delta.String():
				part := event.GetText()
				content.WriteString(part.GetDelta())
			case responses.EventType_response_web_search_call_searching.String():
				logs.L().Ctx(subCtx).Info("web search call searching")
			case responses.EventType_response_web_search_call_in_progress.String():
				logs.L().Ctx(subCtx).Info("web search call in progress")
			case responses.EventType_response_web_search_call_completed.String():
				logs.L().Ctx(subCtx).Info("web search call completed")
			}
			eventType := event.GetEventType()
			if strings.HasSuffix(eventType, ".done") {
				logs.L().Ctx(subCtx).Info("event done",
					zap.String("event", event.GetEventType()), zap.Any("event", event))
			}
			res := &ModelStreamRespReasoning{}
			{
				rc := reasoningContent.String()
				c := content.String()
				res.ReasoningContent = rc
				res.Content = c
			}

			if !yield(res) {
				return
			}
		}
	}, nil
}
