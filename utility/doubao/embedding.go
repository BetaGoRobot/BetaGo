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
	"github.com/BetaGoRobot/BetaGo/utility/history"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	redisdal "github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
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

type ReplyUnit struct {
	ID      string
	Content string
}

const (
	ToolSearchHistory = "search_history"
)

type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Items       []*Property `json:"items,omitempty"`
}
type Parameters struct {
	Type       string               `json:"type"`
	Properties map[string]*Property `json:"properties"`
	Required   []string             `json:"required,omitempty"`
}

type Arguments struct {
	Keywords  []string `json:"keywords"`
	TopK      int      `json:"top_k"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	UserID    string   `json:"user_id"`
}

func (p *Parameters) JSON() []byte {
	return []byte(utility.MustMashal(p))
}

func Tools() []*responses.ResponsesTool {
	p := &Parameters{
		Type: "object",
		Properties: map[string]*Property{
			"keywords": {
				Type:        "array",
				Description: "需要检索的关键词列表",
				Items: []*Property{
					{
						Type:        "string",
						Description: "关键词",
					},
				},
			},
			"user_id": {
				Type:        "string",
				Description: "用户ID",
			},
			"start_time": {
				Type:        "string",
				Description: "开始时间，格式为YYYY-MM-DD HH:MM:SS",
			},
			"end_time": {
				Type:        "string",
				Description: "结束时间，格式为YYYY-MM-DD HH:MM:SS",
			},
			"top_k": {
				Type:        "number",
				Description: "返回的结果数量",
			},
		},
		Required: []string{"keywords"},
	}
	return []*responses.ResponsesTool{
		{
			Union: &responses.ResponsesTool_ToolWebSearch{
				ToolWebSearch: &responses.ToolWebSearch{
					Type:  responses.ToolType_web_search,
					Limit: utility.Ptr[int64](10),
				},
			},
		},
		{
			Union: &responses.ResponsesTool_ToolFunction{
				ToolFunction: &responses.ToolFunction{
					Name:        ToolSearchHistory,
					Type:        responses.ToolType_function,
					Description: utility.Ptr("根据输入的关键词搜索相关的历史对话记录"),
					Parameters:  &responses.Bytes{Value: p.JSON()},
				},
			},
		},
	}
}

func ResponseStreaming(ctx context.Context, sysPrompt, modelID, chatID string, files ...string) (seq iter.Seq[*ModelStreamRespReasoning], err error) {
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
			Tools: Tools(),
			// Store:  utility.Ptr(true),
			Stream: utility.Ptr(true),
			// Caching: &responses.ResponsesCaching{Type: responses.CacheType_enabled.Enum()},
		}
	} else {
		req = &responses.ResponsesRequest{
			Model: modelID,
			Input: &responses.ResponsesInput{Union: &responses.ResponsesInput_StringValue{StringValue: sysPrompt}},
			Store: utility.Ptr(true),
			Tools: Tools(),
			Text: &responses.ResponsesText{
				Format: &responses.TextFormat{
					Type: responses.TextType_json_object,
				},
			},
			// Caching: &responses.ResponsesCaching{Type: responses.CacheType_enabled.Enum()},
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
		_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		defer span.End()
		defer func() { span.RecordError(err) }()
		span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
		span.SetAttributes(attribute.Key("model_id").String(modelID))
		span.SetAttributes(attribute.Key("files").String(strings.Join(files, "\n")))
		content := &strings.Builder{}
		reasoningContent := &strings.Builder{}
		for {
			event, err := resp.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				logs.L().Ctx(ctx).Error("responses error", zap.Error(err))
				return
			}
			if id := event.GetResponse().GetResponse().GetId(); id != "" {
				lastRespID = id
			}

			if responseEvent := event.GetResponse(); responseEvent != nil {
				fmt.Println(event.GetEventType())
				fmt.Println(utility.MustMashal(responseEvent.GetResponse()))
			}
			if fa := event.GetFunctionCallArguments(); fa != nil && fa.GetType() == responses.EventType_response_function_call_arguments_done {
				logs.L().Ctx(ctx).Info("function call arguments", zap.String("arguments", fa.GetArguments()))
				// 调用检索
				args := &Arguments{}
				err = sonic.UnmarshalString(fa.GetArguments(), &args)
				if err != nil {
					return
				}
				searchRes, err := history.HybridSearch(ctx,
					history.HybridSearchRequest{
						QueryText: args.Keywords,
						TopK:      args.TopK,
						UserID:    args.UserID,
						ChatID:    chatID,
					}, EmbeddingText)
				if err != nil {
					return
				}
				span.SetAttributes(attribute.Key("search_res").String(string(utility.MustMashal(searchRes))))
				message := &responses.ResponsesInput{
					Union: &responses.ResponsesInput_ListValue{
						ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{
							{
								Union: &responses.InputItem_FunctionToolCallOutput{
									FunctionToolCallOutput: &responses.ItemFunctionToolCallOutput{
										CallId: fa.GetItemId(),
										Output: string(utility.MustMashal(searchRes)),
										Type:   responses.ItemType_function_call_output,
									},
								},
							},
						}},
					},
				}
				resp, err = client.CreateResponsesStream(ctx, &responses.ResponsesRequest{
					Model:              modelID,
					PreviousResponseId: &lastRespID,
					Input:              message,
				})
				if err != nil {
					return
				}
				continue
			}

			res := &ModelStreamRespReasoning{}
			if part := event.GetReasoningText(); part != nil {
				reasoningContent.WriteString(part.GetDelta())
				span.SetAttributes(attribute.Key("reasoning_content").String(reasoningContent.String()))
			}
			if part := event.GetText(); part != nil {
				content.WriteString(part.GetDelta())
				span.SetAttributes(attribute.Key("content").String(content.String()))
			}
			if part := event.GetResponseWebSearchCallSearching(); part != nil {
				key := "web_search_searching"
				span.SetAttributes(attribute.Key(key).String(part.String()))
				// res.Reply2Show = &ReplyUnit{key, "🔍 开始搜索"}
			}
			if event.GetEventType() == responses.EventType_response_output_item_done.String() &&
				event.GetItem() != nil && event.GetItem().GetItem().GetFunctionWebSearch() != nil {
				searchKeywords := event.GetItem().Item.GetFunctionWebSearch().GetAction().GetQuery()
				res.Reply2Show = &ReplyUnit{"query", "✅ 完成搜索, 关键词：" + searchKeywords}
				span.SetAttributes(attribute.Key("content").String(searchKeywords))
			}
			if part := event.GetResponseWebSearchCallInProgress(); part != nil {
				span.SetAttributes(attribute.Key("web_search_in_progress").String(part.String()))
			}

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
