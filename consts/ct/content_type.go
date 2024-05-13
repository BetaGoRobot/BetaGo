package ct

type ContentType string

const (
	ContentTypeImgPNG    ContentType = "img/png"
	ContentTypeImgJPEG   ContentType = "img/jpeg"
	ContentTypeAudio     ContentType = "audio/mpeg"
	ContentTypePlainText ContentType = "text/plain"
)

func (c *ContentType) String() string {
	return string(*c)
}

func ImgPNG() ContentType {
	return ContentTypeImgPNG
}

func ImgJPEG() ContentType {
	return ContentTypeImgJPEG
}

func Audio() ContentType {
	return ContentTypeAudio
}

func PlainText() ContentType {
	return ContentTypePlainText
}
