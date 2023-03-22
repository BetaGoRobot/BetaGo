package cal

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/httptool"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/lonelyevil/kook"
	"github.com/wcharczuk/go-chart/v2"
	"go.opentelemetry.io/otel/attribute"

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
type DrawChartLabel struct { // chl
	Text []string
	Size int
}

// BuildRequestURL  构建请求URL
//
//	@receiver drawCtx
//	@return string
func (drawCtx *DrawPieAPICtx) BuildRequestURL() string {
	// 构建图类型
	chartTypeStr := string("cht=" + drawCtx.Ct)
	// 构建图标题
	titleSlice := []string{"chtt=" + url.QueryEscape(drawCtx.Title.Text)}
	if drawCtx.Title.Color != "" {
		titleSlice = append(titleSlice, "chts="+url.QueryEscape(drawCtx.Title.Color+","+strconv.Itoa(drawCtx.Title.Size)))
	} else {
		titleSlice = append(titleSlice, "chts="+url.QueryEscape("000000"+","+strconv.Itoa(drawCtx.Title.Size)))
	}
	titleStr := strings.Join(titleSlice, "&")

	// 构建图Legend
	var legendText string
	for i, text := range drawCtx.Legend.Text {
		if i == len(drawCtx.Legend.Text)-1 {
			legendText += text
		} else {
			legendText += text + "|"
		}
	}
	legendSlice := []string{"chdl=" + url.QueryEscape(legendText)}
	legendSlice = append(legendSlice, "chdlp="+drawCtx.Legend.Position)
	if drawCtx.Legend.Size != 0 {
		legendSlice = append(legendSlice, "chdls="+url.QueryEscape("000000,"+strconv.Itoa(drawCtx.Legend.Size)))
	}
	legendText = strings.Join(legendSlice, "&")

	// 构建图Label
	var labelText string
	for i, text := range drawCtx.Label.Text {
		if i != 0 {
			labelText += url.QueryEscape("|" + text)
		} else {
			labelText += url.QueryEscape(text)
		}
	}
	labelText = "chl=" + labelText
	if drawCtx.Label.Size != 0 {
		labelText += "&chlps=" + url.QueryEscape("font.size,"+strconv.Itoa(drawCtx.Label.Size))
	}
	// 构建图数据
	var dataText string
	for i, text := range drawCtx.Data {
		if i != 0 {
			dataText += url.QueryEscape("," + text)
		} else {
			dataText += url.QueryEscape(text)
		}
		if drawCtx.IsDivided && i != len(drawCtx.Data)-1 && (i)%(len(drawCtx.Data)/2) == 0 {
			dataText += "|"
		}
	}
	dataText = "chd=t:" + dataText

	// 构建图颜色
	drawCtx.Color = "chf=ps0-0,lg,45,ffeb3b,0.2,f44336,1"

	// 构建图大小
	sizeText := "chs=" + drawCtx.Size
	QueryStr := strings.Join([]string{chartTypeStr, titleStr, legendText, labelText, dataText, drawCtx.Color, sizeText}, "&")

	return QueryStr
}

// ShowCalHandler 显示时间分布
//
//	@param userID
func ShowCalHandler(ctx context.Context, targetID, quoteID, authorID, guildID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	var (
		userInfo      *kook.User
		cardContainer kook.CardMessageContainer
	)
	if args != nil {
		// 含参数，则获取参数中用户的时间分布
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			userInfo, err = utility.GetUserInfo(userID, guildID)
			if err != nil {
				return
			}
			URL, tmpErr := DrawPieChartWithAPI(ctx, GetUserChannelTimeMap(ctx, userInfo.ID, guildID), userInfo.Nickname)
			if tmpErr != nil {
				// 尝试使用本地绘图
				URL, err = DrawPieChartWithLocal(ctx, GetUserChannelTimeMap(ctx, userInfo.ID, guildID), userInfo.Nickname)
				if err != nil {
					return err
				}
				errorsender.SendErrorInfo(targetID, quoteID, authorID, tmpErr, ctx)
			}
			cardContainer = append(cardContainer,
				kook.CardMessageElementImage{
					Src:  URL,
					Size: string(kook.CardSizeLg),
				},
			)
		}
	} else {
		// 无参数，则获取当前用户的时间分布
		userInfo, err = utility.GetUserInfo(authorID, guildID)
		if err != nil {
			return
		}
		URL, tmpErr := DrawPieChartWithAPI(ctx, GetUserChannelTimeMap(ctx, userInfo.ID, guildID), userInfo.Nickname)
		if tmpErr != nil {
			// 尝试使用本地绘图
			URL, err = DrawPieChartWithLocal(ctx, GetUserChannelTimeMap(ctx, userInfo.ID, guildID), userInfo.Nickname)
			if err != nil {
				return err
			}
			errorsender.SendErrorInfo(targetID, quoteID, authorID, tmpErr, ctx)
		}
		cardContainer = append(cardContainer,
			kook.CardMessageElementImage{
				Src:  URL,
				Size: string(kook.CardSizeLg),
			},
		)
	}
	cardMessageStr, err := kook.CardMessage{&kook.CardMessageCard{
		Theme: kook.CardThemeInfo,
		Size:  kook.CardSizeLg,
		Modules: []interface{}{
			cardContainer,
			&kook.CardMessageDivider{},
			&kook.CardMessageSection{
				Mode: kook.CardMessageSectionModeRight,
				Text: &kook.CardMessageElementKMarkdown{
					Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
				},
				Accessory: kook.CardMessageElementButton{
					Theme: kook.CardThemeSuccess,
					Value: "https://jaeger.kevinmatt.top/trace/" + span.SpanContext().TraceID().String(),
					Click: "link",
					Text:  "链路追踪",
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		return
	}
	_, err = betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    quoteID,
		},
	})
	return
}

