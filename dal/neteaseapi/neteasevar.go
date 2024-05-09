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
		Songs []Song `json:"songs"`
	} `json:"result"`
}

type Song struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Ar   []struct {
		Name string `json:"name"`
	} `json:"ar"`
	Al struct {
		PicURL string `json:"picUrl"`
	} `json:"al"`
}

// SearchMusicRes  搜索音乐返回结果
type SearchMusicRes struct {
	Index      int
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

type SearchLyrics struct {
	Sgc       bool `json:"sgc"`
	Sfy       bool `json:"sfy"`
	Qfy       bool `json:"qfy"`
	TransUser struct {
		ID       int    `json:"id"`
		Status   int    `json:"status"`
		Demand   int    `json:"demand"`
		Userid   int    `json:"userid"`
		Nickname string `json:"nickname"`
		Uptime   int64  `json:"uptime"`
	} `json:"transUser"`
	LyricUser struct {
		ID       int    `json:"id"`
		Status   int    `json:"status"`
		Demand   int    `json:"demand"`
		Userid   int    `json:"userid"`
		Nickname string `json:"nickname"`
		Uptime   int64  `json:"uptime"`
	} `json:"lyricUser"`
	Lrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"lrc"`
	Klyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"klyric"`
	Tlyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"tlyric"`
	Romalrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"romalrc"`
	Code int `json:"code"`
}

type MusicDetail struct {
	Songs []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		Pst  int    `json:"pst"`
		T    int    `json:"t"`
		Ar   []struct {
			ID    int           `json:"id"`
			Name  string        `json:"name"`
			Tns   []interface{} `json:"tns"`
			Alias []interface{} `json:"alias"`
		} `json:"ar"`
		Alia []interface{} `json:"alia"`
		Pop  int           `json:"pop"`
		St   int           `json:"st"`
		Rt   string        `json:"rt"`
		Fee  int           `json:"fee"`
		V    int           `json:"v"`
		Crbt interface{}   `json:"crbt"`
		Cf   string        `json:"cf"`
		Al   struct {
			ID     int           `json:"id"`
			Name   string        `json:"name"`
			PicURL string        `json:"picUrl"`
			Tns    []interface{} `json:"tns"`
			PicStr string        `json:"pic_str"`
			Pic    int64         `json:"pic"`
		} `json:"al"`
		Dt int `json:"dt"`
		H  struct {
			Br   int `json:"br"`
			Fid  int `json:"fid"`
			Size int `json:"size"`
			Vd   int `json:"vd"`
			Sr   int `json:"sr"`
		} `json:"h"`
		M struct {
			Br   int `json:"br"`
			Fid  int `json:"fid"`
			Size int `json:"size"`
			Vd   int `json:"vd"`
			Sr   int `json:"sr"`
		} `json:"m"`
		L struct {
			Br   int `json:"br"`
			Fid  int `json:"fid"`
			Size int `json:"size"`
			Vd   int `json:"vd"`
			Sr   int `json:"sr"`
		} `json:"l"`
		Sq                   interface{}   `json:"sq"`
		Hr                   interface{}   `json:"hr"`
		A                    interface{}   `json:"a"`
		Cd                   string        `json:"cd"`
		No                   int           `json:"no"`
		RtURL                interface{}   `json:"rtUrl"`
		Ftype                int           `json:"ftype"`
		RtUrls               []interface{} `json:"rtUrls"`
		DjID                 int           `json:"djId"`
		Copyright            int           `json:"copyright"`
		SID                  int           `json:"s_id"`
		Mark                 int           `json:"mark"`
		OriginCoverType      int           `json:"originCoverType"`
		OriginSongSimpleData interface{}   `json:"originSongSimpleData"`
		TagPicList           interface{}   `json:"tagPicList"`
		ResourceState        bool          `json:"resourceState"`
		Version              int           `json:"version"`
		SongJumpInfo         interface{}   `json:"songJumpInfo"`
		EntertainmentTags    interface{}   `json:"entertainmentTags"`
		AwardTags            interface{}   `json:"awardTags"`
		Single               int           `json:"single"`
		NoCopyrightRcmd      interface{}   `json:"noCopyrightRcmd"`
		Mv                   int           `json:"mv"`
		Rtype                int           `json:"rtype"`
		Rurl                 interface{}   `json:"rurl"`
		Mst                  int           `json:"mst"`
		Cp                   int           `json:"cp"`
		PublishTime          int           `json:"publishTime"`
	} `json:"songs"`
	Privileges []struct {
		ID                 int         `json:"id"`
		Fee                int         `json:"fee"`
		Payed              int         `json:"payed"`
		St                 int         `json:"st"`
		Pl                 int         `json:"pl"`
		Dl                 int         `json:"dl"`
		Sp                 int         `json:"sp"`
		Cp                 int         `json:"cp"`
		Subp               int         `json:"subp"`
		Cs                 bool        `json:"cs"`
		Maxbr              int         `json:"maxbr"`
		Fl                 int         `json:"fl"`
		Toast              bool        `json:"toast"`
		Flag               int         `json:"flag"`
		PreSell            bool        `json:"preSell"`
		PlayMaxbr          int         `json:"playMaxbr"`
		DownloadMaxbr      int         `json:"downloadMaxbr"`
		MaxBrLevel         string      `json:"maxBrLevel"`
		PlayMaxBrLevel     string      `json:"playMaxBrLevel"`
		DownloadMaxBrLevel string      `json:"downloadMaxBrLevel"`
		PlLevel            string      `json:"plLevel"`
		DlLevel            string      `json:"dlLevel"`
		FlLevel            string      `json:"flLevel"`
		Rscl               interface{} `json:"rscl"`
		FreeTrialPrivilege struct {
			ResConsumable      bool        `json:"resConsumable"`
			UserConsumable     bool        `json:"userConsumable"`
			ListenType         interface{} `json:"listenType"`
			CannotListenReason interface{} `json:"cannotListenReason"`
			PlayReason         interface{} `json:"playReason"`
		} `json:"freeTrialPrivilege"`
		RightSource    int `json:"rightSource"`
		ChargeInfoList []struct {
			Rate          int         `json:"rate"`
			ChargeURL     interface{} `json:"chargeUrl"`
			ChargeMessage interface{} `json:"chargeMessage"`
			ChargeType    int         `json:"chargeType"`
		} `json:"chargeInfoList"`
	} `json:"privileges"`
	Code int `json:"code"`
}

// NetEaseAPIBaseURL 网易云API基础URL
var NetEaseAPIBaseURL = "http://netease-api:3335"

// NetEaseGCtx 网易云全局API调用封装
var NetEaseGCtx = &NetEaseContext{}
