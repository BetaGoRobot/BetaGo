package betagovar

import (
	"fmt"
	"os"
	"sync"

	"github.com/lonelyevil/kook"
	"github.com/lonelyevil/kook/log_adapter/plog"
	"github.com/phuslu/log"
)

var (
	// NotifierChanID 发送消息的频道ID
	NotifierChanID = "8583973157097178"
)

const (
	TimeFormat = "2006-01-02T15:04:05Z07:00"
	// BetaGoUpdateChanID  发送更新消息的频道ID
	BetaGoUpdateChanID = "8937461610423450"

	// ImagePath 图片存储路径
	ImagePath = "/data/images"

	// FontPath  字体存储路径
	FontPath = "/data/fonts"

	// PublicIPURL 获取公网IP的URL
	PublicIPURL = "http://ifconfig.me"
	// DBHostCompose DockerCompose的PGHost
	DBHostCompose = "host=betago-pg"
	// DBHostCluster k8s的PGHost
	DBHostCluster = "host=kubernetes.default"
	// DBHostTest 本地测试的PGHost
	DBHostTest = "host=localhost"
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

// GlobalSession 全局共享session
var GlobalSession = kook.New(os.Getenv("BOTAPI"), plog.NewLogger(&log.Logger{
	Level:  log.DebugLevel,
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
)

func init() {
	if RobotName = os.Getenv("ROBOT_NAME"); RobotName == "" {
		RobotName = "No RobotName Configured"
	}
	if IsTest {
		NotifierChanID = "7419593543056418"
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

func (fc *FlowControlType) Top() (err error) {
	fc.M.RLock()
	if fc.Cnt >= 5 {
		return fmt.Errorf("每秒仅允许5次请求，请求过快，请稍后重试")
	}
	fc.M.RUnlock()
	return
}
