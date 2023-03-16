package admin

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spyzhov/ajson"
)

type content struct {
	GuildID  string
	MsgID    string
	UserName string
	Msg      string
}

var backups = make([]*content, 0)

func backupData(UserName, Msg, MsgID, GuildID string) {
	backups = append(backups, &content{
		UserName: UserName,
		Msg:      Msg,
		MsgID:    MsgID,
		GuildID:  GuildID,
	})
	if len(backups) >= 100 {
		writeBackups(backups)
	}
}

func writeBackups(toWrite []*content) {
	// 写入备份文件中
	f, err := os.OpenFile("/msg-backups/"+time.Now().Format(time.RFC3339)+".json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	tmp, err := json.Marshal(&backups)
	if err != nil {
		panic(err)
	}
	f.Write(tmp)
	f.Write([]byte("\n"))
	backups = make([]*content, 0)
}

func cleaupData() {
	if len(backups) > 0 {
		// 写入备份文件中
		writeBackups(backups)
	}
}

func getStringFromNode(content string) (string, error) {
	res := make([]string, 0)
	nodes, err := ajson.JSONPath([]byte(content), "$..content")
	if err != nil {
		return content, err
	}
	for _, n := range nodes {
		res = append(res, strings.Trim(n.String(), "\""))
	}
	return strings.Join(res, "\n"), nil
}
