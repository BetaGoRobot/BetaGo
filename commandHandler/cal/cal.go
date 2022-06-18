package cal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/cosmanager"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/lonelyevil/khl"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/wcharczuk/go-chart/v2"
)

// ShowCalHandler 显示时间分布
//  @param userID
func ShowCalHandler(targetID, msgID, authorID, guildID string) (err error) {
	userInfo, err := utility.GetUserInfo(authorID, guildID)
	if err != nil {
		return
	}
	cosURL := DrawPieChart(GetUserChannelTimeMap(authorID), userInfo.Nickname)
	cardMessageStr, err := khl.CardMessage{&khl.CardMessageCard{
		Theme: khl.CardThemeInfo,
		Size:  khl.CardSizeLg,
		Modules: []interface{}{
			khl.CardMessageContainer{
				{
					Src:    cosURL,
					Size:   string(khl.CardSizeLg),
					Circle: false,
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		return
	}
	_, err = betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    msgID,
		},
	})
	return
}

// GetUserChannelTimeMap 获取用户在频道的时间
//  @param userID
//  @return map
func GetUserChannelTimeMap(userID string) map[string]time.Duration {
	logs := make([]*dbpack.ChannelLog, 0)
	userInfo, err := utility.GetUserInfo(userID, "")
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
		return nil
	}
	dbpack.GetDbConnection().Table("betago.channel_logs").Where("user_id = ? and is_update = ?", userInfo.ID, true).Find(&logs)
	var chanDiv = make(map[string]time.Duration)
	for _, log := range logs {
		if _, ok := chanDiv[log.ChannelID]; !ok {
			chanDiv[log.ChannelName] = log.LeftTime.Sub(log.JoinedTime)
		} else {
			chanDiv[log.ChannelName] = log.LeftTime.Sub(log.JoinedTime)
		}
	}

	return chanDiv
}

// DrawPieChart 获取频道的时间分布
//  @return {}
func DrawPieChart(inputMap map[string]time.Duration, userName string) string {
	// TODO: Replace with suitable method
	// pie := charts.NewPie()
	// pie.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "频道时间分布", Subtitle: userName}))
	// data := make([]opts.PieData, 0)
	// totalTime := time.Duration(0)
	// for _, v := range inputMap {
	// 	totalTime += v
	// }
	// for k, v := range inputMap {
	// 	data = append(data, opts.PieData{
	// 		Name:  k,
	// 		Value: float64(v.Nanoseconds()) / float64(totalTime.Nanoseconds()),
	// 	})
	// }
	// pie.AddSeries("频道时间分布", data).SetSeriesOptions(
	// 	charts.WithLabelOpts(
	// 		opts.Label{
	// 			Show:      true,
	// 			Formatter: "{b}\n{c}\n{@rate}",
	// 		},
	// 	),
	// 	charts.WithPieChartOpts(
	// 		opts.PieChart{
	// 			Radius: []string{"40%", "75%"},
	// 		},
	// 	),
	// )
	// TODO: Above
	values := make([]chart.Value, 0)
	var totalTime time.Duration
	for _, v := range inputMap {
		totalTime += v
	}
	for k, v := range inputMap {
		timeConv, _ := time.ParseDuration(fmt.Sprintf("%.1fs", v.Seconds()))
		values = append(values, chart.Value{
			Style: chart.Style{
				FontSize:            10,
				TextHorizontalAlign: 2,
				TextVerticalAlign:   4,
				TextWrap:            3,
				TextLineSpacing:     1,
				TextRotationDegrees: 0,
				FontColor:           chart.ColorBlack,
			},
			Label: k + " " + timeConv.String() + " " + fmt.Sprintf("%.2f", float64(v)/float64(totalTime)*100) + "%",
			Value: float64(timeConv),
		})
	}
	// TODO: 绘制频道时间饼状图
	pie := chart.PieChart{
		Title:  userName + "的频道时间分布",
		Width:  256,
		Height: 256,
		Canvas: chart.Style{
			FontColor: chart.ColorWhite,
		},
		SliceStyle: chart.Style{
			FontColor: chart.ColorWhite,
		},
		Font:   utility.GlowSansSC,
		Values: values,
	}
	fileName := time.Now().Format(time.RFC3339) + "_" + userName + "_chtime.png"
	filePath := filepath.Join(betagovar.ImagePath, fileName)
	f, _ := os.Create(filePath)
	defer f.Close()
	pie.Render(chart.PNG, f)

	return cosmanager.UploadFileToCos(filePath)
}
