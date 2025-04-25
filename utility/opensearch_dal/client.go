package opensearchdal

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
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

func InsertData(ctx context.Context, index string, id string, data any) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	req := opensearchapi.IndexReq{
		Index:      index,
		DocumentID: id,
		Body:       opensearchutil.NewJSONReader(data),
	}
	_, err := OpenSearchClient().Index(ctx, req)

	return err
}

func SearchData(ctx context.Context, index string, data any) (*opensearchapi.SearchResp, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	req := &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    opensearchutil.NewJSONReader(data),
	}
	resp, err := OpenSearchClient().Search(ctx, req)

	return resp, err
}
