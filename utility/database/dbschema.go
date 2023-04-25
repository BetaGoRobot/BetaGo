package database

import (
	"log"
	"os"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/kevinmatthe/zaplog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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
		utility.ZapLogger.Error("get db connection error")
		os.Exit(-1)
	}

	// migrate
	db := GetDbConnection()
	err := db.AutoMigrate(&Administrator{}, &CommandInfo{}, &ChannelLogExt{}, &AlertList{}, &ChatContextRecord{}, &ChatRecordLog{}, &DynamicConfig{})
	if err != nil {
		utility.ZapLogger.Error("init", zaplog.Error(err))
	}
	utility.GetReceieverEmailList(betagovar.GlobalDBConn)
	sqlDb, err := db.DB()
	if err != nil {
		log.Panicln(" get sql db error")
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
	if betagovar.GlobalDBConn != nil {
		return betagovar.GlobalDBConn
	}
	var dsn string = " user=postgres password=heyuheng1.22.3 dbname=betago port=5432 sslmode=disable TimeZone=Asia/Shanghai application_name=" + betagovar.RobotName
	if betagovar.IsTest {
		dsn = betagovar.DBHostTest + dsn
	} else if betagovar.IsCluster {
		dsn = betagovar.DBHostCluster + dsn
	} else {
		dsn = betagovar.DBHostCompose + dsn
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "betago.",
			SingularTable: false,
		},
	})
	if err != nil {
		utility.ZapLogger.Error("get db connection error, will try local version", zaplog.Error(err))
		return nil
	}
	betagovar.GlobalDBConn = db
	return betagovar.GlobalDBConn
}
