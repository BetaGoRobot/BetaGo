package templates

import handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"

type CardBaseVars struct {
	RefreshTime     string      `json:"refresh_time"`
	JaegerTraceInfo string      `json:"jaeger_trace_info"`
	JaegerTraceURL  string      `json:"jaeger_trace_url"`
	WithdrawInfo    string      `json:"withdraw_info"`
	WithdrawTitle   string      `json:"withdraw_title"`
	WithdrawConfirm string      `json:"withdraw_confirm"`
	WithdrawObject  WithDrawObj `json:"withdraw_object"`

	RawCmd     *string     `json:"raw_cmd,omitempty"`
	RefreshObj *RefreshObj `json:"refresh_obj,omitempty"`
}

type RefreshObj struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type WithDrawObj struct {
	Type string `json:"type"`
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
type MainTopicOrActivity struct {
	MainTopicOrActivity string `json:"main_topic_or_activity,omitempty"`
}

type KeyConceptAndNoun struct {
	KeyConceptAndNoun string `json:"key_concept_and_noun,omitempty"`
}
type MentionedGroupOrOrganization struct {
	MentionedGroupOrOrganization string `json:"mentioned_group_or_organization,omitempty"`
}

type MentionedPeopleUnit struct {
	MentionedPeople string `json:"mentioned_people,omitempty"`
}

type LocationAndVenue struct {
	LocationAndVenue string `json:"locations_and_venue,omitempty"`
}

type MediaAndWork struct {
	Title string `json:"title,omitempty"`
	Type  string `json:"type,omitempty"`
}
type ChunkMetaData struct {
	Summary string `json:"summary"`

	Intent       string  `json:"intent"`
	Participants []*User `json:"participants,omitempty"`

	Sentiment string       `json:"sentiment"`
	Tones     []*ToneData  `json:"tones,omitempty"`
	Questions []*Questions `json:"questions,omitempty"`

	MsgList            []*MsgLine                         `json:"msg_list,omitempty"`
	PlansAndSuggestion []*handlertypes.PlansAndSuggestion `json:"plans_and_suggestions,omitempty"`

	MainTopicsOrActivities         []*ObjTextArray `json:"main_topics_or_activities,omitempty"`
	KeyConceptsAndNouns            []*ObjTextArray `json:"key_concepts_and_nouns,omitempty"`
	MentionedGroupsOrOrganizations []*ObjTextArray `json:"mentioned_groups_or_organizations,omitempty"`
	MentionedPeople                []*ObjTextArray `json:"mentioned_people,omitempty"`
	LocationsAndVenues             []*ObjTextArray `json:"locations_and_venues,omitempty"`
	MediaAndWorks                  []*MediaAndWork `json:"media_and_works,omitempty"`

	Timestamp string `json:"timestamp"`
	MsgID     string `json:"msg_id"`

	*CardBaseVars
}

type ObjTextArray struct {
	Text string `json:"text,omitempty"`
}

func ToObjTextArray(s string) *ObjTextArray {
	return &ObjTextArray{s}
}

type (
	// 对于wc的卡片，主要涉及几个信息
	WordCountCardVars struct {
		// 1. 用户排行榜、消息/互动频率
		UserList []*UserListItem `json:"user_list"`
		// 2. 词云
		WordCloud any `json:"word_cloud"`
		TimeStamp string
	}
	UserListItem struct {
		Number    int         `json:"number"`
		User      []*UserUnit `json:"user"`
		MsgCnt    int         `json:"msg_cnt"`
		ActionCnt int         `json:"action_cnt"`
	}
	UserUnit struct {
		ID string `json:"id"` // OpenID
	}
)