// ShowCalLocalHandler sc
//
//	@param drawCtx
//	@param targetID
//	@param msgID
//	@param authorID
//	@param guildID
//	@param args
//	@return err
func ShowCalLocalHandler(ctx context.Context, targetID, quoteID, authorID, guildID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	var (
		userInfo      *kook.User
		cardContainer kook.CardMessageContainer
	)
	if args != nil {
		// 含参数，则获取参数中用户的时间分布
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			userInfo, err = utility.GetUserInfo(userID, guildID)
			if err != nil {
				return
			}
			// 尝试使用本地绘图
			URL, err := DrawPieChartWithLocal(ctx, GetUserChannelTimeMap(ctx, userInfo.ID, guildID), userInfo.Nickname)
			if err != nil {
				return err
			}
			cardContainer = append(cardContainer,
				kook.CardMessageElementImage{
					Src:  URL,
					Size: string(kook.CardSizeLg),
				},
			)
		}
	} else {
		// 无参数，则获取当前用户的时间分布
		userInfo, err = utility.GetUserInfo(authorID, guildID)
		if err != nil {
			return
		}
		URL, err := DrawPieChartWithLocal(ctx, GetUserChannelTimeMap(ctx, userInfo.ID, guildID), userInfo.Nickname)
		if err != nil {
			return err
		}
		cardContainer = append(cardContainer,
			kook.CardMessageElementImage{
				Src:  URL,
				Size: string(kook.CardSizeLg),
			},
		)
	}
	cardMessageStr, err := kook.CardMessage{&kook.CardMessageCard{
		Theme: kook.CardThemeInfo,
		Size:  kook.CardSizeLg,
		Modules: []interface{}{
			cardContainer,
			&kook.CardMessageDivider{},
			&kook.CardMessageSection{
				Mode: kook.CardMessageSectionModeRight,
				Text: &kook.CardMessageElementKMarkdown{
					Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
				},
				Accessory: kook.CardMessageElementButton{
					Theme: kook.CardThemeSuccess,
					Value: "https://jaeger.kevinmatt.top/trace/" + span.SpanContext().TraceID().String(),
					Click: "link",
					Text:  "链路追踪",
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		return
	}
	_, err = betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    quoteID,
		},
	})
	return
}

// GetUserChannelTimeMap 获取用户在频道的时间
//
//	@param userID
//	@return map
func GetUserChannelTimeMap(ctx context.Context, userID, guildID string) map[string]time.Duration {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("userID").String(userID))
	defer span.End()

	logs := make([]*utility.ChannelLogExt, 0)
	userInfo, err := utility.GetUserInfo(userID, "")
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err, ctx)
		return nil
	}
	utility.GetDbConnection().
		Table("betago.channel_log_exts").
		Where("user_id = ? and is_update = ? and guild_id = ?", userInfo.ID, true, guildID).
		Order("left_time desc").
		Find(&logs).
		Limit(1000)
	chanDiv := make(map[string]time.Duration)
	var totalTime time.Duration
	for _, log := range logs {
		leftTimeT, _ := time.Parse(betagovar.TimeFormat, log.LeftTime)
		joinTimeT, _ := time.Parse(betagovar.TimeFormat, log.JoinedTime)
		timeCost := leftTimeT.Sub(joinTimeT)
		if timeCost < time.Minute*10 {
			// 10分钟以内的数据忽略
			continue
		}
		chanDiv[log.ChannelName] += timeCost
		if totalTime+timeCost >= time.Hour*24 {
			chanDiv[log.ChannelName] += (time.Hour*24 - totalTime)
			break
		}
		totalTime += timeCost
	}
	return chanDiv
}

