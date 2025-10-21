package doubao

import (
	"context"
	"io"
	"iter"
	"os"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var (
	DOUBAO_EMBEDDING_EPID = os.Getenv("DOUBAO_EMBEDDING_EPID")
	DOUBAO_API_KEY        = os.Getenv("DOUBAO_API_KEY")
	ARK_NORMAL_EPID       = os.Getenv("ARK_NORMAL_EPID")
	ARK_REASON_EPID       = os.Getenv("ARK_REASON_EPID")
	ARK_VISION_EPID       = os.Getenv("ARK_VISION_EPID")

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
		log.Zlog.Error("embeddings error", zap.Error(err))
		return
	}
	embedded = resp.Data[0].Embedding
	tokenUsage = resp.Usage
	return
}

func SingleChat(ctx context.Context, sysPrompt, userPrompt string) (res string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	span.SetAttributes(attribute.Key("user_prompt").String(userPrompt))
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := client.CreateChatCompletion(ctx, model.ChatCompletionRequest{
		Model: ARK_NORMAL_EPID,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: "system",
				Content: &model.ChatCompletionMessageContent{
					StringValue: &sysPrompt,
				},
			},
			{
				Role: "user",
				Content: &model.ChatCompletionMessageContent{
					StringValue: &userPrompt,
				},
			},
		},
	})
	if err != nil {
		log.Zlog.Error("chat error", zap.Error(err))
		return "", err
	}

	return *resp.Choices[0].Message.Content.StringValue, nil
}

func SingleChatPrompt(ctx context.Context, prompt string) (res string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("prompt").String(prompt))
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := client.CreateChatCompletion(ctx, model.ChatCompletionRequest{
		Model: ARK_NORMAL_EPID,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: "system",
				Content: &model.ChatCompletionMessageContent{
					StringValue: &prompt,
				},
			},
		},
	})
	if err != nil {
		log.Zlog.Error("chat error", zap.Error(err))
		return "", err
	}

	return *resp.Choices[0].Message.Content.StringValue, nil
}

func SingleChatModel(ctx context.Context, sysPrompt, userPrompt, modelID string) (res string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	span.SetAttributes(attribute.Key("user_prompt").String(userPrompt))
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := client.CreateChatCompletion(ctx, model.ChatCompletionRequest{
		Model: ARK_NORMAL_EPID,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: "system",
				Content: &model.ChatCompletionMessageContent{
					StringValue: &sysPrompt,
				},
			},
			{
				Role: "user",
				Content: &model.ChatCompletionMessageContent{
					StringValue: &userPrompt,
				},
			},
		},
	})
	if err != nil {
		log.Zlog.Error("chat error", zap.Error(err))
		return "", err
	}

	return *resp.Choices[0].Message.Content.StringValue, nil
}

type ModelStreamRespReasoning struct {
	ReasoningContent string
	Content          string
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
		log.Zlog.Error("chat error", zap.Error(err))
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
					reasoningContent.String(),
					content.String(),
				}) {
					return
				}
			}
		}
	}, nil
}

func ResponseStreaming(ctx context.Context, sysPrompt, modelID string, files ...string) (seq iter.Seq[*ModelStreamRespReasoning], err error) {
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
						ImageUrl: f,
						Detail:   responses.ContentItemImageDetail_auto.Enum(),
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
			Tools: []*responses.ResponsesTool{
				{
					Union: &responses.ResponsesTool_ToolWebSearch{
						ToolWebSearch: &responses.ToolWebSearch{},
					},
				},
			},
			// Thinking: &responses.ResponsesThinking{Type: responses.ThinkingType_auto.Enum()},
			Store:   utility.Ptr(false),
			Stream:  utility.Ptr(true),
			Caching: &responses.ResponsesCaching{Type: responses.CacheType_enabled.Enum()},
			// Reasoning: &responses.ResponsesReasoning{
			// ,
			// Stream: utility.Ptr(true),
			// Reasoning: &responses.ResponsesReasoning{
			// 	Effort: responses.ReasoningEffort_medium,
			// },
		}
	} else {
		req = &responses.ResponsesRequest{
			Model: modelID,
			Input: &responses.ResponsesInput{Union: &responses.ResponsesInput_StringValue{StringValue: sysPrompt}},
			Store: utility.Ptr(false),
			Tools: []*responses.ResponsesTool{
				{
					Union: &responses.ResponsesTool_ToolWebSearch{
						ToolWebSearch: &responses.ToolWebSearch{
							Type: responses.ToolType_web_search,
						},
					},
				},
			},
			Caching: &responses.ResponsesCaching{Type: responses.CacheType_enabled.Enum()},
			Stream:  utility.Ptr(true),
			// Reasoning: &responses.ResponsesReasoning{
			// 	Effort: responses.ReasoningEffort_medium,
			// },
			// Thinking: &responses.ResponsesThinking{Type: responses.ThinkingType_auto.Enum()},
		}
	}
	resp, err := client.CreateResponsesStream(ctx, req)
	if err != nil {
		log.Zlog.Error("responses error", zap.Error(err))
		return nil, err
	}

	return func(yield func(s *ModelStreamRespReasoning) bool) {
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
				log.Zlog.Error("responses error", zap.Error(err))
				return
			}

			if part := event.GetReasoningText(); part != nil {
				reasoningContent.WriteString(part.GetDelta())
			}
			if part := event.GetText(); part != nil {
				content.WriteString(part.GetDelta())
			}
			if part := event.GetResponseWebSearchCallCompleted(); part != nil {
				span.SetAttributes(attribute.Key("web_search_completed").String(part.String()))
			}
			if part := event.GetResponseWebSearchCallSearching(); part != nil {
				span.SetAttributes(attribute.Key("web_search_searching").String(part.String()))
			}
			if part := event.GetResponseWebSearchCallInProgress(); part != nil {
				span.SetAttributes(attribute.Key("web_search_in_progress").String(part.String()))
			}

			rc := reasoningContent.String()
			c := content.String()
			span.SetAttributes(attribute.Key("reasoning_content").String(rc))
			span.SetAttributes(attribute.Key("content").String(c))
			if !yield(&ModelStreamRespReasoning{rc, c}) {
				return
			}
		}
	}, nil
}
