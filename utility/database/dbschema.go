package database

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var dbOnce = &sync.Once{}

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
type Genaral struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// DynamicConfig is the struct of dynamic command info
type DynamicConfig struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value" gorm:"primaryKey"`
}

type RepeatWhitelist struct {
	GuildID string `json:"guild_id" gorm:"primaryKey;autoIncrement:false"`
}

type ReactionWhitelist struct {
	GuildID string `json:"guild_id" gorm:"primaryKey;autoIncrement:false"`
}

type FunctionEnabling struct {
	GuildID  string                  `json:"guild_id" gorm:"primaryKey;autoIncrement:false"`
	Function consts.LarkFunctionEnum `json:"function" gorm:"primaryKey;autoIncrement:false"`
}

type QuoteReplyMsg struct {
	MatchType consts.WordMatchType `json:"match_type" gorm:"primaryKey;index;default:substr"`
	Keyword   string               `json:"keyword" gorm:"primaryKey;index"`
	ReplyNType
}

type QuoteReplyMsgCustom struct {
	GuildID   string               `json:"guil d_id" gorm:"primaryKey;index"`
	MatchType consts.WordMatchType `json:"match_type" gorm:"primaryKey;index;default:substr"`
	Keyword   string               `json:"keyword" gorm:"primaryKey;index"`
	ReplyNType
}

type ReplyNType struct {
	Reply     string           `json:"reply" gorm:"primaryKey;index"`
	ReplyType consts.ReplyType `json:"reply_type" gorm:"primaryKey;index;default:text"`
}
type ReactImageMeterial struct {
	GuildID string `json:"guild_id" gorm:"primaryKey;autoIncrement:false"`
	FileID  string `json:"file_id" gorm:"primaryKey;autoIncrement:false"`
	Type    string `json:"type"`
}

type CopyWritingGeneral struct {
	Endpoint string         `json:"endpoint" gorm:"primaryKey;autoIncrement:false"`
	Content  pq.StringArray `gorm:"type:text[]" json:"content"`
}

type CopyWritingCustom struct {
	Endpoint string         `json:"endpoint" gorm:"primaryKey;autoIncrement:false"`
	GuildID  string         `json:"guild_id" gorm:"primaryKey;autoIncrement:false"`
	Content  pq.StringArray `gorm:"type:text[]" json:"content"`
}

type RepeatWordsRate struct {
	Genaral
	Word string `json:"word" gorm:"primaryKey;autoIncrement:false"`
	Rate int    `json:"rate"`
}

type RepeatWordsRateCustom struct {
	Genaral
	GuildID string `json:"guild_id" gorm:"primaryKey;autoIncrement:false"`
	Word    string `json:"word" gorm:"primaryKey;autoIncrement:false"`
	Rate    int    `json:"rate"`
}

type StickerMapping struct {
	Genaral
	StickerKey string `json:"sticker_key" gorm:"primaryKey;autoIncrement:false;index"`
	ImageKey   string `json:"image_key" gorm:"index"`
}

type InteractionStats struct {
	OpenID     string                 `json:"open_id" gorm:"index"`
	GuildID    string                 `json:"guild_id" gorm:"index"`
	UserName   string                 `json:"user_name"`
	ActionType consts.LarkInteraction `json:"action_type" gorm:"index"`
	CreatedAt  time.Time
}

// MessageLog test
type MessageLog struct {
	MessageID   string `json:"message_id,omitempty" gorm:"primaryKey;index"` // 消息的open_message_id，说明参见：[消息ID说明](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/intro#ac79c1c2)
	RootID      string `json:"root_id,omitempty" gorm:"index"`               // 根消息id，用于回复消息场景，说明参见：[消息ID说明](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/intro#ac79c1c2)
	ParentID    string `json:"parent_id,omitempty" gorm:"index"`             // 父消息的id，用于回复消息场景，说明参见：[消息ID说明](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/intro#ac79c1c2)
	ChatID      string `json:"chat_id,omitempty" gorm:"index"`               // 消息所在的群组 ID
	ThreadID    string `json:"thread_id,omitempty" gorm:"index"`             // 消息所属的话题 ID
	ChatType    string `json:"chat_type,omitempty"`                          // 消息所在的群组类型;;**可选值有**：;- `p2p`：单聊;- `group`： 群组;- `topic_group`：话题群
	MessageType string `json:"message_type,omitempty"`                       // 消息类型

	UserAgent string `json:"user_agent,omitempty"` // 用户代理
	Mentions  string `json:"mentions"`
	RawBody   string `json:"raw_body"`
	Content   string `json:"message_str"`
	FileKey   string `json:"file_key"`
	CreatedAt time.Time
}

type MsgTraceLog struct {
	MsgID     string `json:"msg_id" gorm:"index"`
	TraceID   string `json:"trace_id"`
	CreatedAt time.Time
}

type TemplateVersion struct {
	TemplateID      string `json:"template_id" gorm:"primaryKey;autoIncrement:false"`
	TemplateVersion string `json:"template_version"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
		&ReactionWhitelist{},
		&RepeatWhitelist{},
		&RepeatWordsRate{},
		&QuoteReplyMsg{},
		&QuoteReplyMsgCustom{},
		&FunctionEnabling{},
		&RepeatWordsRateCustom{},
		&ReactImageMeterial{},
		&CopyWritingCustom{},
		&CopyWritingGeneral{},
		&MsgTraceLog{},
		&StickerMapping{},
		&InteractionStats{},
		&TemplateVersion{},
		&MessageLog{},
	)
	if err != nil {
		log.ZapLogger.Error("init", zaplog.Error(err))
	}
	utility.GetReceieverEmailList(consts.GlobalDBConn)
	sqlDb, err := db.DB()
	if err != nil {
		log.ZapLogger.Panic(" get sql db error")
	}
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)
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
	dbOnce.Do(
		func() {
			var dsn string = " user=postgres password=heyuheng1.22.3 dbname=betago port=%s sslmode=disable TimeZone=Asia/Shanghai application_name=" + consts.RobotName
			if consts.IsTest {
				dsn = consts.DBHostTest + fmt.Sprintf(dsn, "15432")
			} else if consts.IsCluster {
				dsn = consts.DBHostCluster + fmt.Sprintf(dsn, "5432")
			} else {
				dsn = consts.DBHostCompose + fmt.Sprintf(dsn, "5432")
			}
			var err error
			consts.GlobalDBConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   "betago.",
					SingularTable: false,
				},
			})
			if err != nil {
				log.ZapLogger.Error("get db connection error, will try local version", zaplog.Error(err))
				return
			}
		},
	)
	return consts.GlobalDBConn
}
