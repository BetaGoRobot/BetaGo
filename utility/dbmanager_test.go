package utility

import (
	"testing"

	"gorm.io/gorm"
)

func TestA(t *testing.T) {
	khl := khlMusicDownload{}
	khl.DownloadMusicDB()
}

// func TestRegistAndBind(t *testing.T) {
// 	RegistAndBind(&khlNetease{KaiheilaID: "123", NetEaseID: "kevinmatt", NetEasePhone: "1111111", NetEasePassword: "adadas"})
// }

func Test_khlMusicDownload_DownloadMusicDB(t *testing.T) {
	type fields struct {
		Model    gorm.Model
		SongID   string
		Filepath string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			music := &khlMusicDownload{
				Model:    tt.fields.Model,
				SongID:   tt.fields.SongID,
				Filepath: tt.fields.Filepath,
			}
			music.DownloadMusicDB()
		})
	}
}

var adminCommandHelper = map[string]string{
	"`addAdmin`":    "添加管理员 \n`@BetaGo` `addAdmin <userID> <userName>`",
	"`removeAdmin`": "移除管理员 \n`@BetaGo` `removeAdmin <userID>`",
}

var userCommandHelper = map[string]string{
	"`searchMusic`": "搜索音乐 \n`@BetaGo` `searchMusic <musicName>`",
	"`getuser`":     "获取用户信息 \n`@BetaGo` `getuser <userID>`",
}

var userNoParamCommandHelper = map[string]string{
	"`help`":    "查看帮助 \n`@BetaGo` `help`",
	"`ping`":    "检查机器人是否运行正常 \n`@BetaGo` `ping`",
	"`roll`":    "掷骰子 \n`@BetaGo` `roll`",
	"`oneword`": "获取一言 \n`@BetaGo` `oneword`",
}

var adminNoParamCommandHelper = map[string]string{
	"`showAdmin`": "显示所有管理员 \n`@BetaGo` `showAdmin`",
}

func TestGetCommandInfo(t *testing.T) {
	for k, v := range adminNoParamCommandHelper {
		GetDbConnection().Table("betago.command_infos").Create(&CommandInfo{
			CommandName:     k,
			CommandDesc:     v,
			CommandParamLen: 0,
			CommandType:     "admin",
		})
	}
	for k, v := range userNoParamCommandHelper {
		GetDbConnection().Table("betago.command_infos").Create(&CommandInfo{
			CommandName:     k,
			CommandDesc:     v,
			CommandParamLen: 0,
			CommandType:     "user",
		})
	}
	for k, v := range adminCommandHelper {
		GetDbConnection().Table("betago.command_infos").Create(&CommandInfo{
			CommandName:     k,
			CommandDesc:     v,
			CommandParamLen: 2,
			CommandType:     "admin",
		})
	}
	for k, v := range userCommandHelper {
		GetDbConnection().Table("betago.command_infos").Create(&CommandInfo{
			CommandName:     k,
			CommandDesc:     v,
			CommandParamLen: 1,
			CommandType:     "user",
		})
	}

	GetCommandInfo("addAdmin")
}

func TestT(t *testing.T) {
	GetCommandInfoWithOpt("command_param_len = 0 and command_type= 'user'")
}
