package neteaseapi

import (
	"net/http"
	"os"
)

// IsTest 是否测试环境
var IsTest = os.Getenv("IS_TEST")

// LoginStatusStruct  登录状态
type LoginStatusStruct struct {
	Data struct {
		Code    int                    `json:"code"`
		Account map[string]interface{} `json:"account"`
		Profile map[string]interface{} `json:"profile"`
	} `json:"data"`
}

// NetEaseContext 网易云API调用封装
type NetEaseContext struct {
	cookies  []*http.Cookie
	err      error
	retryCnt int
	qrStruct struct {
		isOutDated bool
		uniKey     string
		qrBase64   string
	}
	loginType string
}

type dailySongs struct {
	Data struct {
		DailySongs []struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		} `json:"dailySongs"`
	} `json:"data"`
}
type musicList struct {
	Data []struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	} `json:"data"`
}

// MusicInfo 网易云音乐信息
type MusicInfo struct {
	ID   string
	URL  string
	Name string
}

type searchMusic struct {
	Result struct {
		Songs []struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
			Ar   []struct {
				Name string `json:"name"`
			} `json:"ar"`
			Al struct {
				PicURL string `json:"picUrl"`
			} `json:"al"`
		} `json:"songs"`
	} `json:"result"`
}

// SearchMusicRes  搜索音乐返回结果
type SearchMusicRes struct {
	ID         string
	Name       string
	ArtistName string
	SongURL    string
	PicURL     string
}

// GlobRecommendMusicRes  推荐音乐返回结果
type GlobRecommendMusicRes struct {
	Result []struct {
		PicURL string `json:"picUrl"`
		Song   struct {
			Name    string `json:"name"`
			ID      int    `json:"id"`
			Artists []struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"artists"`
		} `json:"song"`
	} `json:"result"`
}

// NetEaseAPIBaseURL 网易云API基础URL
var NetEaseAPIBaseURL = "http://netease-api:3335"

// NetEaseGCtx 网易云全局API调用封装
var NetEaseGCtx = &NetEaseContext{}
