package tools

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/ark"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/utils"
	"go.uber.org/zap"
)

type fcFunc func(context.Context, *FunctionCallMeta, string) (any, error)

func CallFunction(ctx context.Context, argEvent *responses.FunctionCallArgumentsDoneEvent, meta *FunctionCallMeta, modelID, lastRespID string, f fcFunc) (resp *utils.ResponsesStreamReader, err error) {
	searchRes, err := f(ctx, meta, argEvent.GetArguments())
	if err != nil {
		return
	}
	logs.L().Ctx(ctx).Info("called fc history_search search_res", zap.String("search_res", string(utility.MustMashal(searchRes))))
	message := &responses.ResponsesInput{
		Union: &responses.ResponsesInput_ListValue{
			ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{
				{
					Union: &responses.InputItem_FunctionToolCallOutput{
						FunctionToolCallOutput: &responses.ItemFunctionToolCallOutput{
							CallId: argEvent.GetItemId(),
							Output: string(utility.MustMashal(searchRes)),
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
