package consts

type WordMatchType string

const (
	MatchTypeSubStr WordMatchType = "substr"
	MatchTypeRegex  WordMatchType = "regex"
	MatchTypeFull   WordMatchType = "full"
)
