package opensearchdal

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"go.uber.org/zap"
)

var opensearchClient *opensearchapi.Client

var opensearchDomain = os.Getenv("OPENSEARCH_DOMAIN")

func OpenSearchClient() *opensearchapi.Client {
	if opensearchClient == nil {
		var err error
		opensearchClient, err = opensearchapi.NewClient(opensearchapi.Config{
			Client: opensearch.Config{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
				Addresses: []string{
					"https://" + opensearchDomain + ":9200",
					"https://" + opensearchDomain + ":9200",
				},
				Username: os.Getenv("OPENSEARCH_USERNAME"),
				Password: os.Getenv("OPENSEARCH_PASSWORD"),
			},
		})
		if err != nil {
			panic(err)
		}
	}
	return opensearchClient
}

func InsertData(ctx context.Context, index string, id string, data any) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	index += "-" + time.Now().In(utility.UTCPlus8Loc()).Format("2006-01-02")
	req := opensearchapi.IndexReq{
		Index:      index,
		DocumentID: id,
		Body:       opensearchutil.NewJSONReader(data),
	}
	resp, err := OpenSearchClient().Index(ctx, req)
	if err != nil {
		logs.L().Ctx(ctx).Error("Index error", zap.Error(err), zap.Any("resp", resp))
		return err
	}
	return nil
}

func SearchData(ctx context.Context, index string, data any) (resp *opensearchapi.SearchResp, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	req := &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    opensearchutil.NewJSONReader(data),
	}
	resp, err = OpenSearchClient().Search(ctx, req)

	return resp, err
}

func SearchDataStr(ctx context.Context, index string, data string) (resp *opensearchapi.SearchResp, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	req := &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    strings.NewReader(data),
	}
	resp, err = OpenSearchClient().Search(ctx, req)
	return resp, err
}
