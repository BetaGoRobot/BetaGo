package miniohelper

import (
	"context"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts/ct"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"

	"github.com/kevinmatthe/zaplog"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type minioManager struct {
	context.Context
	span        trace.Span
	bucketName  string
	objName     string
	err         error
	file        io.ReadCloser
	expiration  *time.Time
	contentType ct.ContentType
}

func Client() *minioManager {
	return &minioManager{
		Context: context.Background(),
	}
}

func (m *minioManager) SetContext(ctx context.Context) *minioManager {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, "UploadToMinio")
	m.Context = ctx
	m.span = span
	return m
}

func (m *minioManager) SetBucketName(bucketName string) *minioManager {
	m.span.SetAttributes(attribute.Key("bucketName").String(bucketName))
	m.bucketName = bucketName
	return m
}

func (m *minioManager) SetFileFromURL(url string) *minioManager {
	m.span.SetAttributes(attribute.Key("url").String(url))

	resp, err := requests.Req().SetContext(m.Context).SetDoNotParseResponse(true).Get(url)
	if err != nil {
		log.ZapLogger.Error("Get file failed", zaplog.Error(err))
		m.err = err
		m.span.SetStatus(2, err.Error())
		return m
	}
	m.file = resp.RawResponse.Body
	return m
}

func (m *minioManager) SetFileFromReader(r io.ReadCloser) *minioManager {
	m.file = r
	return m
}

func (m *minioManager) SetFileFromString(s string) *minioManager {
	m.file = io.NopCloser(strings.NewReader(s))
	return m
}

func (m *minioManager) SetFileFromPath(path string) *minioManager {
	m.span.SetAttributes(attribute.Key("path").String(path))

	reader, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	r := io.NopCloser(reader)
	m.file = r
	return m
}

func (m *minioManager) SetObjName(objName string) *minioManager {
	m.span.SetAttributes(attribute.Key("objName").String(objName))

	m.objName = objName
	return m
}

func (m *minioManager) SetContentType(contentType ct.ContentType) *minioManager {
	m.span.SetAttributes(attribute.Key("contentType").String(contentType.String()))

	m.contentType = contentType
	return m
}

func (m *minioManager) SetExpiration(expiration time.Time) *minioManager {
	m.span.SetAttributes(attribute.Key("expiration").String(expiration.Format(time.RFC3339)))
	m.expiration = &expiration
	return m
}

func (m *minioManager) Upload() (u *url.URL, err error) {
	defer m.span.End()
	opts := minio.PutObjectOptions{
		ContentType: m.contentType.String(),
	}
	if m.expiration != nil {
		opts.Expires = *m.expiration
	}
	u, err = m.tryGetFile()
	if err != nil {
		m.span.SetAttributes(attribute.Bool("hit_cache", false))
		log.ZapLogger.Warn("tryGetFile failed", zaplog.Error(err))
		err = m.uploadFile(opts)
		if err != nil {
			log.ZapLogger.Error("uploadFile failed", zaplog.Error(err))
			return
		}
		return m.presignURL()
	}
	m.span.SetAttributes(attribute.Bool("hit_cache", true))
	return
}

func (m *minioManager) tryGetFile() (u *url.URL, err error) {
	shareURL, err := minioTryGetFile(m, m.bucketName, m.objName)
	if err != nil {
		if e, ok := err.(minio.ErrorResponse); ok {
			err = nil
			log.ZapLogger.Warn(e.Error())
		} else {
			log.ZapLogger.Error(err.Error())
			return
		}
	}
	if shareURL != nil {
		u = shareURL
		return
	}
	return
}

func (m *minioManager) uploadFile(opts minio.PutObjectOptions) (err error) {
	err = minioUploadReader(m, m.bucketName, m.file, m.objName, opts)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	return
}

func (m *minioManager) presignURL() (u *url.URL, err error) {
	return presignObj(m, m.bucketName, m.objName)
}
