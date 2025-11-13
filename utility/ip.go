package utility

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"go.uber.org/zap"
)

// GetPubIP 获取公网ip
//
//	@return ip
//	@return err
func GetPubIP() (ipv4, ipv6 string, err error) {
	ctx := context.Background()
	ipv4, err = GetIpv4()
	if err != nil {
		logs.L().Ctx(ctx).Warn("获取ipv4失败", zap.Error(err))
	}
	ipv6, err = GetIpv6()
	if err != nil {
		logs.L().Ctx(ctx).Warn("获取ipv6失败", zap.Error(err))
	}
	logs.L().Info("获取ip结果", zap.String("ipv4", ipv4), zap.String("ipv6", ipv6))
	return
}

func GetIpv4() (ip string, err error) {
	ctx := context.Background()
	resp, err := requests.Req().Get("https://v4.ident.me/")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			logs.L().Ctx(ctx).Error("获取ip失败", zap.Error(err))
		} else {
			logs.L().Ctx(ctx).Error("获取ip失败", zap.Int("StatusCode", resp.StatusCode()))
		}
		return
	}
	return string(resp.Body()), err
}

func GetIpv6() (ip string, err error) {
	ctx := context.Background()
	resp, err := requests.Req().Get("https://v6.ident.me/")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			logs.L().Ctx(ctx).Error("获取ip失败", zap.Error(err))
		} else {
			logs.L().Ctx(ctx).Error("获取ip失败", zap.Int("StatusCode", resp.StatusCode()))
		}
		return
	}
	return string(resp.Body()), err
}
