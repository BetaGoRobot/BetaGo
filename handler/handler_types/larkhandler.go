package handlertypes

import (
	"time"

	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type MessageLog struct {
	MessageID   string `json:"message_id,omitempty" `  // 消息的open_message_id，说明参见：[消息ID说明](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/intro#ac79c1c2)
	RootID      string `json:"root_id,omitempty"`      // 根消息id，用于回复消息场景，说明参见：[消息ID说明](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/intro#ac79c1c2)
	ParentID    string `json:"parent_id,omitempty"`    // 父消息的id，用于回复消息场景，说明参见：[消息ID说明](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/intro#ac79c1c2)
	ChatID      string `json:"chat_id,omitempty"`      // 消息所在的群组 ID
	ThreadID    string `json:"thread_id,omitempty"`    // 消息所属的话题 ID
	ChatType    string `json:"chat_type,omitempty"`    // 消息所在的群组类型;;**可选值有**：;- `p2p`：单聊;- `group`： 群组;- `topic_group`：话题群
	MessageType string `json:"message_type,omitempty"` // 消息类型

	UserAgent string `json:"user_agent,omitempty"` // 用户代理
	Mentions  string `json:"mentions"`
	RawBody   string `json:"raw_body"`
	Content   string `json:"message_str"`
	FileKey   string `json:"file_key"`
	TraceID   string `json:"trace_id"`
	CreatedAt time.Time
}

type WordWithTag struct {
	Word string `json:"word"`
	Tag  string `json:"tag"`
}
type MessageIndex struct {
	*MessageLog
	ChatName             string         `json:"chat_name"`
	CreateTime           string         `json:"create_time"`
	CreateTimeV2         string         `json:"create_time_v2"`
	Message              []float32      `json:"message"`
	UserID               string         `json:"user_id"`
	UserName             string         `json:"user_name"`
	RawMessage           string         `json:"raw_message"`
	RawMessageJieba      string         `json:"raw_message_jieba"`
	RawMessageJiebaArray []string       `json:"raw_message_jieba_array"`
	RawMessageJiebaTag   []*WordWithTag `json:"raw_message_jieba_tag"`
	TokenUsage           model.Usage    `json:"token_usage"`
	IsCommand            bool           `json:"is_command"`
	MainCommand          string         `json:"main_command"`
}

type CardActionIndex struct {
	*callback.CardActionTriggerEvent
	ChatName    string         `json:"chat_name"`
	CreateTime  string         `json:"create_time"`
	UserID      string         `json:"user_id"`
	UserName    string         `json:"user_name"`
	ActionValue map[string]any `json:"action_value"`
}
