package utility

import (
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"github.com/kevinmatthe/zaplog"
)

// GetPubIP 获取公网ip
//
//	@return ip
//	@return err
func GetPubIP() (ip string, err error) {
	resp, err := requests.Req().Get("http://ifconfig.me")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			log.ZapLogger.Error("获取ip失败", zaplog.Error(err))
		} else {
			log.ZapLogger.Error("获取ip失败", zaplog.Int("StatusCode", resp.StatusCode()))
		}
	}
	data := resp.Body()
	ip = string(data)
	log.ZapLogger.Info("获取ip成功", zaplog.String("data", ip))
	return
}
