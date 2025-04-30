package larkmsgutils

type msgTypeCons interface {
	textMsg | imageMsg | fileMsg | stickerMsg | postMsg
}

// text类型的消息
type textMsg struct {
	Text string `json:"text"`
}

// image类型的消息
type imageMsg struct {
	ImageKey string `json:"image_key"`
}

// file类型的消息
type fileMsg struct {
	FileKey string `json:"file_key"`
}

// 表情包类型的消息
type stickerMsg struct {
	FileKey string `json:"file_key"`
}

// 对于收到的post类型消息，可以通过这样的方式来解析其中的内容
type postMsg struct {
	Title   string           `json:"title"`
	Content [][]*contentData `json:"content"`
}

type contentData struct {
	Tag      string `json:"tag"`
	Text     string `json:"text"`
	ImageKey string `json:"image_key"`
	FileKey  string `json:"file_key"`
	UserID   string `json:"user_id"`
}

type Item struct {
	Tag     string `json:"tag"` // image text
	Content string `json:"content"`
}
