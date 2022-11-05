package utility

import (
	"log"
	"os"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/heyuhengmatt/zaplog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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

// ChannelLog  is the struct of channel log
type ChannelLog struct {
	UserID      string    `json:"user_id" gorm:"primaryKey"`
	UserName    string    `json:"user_name"`
	ChannelID   string    `json:"channel_id"`
	ChannelName string    `json:"channel_name"`
	JoinedTime  time.Time `json:"joined_time"`
	LeftTime    time.Time `json:"left_time"`
	ISUpdate    bool      `json:"is_update"`
}

// AlertList  is the struct of alert config
type AlertList struct {
	EmailAddress string
}

var (
	isTest = os.Getenv("IS_TEST")

	// globalDBConn  is the global db connection
	globalDBConn *gorm.DB
)

func init() {
	InitLogger()
	// try get db conn
	if GetDbConnection() == nil {
		ZapLogger.Error("get db connection error")
		os.Exit(-1)
	}

	// migrate
	db := GetDbConnection()
	err := db.AutoMigrate(&Administrator{}, &CommandInfo{}, &ChannelLog{}, &AlertList{})
	if err != nil {
		ZapLogger.Error("init", zaplog.Error(err))
	}
	getReceieverEmailList()
	sqlDb, err := db.DB()
	if err != nil {
		log.Panicln(" get sql db error")
	}
	sqlDb.SetMaxIdleConns(2)
	sqlDb.SetMaxOpenConns(5)
	sqlDb.SetConnMaxLifetime(time.Minute * 10)

	// 启动时，清空所有的超时Channel log
	db.Model(&ChannelLog{}).Where("left_time < joined_time").Delete(&ChannelLog{})

	// 标记所有update为true
	db.Model(&ChannelLog{}).Where("is_update = false").Update("is_update", true).Debug()
}

// GetDbConnection  returns the db connection
//
//	@return *gorm.DB
func GetDbConnection() *gorm.DB {
	if globalDBConn != nil {
		return globalDBConn
	}
	var dsn string
	if isTest == "true" {
		dsn = "host=localhost user=postgres password=heyuheng1.22.3 dbname=betago port=5432 sslmode=disable TimeZone=Asia/Shanghai application_name=" + betagovar.RobotName
	} else {
		dsn = "host=betago-pg user=postgres password=heyuheng1.22.3 dbname=betago port=5432 sslmode=disable TimeZone=Asia/Shanghai application_name=" + betagovar.RobotName
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "betago.",
			SingularTable: false,
		},
	})
	if err != nil {
		ZapLogger.Error("get db connection error, will try local version", zaplog.Error(err))
		dsn = "host=192.168.31.32 user=postgres password=heyuheng1.22.3 dbname=betago port=5432 sslmode=disable TimeZone=Asia/Shanghai application_name=" + betagovar.RobotName
		return nil
	}
	globalDBConn = db
	return globalDBConn
}
