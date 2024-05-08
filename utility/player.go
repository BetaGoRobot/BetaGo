package utility

import (
	"net/url"
	"os"
	"strconv"
)

func BuildURL(lyricsURL, musicURL, pictureURL, album, title, artist string, duration int) string {
	u := &url.URL{}
	u.Host = os.Getenv("PLAYER_URL")
	u.Scheme = "https"
	q := u.Query()
	q.Add("lyrics", lyricsURL)
	q.Add("music", musicURL)
	q.Add("picture", pictureURL)
	q.Add("duration", strconv.Itoa(duration))
	q.Add("album", album)
	q.Add("title", title)
	q.Add("artists", artist)
	u.RawQuery = q.Encode()
	return u.String()
}
