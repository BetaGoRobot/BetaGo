package copywriting

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm/clause"
)

const (
	ImgAdd               = "image_add"
	ImgAddRespAlreadyAdd = "image_add_resp_already_add"
	ImgAddRespAddSuccess = "image_add_resp_add_success"

	ImgNotStickerOrIMG = "image_not_sticker_or_img"
	ImgNotAnyValidArgs = "image_not_any_valid_args"
	ImgQuoteNoParent   = "image_quote_no_parent"
)

func GetCopyWritings(ctx context.Context, chatID, endPoint string) []string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// custom copy writing
	customRes, hitCache := database.FindByCacheFunc(
		database.CopyWritingCustom{
			GuildID: chatID, Endpoint: endPoint,
		}, func(d database.CopyWritingCustom) string {
			return d.Endpoint + d.GuildID
		},
	)
	span.SetAttributes(attribute.Bool("CopyWritingCustom hitCache", hitCache))
	if len(customRes) != 0 && len(customRes[0].Content) != 0 {
		return customRes[0].Content
	} else {
		database.GetDbConnection().Clauses(clause.OnConflict{DoNothing: true}).Create(&database.CopyWritingCustom{
			GuildID: chatID, Endpoint: endPoint, Content: []string{},
		})
	}

	// default copy writing
	generalRes, hitCache := database.FindByCacheFunc(
		database.CopyWritingGeneral{Endpoint: endPoint}, func(d database.CopyWritingGeneral) string {
			return d.Endpoint
		},
	)

	span.SetAttributes(attribute.Bool("CopyWritingGeneral hitCache", hitCache))
	if len(generalRes) != 0 && len(generalRes[0].Content) != 0 {
		return generalRes[0].Content
	} else {
		database.GetDbConnection().Clauses(clause.OnConflict{DoNothing: true}).Create(&database.CopyWritingGeneral{
			Endpoint: endPoint, Content: []string{},
		})
	}

	return []string{endPoint}
}

func GetSampleCopyWritings(ctx context.Context, chatID, endPoint string) string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	return utility.SampleSlice(GetCopyWritings(ctx, chatID, endPoint))
}
