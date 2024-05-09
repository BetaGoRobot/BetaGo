package requests

import (
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	client      = resty.New()
	clientProxy = resty.New().SetProxy(os.Getenv("PRIVATE_PROXY"))
)

func Req() *resty.Request {
	return client.R()
}

func ReqTimestamp() *resty.Request {
	return client.R().SetQueryParam("timestamp", strconv.Itoa(int(time.Now().Unix())))
}

func ReqProxy() *resty.Request {
	return clientProxy.R()
}
