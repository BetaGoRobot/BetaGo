package database

import (
	"fmt"
	"os"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// LarkImg is
type LarkImg struct {
	gorm.Model
	SongID string `json:"song_id" gorm:"primaryKey;autoIncrement:false"`
	ImgKey string `json:"img_key" gorm:"primaryKey;autoIncrement:false"`
}

// ChatContextRecord is
type ChatContextRecord struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// Administrator is the struct of administrator
type Administrator struct {
	gorm.Model
	UserID   int64  `json:"user_id" gorm:"primaryKey"`
	UserName string `json:"user_name"`
	Level    int64  `json:"level"`
}

// CommandInfo is the struct of command info
type CommandInfo struct {
	CommandName     string    `json:"command_name" gorm:"primaryKey;autoIncrement:false"`
	CommandDesc     string    `json:"command_desc"`
	CommandParamLen int       `json:"command_param_len"`
	CommandType     string    `json:"command_type"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DynamicConfig is the struct of dynamic command info
type DynamicConfig struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value" gorm:"primaryKey"`
}

type RepeatWhitelist struct {
	GuildID string `json:"guild_id" gorm:"primaryKey"`
}

type RepeatWordsRate struct {
	Word string `json:"word" gorm:"primaryKey"`
	Rate int    `json:"rate"`
}

// ChannelLogExt  is the struct of channel log
type ChannelLogExt struct {
	UserID      string `json:"user_id" gorm:"primaryKey"`
	UserName    string `json:"user_name"`
	ChannelID   string `json:"channel_id" gorm:"primaryKey"`
	ChannelName string `json:"channel_name"`
	JoinedTime  string `json:"joined_time" gorm:"primaryKey"`
	LeftTime    string `json:"left_time"`
	ISUpdate    bool   `json:"is_update"`
	MsgID       string `json:"msg_id"`
	GuildID     string `json:"guild_id" `
}

// AlertList  is the struct of alert config
type AlertList struct {
	EmailAddress string
}

// ChatRecordLog 存储chat的对话记录
type ChatRecordLog struct {
	AuthorID  string `json:"user_id" gorm:"primaryKey"`
	RecordStr string `json:"record_str"`
}

var (
	isTest    = os.Getenv("IS_TEST")
	isCluster = os.Getenv("IS_CLUSTER")
)

func init() {
	// try get db conn
	if GetDbConnection() == nil {
		log.ZapLogger.Error("get db connection error")
		os.Exit(-1)
	}

	// migrate
	db := GetDbConnection()
	err := db.AutoMigrate(
		&Administrator{},
		&CommandInfo{},
		&ChannelLogExt{},
		&AlertList{},
		&ChatContextRecord{},
		&ChatRecordLog{},
		&DynamicConfig{},
		&LarkImg{},
		// &RepeatWhitelist{},
		// &RepeatWordsRate{},
	)
	if err != nil {
		log.ZapLogger.Error("init", zaplog.Error(err))
	}
	utility.GetReceieverEmailList(consts.GlobalDBConn)
	sqlDb, err := db.DB()
	if err != nil {
		log.ZapLogger.Panic(" get sql db error")
	}
	sqlDb.SetMaxIdleConns(2)
	sqlDb.SetMaxOpenConns(5)
	sqlDb.SetConnMaxLifetime(time.Minute * 10)

	// // 启动时，清空所有的超时Channel log
	// db.Model(&ChannelLogExt{}).Where("left_time < joined_time").Delete(&ChannelLogExt{})

	// // 标记所有update为true
	// db.Model(&ChannelLogExt{}).Where("is_update = false").Update("is_update", true).Debug()
}

// GetDbConnection  returns the db connection
//
//	@return *gorm.DB
func GetDbConnection() *gorm.DB {
	if consts.GlobalDBConn != nil {
		return consts.GlobalDBConn
	}
	var dsn string = " user=postgres password=heyuheng1.22.3 dbname=betago port=%s sslmode=disable TimeZone=Asia/Shanghai application_name=" + consts.RobotName
	if consts.IsTest {
		dsn = consts.DBHostTest + fmt.Sprintf(dsn, "3033")
	} else if consts.IsCluster {
		dsn = consts.DBHostCluster + fmt.Sprintf(dsn, "5432")
	} else {
		dsn = consts.DBHostCompose + fmt.Sprintf(dsn, "5432")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "betago.",
			SingularTable: false,
		},
	})
	if err != nil {
		log.ZapLogger.Error("get db connection error, will try local version", zaplog.Error(err))
		return nil
	}
	consts.GlobalDBConn = db
	return consts.GlobalDBConn
}
