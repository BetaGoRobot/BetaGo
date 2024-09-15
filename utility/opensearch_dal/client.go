package opensearchdal

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

var opensearchClient *opensearchapi.Client

var opensearchDomain = os.Getenv("OPENSEARCH_DOMAIN")

func init() {
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

func InsertData(ctx context.Context, index string, id string, data any) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	req := opensearchapi.IndexReq{
		Index:      index,
		DocumentID: id,
		Body:       opensearchutil.NewJSONReader(data),
	}
	_, err := opensearchClient.Index(ctx, req)

	return err
}