// DrawPieChartWithAPI 本地获取频道的时间分布
//
//	@param inputMap
//	@param userName
//	@return string
func DrawPieChartWithAPI(ctx context.Context, inputMap map[string]time.Duration, userName string) (string, error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	apiURL := "https://image-charts.com/chart?"
	drawCtx := &DrawPieAPICtx{
		Ct: "p3",
		Title: DrawChartTitle{
			Text: userName + "的频道时间分布(Last 24h)",
			Size: 35,
		},
		Label: DrawChartLabel{
			Text: []string{},
			Size: 25,
		},
		Legend: DrawChartLegend{
			Text:     []string{},
			Size:     15,
			Position: "t",
		},
		Size: "600x600",
		// IsDivided: true,
	}
	var (
		totalTime time.Duration
		tmpSlice  = make([]struct {
			k string
			v time.Duration
		}, 0)
	)

	for k, v := range inputMap {
		totalTime += v
		tmpSlice = append(tmpSlice, struct {
			k string
			v time.Duration
		}{k, v})
	}
	sort.Slice(tmpSlice, func(i, j int) bool {
		return tmpSlice[i].v > tmpSlice[j].v
	})
	for _, item := range tmpSlice {
		k := item.k
		v := item.v
		tmp := fmt.Sprintf("%.1fm", v.Minutes())
		timeConv, _ := time.ParseDuration(tmp)
		percent := float64(v) / float64(totalTime) * 100
		percentStr := fmt.Sprintf("%.1f", float64(v)/float64(totalTime)*100) + "%"
		timeConvWithPercent := timeConv.String() + "\n" + percentStr
		if percent >= 10 {
			// %5实质为最大1个小时的值
			drawCtx.Label.Text = append(drawCtx.Label.Text, k+"\n"+timeConvWithPercent)
		} else {
			drawCtx.Label.Text = append(drawCtx.Label.Text, percentStr)
		}
		drawCtx.Legend.Text = append(drawCtx.Legend.Text, k+"-"+percentStr)
		drawCtx.Data = append(drawCtx.Data, fmt.Sprintf("%.1f", float64(v)/float64(totalTime)*100))
	}
	apiURL += drawCtx.BuildRequestURL()
	resp, err := httptool.GetWithParams(httptool.RequestInfo{URL: apiURL})
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		err = fmt.Errorf("Request Error with status code `%d`", resp.StatusCode)
		return "", err
	}
	defer resp.Body.Close()
	filePath := filepath.Join(betagovar.ImagePath, time.Now().Format(time.RFC3339)+"_"+userName+"_chtime.png")
	f, _ := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return "", err
	}
	defer os.Remove(filePath)
	fileURL, err := utility.UploadFileToCos(filePath)
	if err != nil {
		return "", err
	}
	return fileURL, err
}

// DrawPieChartWithLocal 本地获取频道的时间分布
//
//	@return {}
func DrawPieChartWithLocal(ctx context.Context, inputMap map[string]time.Duration, userName string) (string, error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	if len(inputMap) == 0 {
		return "", fmt.Errorf("No Data Found")
	}
	values := make([]chart.Value, 0)
	var totalTime time.Duration
	for _, v := range inputMap {
		totalTime += v
	}
	tmpSlice := make([]struct {
		k string
		v time.Duration
	}, 0)
	for k, v := range inputMap {
		tmpSlice = append(tmpSlice, struct {
			k string
			v time.Duration
		}{k, v})
	}
	sort.Slice(tmpSlice, func(i, j int) bool {
		return tmpSlice[i].v > tmpSlice[j].v
	})
	for _, item := range tmpSlice {
		timeConv, _ := time.ParseDuration(fmt.Sprintf("%.1fs", item.v.Seconds()))
		values = append(values,
			chart.Value{
				Style: chart.Style{
					FontSize:            30,
					TextHorizontalAlign: 2,
					TextVerticalAlign:   4,
					TextWrap:            1,
					TextLineSpacing:     1,
					TextRotationDegrees: 0,
					FontColor:           chart.ColorBlack,
				},
				Label: item.k + " " + timeConv.String(),
				Value: float64(float64(item.v) / float64(totalTime) * 100),
			})
	}

	// TODO: 绘制频道时间饼状图
	pie := chart.BarChart{
		Title:  userName + "的频道时间分布",
		Width:  512,
		Height: 512,
		Font:   utility.MicrosoftYaHei,
		Bars:   values,
	}

	fileName := time.Now().Format(time.RFC3339) + "_" + userName + "_chtime.png"
	filePath := filepath.Join(betagovar.ImagePath, fileName)
	f, _ := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666)
	defer f.Close()
	err := pie.Render(chart.PNG, f)

	fileURL, err := utility.UploadFileToCos(filePath)
	if err != nil {
		return "", err
	}
	return fileURL, err
}
