package admin

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// ShowAdminHandler 显示管理员
//
//	@param targetID
//	@param quoteID
//	@return err
func ShowAdminHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	admins := make([]database.Administrator, 0)
	database.GetDbConnection().Table("betago.administrators").Find(&admins).Order("level DESC")
	modules := make([]interface{}, 0)
	modules = append(modules,
		kook.CardMessageSection{
			Text: kook.CardMessageParagraph{
				Cols: 3,
				Fields: []interface{}{
					kook.CardMessageElementKMarkdown{
						Content: "**用户名**",
					},
					kook.CardMessageElementKMarkdown{
						Content: "**用户ID**",
					},
					kook.CardMessageElementKMarkdown{
						Content: "**管理等级**",
					},
				},
			},
		})
	for _, admin := range admins {
		info, err := utility.GetUserInfo(strconv.Itoa(int(admin.UserID)), "")
		if err != nil {
			return err
		}
		modules = append(modules,
			kook.CardMessageSection{
				Text: kook.CardMessageParagraph{
					Cols: 3,
					Fields: []interface{}{
						kook.CardMessageElementKMarkdown{
							Content: fmt.Sprintf("`%s`", info.Nickname),
						},
						kook.CardMessageElementKMarkdown{
							Content: strconv.Itoa(int(admin.UserID)),
						},
						kook.CardMessageElementKMarkdown{
							Content: strconv.Itoa(int(admin.Level)),
						},
					},
				},
			},
		)
	}
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: append(
				modules,
				&kook.CardMessageDivider{},
				utility.GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
			),
		},
	}.BuildMessage()
	if err != nil {
		return
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
	return
}

// AddAdminHandler 增加管理员
//
//	@param userID
//	@param userName
//	@param QuoteID
//	@param TargetID
//	@return err PASS
func AddAdminHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	var (
		succUserID []string
		ec         utility.ErrorCollector
	)
	if len(args) != 0 {
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			// 先检验是否存在
			if database.GetDbConnection().
				Table("betago.administrators").
				Where("user_id = ?", utility.MustAtoI(userID)).
				Find(&database.Administrator{}).
				RowsAffected != 0 {
				// 存在则不处理，返回信息
				return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 已经是管理员了`, userID))
			}
			userInfo, err := utility.GetUserInfo(userID, "")
			if err != nil {
				return err
			}
			// 创建管理员
			dbRes := database.GetDbConnection().Table("betago.administrators").
				Create(
					&database.Administrator{
						UserID:   int64(utility.MustAtoI(userID)),
						UserName: userInfo.Nickname,
						Level:    1,
					},
				)
			if dbRes.Error != nil {
				ec.Collect(dbRes.Error)
				continue
			}
			succUserID = append(succUserID, userID)
		}
		if !ec.NoError() {
			return ec.GenErr()
		}
	} else {
		return fmt.Errorf("请输入用户ID")
	}
	var succStr string
	for _, userID := range succUserID {
		succStr += fmt.Sprintf("(met)%s(met) 已被设置为管理员\n", userID)
	}
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: "指令执行成功~~",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: succStr,
					},
				},
				&kook.CardMessageDivider{},
				utility.GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
			},
		},
	}.BuildMessage()
	if err != nil {
		ec.Collect(err)
		return ec.GenErr()
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
	if !ec.NoError() {
		err = ec.GenErr()
	}
	return
}

// RemoveAdminHandler 删除管理员
//
//	@param userID
//	@param targetUserID
//	@param QuoteID
//	@param TargetID
//	@return err PASS
func RemoveAdminHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	var (
		ec         utility.ErrorCollector
		succUserID []string
	)
	if len(args) != 0 {
		// 参数有效
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			// 先检验是否存在
			if !database.CheckIsAdmin(userID) {
				// 不存在则不处理，返回信息
				return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 不是管理员`, userID))
			}
			// 等级校验
			if userLevel, targetLevel := database.GetAdminLevel(authorID), database.GetAdminLevel(userID); userLevel <= targetLevel && userID != authorID {
				// 等级不足，无权限操作
				err = fmt.Errorf("您的等级小于或等于目标用户，无权限操作")
				return
			}
			// 删除管理员
			dbRes := database.GetDbConnection().Table("betago.administrators").
				Where("user_id = ?", utility.MustAtoI(userID)).
				Unscoped().
				Delete(&database.Administrator{})
			if dbRes.Error != nil {
				ec.Collect(dbRes.Error)
				continue
			}
			succUserID = append(succUserID, userID)
		}
	} else {
		return fmt.Errorf("参数不足")
	}

	var succStr string
	for _, userID := range succUserID {
		succStr += fmt.Sprintf("(met)%s(met) 管理员已被移除\n", userID)
	}
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: "指令执行成功~~",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: succStr,
					},
				},
				&kook.CardMessageDivider{},
				utility.GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
			},
		},
	}.BuildMessage()
	if err != nil {
		ec.Collect(err)
		return ec.GenErr()
	}
	betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    quoteID,
		},
	})
	if !ec.NoError() {
		err = ec.GenErr()
	}
	return
}

