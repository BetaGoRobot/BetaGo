package qqmusicapi

var (
	qqmusicBaseURL    = "http://qqmusic-api:3300"
	qqmusicPicBaseURL = "https://y.gtimg.cn/music/photo_new/T002R300x300M000"
)

// QQmusicContext QQ音乐API调用封装
type QQmusicContext struct {
	err error
}

// QQmusicSearchResponse QQ音乐搜索结果
type QQmusicSearchResponse struct {
	Data struct {
		List []struct {
			Albumid   int    `json:"albumid"`
			Albummid  string `json:"albummid"`
			Albumname string `json:"albumname"`
			Singer    []struct {
				Name string `json:"name"`
			} `json:"singer"`
			Songid      int    `json:"songid"`
			Songmid     string `json:"songmid"`
			Songname    string `json:"songname"`
			StrMediaMid string `json:"strMediaMid"`
		} `json:"list"`
	} `json:"data"`
}

// SearchMusicRes QQ音乐搜索的结果
type SearchMusicRes struct {
	ID         string
	Name       string
	ArtistName string
	SongURL    string
	PicURL     string
}

// MusicInfo 音乐信息
type MusicInfo struct {
	ID   string
	URL  string
	Name string
}

// MusicURLId 获取到的url
type MusicURLId struct {
	SongURL string `json:"data"`
	Result  int    `json:"result"`
}
