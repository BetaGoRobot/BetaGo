package shorter

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/bytedance/sonic"
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

func GenAKA(ctx context.Context, u *url.URL) (newURL *url.URL) {
	expires := ExpireTime{
		Value: 30,
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
		logs.L.Error().Ctx(ctx).Err(err).Msg("Marshal failed")
		return
	}
	r, err := consts.HttpClient.R().
		SetHeader("X-API-KEY", os.Getenv("KUTT_API_KEY")).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post("https://kutt.kmhomelab.cn/api/links")
	if err != nil || (r.StatusCode() != 200 && r.StatusCode() != 201) {
		logs.L.Error().Ctx(ctx).Err(err).Int("status_code", r.StatusCode()).Msg("Post failed")
		return
	}
	resp := &KuttResp{}
	err = sonic.Unmarshal(r.Body(), resp)
	if err != nil {
		logs.L.Error().Ctx(ctx).Err(err).Msg("Unmarshal failed")
		return
	}
	newURL, err = url.Parse(resp.Link)
	if err != nil {
		logs.L.Error().Ctx(ctx).Err(err).Msg("Parse url failed")
		return
	}
	logs.L.Info().Ctx(ctx).Str("new_url", newURL.String()).Str("old_url", oldURL).Msg("GenAKA with url")
	return
}

func GenAKAKutt(ctx context.Context, u *url.URL, expires ExpireTime) (newURL *url.URL) {
	oldURL := u.String()
	req := &KuttRequest{
		Target:   oldURL,
		ExpireIn: fmt.Sprintf("%d%s", expires.Value, expires.Unit),
		Reuse:    true,
	}
	reqBody, err := sonic.Marshal(req)
	if err != nil {
		logs.L.Error().Ctx(ctx).Err(err).Msg("Marshal failed")
		return
	}
	r, err := consts.HttpClient.R().
		SetHeader("X-API-KEY", os.Getenv("KUTT_API_KEY")).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post("https://kutt.kmhomelab.cn/api/links")
	if err != nil || (r.StatusCode() != 200 && r.StatusCode() != 201) {
		logs.L.Error().Ctx(ctx).Err(err).Int("status_code", r.StatusCode()).Msg("Post failed")
		return
	}
	resp := &KuttResp{}
	err = sonic.Unmarshal(r.Body(), resp)
	if err != nil {
		logs.L.Error().Ctx(ctx).Err(err).Msg("Unmarshal failed")
		return
	}
	newURL, err = url.Parse(resp.Link)
	if err != nil {
		logs.L.Error().Ctx(ctx).Err(err).Msg("Parse url failed")
		return
	}
	logs.L.Info().Ctx(ctx).Str("new_url", newURL.String()).Str("old_url", oldURL).Msg("GenAKA with url")
	return
}

func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
