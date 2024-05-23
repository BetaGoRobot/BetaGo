package miniohelper

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/utility"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"

	"github.com/kevinmatthe/zaplog"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type minioUploadStage func(m *MinioManager)

// MinioManager minio上传管理上下文
type MinioManager struct {
	context.Context
	span           trace.Span
	bucketName     string
	objName        string
	err            error
	file           io.ReadCloser
	expiration     *time.Time
	contentType    ct.ContentType
	inputTransFunc minioUploadStage
	needAKA        bool
}

// Client 返回一个新的minioManager Client
func Client() *MinioManager {
	return &MinioManager{
		Context: context.Background(),
		needAKA: true,
	}
}

// SetNeedAKA 设置上下文
//
//	@receiver m *MinioManager
//	@param ctx context.Context
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:13
func (m *MinioManager) SetNeedAKA(needAKA bool) *MinioManager {
	m.span.SetAttributes(attribute.Bool("needAKA", needAKA))
	m.needAKA = needAKA
	return m
}

// SetContext  设置上下文
//
//	@receiver m *MinioManager
//	@param ctx context.Context
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:13
func (m *MinioManager) SetContext(ctx context.Context) *MinioManager {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, "UploadToMinio")
	m.Context = ctx
	m.span = span
	return m
}

// SetBucketName  设置bucketName
//
//	@receiver m *MinioManager
//	@param bucketName string
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:18
func (m *MinioManager) SetBucketName(bucketName string) *MinioManager {
	m.span.SetAttributes(attribute.Key("bucketName").String(bucketName))
	m.bucketName = bucketName
	return m
}

// SetFileFromURL  从url设置文件
//
//	@receiver m *MinioManager
//	@param url string
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:24
func (m *MinioManager) SetFileFromURL(url string) *MinioManager {
	m.inputTransFunc = func(m *MinioManager) {
		m.span.SetAttributes(attribute.Key("url").String(url))
		resp, err := requests.Req().SetContext(m.Context).SetDoNotParseResponse(true).Get(url)
		if err != nil {
			log.ZapLogger.Error("Get file failed", zaplog.Error(err))
			m.err = err
			m.span.SetStatus(2, err.Error())
			return
		}
		m.file = resp.RawResponse.Body
	}
	return m
}

// SetFileFromReader  从reader设置文件
//
//	@receiver m *MinioManager
//	@param r io.ReadCloser
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:30
func (m *MinioManager) SetFileFromReader(r io.ReadCloser) *MinioManager {
	m.file = r
	return m
}

// SetFileFromString  从字符串设置文件
//
//	@receiver m *MinioManager
//	@param s string
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:35
func (m *MinioManager) SetFileFromString(s string) *MinioManager {
	m.file = io.NopCloser(strings.NewReader(s))
	return m
}

// SetFileFromPath  从路径设置文件
//
//	@receiver m *MinioManager
//	@param path string
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:39
func (m *MinioManager) SetFileFromPath(path string) *MinioManager {
	m.inputTransFunc = func(m *MinioManager) {
		m.span.SetAttributes(attribute.Key("path").String(path))
		reader, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		r := io.NopCloser(reader)
		m.file = r
	}

	return m
}

// SetObjName 设置objName
//
//	@receiver m *MinioManager
//	@param objName string
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:44
func (m *MinioManager) SetObjName(objName string) *MinioManager {
	m.span.SetAttributes(attribute.Key("objName").String(objName))

	m.objName = objName
	return m
}

// SetContentType  设置contentType
//
//	@receiver m *MinioManager
//	@param contentType ct.ContentType
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:54
func (m *MinioManager) SetContentType(contentType ct.ContentType) *MinioManager {
	m.span.SetAttributes(attribute.Key("contentType").String(contentType.String()))

	m.contentType = contentType
	return m
}

// SetExpiration  设置过期时间
//
//	@receiver m *MinioManager
//	@param expiration time.Time
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:59
func (m *MinioManager) SetExpiration(expiration time.Time) *MinioManager {
	m.span.SetAttributes(attribute.Key("expiration").String(expiration.Format(time.RFC3339)))
	m.expiration = &expiration
	return m
}

func (m *MinioManager) addTracePresigned(u *url.URL) {
	if u != nil {
		if url := u.String(); url != "" {
			m.span.SetAttributes(attribute.String("presigned_url", url))
		} else {
			log.ZapLogger.Error("presigned url is empty")
		}
	} else {
		log.ZapLogger.Error("presigned url is nil")
	}
}

func (m *MinioManager) addTraceCached(hit bool) {
	m.span.SetAttributes(attribute.Bool("hit_cache", hit))
}

// Upload  上传文件
//
//	@receiver m *MinioManager
//	@return u *url.URL
//	@return err error
//	@author heyuhengmatt
//	@update 2024-05-13 01:55:04
func (m *MinioManager) Upload() (u *url.URL, err error) {
	u = new(url.URL)
	defer m.span.End()
	if m.file != nil {
		defer m.file.Close()
	}
	opts := minio.PutObjectOptions{
		ContentType: m.contentType.String(),
	}
	if m.expiration != nil {
		opts.Expires = *m.expiration
	}
	u, err = m.tryGetFile()
	if err != nil {
		m.addTraceCached(false)
		if m.inputTransFunc != nil {
			m.inputTransFunc(m)
		}
		log.ZapLogger.Warn("tryGetFile failed", zaplog.Error(err))
		err = m.uploadFile(opts)
		if err != nil {
			log.ZapLogger.Error("uploadFile failed", zaplog.Error(err))
			return
		}
		return m.presignURL()
	}
	m.addTraceCached(true)
	return
}

// 此函数会修改入参，不返回err外的值
func (m *MinioManager) tryGetFile() (shareURL *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(m, utility.GetCurrentFunc())
	defer span.End()

	if MinioCheckFileExists(ctx, m.bucketName, m.objName) {
		return m.presignURL()
	}
	return nil, errors.New("file not exists")
}

func (m *MinioManager) uploadFile(opts minio.PutObjectOptions) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(m, utility.GetCurrentFunc())
	defer span.End()

	err = minioUploadReader(ctx, m.bucketName, m.file, m.objName, opts)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	return
}

func (m *MinioManager) presignURL() (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(m, utility.GetCurrentFunc())
	defer span.End()
	defer m.addTracePresigned(u)

	return presignObj(ctx, m.bucketName, m.objName, m.needAKA)
}

// Run 启动上传
//
//	@receiver m *MinioManager
//	@param ctx context.Context
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:13
func (m *MinioManager) Run(ctx context.Context) *MinioManager {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, "UploadToMinio")
	m.Context = ctx
	m.span = span
	return m
}
