package helper

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// TryPanic 1
//
//	@param ctx
//	@param targetID
//	@param quoteID
//	@param authorID
//	@param args
//	@return err
func TryPanic(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.End()

	panic("try panic")
}

// CommandRouter  1
//
//	@param ctx
//	@param targetID
//	@param quoteID
//	@param authorID
//	@param args
//	@return err
func CommandRouter(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	if utility.CheckIsAdmin(authorID) {
		return AdminCommandHelperHandler(ctx, targetID, quoteID, authorID, args...)
	}
	return UserCommandHelperHandler(ctx, targetID, quoteID, authorID, args...)
}

// AdminCommandHelperHandler 查看帮助
//
//	@param targetID
//	@param quoteID
//	@param authorID
//	@return err
func AdminCommandHelperHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.End()

	if len(args) == 1 {
		commandInfo := utility.CommandInfo{}
		var cardMessageStr string
		if utility.GetDbConnection().Table("betago.command_infos").Where("command_name = ?", "`"+strings.ToUpper(args[0])+"`").Find(&commandInfo).RowsAffected == 0 {
			return fmt.Errorf("没有找到指令: %s", args[0])
		}
		cardMessageStr, err = kook.CardMessage{&kook.CardMessageCard{
			Theme: "info",
			Size:  kook.CardSizeLg,
			Modules: []interface{}{
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeLeft,
					Text: &kook.CardMessageElementKMarkdown{
						Content: commandInfo.CommandName + ": " + commandInfo.CommandDesc,
					},
				},
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeLeft,
					Text: &kook.CardMessageElementKMarkdown{
						Content: "TraceID: " + span.SpanContext().TraceID().String(),
					},
				},
			},
		}}.BuildMessage()
		if err != nil {
			return err
		}
		_, err = betagovar.GlobalSession.MessageCreate(
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
	// title := "嗨，你可以使用的指令如下:"
	// !对无参数指令，使用Button展示
	var commandInfoList []*utility.CommandInfo
	if utility.GetDbConnection().Table("betago.command_infos").Where("command_param_len=0").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}

	modules := []interface{}{
		kook.CardMessageHeader{
			Text: kook.CardMessageElementText{
				Content: "遇到什么问题了吗？看看下面的命令指南吧~" + emoji.SmilingFaceWithHalo.String(),
				Emoji:   true,
			},
		},
		kook.CardMessageHeader{
			Text: kook.CardMessageElementText{
				Content: string(emoji.ComputerMouse) + "无参数指令:",
				Emoji:   true,
			},
		},
		kook.CardMessageActionGroup{},
	}
	count := 0
	for _, commandInfo := range commandInfoList {
		count++
		if count%4 == 0 {
			modules = append(modules, kook.CardMessageActionGroup{})
		}
		modules[count/4+2] = append(modules[count/4+2].(kook.CardMessageActionGroup),
			kook.CardMessageElementButton{
				Theme: kook.CardThemeSuccess,
				Value: strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")),
				Click: string(kook.CardMessageElementButtonClickReturnVal),
				Text:  strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")) + ">" + getShortDesc(commandInfo.CommandDesc),
			},
		)
	}

	// !对有参指令，使用文本展示
	commandInfoList = make([]*utility.CommandInfo, 0)
	if utility.GetDbConnection().Table("betago.command_infos").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}
	modules = append(modules,
		kook.CardMessageHeader{
			Text: kook.CardMessageElementText{
				Content: emoji.Keyboard.String() + "含参数指令:",
				Emoji:   true,
			},
		},
		kook.CardMessageSection{
			Text: kook.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					kook.CardMessageElementKMarkdown{
						Content: "**指令名称**",
					},
					kook.CardMessageElementKMarkdown{
						Content: "**指令功能**",
					},
				},
			},
		},
	)
	for _, commandInfo := range commandInfoList {
		modules = append(modules, kook.CardMessageSection{
			Text: kook.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					kook.CardMessageElementKMarkdown{
						Content: commandInfo.CommandName,
					},
					kook.CardMessageElementKMarkdown{
						Content: commandInfo.CommandDesc,
					},
				},
			},
		})
	}

	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: append(
				modules,
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeLeft,
					Text: &kook.CardMessageElementKMarkdown{
						Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
					},
					Accessory: kook.CardMessageElementButton{
						Theme: kook.CardThemeWarning,
						Value: "https://jaeger.kevinmatt.top/trace/" + span.SpanContext().TraceID().String(),
						Click: "link",
						Text:  "链路追踪",
					},
				},
			),
		},
	}.BuildMessage()
	if err != nil {
		err = fmt.Errorf("building cardMessage error %s", err.Error())
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
			TempTargetID: authorID,
		},
	)
	return
}

