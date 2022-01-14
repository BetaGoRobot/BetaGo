package main

import "net/http"

// NetEaseContext 网易云API调用封装
type NetEaseContext struct {
	cookies []*http.Cookie
	err     error
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
type musicInfo struct {
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
type searchMusicRes struct {
	ID         string
	Name       string
	ArtistName string
	SongURL    string
	PicURL     string
}

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
const NetEaseAPIBaseURL = "http://localhost:3335"
