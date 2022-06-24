package httptool

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
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
		log.Printf("%#v", resp)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(data))
	resp, err = PostWithTimestamp(
		RequestInfo{
			URL:     NetEaseAPIBaseURL + "/login/status",
			Params:  map[string][]string{},
			Cookies: resp.Cookies(),
		},
	)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("%#v", resp)
	}
	data, _ = ioutil.ReadAll(resp.Body)
	log.Println(string(data))
}
