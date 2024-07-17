package larkhandler

// MessageHandler  is the main handler for lark messages
var MessageHandler = &LarkMsgProcessor{}

func init() {
	MessageHandler = MessageHandler.
		AddParallelStages(&RepeatMsgOperator{}).
		AddParallelStages(&ReactMsgOperator{}).
		AddParallelStages(&WordReplyMsgOperator{}).
		AddParallelStages(&MusicMsgOperator{})
}
