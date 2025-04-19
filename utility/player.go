package utility

import (
	"net/url"
	"os"
	"strings"
)

func BuildURL(jsonURL string) string {
	u := &url.URL{}
	u.Host = os.Getenv("PLAYER_URL")
	u.Scheme = "https"
	q := u.Query()
	q.Add("target", strings.TrimPrefix(jsonURL, "https://kutt.kmhomelab.cn/"))
	u.RawQuery = q.Encode()
	return u.String()
}
