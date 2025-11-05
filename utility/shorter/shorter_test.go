package shorter

import (
	"context"
	"net/url"
	"testing"
)

func TestGenAKAKutt(t *testing.T) {
	u := &url.URL{
		Scheme: "https",
		Host:   "beta.betagov.cn",
		Path:   "/api/v1/oss/object",
	}
	GenAKAKutt(context.Background(), u, ExpireTime{1, TimeUnitsMinute})
}
