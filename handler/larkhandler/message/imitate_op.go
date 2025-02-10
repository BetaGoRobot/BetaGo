package message

import (
	"context"
	"fmt"
	"strings"

	"github.com/defensestation/osquery"
	"github.com/pkg/errors"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/lark_command/handlers"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var _ Op = &ChatMsgOperator{}

// ChatMsgOperator  RepeatMsg Op
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:51
type ChatMsgOperator struct {
	OpBase
}

// PreRun Repeat
//
//	@receiver r *ImitateMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:35
func (r *ChatMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	// 先判断群聊的功能启用情况
	if !larkutils.CheckFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomRepeat) {
		return errors.Wrap(consts.ErrStageSkip, "ImitateMsgOperator: Not enabled")
	}
	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "ImitateMsgOperator: Is Command")
	}
	return
}

// Run Repeat
//
//	@receiver r *ImitateMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:41
func (r *ChatMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// 开始摇骰子, 默认概率10%
	realRate := utility.MustAtoI(utility.GetEnvWithDefault("IMITATE_DEFAULT_RATE", "10"))
	// 群聊定制化
	config, hitCache := database.FindByCacheFunc(
		database.ImitateRateCustom{
			GuildID: *event.Event.Message.ChatId,
		},
		func(d database.ImitateRateCustom) string {
			return d.GuildID
		},
	)
	span.SetAttributes(attribute.Bool("ImitateRateCustom hitCache", hitCache))

	if len(config) != 0 {
		realRate = config[0].Rate
	}

	if utility.Probability(float64(realRate) / 100) {
		// sendMsg
		textMsgBuilder := larkim.NewTextMsgBuilder()

		res, err := GenerateChat(ctx, event)
		if err != nil {
			return err
		}
		textMsgBuilder.Text(res)
		err = larkutils.CreateMsgText(ctx, res, *event.Event.Message.MessageId, *event.Event.Message.ChatId)
		if err != nil {
			return err
		}
	}
	return
}

func GenerateChat(ctx context.Context, event *larkim.P2MessageReceiveV1) (res string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// 获取最近30条消息
	size := 30
	chatID := *event.Event.Message.ChatId
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("chat_id", chatID),
			),
		).
		SourceIncludes("raw_message", "mentions", "create_time").
		Size(uint64(size*3)).
		Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(
		context.Background(),
		"lark_msg_index",
		query)
	if err != nil {
		panic(err)
	}
	messageList := handlers.FilterMessage(resp.Hits.Hits, size)
	// 生成消息
	sysPrompt := `# 角色
你是一个擅长参与话题讨论的人，能够依据历史的发言信息参与讨论，并且能够根据历史的发言信息生成新的话题。


# 流程
1. 综合历史发言参考，以相似的语气生成回答
2. 基于输入的近期文本，判断是否需要回复
3. 如果需要回复，按照他的语气进行回复

# 限制
1. 回复内容和历史对话的平均长度基本一致
2. 禁止拼接历史对话
3. 回复的文本需要跟最近几次输入存在关联关系
4. 仅回复一条`
	userPrompt := `历史发言输入: %s 请给出模仿的输出:`
	latestMsg := strings.Join(messageList, "\n- ")
	userPrompt = fmt.Sprintf(userPrompt, latestMsg)
	res, err = doubao.SingleChat(ctx, sysPrompt, userPrompt)
	if err != nil {
		return
	}
	span.SetAttributes(attribute.String("res", res))
	res = strings.Trim(res, "\n")
	return
}
