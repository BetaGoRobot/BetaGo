package utility

import (
	"bytes"
	"context"
	"io"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/h2non/bimg"
	"github.com/kevinmatthe/zaplog"
)

func ResizeIMGFromReader(ctx context.Context, r io.ReadCloser) io.ReadCloser {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, GetCurrentFunc())
	defer span.End()

	data, err := io.ReadAll(r)
	defer r.Close()
	if err != nil {
		log.ZapLogger.Error("read image error", zaplog.Error(err))
	}
	return io.NopCloser(bytes.NewReader(resizeIMG(data)))
}

func resizeIMG(input []byte) (output []byte) {
	newImage, err := bimg.NewImage(input).Resize(512, 512)
	if err != nil {
		log.ZapLogger.Error("resize image error", zaplog.Error(err))
		return nil
	}
	return newImage
}
