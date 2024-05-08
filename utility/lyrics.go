package utility

import (
	"strconv"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	"github.com/kevinmatthe/zaplog"
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
			log.ZapLogger.Warn("match string error", zaplog.String("lyric", lyric))
			continue
		} else if matched {
			m, err := rePattern.FindStringMatch(lyric)
			if err != nil {
				log.ZapLogger.Warn("find string match error", zaplog.String("lyric", lyric))
				continue
			}
			group := m.Groups()
			if len(group) < 3 {
				log.ZapLogger.Warn("group length less than 3", zaplog.String("lyric", lyric))
				continue
			}
			minuteStr := group[1].String()
			secondStr := group[2].String()
			content := group[3].String()
			minute, err := strconv.Atoi(minuteStr)
			if err != nil {
				log.ZapLogger.Warn("convert minute to int error", zaplog.String("minute", minuteStr))
				continue
			}
			second, err := strconv.ParseFloat(secondStr, 64)
			if err != nil {
				log.ZapLogger.Warn("convert second to float error", zaplog.String("second", secondStr))
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
		log.ZapLogger.Error("marshal string error", zaplog.Error(err))
		return
	}
	return
}
