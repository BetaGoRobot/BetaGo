package doubao

import (
	"context"
	"os"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var (
	DOUBAO_EMBEDDING_EPID = os.Getenv("DOUBAO_EMBEDDING_EPID")
	DOUBAO_32K_EPID       = os.Getenv("DOUBAO_32K_EPID")
	DOUBAO_API_KEY        = os.Getenv("DOUBAO_API_KEY")
)

var client = arkruntime.NewClientWithApiKey(DOUBAO_API_KEY)

// EmbeddingText returns the embedding of the input text.
//
//	@param ctx
//	@param input
//	@return embedded
//	@return err
func EmbeddingText(ctx context.Context, input string) (embedded []float32, tokenUsage model.Usage, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("input").String(input))
	defer span.End()

	req := model.EmbeddingRequestStrings{
		Input: []string{input},
		Model: DOUBAO_EMBEDDING_EPID,
	}
	resp, err := client.CreateEmbeddings(ctx, req)
	if err != nil {
		log.ZapLogger.Error("embeddings error", zap.Error(err))
		return
	}
	embedded = resp.Data[0].Embedding
	tokenUsage = resp.Usage
	return
}

func SingleChat(ctx context.Context, sysPrompt, userPrompt string) (string, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
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

func SingleChatModel(ctx context.Context, sysPrompt, userPrompt, modelID string) (string, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
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
