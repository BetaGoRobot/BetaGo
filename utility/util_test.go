package utility

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetOutBoundIP(t *testing.T) {
	fmt.Println(GetOutBoundIP())
	responseClient, errClient := http.Get("http://ip.dhcp.cn/?ip") // 获取外网 IP
	if errClient != nil {
		fmt.Printf("获取外网 IP 失败，请检查网络\n")
		panic(errClient)
	}
	// 程序在使用完 response 后必须关闭 response 的主体。
	defer responseClient.Body.Close()

	body, _ := ioutil.ReadAll(responseClient.Body)
	clientIP := fmt.Sprintf("%s", string(body))
	print(clientIP)
}
