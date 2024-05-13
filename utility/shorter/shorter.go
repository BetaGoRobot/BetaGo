package shorter

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
)

func GenAKA(u *url.URL) (newURL *url.URL) {
	oldURL := u.String()
	slug := GetRandomString(8)
	r, err := consts.HttpClient.R().
		SetHeader("Authorization", os.Getenv("LYNX_API_KEY")).
		SetCookie(&http.Cookie{
			Name:  "token",
			Value: os.Getenv("LYNX_TOKEN"),
		}).
		SetFormData(
			map[string]string{
				"slug":        slug,
				"destination": oldURL,
			},
		).Post("https://aka.kmhomelab.cn/api/link")
	if err != nil || r.StatusCode() != 200 {
		log.ZapLogger.Error("Post failed", zaplog.Error(err), zaplog.Int("status_code", r.StatusCode()))
		return
	}
	newURL, err = url.Parse(slug)
	if err != nil {
		log.ZapLogger.Error("Parse url failed", zaplog.Error(err))
		return
	}
	log.ZapLogger.Info("GenAKA with url", zaplog.String("new_url", newURL.String()), zaplog.String("old_url", oldURL))
	newURL.Host = "aka.kmhomelab.cn"
	newURL.Scheme = "https"
	return
}

func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
