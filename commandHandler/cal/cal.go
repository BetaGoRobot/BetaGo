package cal

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/cosmanager"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/BetaGoRobot/BetaGo/httptool"
	"github.com/lonelyevil/khl"

	"github.com/BetaGoRobot/BetaGo/utility"
)

// ChartType  图表类型
// p: 2 dimension
// pc: more than 1 dimension
// pd: donut
type ChartType string

// AnimationTime 动画时长
type AnimationTime int

// AnimationType 动画类型
type AnimationType string

// DrawPieAPICtx 画饼图的上下文
type DrawPieAPICtx struct {
	Ct        ChartType // bvs,p,ls
	Title     DrawChartTitle
	Label     DrawChartLabel
	Legend    DrawChartLegend
	Data      []string
	IsDivided bool
	Color     string
	Size      string
}

// DrawChartTitle 画图标题
type DrawChartTitle struct {
	Text  string
	Size  int
	Color string // RGB,eg: FF0000,20,r
}

// DrawChartLegend  画图标签
type DrawChartLegend struct {
	Text     []string
	Size     int
	Color    string
	Position string //chdlp:b,t,l,r
}

// DrawChartLabel 画图标签
type DrawChartLabel struct { //chl
	Text []string
	Size int
}

// BuildRequestURL  构建请求URL
//  @receiver ctx
//  @return string
func (ctx *DrawPieAPICtx) BuildRequestURL() string {

	// 构建图类型
	chartTypeStr := string("cht=" + ctx.Ct)
	// 构建图标题
	titleSlice := []string{"chtt=" + url.QueryEscape(ctx.Title.Text)}
	if ctx.Title.Color != "" {
		titleSlice = append(titleSlice, "chts="+url.QueryEscape(ctx.Title.Color+","+strconv.Itoa(ctx.Title.Size)))
	} else {
		titleSlice = append(titleSlice, "chts="+url.QueryEscape("000000"+","+strconv.Itoa(ctx.Title.Size)))
	}
	titleStr := strings.Join(titleSlice, "&")

	// 构建图Legend
	var legendText string
	for i, text := range ctx.Legend.Text {
		if i == len(ctx.Legend.Text)-1 {
			legendText += text
		} else {
			legendText += text + "|"
		}
	}
	legendSlice := []string{"chdl=" + url.QueryEscape(legendText)}
	legendSlice = append(legendSlice, "chdlp="+ctx.Legend.Position)
	if ctx.Legend.Size != 0 {
		legendSlice = append(legendSlice, "chdls="+url.QueryEscape("000000,"+strconv.Itoa(ctx.Legend.Size)))
	}
	legendText = strings.Join(legendSlice, "&")

	// 构建图Label
	var labelText string
	for i, text := range ctx.Label.Text {
		if i != 0 {
			labelText += url.QueryEscape("|" + text)
		} else {
			labelText += url.QueryEscape(text)
		}
	}
	labelText = "chl=" + labelText
	if ctx.Label.Size != 0 {
		labelText += "&chlps=" + url.QueryEscape("font.size,"+strconv.Itoa(ctx.Label.Size))
	}
	// 构建图数据
	var dataText string
	for i, text := range ctx.Data {
		if i != 0 {
			dataText += url.QueryEscape("," + text)
		} else {
			dataText += url.QueryEscape(text)
		}
		if ctx.IsDivided && i != len(ctx.Data)-1 && (i)%(len(ctx.Data)/2) == 0 {
			dataText += "|"
		}
	}
	dataText = "chd=t:" + dataText

	// 构建图颜色
	ctx.Color = "chf=ps0-0,lg,45,ffeb3b,0.2,f44336,1"

	// 构建图大小
	sizeText := "chs=" + ctx.Size
	QueryStr := strings.Join([]string{chartTypeStr, titleStr, legendText, labelText, dataText, ctx.Color, sizeText}, "&")

	return QueryStr
}

// ShowCalHandler 显示时间分布
//  @param userID
func ShowCalHandler(targetID, msgID, authorID, guildID string, args ...string) (err error) {
	var (
		userInfo      *khl.User
		cardContainer khl.CardMessageContainer
	)
	if args != nil {
		// 含参数，则获取参数中用户的时间分布
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			userInfo, err = utility.GetUserInfo(userID, guildID)
			if err != nil {
				return
			}
			cardContainer = append(cardContainer,
				khl.CardMessageElementImage{
					Src:  DrawPieChartWithAPI(GetUserChannelTimeMap(userID), userInfo.Nickname),
					Size: string(khl.CardSizeLg),
				},
			)
		}
	} else {
		// 无参数，则获取当前用户的时间分布
		userInfo, err = utility.GetUserInfo(authorID, guildID)
		if err != nil {
			return
		}
		cardContainer = append(cardContainer,
			khl.CardMessageElementImage{
				Src:  DrawPieChartWithAPI(GetUserChannelTimeMap(authorID), userInfo.Nickname),
				Size: string(khl.CardSizeLg),
			},
		)
	}
	cardMessageStr, err := khl.CardMessage{&khl.CardMessageCard{
		Theme: khl.CardThemeInfo,
		Size:  khl.CardSizeLg,
		Modules: []interface{}{
			cardContainer,
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
			chanDiv[log.ChannelName] += log.LeftTime.Sub(log.JoinedTime)
		} else {
			chanDiv[log.ChannelName] += log.LeftTime.Sub(log.JoinedTime)
		}
	}

	return chanDiv
}

// DrawPieChartWithAPI 获取频道的时间分布
//  @param inputMap
//  @param userName
//  @return string
func DrawPieChartWithAPI(inputMap map[string]time.Duration, userName string) string {
	apiURL := "https://image-charts.com/chart?"
	ctx := &DrawPieAPICtx{
		Ct: "p3",
		Title: DrawChartTitle{
			Text: userName + "的频道时间分布",
			Size: 40,
		},
		Label: DrawChartLabel{
			Text: []string{},
			Size: 30,
		},
		Legend: DrawChartLegend{
			Text:     []string{},
			Size:     20,
			Position: "r",
		},
		Size: "812x812",
		// IsDivided: true,
	}
	var totalTime time.Duration
	for _, v := range inputMap {
		totalTime += v
	}
	for k, v := range inputMap {
		timeConv, _ := time.ParseDuration(fmt.Sprintf("%.1fs", v.Seconds()))
		ctx.Label.Text = append(ctx.Label.Text, k+"\n"+timeConv.String()+"\n"+fmt.Sprintf("%.2f", float64(v)/float64(totalTime)*100)+"%")
		ctx.Legend.Text = append(ctx.Legend.Text, k)
		ctx.Data = append(ctx.Data, fmt.Sprintf("%.1f", float64(v)/float64(totalTime)*100))
	}
	apiURL += ctx.BuildRequestURL()
	resp, err := httptool.GetWithParams(httptool.RequestInfo{URL: apiURL})
	if err != nil {
		log.Fatal(err)
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
		return ""
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		log.Fatal(err)
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
		return ""
	}
	defer resp.Body.Close()
	filePath := filepath.Join(betagovar.ImagePath, time.Now().Format(time.RFC3339)+"_"+userName+"_chtime.png")
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
		return ""
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
		return ""
	}
	return cosmanager.UploadFileToCos(filePath)
}
