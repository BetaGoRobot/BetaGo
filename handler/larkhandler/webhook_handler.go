package larkhandler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi/neteaselark"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/message"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"go.uber.org/zap"

	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkimg"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func WebHookHandler(ctx context.Context, cardAction *callback.CardActionTriggerEvent) (resp *callback.CardActionTriggerResponse, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, cardAction.Event.Context.OpenMessageID)
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(cardAction)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	metaData := handlerbase.NewBaseMetaDataWithChatIDUID(ctx, cardAction.Event.Context.OpenChatID, cardAction.Event.Operator.OpenID)
	// 记录一下操作记录
	defer func() { go larkutils.RecordCardAction2Opensearch(ctx, cardAction) }()
	if len(cardAction.Event.Action.FormValue) > 0 {
		go HandleSubmit(ctx, cardAction)
	} else if buttonType, ok := cardAction.Event.Action.Value["type"]; ok {
		switch buttonType {
		case "song":
			if musicID, ok := cardAction.Event.Action.Value["id"]; ok {
				go SendMusicCard(ctx, metaData, musicID.(string), cardAction.Event.Context.OpenMessageID, 1)
			}
		case "album":
			if albumID, ok := cardAction.Event.Action.Value["id"]; ok {
				_ = albumID
				go SendAlbumCard(ctx, metaData, albumID.(string), cardAction.Event.Context.OpenMessageID)
			}
		case "lyrics":
			if musicID, ok := cardAction.Event.Action.Value["id"]; ok {
				go HandleFullLyrics(ctx, metaData, musicID.(string), cardAction.Event.Context.OpenMessageID)
			}
		case "withdraw":
			// 撤回消息
			go HandleWithDraw(ctx, cardAction)
		case "refresh":
			if musicID, ok := cardAction.Event.Action.Value["id"]; ok {
				go HandleRefreshMusic(ctx, musicID.(string), cardAction.Event.Context.OpenMessageID)
			}
		case "refresh_obj":
			// 通用的卡片刷新结构，重点是记录触发的command重新触发？
			go HandleRefreshObj(ctx, cardAction)
		}
	}
	// 无返回值示例
	return nil, nil
}

func GetCardMusicByPage(ctx context.Context, musicID string, page int) *templates.TemplateCardContent {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	const (
		maxSingleLineLen = 48
		maxPageSize      = 18
	)
	musicURL, err := neteaseapi.NetEaseGCtx.GetMusicURL(ctx, musicID)
	if err != nil {
		logs.L().Ctx(ctx).Error("Failed to get music URL", zap.Error(err))
		return nil
	}

	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]
	picURL := songDetail.Al.PicURL
	imageKey, ossURL, err := larkimg.UploadPicAllinOne(ctx, picURL, musicID, true)
	if err != nil {
		logs.L().Ctx(ctx).Error("Failed to upload picture", zap.Error(err))
		return nil
	}

	lyrics, lyricsURL := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyrics = larkutils.TrimLyrics(lyrics)

	artistNameList := make([]map[string]string, 0)
	for _, ar := range songDetail.Ar {
		artistNameList = append(artistNameList, map[string]string{"name": ar.Name})
	}

	type resultURL struct {
		Title      string
		LyricsURL  string
		MusicURL   string
		PictureURL string
		Album      string
		Artist     []map[string]string
		Duration   int
	}

	targetURL := &resultURL{
		Title:      songDetail.Name,
		LyricsURL:  lyricsURL,
		MusicURL:   musicURL,
		PictureURL: ossURL,
		Album:      songDetail.Al.Name,
		Artist:     artistNameList,
		Duration:   songDetail.Dt,
	}

	u, err := miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromString(utility.MustMashal(targetURL)).
		SetObjName("info/" + musicID + ".json").
		SetContentType(ct.ContentTypePlainText).
		Overwrite().
		Upload()
	if err != nil {
		logs.L().Ctx(ctx).Error("Failed to upload to minio", zap.Error(err))
		return nil
	}

	playerURL := utility.BuildURL(u.String())
	// eg: page = 1
	quotaRemain := maxPageSize
	lyricList := strings.Split(lyrics, "\n")
	newList := make([]string, 0)
	curPage := 1
	for _, l := range lyricList {
		quotaRemain--
		if len(l) > maxSingleLineLen {
			quotaRemain--
		}
		if quotaRemain <= 0 {
			curPage++
			quotaRemain = maxPageSize
			if curPage > page {
				break
			}
		}
		if curPage == page {
			newList = append(newList, l)
		}
	}

	lyrics = strings.Join(newList, "\n")

	return templates.NewCardContent(
		ctx,
		templates.SingleSongDetailTemplate,
	).
		AddVariable("lyrics", lyrics).
		AddVariable("title", songDetail.Name).
		AddVariable("sub_title", songDetail.Ar[0].Name).
		AddVariable("imgkey", imageKey).
		AddVariable("player_url", playerURL).
		AddVariable("full_lyrics_button", map[string]string{"type": "lyrics", "id": musicID}).
		AddVariable("refresh_id", map[string]string{"type": "refresh", "id": musicID})
}

