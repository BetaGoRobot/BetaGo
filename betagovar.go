package main

type cardMessageModule struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Src   string `json:"src"`
	Cover string `json:"cover"`
}

type cardMessageTextModule struct {
	Type string `json:"type"`
	Text struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"text"`
}
