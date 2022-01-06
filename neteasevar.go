package main

type loginJSON struct {
	Account struct {
		AnonimousUser      bool   `json:"anonimousUser"`
		Ban                int    `json:"ban"`
		BaoyueVersion      int    `json:"baoyueVersion"`
		CreateTime         int64  `json:"createTime"`
		DonateVersion      int    `json:"donateVersion"`
		ID                 int    `json:"id"`
		Salt               string `json:"salt"`
		Status             int    `json:"status"`
		TokenVersion       int    `json:"tokenVersion"`
		Type               int    `json:"type"`
		Uninitialized      bool   `json:"uninitialized"`
		UserName           string `json:"userName"`
		VipType            int    `json:"vipType"`
		ViptypeVersion     int64  `json:"viptypeVersion"`
		WhitelistAuthority int    `json:"whitelistAuthority"`
	} `json:"account"`
	Bindings []struct {
		BindingTime  int64  `json:"bindingTime"`
		Expired      bool   `json:"expired"`
		ExpiresIn    int64  `json:"expiresIn"`
		ID           int64  `json:"id"`
		RefreshTime  int    `json:"refreshTime"`
		TokenJSONStr string `json:"tokenJsonStr"`
		Type         int    `json:"type"`
		URL          string `json:"url"`
		UserID       int    `json:"userId"`
	} `json:"bindings"`
	Code      int    `json:"code"`
	Cookie    string `json:"cookie"`
	LoginType int    `json:"loginType"`
	Profile   struct {
		AccountStatus      int         `json:"accountStatus"`
		AuthStatus         int         `json:"authStatus"`
		Authority          int         `json:"authority"`
		AvatarDetail       interface{} `json:"avatarDetail"`
		AvatarImgID        int64       `json:"avatarImgId"`
		AvatarImgIDStr     string      `json:"avatarImgIdStr"`
		AvatarImgID_Str    string      `json:"avatarImgId_str"`
		AvatarURL          string      `json:"avatarUrl"`
		BackgroundImgID    int64       `json:"backgroundImgId"`
		BackgroundImgIDStr string      `json:"backgroundImgIdStr"`
		BackgroundURL      string      `json:"backgroundUrl"`
		Birthday           int64       `json:"birthday"`
		City               int         `json:"city"`
		DefaultAvatar      bool        `json:"defaultAvatar"`
		Description        string      `json:"description"`
		DetailDescription  string      `json:"detailDescription"`
		DjStatus           int         `json:"djStatus"`
		EventCount         int         `json:"eventCount"`
		ExpertTags         interface{} `json:"expertTags"`
		Experts            struct {
		} `json:"experts"`
		Followed                  bool        `json:"followed"`
		Followeds                 int         `json:"followeds"`
		Follows                   int         `json:"follows"`
		Gender                    int         `json:"gender"`
		Mutual                    bool        `json:"mutual"`
		Nickname                  string      `json:"nickname"`
		PlaylistBeSubscribedCount int         `json:"playlistBeSubscribedCount"`
		PlaylistCount             int         `json:"playlistCount"`
		Province                  int         `json:"province"`
		RemarkName                interface{} `json:"remarkName"`
		Signature                 string      `json:"signature"`
		UserID                    int         `json:"userId"`
		UserType                  int         `json:"userType"`
		VipType                   int         `json:"vipType"`
	} `json:"profile"`
	Token string `json:"token"`
}
