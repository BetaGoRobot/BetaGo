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

type MinioManager struct {
	context.Context
	span        trace.Span
	bucketName  string
	objName     string
	err         error
	file        io.ReadCloser
	expiration  *time.Time
	contentType ct.ContentType
}

func Client() *MinioManager {
	return &MinioManager{
		Context: context.Background(),
	}
}

func (m *MinioManager) SetContext(ctx context.Context) *MinioManager {
	m.Context = ctx
	ctx, span := otel.BetaGoOtelTracer.Start(m, "UploadToMinio")
	m.span = span
	return m
}

func (m *MinioManager) SetBucketName(bucketName string) *MinioManager {
	m.span.SetAttributes(attribute.Key("bucketName").String(bucketName))
	m.bucketName = bucketName
	return m
}

func (m *MinioManager) SetFileFromURL(url string) *MinioManager {
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

func (m *MinioManager) SetFileFromReader(r io.ReadCloser) *MinioManager {
	m.file = r
	return m
}

func (m *MinioManager) SetFileFromString(s string) *MinioManager {
	m.file = io.NopCloser(strings.NewReader(s))
	return m
}

func (m *MinioManager) SetFileFromPath(path string) *MinioManager {
	m.span.SetAttributes(attribute.Key("path").String(path))

	reader, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	r := io.NopCloser(reader)
	m.file = r
	return m
}

func (m *MinioManager) SetObjName(objName string) *MinioManager {
	m.span.SetAttributes(attribute.Key("objName").String(objName))

	m.objName = objName
	return m
}

func (m *MinioManager) SetContentType(contentType ct.ContentType) *MinioManager {
	m.span.SetAttributes(attribute.Key("contentType").String(contentType.String()))

	m.contentType = contentType
	return m
}

func (m *MinioManager) SetExpiration(expiration time.Time) *MinioManager {
	m.span.SetAttributes(attribute.Key("expiration").String(expiration.Format(time.RFC3339)))
	m.expiration = &expiration
	return m
}

func (m *MinioManager) Upload() (u *url.URL, err error) {
	defer m.span.End()
	opts := minio.PutObjectOptions{
		ContentType: m.contentType.String(),
	}
	if m.expiration != nil {
		opts.Expires = *m.expiration
	}
	u, err = m.tryGetFile()
	if err != nil {
		log.ZapLogger.Warn("tryGetFile failed", zaplog.Error(err))
	}
	err = m.uploadFile(opts)
	if err != nil {
		log.ZapLogger.Error("uploadFile failed", zaplog.Error(err))
		return
	}
	return m.presignURL()
}

func (m *MinioManager) tryGetFile() (u *url.URL, err error) {
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

func (m *MinioManager) uploadFile(opts minio.PutObjectOptions) (err error) {
	minioUploadReader(m, m.bucketName, m.file, m.objName, opts)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	return
}

func (m *MinioManager) presignURL() (u *url.URL, err error) {
	return presignObj(m, m.bucketName, m.objName)
}