func SendMusicCard(ctx context.Context, metaData *handlerbase.BaseMetaData, musicID string, msgID string, page int) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	card := GetCardMusicByPage(ctx, musicID, page)
	err := larkutils.ReplyCard(ctx, card, msgID, "_music"+musicID, utility.GetIfInthread(ctx, metaData, env.MusicCardInThread))
	if err != nil {
		return
	}
}

func SendAlbumCard(ctx context.Context, metaData *handlerbase.BaseMetaData, albumID string, msgID string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("albumID").String(albumID))
	defer span.End()

	albumDetails, err := neteaseapi.NetEaseGCtx.GetAlbumDetail(ctx, albumID)
	if err != nil {
		logs.L().Ctx(ctx).Error("Failed to get album detail", zap.Error(err))
		return
	}
	searchRes := neteaseapi.SearchMusic{Result: *albumDetails}

	result, err := neteaseapi.NetEaseGCtx.AsyncGetSearchRes(ctx, searchRes)
	if err != nil {
		return
	}
	cardContent, err := neteaselark.BuildMusicListCard(ctx,
		result,
		neteaselark.MusicItemNoTrans,
		neteaseapi.CommentTypeSong,
	)
	if err != nil {
		return
	}
	err = larkutils.ReplyCard(ctx, cardContent, msgID, "_album", utility.GetIfInthread(ctx, metaData, env.MusicCardInThread))
	if err != nil {
		return
	}
}

func HandleFullLyrics(ctx context.Context, metaData *handlerbase.BaseMetaData, musicID, msgID string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("musicID").String(musicID))
	defer span.End()
	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]

	imgKey, _, err := larkimg.UploadPicAllinOne(ctx, songDetail.Al.PicURL, musicID, true)
	lyric, _ := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyric = larkutils.TrimLyrics(lyric)
	sp := strings.Split(lyric, "\n")
	left := strings.Join(sp[:len(sp)/2], "\n")
	right := strings.Join(sp[len(sp)/2+1:], "\n")

	cardContent := templates.NewCardContent(
		ctx,
		templates.FullLyricsTemplate,
	).
		AddVariable("left_lyrics", left).
		AddVariable("right_lyrics", right).
		AddVariable("title", songDetail.Name).
		AddVariable("sub_title", songDetail.Ar[0].Name).
		AddVariable("imgkey", imgKey)
	err = larkutils.ReplyCard(ctx, cardContent, msgID, "_music", utility.GetIfInthread(ctx, metaData, env.MusicCardInThread))
	if err != nil {
		return
	}
}

