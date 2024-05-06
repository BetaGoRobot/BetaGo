package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
)

func main() {
	cookieStr := `MUSIC_U=00F3636064E38E833BA8B5EFEDCE13663DB3D7E0CAB447F474785F3FF442E51E7CCBF527D55FED0D28A884FB984F538EBD3EE8ED0425A90A8300A7376C5B78C8F96E47F2A9CF4146D20A8F693DDEFC379D56467FB7DE596F9F9387D4B15E2A675D5166E295DAAE2423D3F520D605EEEA4405DEB5094B66E47E521BBC9CC6E6642DC2D60B80EB98FFF30B21B5DCBBCA17ACCD476408E85566A44C96BCAB8B7EA55B1DB12FA5560B60798553F2492CA36A2A107EF4200FB5032B613A5AFED3263A8E252D0082D495C534845301E0FEF3154E94C99DEF537C18E629359C68FC45D7B651B563E6DB55428A07C4B4D9ADA4112D4FAF88F9325D78E5D315806960904BD3062DFE611D33C9E4331224D83CFD62C631D3934534A37D68A549A51E492EC0201009FB95D7FB2D16B8FE8346DD1551E3361221501BF4546552CE65398525086CD8E5E50D243D4601D9C445A2D8C7B9BA; NMTID=00ONIob4G4hHnVwrEPTsl01US2PhskAAAGPTUt38w; __csrf=29817d9e4af42edab8ba459f9236e12a`
	req := betagovar.HttpClient.R().
		SetHeader("Cookie", cookieStr)
	fmt.Println(req.Cookies)
	r, err := req.Post("http://v4.kmhomelab.cn:3335/login/status?timestamp=" + strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		panic(err)
	}
	fmt.Println(r.Header())
	fmt.Println(string(r.Body()))
	_ = r
}
