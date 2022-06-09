package dbpack

import (
	"fmt"
	"log"
	"strconv"

	"gorm.io/gorm"
)

type khlNetease struct {
	gorm.Model
	KaiheilaID      string `gorm:"primaryKey;autoIncrement:false"`
	NetEaseID       string `gorm:"primaryKey;autoIncrement:false"`
	NetEasePhone    string
	NetEasePassword string
}

type khlMusicDownload struct {
	gorm.Model
	SongID   string `gorm:"primaryKey;autoIncrement:false"`
	Filepath string `gorm:"primaryKey;autoIncrement:false"`
}

func (music *khlMusicDownload) DownloadMusicDB() {
	db := GetDbConnection()
	err := db.AutoMigrate(&khlMusicDownload{})
	if err != nil {
		log.Println(err.Error())
	}
}

// CheckIsAdmin 检查是否是管理员
//  @param userID
//  @return isAdmin
func CheckIsAdmin(userID string) (isAdmin bool) {
	db := GetDbConnection()
	userIDInt, _ := strconv.Atoi(userID)
	admin := Administrator{
		UserID: int64(userIDInt),
	}
	res := db.Table("betago.administrators").Where("user_id = ?", userIDInt).Find(&admin)
	if res.RowsAffected == 0 {
		return false
	}
	return true
}

// GetAdminLevel 获取管理员等级
//  @param userID
//  @return level
func GetAdminLevel(userID string) int {
	db := GetDbConnection()
	userIDInt, _ := strconv.Atoi(userID)
	admin := Administrator{
		UserID: int64(userIDInt),
	}
	res := db.Table("betago.administrators").First(&admin, "user_id = ?", userID)
	if res.RowsAffected == 0 {
		// 不存在该管理员，返回预设值-1
		return -1
	}
	return int(admin.Level)
}

// GetCommandInfo 获取命令信息
//  @param command
//  @return info
func GetCommandInfo(command string) (commandInfoList []*CommandInfo, err error) {
	db := GetDbConnection()
	command = "`" + command + "`"
	commandInfoList = []*CommandInfo{
		{
			CommandName: command,
		},
	}
	res := db.Table("betago.command_infos").Where("command_name = ?", command).Find(&commandInfoList)
	if res.RowsAffected == 0 {
		err = fmt.Errorf("command %s not found", command)
		return
	}
	return
}

// GetCommandInfoWithOpt 获取命令信息
//  @param option
//  @return info
func GetCommandInfoWithOpt(optionf string) (commandInfoList []*CommandInfo, err error) {
	if GetDbConnection().Table("betago.command_infos").Where(optionf).Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("option %s  not found", optionf)
		return
	}
	return
}