// DeleteAllMessageHandler 删除频道内所有消息
//
//	@param TargetID
//	@param QuoteID
//	@param authorID
//	@param args
//	@return err
func DeleteAllMessageHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	time.Sleep(1 * time.Second)
	var (
		ec         utility.ErrorCollector
		messageNum int
	)
	if !database.CheckIsAdmin(authorID) {
		// 不存在则不处理，返回信息
		return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 不是管理员`, authorID))
	}
	defer cleaupData()
	if len(args) != 0 {
		messageNum, err = strconv.Atoi(args[0])
		if err != nil {
			return
		}
	}
	// curMsgID := quoteID
	if messageNum > 50 {
		cnt := 0
		for i := 0; i < messageNum/100; i++ {
			ms, err := betagovar.GlobalSession.MessageList(
				targetID,
				kook.MessageListWithPageSize(100),
				kook.MessageListWithFlag(kook.MessageListFlagBefore),
			)
			if err != nil {
				ec.Collect(err)
			}
			for i := 0; i < len(ms) && cnt <= messageNum; i++ {
				err := betagovar.GlobalSession.MessageDelete(ms[i].ID)
				backupData(ms[i].Author.Username, ms[i].Content, ms[i].ID, targetID)
				ec.Collect(err)
				// curMsgID = ms[i].ID
				cnt++
			}
		}
	} else {
		ms, err := betagovar.GlobalSession.MessageList(
			targetID,
			kook.MessageListWithPageSize(100),
			kook.MessageListWithFlag(kook.MessageListFlagBefore),
		)
		ms = ms[:messageNum]
		if err != nil {
			ec.Collect(err)
		}
		ms = ms[:messageNum]
		if len(ms) > 50 || len(ms) > messageNum || messageNum <= 0 {
			err = fmt.Errorf("若全部删除，需要删除的消息数量>50，高危操作，请确认后`指定需要删除的消息数量`完成操作")
			return err
		}
		for i := 0; i < len(ms); i++ {
			err := betagovar.GlobalSession.MessageDelete(ms[i].ID)
			ec.Collect(err)
			msg := ms[i].Content
			backupData(ms[i].Author.Username, msg, ms[i].ID, targetID)
		}
	}
	err = ec.CheckError()
	return
}

// ReconnectHandler 重连
//
//	@param TargetID
//	@param QuoteID
//	@param authorID
//	@param args
//	@return err
func ReconnectHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	if !database.CheckIsAdmin(authorID) {
		// 不存在则不处理，返回信息
		return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 不是管理员`, authorID))
	}
	betagovar.ReconnectChan <- "reconnect"
	return
}

// RestartHandler
//
//	@param ctx
//	@param targetID
//	@param quoteID
//	@param authorID
//	@param args
func RestartHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	if authorID != "" && !database.CheckIsAdmin(authorID) {
		// 不存在则不处理，返回信息
		return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 不是管理员`, authorID))
	}
	if quoteID != "" && authorID != "" && targetID != "" {
		redis.GetRedisClient().Set(context.Background(), "RestartMsgID", quoteID, -1)
		redis.GetRedisClient().Set(context.Background(), "RestartTargetID", targetID, -1)
		redis.GetRedisClient().Set(context.Background(), "RestartAuthorID", authorID, -1)
	}

	os.Exit(0)
	return
}
