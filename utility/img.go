package utility

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/h2non/bimg"
	"github.com/kevinmatthe/zaplog"
)

func ResizeIMGFromReader(ctx context.Context, r io.ReadCloser) (output []byte) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, GetCurrentFunc())
	defer span.End()
	imgBody, err := io.ReadAll(r)
	if err != nil {
		log.ZapLogger.Error("read image error", zaplog.Error(err))
		return
	}
	newImage, err := bimg.NewImage(imgBody).Resize(512, 512)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return newImage
}
