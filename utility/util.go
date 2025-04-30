package utility

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/bytedance/sonic"
	"github.com/golang/freetype/truetype"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
	"github.com/lonelyevil/kook/log_adapter/plog"
	p_log "github.com/phuslu/log"

	"go.opentelemetry.io/otel/trace"
)

// MicrosoftYaHei  字体类型,未来荧黑
var MicrosoftYaHei *truetype.Font

func init() {
	InitGlowSansSCFontType()
}

func GenerateTraceButton(traceInfo string) (b kook.CardMessageElementButton) {
	return kook.CardMessageElementButton{
		Theme: kook.CardThemeInfo,
		Value: "https://jaeger.kevinmatt.top/trace/" + traceInfo,
		Click: "link",
		Text:  url.PathEscape("TraceID:" + traceInfo),
	}
}

func GenerateTraceButtonSection(traceInfo string) kook.CardMessageSection {
	return kook.CardMessageSection{
		Mode: kook.CardMessageSectionModeRight,
		Text: &kook.CardMessageElementKMarkdown{
			Content: "TraceID: `" + traceInfo + "`",
		},
		Accessory: kook.CardMessageElementButton{
			Theme: kook.CardThemeInfo,
			Value: "https://jaeger.kevinmatt.top/trace/" + traceInfo,
			Click: "link",
			Text:  "链路追踪",
		},
	}
}

// InitGlowSansSCFontType 初始化字体类型
func InitGlowSansSCFontType() {
	fontFile := filepath.Join(consts.FontPath, "Microsoft Yahei.ttf")
	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		log.Zlog.Info("errot init font", zaplog.Error(err))
		return
	}
	MicrosoftYaHei, err = truetype.Parse(fontBytes)
	if err != nil {
		log.Zlog.Info("errot init font", zaplog.Error(err))
		return
	}
}

// GetOutBoundIP 获取机器人部署的当前ip
func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "101.132.154.52:80")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

// GetCurrentTime 获取当前时间
func GetCurrentTime() (localTime string) {
	timeLocal, _ := time.LoadLocation("Asia/Shanghai")
	time.Local = timeLocal
	localTime = time.Now().Local().Format("2006-01-02 15:04:05")
	return
}

// ForDebug 用于测试
func ForDebug(test ...any) {
	return
}

// IsInSlice 判断机器人是否被at到
//
//	@param target
//	@param slice
//	@return bool
func IsInSlice(target string, slice []string) bool {
	for i := range slice {
		if slice[i] == target {
			return true
		}
	}
	return false
}

// MustAtoI 将字符串转换为int
//
//	@param str
//	@return int
func MustAtoI(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return i
	}
	return i
}

// GetUserInfo 获取用户信息
//
//	@param userID
//	@param guildID
//	@return userInfo
func GetUserInfo(userID, guildID string) (userInfo *kook.User, err error) {
	if guildID != "" {
		userInfo, err = consts.GlobalSession.UserView(userID, kook.UserViewWithGuildID(guildID))
	} else {
		userInfo, err = consts.GlobalSession.UserView(userID)
	}
	if err != nil {
		return
	}
	return
}

// GetGuildIDFromChannelID 通过ChannelID获取GuildID
//
//	@param channelID
//	@return GuildID
func GetGuildIDFromChannelID(channelID string) (GuildID string) {
	c, err := consts.GlobalSession.ChannelView(channelID)
	if err != nil {
		log.Zlog.Error("Error getting guild", zaplog.Error(err))
	}
	return c.GuildID
}

// GetGuildInfo 获取公会信息
//
//	@param guildID
//	@return guildInfo
//	@return err
func GetGuildInfo(guildID string) (guildInfo *kook.Guild, err error) {
	guildInfo, err = consts.GlobalSession.GuildView(guildID)
	if err != nil {
		return
	}
	return
}

// GetChannnelInfo  获取频道信息
//
//	@param channelID
//	@return channelInfo
//	@return err
func GetChannnelInfo(channelID string) (channelInfo *kook.Channel, err error) {
	channelInfo, err = consts.GlobalSession.ChannelView(channelID)
	if err != nil {
		return
	}
	return
}