func HandleWithDraw(ctx context.Context, cardAction *callback.CardActionTriggerEvent) {
	// 伪撤回
	userID := cardAction.Event.Operator.OpenID
	msgID := cardAction.Event.Context.OpenMessageID
	if consts.WITHDRAW_REPLACE {
		cardContent := cardutil.NewCardBuildHelper().
			SetContent(fmt.Sprintf("这条消息被%s撤回啦！", larkutils.AtUserString(userID))).Build(ctx)
		err := larkutils.PatchCard(ctx, cardContent, msgID)
		if err != nil {
			logs.L().Ctx(ctx).Error("Failed to patch card", zap.Error(err))
		}
	} else {
		// 撤回消息
		resp, err := lark.LarkClient.Im.Message.Delete(ctx, larkim.NewDeleteMessageReqBuilder().MessageId(msgID).Build())
		if err != nil {
			logs.L().Ctx(ctx).Error("Failed to delete message", zap.Error(err))
			return
		}
		if !resp.Success() {
			logs.L().Ctx(ctx).Error("Delete message error", zap.String("error", resp.Error()))
		}
	}
}

func HandleRefreshMusic(ctx context.Context, musicID, msgID string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("musicID").String(musicID))
	defer span.End()

	card := GetCardMusicByPage(ctx, musicID, 1)
	resp, err := lark.LarkClient.Im.V1.Message.Patch(
		ctx, larkim.NewPatchMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewPatchMessageReqBodyBuilder().
					Content(card.String()).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		logs.L().Ctx(ctx).Error("Refresh music card error", zap.Error(err))
		return
	}
	return
}

func HandleRefreshObj(ctx context.Context, cardAction *callback.CardActionTriggerEvent) {
	srcCmd := cardAction.Event.Action.Value["command"].(string)
	msgID := cardAction.Event.Context.OpenMessageID

	data := new(larkim.P2MessageReceiveV1)
	data.Event = new(larkim.P2MessageReceiveV1Data)
	data.Event.Message = new(larkim.EventMessage)
	data.Event.Message.MessageId = utility.Ptr(msgID)
	data.Event.Message.ChatId = new(string)
	*data.Event.Message.ChatId = cardAction.Event.Context.OpenChatID

	err := message.ExecuteFromRawCommand(
		ctx,
		data,
		&handlerbase.BaseMetaData{
			Refresh: true,
		},
		srcCmd,
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("Refresh obj error", zap.Error(err))
	}
}

func HandleSubmit(ctx context.Context, cardAction *callback.CardActionTriggerEvent) {
	// 移除 --st=xxx --et=xxx这样的参数
	srcCmd := cardAction.Event.Action.Value["command"].(string)
	srcCmd = utility.RemoveArgFromStr(srcCmd, "st", "et")
	stStr, _ := cardAction.Event.Action.FormValue["start_time_picker"].(string)
	etStr, _ := cardAction.Event.Action.FormValue["end_time_picker"].(string)
	st, err := time.ParseInLocation("2006-01-02 15:04 -0700", stStr, utility.UTCPlus8Loc())
	if err != nil {
		logs.L().Ctx(ctx).Error("Failed to parse start time", zap.Error(err))
	}
	et, err := time.ParseInLocation("2006-01-02 15:04 -0700", etStr, utility.UTCPlus8Loc())
	if err != nil {
		logs.L().Ctx(ctx).Error("Failed to parse end time", zap.Error(err))
	}

	srcCmd += fmt.Sprintf(" --st=\"%s\" --et=\"%s\"", st.In(utility.UTCPlus8Loc()).Format(time.RFC3339), et.In(utility.UTCPlus8Loc()).Format(time.RFC3339))
	msgID := cardAction.Event.Context.OpenMessageID

	data := new(larkim.P2MessageReceiveV1)
	data.Event = new(larkim.P2MessageReceiveV1Data)
	data.Event.Message = new(larkim.EventMessage)
	data.Event.Message.MessageId = utility.Ptr(msgID)
	data.Event.Message.ChatId = new(string)
	*data.Event.Message.ChatId = cardAction.Event.Context.OpenChatID

	err = message.ExecuteFromRawCommand(
		ctx,
		data,
		&handlerbase.BaseMetaData{
			Refresh: true,
		},
		srcCmd,
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("Refresh obj error", zap.Error(err))
	}
}
