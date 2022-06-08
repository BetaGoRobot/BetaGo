package betagovar

import (
	"os"

	"github.com/lonelyevil/khl"
	"github.com/lonelyevil/khl/log_adapter/plog"
	"github.com/phuslu/log"
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
