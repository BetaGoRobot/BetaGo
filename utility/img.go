package utility

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/h2non/bimg"
	"go.uber.org/zap"
)

func ResizeIMGFromReader(ctx context.Context, r io.ReadCloser) (output []byte) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, GetCurrentFunc())
	defer span.End()
	imgBody, err := io.ReadAll(r)
	if err != nil {
		logs.L().Ctx(ctx).Error("read image error", zap.Error(err))
		return
	}
	newImage, err := bimg.NewImage(imgBody).Resize(512, 512)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return newImage
}
