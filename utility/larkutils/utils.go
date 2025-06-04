package larkutils

import (
	"strings"
)

func TrimLyrics(lyrics string) string {
	lyricsList := strings.Split(lyrics, "\n")
	for index, lyric := range lyricsList {
		right := strings.Index(lyric, "]")
		lyricsList[index] = lyric[right+1:]
	}
	return strings.Join(lyricsList, "\n")
}
