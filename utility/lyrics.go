package utility

import (
	"strconv"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	"go.uber.org/zap"
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
			logs.L().Warn("match string error", zap.Error(err))
			continue
		} else if matched {
			m, err := rePattern.FindStringMatch(lyric)
			if err != nil {
				logs.L().Warn("find string match error", zap.Error(err))
				continue
			}
			group := m.Groups()
			if len(group) < 3 {
				logs.L().Warn("group length less than 3", zap.Error(err))
				continue
			}
			minuteStr := group[1].String()
			secondStr := group[2].String()
			content := group[3].String()
			minute, err := strconv.Atoi(minuteStr)
			if err != nil {
				logs.L().Warn("convert minute to int error", zap.Error(err))
				continue
			}
			second, err := strconv.ParseFloat(secondStr, 64)
			if err != nil {
				logs.L().Warn("convert second to float error", zap.Error(err))
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
		logs.L().Warn("marshal string error", zap.Error(err))
		return
	}
	return
}
