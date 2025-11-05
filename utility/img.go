package utility

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/h2non/bimg"
)

func ResizeIMGFromReader(ctx context.Context, r io.ReadCloser) (output []byte) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, GetCurrentFunc())
	defer span.End()
	imgBody, err := io.ReadAll(r)
	if err != nil {
		logs.L.Error(ctx, "read image error", "error", err)
		return
	}
	newImage, err := bimg.NewImage(imgBody).Resize(512, 512)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return newImage
}
