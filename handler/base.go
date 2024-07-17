package handler

type BotMsgProcessor interface {
	RunStages() error
	RunParallelStages() error
}
