package templates

type JaegerBase struct {
	RefreshTime     string `json:"refresh_time"`
	JaegerTraceInfo string `json:"jaeger_trace_info"`
	JaegerTraceURL  string `json:"jaeger_trace_url"`
	WithdrawInfo    string `json:"withdraw_info"`
	WithdrawTitle   string `json:"withdraw_title"`
	WithdrawConfirm string `json:"withdraw_confirm"`
	WithdrawObject  struct {
		Type string `json:"type"`
	} `json:"withdraw_object"`
}

type User struct {
	ID string `json:"id"`
}

type ToneData struct {
	Tone string `json:"tone"`
}

type Questions struct {
	Question string `json:"question"`
}

type MsgLine struct {
	Time    string `json:"time"`
	User    *User  `json:"user,omitempty"`
	Content string `json:"content"`
}

type ChunkMetaData struct {
	Summary string `json:"summary"`

	Intent       string  `json:"intent"`
	Participants []*User `json:"participants,omitempty"`

	Sentiment string       `json:"sentiment"`
	Tones     []*ToneData  `json:"tones,omitempty"`
	Questions []*Questions `json:"questions,omitempty"`

	MsgList []*MsgLine `json:"msg_list,omitempty"`

	Timestamp string `json:"timestamp"`
	MsgID     string `json:"msg_id"`

	*JaegerBase
}

func (c *ChunkMetaData) WithJaegerBase(jaegerBase *JaegerBase) {
	c.JaegerBase = jaegerBase
}
