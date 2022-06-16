package betagovar

import (
	"os"

	"github.com/lonelyevil/khl"
	"github.com/lonelyevil/khl/log_adapter/plog"
	"github.com/phuslu/log"
)

const (
	// BetaGoUpdateChanID  发送更新消息的频道ID
	BetaGoUpdateChanID = "8937461610423450"

	// NotifierChanID 发送消息的频道ID
	NotifierChanID = "8583973157097178"
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
var GlobalSession = khl.New(os.Getenv("BOTAPI"), plog.NewLogger(&log.Logger{
	Level:  log.TraceLevel,
	Writer: &log.ConsoleWriter{},
}))

// var
//  @param CommitMessage
var (
	CommitMessage = os.Getenv("COM_MES")
	HTMLURL       = os.Getenv("HTML_URL")
	CommentsURL   = os.Getenv("COM_URL")
	RobotName     = os.Getenv("ROBOT_NAME")
	RobotID       = os.Getenv("ROBOT_ID")
	TestChanID    = os.Getenv("TEST_CHAN_ID")
)

func init() {
	if RobotName = os.Getenv("ROBOT_NAME"); RobotName == "" {
		RobotName = "No RobotName Configured"
	}
}
