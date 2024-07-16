package larkhandler

// MainLarkHandler  is the main handler for lark messages
var MainLarkHandler = &LarkMsgProcessor{}

func init() {
	MainLarkHandler = MainLarkHandler.
		AddParallelStages(&RepeatMsgOperator{}).
		AddParallelStages(&ReactMsgOperator{})
}
