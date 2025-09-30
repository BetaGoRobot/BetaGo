package chunking

var translations = map[string]string{
	// 核心意图 (Intent)
	"SOCIAL_COORDINATION":          "社交协调",
	"INFORMATION_SHARING":          "信息分享",
	"SEEKING_HELP_OR_ADVICE":       "寻求帮助",
	"DEBATE_OR_DISCUSSION":         "辩论讨论",
	"EMOTIONAL_SHARING_OR_SUPPORT": "情感分享",
	"REQUESTING_RECOMMENDATION":    "请求推荐",
	"CASUAL_CHITCHAT":              "日常闲聊",

	// 情感倾向 (Sentiment)
	"POSITIVE": "积极",
	"NEGATIVE": "消极",
	"NEUTRAL":  "中性",
	"MIXED":    "复杂/混合",

	// 语气 (Tones)
	"HUMOROUS":      "幽默",
	"SUPPORTIVE":    "支持",
	"CURIOUS":       "好奇",
	"EXCITED":       "兴奋",
	"URGENT":        "紧急",
	"FORMAL":        "正式",
	"INFORMAL":      "非正式",
	"SARCASTIC":     "讽刺",
	"ARGUMENTATIVE": "争论",
	"NOSTALGIC":     "怀旧",

	// 文艺作品类型 (Media Type)
	"MOVIE":   "电影",
	"BOOK":    "书籍",
	"MUSIC":   "音乐",
	"GAME":    "游戏",
	"TV_SHOW": "电视节目",

	// 外部资源类型 (Resource Type)
	"URL":   "链接",
	"FILE":  "文件",
	"IMAGE": "图片",
	"VIDEO": "视频",

	// 对话流模式 (Conversation Flow)
	"Q&A":             "问答模式",
	"NARRATIVE":       "叙事模式",
	"BRAINSTORMING":   "头脑风暴",
	"DEBATE":          "辩论模式",
	"TOPIC_SWITCHING": "话题切换",

	// 社交动态 (Social Dynamics)
	"JOKING":           "开玩笑",
	"AGREEMENT":        "达成一致",
	"DISAGREEMENT":     "存在分歧",
	"OFFERING_SUPPORT": "提供支持",
	"PERSUASION":       "说服/劝说",
}

// Translate 将预定义的英文 key 翻译成中文。
// 如果找不到对应的翻译，将原样返回 key。
func Translate(key string) string {
	if val, ok := translations[key]; ok {
		return val
	}
	return key // Fallback to the original key if no translation is found
}