// Struct2Map  将结构体转换为map
//
//	@param obj
//	@return map
func Struct2Map(obj any) map[string]any {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	data := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		if timestamp, ok := v.Field(i).Interface().(kook.MilliTimeStamp); ok {
			data[t.Field(i).Name] = time.Unix(int64(timestamp)/1000, 0).Format("2006-01-02 15:04:05")
		} else {
			data[t.Field(i).Name] = v.Field(i).Interface()
		}
	}
	return data
}

// BuildCardMessageCols 创建卡片消息的列
func BuildCardMessageCols(titleK, titleV string, kvMap map[string]any) (res []any, err error) {
	sectionElements := []any{
		kook.CardMessageSection{
			Text: kook.CardMessageParagraph{
				Cols: 2,
				Fields: []any{
					kook.CardMessageElementKMarkdown{
						Content: titleK,
					},
					kook.CardMessageElementKMarkdown{
						Content: titleV,
					},
				},
			},
		},
	}
	for k, v := range kvMap {
		if fmt.Sprint(v) == "" {
			continue
		}
		if strings.HasPrefix(fmt.Sprint(v), "http") {
			sectionElements = append(sectionElements,
				kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeRight,
					Text: kook.CardMessageParagraph{
						Cols: 2,
						Fields: []any{
							kook.CardMessageElementKMarkdown{
								Content: "**" + k + "**",
							},
						},
					},
					Accessory: kook.CardMessageElementImage{
						Src:    fmt.Sprint(v),
						Size:   "sm",
						Circle: true,
					},
				},
			)
		} else {
			sectionElements = append(sectionElements,
				kook.CardMessageSection{
					Text: kook.CardMessageParagraph{
						Cols: 2,
						Fields: []any{
							kook.CardMessageElementKMarkdown{
								Content: "**" + k + "**",
							},
							kook.CardMessageElementKMarkdown{
								Content: "`" + fmt.Sprintf("%#v", v) + "`",
							},
						},
					},
				},
			)
		}

	}
	res = sectionElements
	return
}

// Reconnect 重建链接
func Reconnect() (err error) {
	err = consts.GlobalSession.Close()
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	consts.GlobalSession = kook.New(os.Getenv("BOTAPI"), plog.NewLogger(&p_log.Logger{
		Level:  p_log.InfoLevel,
		Writer: &p_log.ConsoleWriter{},
	}))

	err = consts.GlobalSession.Open()
	// retryCnt := 0
	// for err != nil {
	// 	time.Sleep(100 * time.Millisecond)
	// 	betagovar.GlobalSession.Close()
	// 	betagovar.GlobalSession = kook.New(os.Getenv("BOTAPI"), plog.NewLogger(&log.Logger{
	// 		Level:  log.InfoLevel,
	// 		Writer: &log.ConsoleWriter{},
	// 	}))
	// 	err = betagovar.GlobalSession.Open()
	// 	if err != nil {
	// 		gotify.SendMessage("", "Reconnect failed, error is "+err.Error(), 7)
	// 	}
	// 	if retryCnt++; retryCnt == 5 {
	// 		return fmt.Errorf("reconnect to kook server reaches max retry cnt 5, need restart or try again" + err.Error())
	// 	}
	// }
	log.Zlog.Info("Reconnecting successfully")
	time.Sleep(time.Second * 5)
	return
}

// GetFuncFromInstance 1
//
//	@return string
func GetFuncFromInstance(ctxFunc any) string {
	r := strings.Split(runtime.FuncForPC(reflect.ValueOf(ctxFunc).Pointer()).Name(), "/")
	return r[len(r)-1]
}

var pcCache = &sync.Map{}

// GetCurrentFunc 1
//
//	@return string
func GetCurrentFunc() string {
	pc, _, _, _ := runtime.Caller(1)
	if name, ok := pcCache.Load(pc); ok {
		return name.(string)
	}

	name, _ := pcCache.LoadOrStore(pc, runtime.FuncForPC(pc).Name())
	return name.(string)
}

