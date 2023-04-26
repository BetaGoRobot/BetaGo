package dailyrate

import (
	"context"
	"fmt"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

type timeCostStru struct {
	UserID   string
	UserName string
	TimeCost time.Duration
}

func GetRateHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	title := "勤恳在线排行榜"
	// quoteID := ""
	matchSelect := database.GetDbConnection().
		Debug().
		Table("betago.channel_log_exts").
		Select(`
			user_id,
			user_name,
			(
				SUM(
					EXTRACT(
						epoch
						FROM
						to_timestamp(left_time, 'YYYY\-MM\-DD\THH24\:MI\.MS') - to_timestamp(joined_time, 'YYYY-MM-DD\THH24\:MI\.MS')
					)
				)* 1000 * 1000 * 1000
			):: bigint as time_cost
		`).Where("is_update = true").
		Where(`EXTRACT(
					epoch
					FROM
					to_timestamp(left_time, 'YYYY\-MM\-DD\THH24\:MI\.MS')
				) > EXTRACT(
					epoch
					FROM
					now() - interval '32 hours'
				)`).
		Group("user_id, user_name").
		Order("time_cost DESC")
	if len(args) != 0 {
		title = "近24小时在线时长"
		// quoteID = msgID
		matchSelect = matchSelect.Where("user_id = ?", args[0])
	}
	// 获取24小时的整体在线情况
	timeCostList := make([]*timeCostStru, 0)
	matchSelect.Limit(3).Find(&timeCostList)
	if len(timeCostList) != 0 {
		modules := make([]interface{}, 0)
		modules = append(modules, betagovar.CardMessageTextModule{
			Type: "header",
			Text: struct {
				Type    string "json:\"type\""
				Content string "json:\"content\""
			}{"plain-text", title + emoji.DesktopComputer.String()},
		})
		modules = append(modules,
			kook.CardMessageSection{
				Text: kook.CardMessageParagraph{
					Cols: 3,
					Fields: []interface{}{
						kook.CardMessageElementKMarkdown{
							Content: "**用户ID**",
						},
						kook.CardMessageElementKMarkdown{
							Content: "**用户名**",
						},
						kook.CardMessageElementKMarkdown{
							Content: "**语音时长**",
						},
					},
				},
			},
		)
		for _, user := range timeCostList {
			modules = append(modules,
				kook.CardMessageSection{
					Text: kook.CardMessageParagraph{
						Cols: 3,
						Fields: []interface{}{
							kook.CardMessageElementKMarkdown{
								Content: user.UserID,
							},
							kook.CardMessageElementKMarkdown{
								Content: fmt.Sprintf("`%s`", user.UserName),
							},
							kook.CardMessageElementKMarkdown{
								Content: user.TimeCost.String(),
							},
						},
					},
				},
			)
		}
		cardMessageStr, err := utility.BuildCardMessage(
			"secondary",
			"lg",
			"",
			quoteID,
			span,
			modules...,
		)
		if err != nil {
			return err
		}
		betagovar.GlobalSession.MessageCreate(
			&kook.MessageCreate{
				MessageCreateBase: kook.MessageCreateBase{
					Type:     kook.MessageTypeCard,
					TargetID: targetID,
					Content:  cardMessageStr,
					Quote:    quoteID,
				},
			},
		)
	}
	return
}
