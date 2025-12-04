package tools

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/ark"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/utils"
	"go.uber.org/zap"
)

type fcFunc func(context.Context, *FunctionCallMeta, string) (any, error)

func CallFunction(ctx context.Context, argEvent *responses.FunctionCallArgumentsDoneEvent, meta *FunctionCallMeta, modelID, lastRespID string, f *FunctionCallUnit) (resp *utils.ResponsesStreamReader, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	logs.L().Ctx(ctx).Info("calling function",
		zap.String("function_name", f.FunctionName),
		zap.String("desc", f.Description),
		zap.String("valid params", string(f.Parameters.JSON())),
		zap.String("arguments", argEvent.GetArguments()),
	)
	callRes, err := f.Function(ctx, meta, argEvent.GetArguments())
	if err != nil {
		return
	}
	logs.L().Ctx(ctx).Info("called function",
		zap.String("function_name", f.FunctionName),
		zap.String("desc", f.Description),
		zap.String("valid params", string(f.Parameters.JSON())),
		zap.String("arguments", argEvent.GetArguments()),
	)
	message := &responses.ResponsesInput{
		Union: &responses.ResponsesInput_ListValue{
			ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{
				{
					Union: &responses.InputItem_FunctionToolCallOutput{
						FunctionToolCallOutput: &responses.ItemFunctionToolCallOutput{
							CallId: argEvent.GetItemId(),
							Output: string(utility.MustMashal(callRes)),
							Type:   responses.ItemType_function_call_output,
						},
					},
				},
			}},
		},
	}
	resp, err = ark.Cli().CreateResponsesStream(ctx, &responses.ResponsesRequest{
		Model:              modelID,
		PreviousResponseId: &lastRespID,
		Input:              message,
	})
	if err != nil {
		return
	}
	return
}
