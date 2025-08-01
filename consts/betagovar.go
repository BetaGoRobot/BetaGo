package consts

import (
	"fmt"
	"os"
	"sync"

	"github.com/lonelyevil/kook"
	"github.com/lonelyevil/kook/log_adapter/plog"
	"github.com/phuslu/log"
)

// NotifierChanID 发送消息的频道ID
var (
	NotifierChanID = "2472013302648680"
	BotIdentifier  = os.Getenv("BOT_IDENTIFY")
)

// Database
const (
	DBHostCompose = "host=betago_pg"
	// DBHostCluster k8s的PGHost
	DBHostCluster = "host=kubernetes.default"
	// DBHostTest 本地测试的PGHost
	DBHostTest = "host=localhost"
)

// netease

const (
	NetEaseAPIDomainCompose = "http://n"
)

const (
	TimeFormat = "2006-01-02T15:04:05"
	// BetaGoUpdateChanID  发送更新消息的频道ID
	BetaGoUpdateChanID = "6422768213722929"

	// ImagePath 图片存储路径
	ImagePath = "/data/images"

	// ChatPath 对话信息存储路径
	ChatPath = "/data/chat/chat.dump"

	// FontPath  字体存储路径
	FontPath = "/data/fonts"

	// PublicIPURL 获取公网IP的URL
	PublicIPURL = "http://ifconfig.me"
	// DBHostCompose DockerCompose的PGHost

	// SelfCheckMessage 自我健康检查
	SelfCheckMessage = "self check message"
)

// CardMessageModule khl cardmessage
type CardMessageModule struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Src   string `json:"src"`
	Cover string `json:"cover"`
}

// CardMessageTextModule khl cardmessage Text
type CardMessageTextModule struct {
	Type string `json:"type"`
	Text struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"text"`
}

// CardMessageColModule  khl cardmessage Col
type CardMessageColModule struct {
	Type string `json:"type"`
	Text []struct {
		Type   string `json:"type"`
		Cols   int    `json:"cols"`
		Fields []struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
	}
}

// SelfCheckChan 自我检查通道
var SelfCheckChan = make(chan string, 1)

// ReconnectChan 1
var ReconnectChan = make(chan string)

// GlobalSession 全局共享session
var GlobalSession = kook.New(os.Getenv("BOTAPI"), plog.NewLogger(&log.Logger{
	Level:  log.InfoLevel,
	Writer: &log.ConsoleWriter{},
}))

// 环境变量
var (
	CommitMessage = os.Getenv("COM_MES")
	HTMLURL       = os.Getenv("HTML_URL")
	CommentsURL   = os.Getenv("COM_URL")
	RobotName     = os.Getenv("ROBOT_NAME")
	RobotID       = os.Getenv("ROBOT_ID")
	TestChanID    = os.Getenv("TEST_CHAN_ID")
	BetaGoTest    = os.Getenv("IS_TEST") == "true"
	IsTest        = os.Getenv("IS_TEST") == "true"
	IsCluster     = os.Getenv("IS_CLUSTER") == "true"
	IsCompose     = os.Getenv("IS_COMPOSE") == "true"
	CommandPrefix = "(met)" + RobotID + "(met)"
)

func init() {
	if RobotName = os.Getenv("ROBOT_NAME"); RobotName == "" {
		RobotName = "No RobotName Configured"
	}
	if IsTest {
		NotifierChanID = "4988093461275944"
	}
}

type FlowControlType struct {
	M   sync.RWMutex
	Cnt int
}

var FlowControl = &FlowControlType{}

func (fc *FlowControlType) Add() {
	fc.M.Lock()
	fc.Cnt++
	fc.M.Unlock()
}

func (fc *FlowControlType) Sub() {
	fc.M.Lock()
	fc.Cnt--
	fc.M.Unlock()
}

var ErrorOverReq = fmt.Errorf("每秒仅允许5次请求，请求过快，请稍后重试")

func (fc *FlowControlType) Top() (err error) {
	fc.M.RLock()
	if fc.Cnt >= 5 {
		return ErrorOverReq
	}
	fc.M.RUnlock()
	return
}

const (
	LarkMsgIndex        = "lark_msg_index_jieba"
	LarkCardActionIndex = "lark_card_action_index"
)
