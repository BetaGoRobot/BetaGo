package consts

import "os"

type LarkFunctionEnum string

const (
	LarkFunctionWordReply    LarkFunctionEnum = "word_reply"
	LarkFunctionRandomReact  LarkFunctionEnum = "random_react"
	LarkFunctionRandomRepeat LarkFunctionEnum = "random_repeat"
)

const (
	BotOpenID = "ou_8817f540f718affd21718f415b81597f"
)

const (
	LarkResourceTypeImage   string = "image"
	LarkResourceTypeSticker string = "sticker"
)

type LarkInteraction string

const (
	LarkInteractionSendMsg     LarkInteraction = "send_msg"
	LarkInteractionAddReaction LarkInteraction = "add_reaction"
	LarkInteractionCallBot     LarkInteraction = "call_bot"
)

const (
	MODEL_TYPE_REASON = "reason"
	MODEL_TYPE_NORMAL = "normal"
)

var WITHDRAW_REPLACE = os.Getenv("WITHDRAW_REPLACE") != ""
