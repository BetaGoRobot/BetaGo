package doubao

import (
	"context"
	"io"
	"iter"
	"os"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var (
	DOUBAO_EMBEDDING_EPID = os.Getenv("DOUBAO_EMBEDDING_EPID")
	DOUBAO_32K_EPID       = os.Getenv("DOUBAO_32K_EPID")
	DOUBAO_API_KEY        = os.Getenv("DOUBAO_API_KEY")
	DOUBAO_THINK_EPID     = os.Getenv("DOUBAO_THINK_EPID")
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
		log.ZapLogger.Error("embeddings error", zap.Error(err))
		return
	}
	embedded = resp.Data[0].Embedding
	tokenUsage = resp.Usage
	return
}

func SingleChat(ctx context.Context, sysPrompt, userPrompt string) (string, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	span.SetAttributes(attribute.Key("user_prompt").String(userPrompt))
	defer span.End()

	resp, err := client.CreateChatCompletion(ctx, model.ChatCompletionRequest{
		Model: DOUBAO_32K_EPID,
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
		log.ZapLogger.Error("chat error", zap.Error(err))
		return "", err
	}

	return *resp.Choices[0].Message.Content.StringValue, nil
}

func SingleChatPrompt(ctx context.Context, prompt string) (string, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("prompt").String(prompt))
	defer span.End()

	resp, err := client.CreateChatCompletion(ctx, model.ChatCompletionRequest{
		Model: DOUBAO_32K_EPID,
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
		log.ZapLogger.Error("chat error", zap.Error(err))
		return "", err
	}

	return *resp.Choices[0].Message.Content.StringValue, nil
}

func SingleChatModel(ctx context.Context, sysPrompt, userPrompt, modelID string) (string, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	span.SetAttributes(attribute.Key("user_prompt").String(userPrompt))
	defer span.End()

	resp, err := client.CreateChatCompletion(ctx, model.ChatCompletionRequest{
		Model: DOUBAO_32K_EPID,
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
		log.ZapLogger.Error("chat error", zap.Error(err))
		return "", err
	}

	return *resp.Choices[0].Message.Content.StringValue, nil
}

type ModelStreamRespReasoning struct {
	ReasoningContent string
	Content          string
}

func SingleChatStreamingPrompt(ctx context.Context, sysPrompt, modelID string) (iter.Seq[*ModelStreamRespReasoning], error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("sys_prompt").String(sysPrompt))
	defer span.End()
	req := model.CreateChatCompletionRequest{
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

	r, err := client.CreateChatCompletionStream(ctx, req, arkruntime.WithCustomHeader("x-is-encrypted", "true"))
	if err != nil {
		log.ZapLogger.Error("chat error", zap.Error(err))
		return nil, err
	}

	return func(yield func(*ModelStreamRespReasoning) bool) {
		content := &strings.Builder{}
		reasoningContent := &strings.Builder{}
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
