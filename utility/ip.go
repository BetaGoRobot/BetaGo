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
	ctx := context.Background()
	ipv4, err = GetIpv4()
	if err != nil {
		logs.L.Warn().Ctx(ctx).Err(err).Msg("获取ipv4失败")
	}
	ipv6, err = GetIpv6()
	if err != nil {
		logs.L.Warn().Ctx(ctx).Err(err).Msg("获取ipv6失败")
	}
	logs.L.Info().Ctx(ctx).Str("ipv4", ipv4).Str("ipv6", ipv6).Msg("获取ip结果")
	return
}

func GetIpv4() (ip string, err error) {
	ctx := context.Background()
	resp, err := requests.Req().Get("https://v4.ident.me/")
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			logs.L.Error().Ctx(ctx).Err(err).Msg("获取ip失败")
		} else {
			logs.L.Error().Ctx(ctx).Int("StatusCode", resp.StatusCode()).Msg("获取ip失败")
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
			logs.L.Error().Ctx(ctx).Err(err).Msg("获取ip失败")
		} else {
			logs.L.Error().Ctx(ctx).Int("StatusCode", resp.StatusCode()).Msg("获取ip失败")
		}
		return
	}
	return string(resp.Body()), err
}
