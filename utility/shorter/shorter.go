package shorter

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
)

type TimeUnit string

const (
	TimeUnitsMinute TimeUnit = "minutes"
	TimeUnitsHour   TimeUnit = "hours"
	TimeUnitsDay    TimeUnit = "days"
)

type ExpireTime struct {
	Value int
	Unit  TimeUnit
}
type KuttRequest struct {
	Target      string `json:"target"`
	Description string `json:"description"`
	ExpireIn    string `json:"expire_in"`
	Password    string `json:"password"`
	Customurl   string `json:"customurl"`
	Reuse       bool   `json:"reuse"`
	Domain      string `json:"domain"`
}

type KuttResp struct {
	Address     string    `json:"address"`
	Banned      bool      `json:"banned"`
	CreatedAt   time.Time `json:"created_at"`
	ID          string    `json:"id"`
	Link        string    `json:"link"`
	Password    bool      `json:"password"`
	Target      string    `json:"target"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
	VisitCount  int       `json:"visit_count"`
}

// func GenAKA(u *url.URL) (newURL *url.URL) { this is deprecated for LYNX
// 	oldURL := u.String()
// 	slug := GetRandomString(8)
// 	r, err := consts.HttpClient.R().
// 		SetHeader("Authorization", os.Getenv("LYNX_API_KEY")).
// 		SetCookie(&http.Cookie{
// 			Name:  "token",
// 			Value: os.Getenv("LYNX_TOKEN"),
// 		}).
// 		SetFormData(
// 			map[string]string{
// 				"slug":        slug,
// 				"destination": oldURL,
// 			},
// 		).Post("https://aka.kmhomelab.cn/api/link")
// 	if err != nil || r.StatusCode() != 200 {
// 		log.ZapLogger.Error("Post failed", zaplog.Error(err), zaplog.Int("status_code", r.StatusCode()))
// 		return
// 	}
// 	newURL, err = url.Parse(slug)
// 	if err != nil {
// 		log.ZapLogger.Error("Parse url failed", zaplog.Error(err))
// 		return
// 	}
// 	log.ZapLogger.Info("GenAKA with url", zaplog.String("new_url", newURL.String()), zaplog.String("old_url", oldURL))
// 	newURL.Host = "aka.kmhomelab.cn"
// 	newURL.Scheme = "https"
// 	return
// }

func GenAKA(u *url.URL) (newURL *url.URL) {
	expires := ExpireTime{
		Value: 7,
		Unit:  TimeUnitsDay,
	}
	oldURL := u.String()
	req := &KuttRequest{
		Target:   oldURL,
		ExpireIn: fmt.Sprintf("%d%s", expires.Value, expires.Unit),
		Reuse:    true,
	}
	reqBody, err := sonic.Marshal(req)
	if err != nil {
		log.Zlog.Error("Marshal failed", zaplog.Error(err))
		return
	}
	r, err := consts.HttpClient.R().
		SetHeader("X-API-KEY", os.Getenv("KUTT_API_KEY")).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post("https://kutt.kmhomelab.cn/api/links")
	if err != nil || (r.StatusCode() != 200 && r.StatusCode() != 201) {
		log.Zlog.Error("Post failed", zaplog.Error(err), zaplog.Int("status_code", r.StatusCode()))
		return
	}
	resp := &KuttResp{}
	err = sonic.Unmarshal(r.Body(), resp)
	if err != nil {
		log.Zlog.Error("Unmarshal failed", zaplog.Error(err))
		return
	}
	newURL, err = url.Parse(resp.Link)
	if err != nil {
		log.Zlog.Error("Parse url failed", zaplog.Error(err))
		return
	}
	log.Zlog.Info("GenAKA with url", zaplog.String("new_url", newURL.String()), zaplog.String("old_url", oldURL))
	return
}

func GenAKAKutt(u *url.URL, expires ExpireTime) (newURL *url.URL) {
	oldURL := u.String()
	req := &KuttRequest{
		Target:   oldURL,
		ExpireIn: fmt.Sprintf("%d%s", expires.Value, expires.Unit),
		Reuse:    true,
	}
	reqBody, err := sonic.Marshal(req)
	if err != nil {
		log.Zlog.Error("Marshal failed", zaplog.Error(err))
		return
	}
	r, err := consts.HttpClient.R().
		SetHeader("X-API-KEY", os.Getenv("KUTT_API_KEY")).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post("https://kutt.kmhomelab.cn/api/links")
	if err != nil || (r.StatusCode() != 200 && r.StatusCode() != 201) {
		log.Zlog.Error("Post failed", zaplog.Error(err), zaplog.Int("status_code", r.StatusCode()))
		return
	}
	resp := &KuttResp{}
	err = sonic.Unmarshal(r.Body(), resp)
	if err != nil {
		log.Zlog.Error("Unmarshal failed", zaplog.Error(err))
		return
	}
	newURL, err = url.Parse(resp.Link)
	if err != nil {
		log.Zlog.Error("Parse url failed", zaplog.Error(err))
		return
	}
	log.Zlog.Info("GenAKA with url", zaplog.String("new_url", newURL.String()), zaplog.String("old_url", oldURL))
	return
}

func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
