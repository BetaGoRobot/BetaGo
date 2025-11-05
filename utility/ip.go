package utility

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
)

// GetPubIP 获取公网ip
//
//	@return ip
//	@return err
func GetPubIP() (ipv4, ipv6 string, err error) {
	ipv4, err = GetIpv4()
	if err != nil {
		logs.L.Warn(context.Background(), "获取ipv4失败", "error", err)
	}
	ipv6, err = GetIpv6()
	if err != nil {
		logs.L.Warn(context.Background(), "获取ipv6失败", "error", err)
	}
	logs.L.Info(context.Background(), "获取ip结果", "ipv4", ipv4, "ipv6", ipv6)
	return
}

func GetIpv4() (ip string, err error) {
	resp, err := requests.Req().Get("https://v4.ident.me/")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			logs.L.Error(context.Background(), "获取ip失败", "error", err)
		} else {
			logs.L.Error(context.Background(), "获取ip失败", "StatusCode", resp.StatusCode())
		}
		return
	}
	return string(resp.Body()), err
}

func GetIpv6() (ip string, err error) {
	resp, err := requests.Req().Get("https://v6.ident.me/")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			logs.L.Error(context.Background(), "获取ip失败", "error", err)
		} else {
			logs.L.Error(context.Background(), "获取ip失败", "StatusCode", resp.StatusCode())
		}
		return
	}
	return string(resp.Body()), err
}
