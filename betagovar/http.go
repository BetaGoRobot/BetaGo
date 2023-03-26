package betagovar

import (
	"os"

	"github.com/go-resty/resty/v2"
)

var (
	HttpClient          = resty.New()
	HttpClientWithProxy = resty.New().SetProxy(os.Getenv("PRIVATE_PROXY"))
)
