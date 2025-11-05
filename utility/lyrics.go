package utility

import (
	"strconv"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
)

type lyricMap struct {
	Lyrics []*lyricLine `json:"lyrics"`
}
type lyricLine struct {
	Line string `json:"line"`
	Time int    `json:"time"`
}

var rePattern = regexp2.MustCompile(`\[(.*):(.*)\](.*)`, 0)

func ExtractLyrics(lyric string) (s string, err error) {
	lyricList := strings.Split(lyric, "\n")
	newLyrics := make([]*lyricLine, 0)
	newLyrics = append(newLyrics,
		&lyricLine{
			Line: "",
			Time: -1,
		},
	)
	for _, lyric := range lyricList {
		if matched, err := rePattern.MatchString(lyric); err != nil {
			logs.L.Warn().Err(err).Msg("match string error")
			continue
		} else if matched {
			m, err := rePattern.FindStringMatch(lyric)
			if err != nil {
				logs.L.Warn().Err(err).Msg("find string match error")
				continue
			}
			group := m.Groups()
			if len(group) < 3 {
				logs.L.Warn().Err(err).Msg("group length less than 3")
				continue
			}
			minuteStr := group[1].String()
			secondStr := group[2].String()
			content := group[3].String()
			minute, err := strconv.Atoi(minuteStr)
			if err != nil {
				logs.L.Warn().Err(err).Msg("convert minute to int error")
				continue
			}
			second, err := strconv.ParseFloat(secondStr, 64)
			if err != nil {
				logs.L.Warn().Err(err).Msg("convert second to float error")
				continue
			}

			timeMs := minute*60*1000 + int(second*1000)

			newLyrics = append(newLyrics, &lyricLine{
				Line: content,
				Time: timeMs,
			})
		}
	}
	s, err = sonic.MarshalString(lyricMap{newLyrics})
	if err != nil {
		logs.L.Error().Err(err).Msg("marshal string error")
		return
	}
	return
}