// BuildCardMessage 1
//
//	@return string
func BuildCardMessage(theme, size, title, quoteID string, span any, modules ...any) (cardMessageStr string, err error) {
	var inputCommand any
	cardMessageCard := &kook.CardMessageCard{
		Theme:   kook.CardTheme(theme),
		Size:    "lg",
		Modules: make([]any, 0),
	}
	cardMessage := kook.CardMessage{cardMessageCard}
	if quoteID != "" {
		m, err := consts.GlobalSession.MessageView(quoteID)
		if err != nil {
			log.Zlog.Error("MessageView Error", zaplog.Error(err))
		}
		prevCardMessage := make(kook.CardMessage, 0)
		err = json.UnmarshalFromString(m.Content, &prevCardMessage)
		if err != nil {
			// 不是卡片消息
			if len(m.MentionInfo.MentionPart) > 0 {
				m.Content = "@" + m.MentionInfo.MentionPart[0].Username + m.Content[strings.LastIndex(m.Content, "(met)")+5:]
			}
			m.Content = "`" + m.Content + "`"
			inputCommand = kook.CardMessageSection{
				Mode: kook.CardMessageSectionModeLeft,
				Text: kook.CardMessageElementKMarkdown{
					Content: m.Content,
				},
			}
		} else {
			inputCommand = prevCardMessage[0].Modules
		}
	}
	var titleModule any
	if title != "" {
		titleModule = &kook.CardMessageHeader{
			Text: kook.CardMessageElementText{
				Content: title,
				Emoji:   false,
			},
		}
	}

	var traceModule any
	if span != nil {
		if spanTrace, ok := span.(trace.Span); ok {
			traceModule = GenerateTraceButtonSection(spanTrace.SpanContext().TraceID().String())
		} else if spanStr, ok := span.(string); ok {
			traceModule = GenerateTraceButtonSection(spanStr)
		}
	}
	resModules := make([]any, 0)
	if inputCommand != nil {
		resModules = append(resModules,
			kook.CardMessageSection{
				Text: kook.CardMessageElementText{
					Content: "你的输入：",
				},
			})
		if _, ok := inputCommand.([]any); ok {
			resModules = append(resModules, inputCommand.([]any)...)
		} else {
			resModules = append(resModules, inputCommand)
		}
	}
	resModules = append(resModules, kook.CardMessageDivider{})
	if titleModule != nil {
		resModules = append(resModules, titleModule, kook.CardMessageDivider{})
	}
	resModules = append(resModules, modules...)
	if traceModule != nil {
		resModules = append(resModules, kook.CardMessageDivider{}, traceModule)
	}
	cardMessageCard.Modules = resModules

	return cardMessage.BuildMessage()
}

func RemovePostStyle(content string) (res string) {
	m := make(map[string]any)
	err := sonic.UnmarshalString(content, &m)
	if err != nil {
		return content
	}
	s, err := sonic.MarshalString(map[string]any{"zh_cn": removeStyleKey(m)})
	if err != nil {
		return content
	}
	return s
}

// removeStyleKey 递归地从 map 或 slice 中删除 key 为 "style" 的键
func removeStyleKey(data any) any {
	switch v := data.(type) {
	case map[string]any:
		for key, val := range v {
			if key == "style" {
				delete(v, key)
			} else {
				v[key] = removeStyleKey(val)
			}
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = removeStyleKey(item)
		}
		return v
	default:
		return data
	}
}

func UnmarshallString[T any](s string) (*T, error) {
	t := new(T)
	err := sonic.UnmarshalString(s, &t)
	if err != nil {
		return t, err
	}
	return t, nil
}

func MustUnmarshallString[T any](s string) *T {
	t := new(T)
	err := sonic.UnmarshalString(s, &t)
	if err != nil {
		panic(err)
	}
	return t
}

func UnmarshallStringPre[T any](s string, val *T) error {
	err := sonic.UnmarshalString(s, &val)
	if err != nil {
		return err
	}
	return nil
}
