package reaction

import (
	"context"

	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// Handler  消息处理器
var Handler = &handlerbase.Processor[larkim.P2MessageReactionCreatedV1, handlerbase.BaseMetaData]{}

type (
	OpBase = handlerbase.OperatorBase[larkim.P2MessageReactionCreatedV1, handlerbase.BaseMetaData]
	Op     = handlerbase.Operator[larkim.P2MessageReactionCreatedV1, handlerbase.BaseMetaData]
)

func larkDeferFunc(ctx context.Context, err error, event *larkim.P2MessageReactionCreatedV1, metaData *handlerbase.BaseMetaData) {
	larkutils.SendRecoveredMsg(ctx, err, *event.Event.MessageId)
}

func init() {
	Handler = Handler.
		OnPanic(larkDeferFunc).
		AddParallelStages(&FollowReactionOperator{}).
		AddParallelStages(&RecordReactionOperator{})
}
