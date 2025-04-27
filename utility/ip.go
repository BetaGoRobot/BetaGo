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
func GetPubIP() (ipv4, ipv6 string, err error) {
	ipv4, err = GetIpv4()
	if err != nil {
		log.Zlog.Warn("获取ipv4失败", zaplog.Error(err))
	}
	ipv6, err = GetIpv6()
	if err != nil {
		log.Zlog.Warn("获取ipv6失败", zaplog.Error(err))
	}
	log.Zlog.Info("获取ip结果", zaplog.String("ipv4", ipv4), zaplog.String("ipv6", ipv6))
	return
}

func GetIpv4() (ip string, err error) {
	resp, err := requests.Req().Get("https://v4.ident.me/")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			log.Zlog.Error("获取ip失败", zaplog.Error(err))
		} else {
			log.Zlog.Error("获取ip失败", zaplog.Int("StatusCode", resp.StatusCode()))
		}
		return
	}
	return string(resp.Body()), err
}

func GetIpv6() (ip string, err error) {
	resp, err := requests.Req().Get("https://v6.ident.me/")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			log.Zlog.Error("获取ip失败", zaplog.Error(err))
		} else {
			log.Zlog.Error("获取ip失败", zaplog.Int("StatusCode", resp.StatusCode()))
		}
		return
	}
	return string(resp.Body()), err
}
