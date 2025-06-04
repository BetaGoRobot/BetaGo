package crontask

import (
	"context"

	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
)

type CronTaskEvent struct{}

// Handler  消息处理器
var Handler = &handlerbase.Processor[CronTaskEvent, handlerbase.BaseMetaData]{}

type (
	OpBase = handlerbase.OperatorBase[CronTaskEvent, handlerbase.BaseMetaData]
	Op     = handlerbase.Operator[CronTaskEvent, handlerbase.BaseMetaData]
)

func larkDeferFunc(ctx context.Context, err error, event *CronTaskEvent, metaData *handlerbase.BaseMetaData) {
	// larkutils.SendRecoveredMsgUserID(ctx, err, metaData.UserID)
}

func init() {
	Handler = Handler.
		OnPanic(larkDeferFunc).
		AddParallelStages(&CronTaskRunReactionOperator{})
	Handler.Run()
}
