package consts

type WordMatchType string

const (
	MatchTypeSubStr WordMatchType = "substr"
	MatchTypeRegex  WordMatchType = "regex"
	MatchTypeFull   WordMatchType = "full"
)

type ReplyType string

const (
	ReplyTypeText    ReplyType = "text"
	ReplyTypeImg     ReplyType = "img"
	ReplyTypeSticker ReplyType = "sticker"
)
