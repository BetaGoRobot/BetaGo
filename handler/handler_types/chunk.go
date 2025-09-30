package handlertypes

type MessageChunkLogV3 struct {
	ID               string `json:"id"`
	Summary          string `json:"summary"`
	Intent           string `json:"intent"`
	SentimentAndTone struct {
		Sentiment string   `json:"sentiment"`
		Tones     []string `json:"tones"`
	} `json:"sentiment_and_tone"`
	Entities            *Entities            `json:"entities"`
	InteractionAnalysis *InteractionAnalysis `json:"interaction_analysis"`
	Outcomes            *Outcome             `json:"outcomes"`

	ConversationEmbedding []float32 `json:"conversation_embedding"`

	MsgList   []string `json:"msg_list"`
	UserIDs   []string `json:"user_ids"`
	GroupID   string   `json:"group_id"`
	Timestamp string   `json:"timestamp"`
	MsgIDs    []string `json:"msg_ids"`
}

type PlansAndSuggestion struct {
	ActivityOrSuggestion string  `json:"activity_or_suggestion"`
	Proposer             *User   `json:"proposer"`
	ParticipantsInvolved []*User `json:"participants_involved"`
	Timing               *Timing `json:"timing"`
}

type Participant struct {
	*User
	MessageCount int `json:"message_count"`
}

type User struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type Outcome struct {
	ConclusionsOrAgreements    []string              `json:"conclusions_or_agreements"`
	PlansAndSuggestions        []*PlansAndSuggestion `json:"plans_and_suggestions"`
	OpenThreadsOrPendingPoints []string              `json:"open_threads_or_pending_points"`
}
type Timing struct {
	RawText        string `json:"raw_text,omitempty"`
	NormalizedDate string `json:"normalized_date,omitempty"`
}

type Entities struct {
	MainTopicsOrActivities         []string        `json:"main_topics_or_activities"`
	KeyConceptsAndNouns            []string        `json:"key_concepts_and_nouns"`
	MentionedGroupsOrOrganizations []string        `json:"mentioned_groups_or_organizations"`
	MentionedPeople                []string        `json:"mentioned_people"`
	LocationsAndVenues             []string        `json:"locations_and_venues"`
	MediaAndWorks                  []*MediaAndWork `json:"media_and_works"`
	Resources                      []any           `json:"resources"`
}

type MediaAndWork struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

type InteractionAnalysis struct {
	Participants        []*Participant `json:"participants"`
	ConversationFlow    string         `json:"conversation_flow"`
	SocialDynamics      []string       `json:"social_dynamics"`
	IsQuestionPresent   bool           `json:"is_question_present"`
	UnresolvedQuestions []string       `json:"unresolved_questions"`
}
