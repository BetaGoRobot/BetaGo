package larkhandler

import "errors"

var (
	ErrStageSkip  = errors.New("Stage skip")
	ErrStageError = errors.New("Stage error")
	ErrStageWarn  = errors.New("Stage warn")
)
