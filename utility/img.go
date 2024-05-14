package utility

import (
	"bytes"
	"context"
	"image"
	"io"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/disintegration/imaging"
	"github.com/kevinmatthe/zaplog"
)

func ResizeIMGFromReader(ctx context.Context, r io.ReadCloser) (output io.ReadCloser) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, GetCurrentFunc())
	defer span.End()

	defer r.Close()
	src, _, err := image.Decode(r)
	if err != nil {
		log.ZapLogger.Error("read image error", zaplog.Error(err))
		return
	}
	imaging.Resize(src, 512, 512, imaging.Lanczos)
	buff := bytes.Buffer{}
	err = imaging.Encode(&buff, src, imaging.JPEG, imaging.JPEGQuality(95))
	if err != nil {
		log.ZapLogger.Error("encode image error", zaplog.Error(err))
		return
	}
	return io.NopCloser(&buff)
}