func getShortDesc(fullDesc string) (short string) {
	short, _, _ = strings.Cut(strings.Trim(fullDesc, "\n"), "\n")
	return strings.Trim(short, "**")
}

// UserCommandHelperHandler 查看帮助
//
//	@param targetID
//	@param quoteID
//	@param authorID
//	@return err
func UserCommandHelperHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.End()

	// 帮助信息
	if len(args) == 1 {
		commandInfo := utility.CommandInfo{}
		var cardMessageStr string
		if utility.GetDbConnection().Table("betago.command_infos").Where("command_name = ? and command_type = ?", "`"+strings.ToUpper(args[0])+"`", "user").Find(&commandInfo).RowsAffected == 0 {
			return fmt.Errorf("没有找到指令: %s", args[0])
		}
		cardMessageStr, err = kook.CardMessage{&kook.CardMessageCard{
			Theme: "info",
			Size:  kook.CardSizeLg,
			Modules: []interface{}{
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeLeft,
					Text: &kook.CardMessageElementKMarkdown{
						Content: commandInfo.CommandName + ": " + commandInfo.CommandDesc,
					},
				},
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeLeft,
					Text: &kook.CardMessageElementKMarkdown{
						Content: "TraceID: " + span.SpanContext().TraceID().String(),
					},
				},
			},
		}}.BuildMessage()
		if err != nil {
			return err
		}
		_, err = betagovar.GlobalSession.MessageCreate(
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
	// !对无参数指令，使用Button展示
	var commandInfoList []*utility.CommandInfo
	if utility.GetDbConnection().Table("betago.command_infos").Where("command_param_len=0 and command_type='user'").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}

	modules := []interface{}{
		kook.CardMessageHeader{
			Text: kook.CardMessageElementText{
				Content: string(emoji.ComputerMouse) + "无参数指令:",
				Emoji:   true,
			},
		},
		kook.CardMessageActionGroup{},
	}
	count := 0
	for _, commandInfo := range commandInfoList {
		count++
		if count%4 == 0 {
			modules = append(modules, kook.CardMessageActionGroup{})
		}
		modules[count/4+1] = append(modules[count/4+1].(kook.CardMessageActionGroup),
			kook.CardMessageElementButton{
				Theme: kook.CardThemeSuccess,
				Value: strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")),
				Click: string(kook.CardMessageElementButtonClickReturnVal),
				Text:  strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")) + ">" + getShortDesc(commandInfo.CommandDesc),
			},
		)
	}

	// !对有参指令，使用文本展示
	commandInfoList = make([]*utility.CommandInfo, 0)
	if utility.GetDbConnection().Table("betago.command_infos").Where("command_type='user'").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}
	modules = append(modules,
		kook.CardMessageHeader{
			Text: kook.CardMessageElementText{
				Content: emoji.Keyboard.String() + "含参数指令:",
				Emoji:   true,
			},
		},
		kook.CardMessageSection{
			Text: kook.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					kook.CardMessageElementKMarkdown{
						Content: "**指令名称**",
					},
					kook.CardMessageElementKMarkdown{
						Content: "**指令功能**",
					},
				},
			},
		},
	)
	for _, commandInfo := range commandInfoList {
		modules = append(modules, kook.CardMessageSection{
			Text: kook.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					kook.CardMessageElementKMarkdown{
						Content: commandInfo.CommandName,
					},
					kook.CardMessageElementKMarkdown{
						Content: commandInfo.CommandDesc,
					},
				},
			},
		})
	}

	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: append(
				modules,
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeLeft,
					Text: &kook.CardMessageElementKMarkdown{
						Content: "TraceID: " + span.SpanContext().TraceID().String(),
					},
				},
			),
		},
	}.BuildMessage()
	if err != nil {
		err = fmt.Errorf("building cardMessage error %s", err.Error())
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
			TempTargetID: authorID,
		},
	)
	return
}
