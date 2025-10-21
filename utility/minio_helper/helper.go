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
	"github.com/BetaGoRobot/go_utils/reflecting"

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
	domain         string
	overwrite      bool
}

// Client 返回一个新的minioManager Client
func Client() *MinioManager {
	return &MinioManager{
		Context: context.Background(),
		needAKA: true,
		domain:  "kutt.kmhomelab.cn",
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

// Overwrite 是否覆盖文件
//
//	@receiver m *MinioManager
//	@param ctx context.Context
//	@return *MinioManager
//	@author heyuhengmatt
//	@update 2024-05-13 01:54:13
func (m *MinioManager) Overwrite() *MinioManager {
	m.overwrite = true
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
			log.Zlog.Error("Get file failed", zaplog.Error(err))
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

// SetV4 to be filled SetV4  设置contentType
//
//	@receiver m *MinioManager
//	@return *MinioManager
//	@author kevinmatthe
//	@update 2025-04-28 21:07:40
func (m *MinioManager) SetV4() *MinioManager {
	m.span.SetAttributes(attribute.Key("stack").String("V4"))
	m.domain = "kutt.kmhomelab.online:2443"
	return m
}

// SetV6  设置V6
//
//	@receiver m *MinioManager
//	@return *MinioManager
//	@author kevinmatthe
//	@update 2025-04-28 21:07:40
func (m *MinioManager) SetV6() *MinioManager {
	m.span.SetAttributes(attribute.Key("stack").String("V6"))
	m.domain = "kutt.kmhomelab.cn"
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
			log.Zlog.Error("presigned url is empty")
		}
	} else {
		log.Zlog.Error("presigned url is nil")
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
	if !m.overwrite {
		u, err = m.TryGetFile()
		if err != nil {
			u, err = m.UploadFileOverwrite(opts)
			if err != nil {
				return
			}
		}
	} else {
		u, err = m.UploadFileOverwrite(opts)
		if err != nil {
			return
		}
	}

	m.addTraceCached(true)
	return
}

func (m *MinioManager) UploadFileOverwrite(opts minio.PutObjectOptions) (u *url.URL, err error) {
	m.addTraceCached(false)
	if m.inputTransFunc != nil {
		m.inputTransFunc(m)
	}
	log.Zlog.Warn("tryGetFile failed", zaplog.Error(err))
	err = m.UploadFile(opts)
	if err != nil {
		log.Zlog.Error("uploadFile failed", zaplog.Error(err))
		return
	}
	return m.PresignURL()
}

// 此函数会修改入参，不返回err外的值
func (m *MinioManager) TryGetFile() (shareURL *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(m, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	if MinioCheckFileExists(ctx, m.bucketName, m.objName) {
		return m.PresignURL()
	}
	return nil, errors.New("file not exists")
}

func (m *MinioManager) UploadFile(opts minio.PutObjectOptions) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(m, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	err = minioUploadReader(ctx, m.bucketName, m.file, m.objName, opts)
	if err != nil {
		log.Zlog.Error(err.Error())
		return
	}
	return
}

func (m *MinioManager) PresignURL() (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(m, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	u, err = presignObjInner(ctx, m.bucketName, m.objName)
	if err != nil {
		return
	}
	m.span.SetAttributes(attribute.String("presigned_url", u.String()))
	if m.needAKA {
		u = shortenURL(ctx, u)
		u.Host = m.domain
		m.span.SetAttributes(attribute.String("presigned_url_shortened", u.String()))
	}
	return
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
