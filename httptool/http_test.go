package httptool

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/heyuhengmatt/zaplog"
)

const NetEaseAPIBaseURL = "http://localhost:3335"

func TestPostWithParamsWithTimestamp(t *testing.T) {
	resp, err := PostWithTimestamp(RequestInfo{
		URL: NetEaseAPIBaseURL + "/login/cellphone",
		Params: map[string][]string{
			"phone":    {os.Getenv("NETEASE_PHONE")},
			"password": {os.Getenv("NETEASE_PASSWORD")},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		zapLogger.Error("登录失败", zaplog.Error(err))
	}
	data, _ := ioutil.ReadAll(resp.Body)
	zapLogger.Error("登录成功", zaplog.Error(err), zaplog.String("data", string(data)))
	resp, err = PostWithTimestamp(
		RequestInfo{
			URL:     NetEaseAPIBaseURL + "/login/status",
			Params:  map[string][]string{},
			Cookies: resp.Cookies(),
		},
	)
	if err != nil || resp.StatusCode != 200 {
		zapLogger.Error("获取登录状态失败", zaplog.Error(err))
	}
	data, _ = ioutil.ReadAll(resp.Body)
	zapLogger.Info("获取登录状态成功", zaplog.String("data", string(data)))
}
