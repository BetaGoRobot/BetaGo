package consts

import "errors"

var (
	ErrStageSkip         = errors.New("Stage Skip")
	ErrStageError        = errors.New("Stage Error")
	ErrStageWarn         = errors.New("Stage Warn")
	ErrCommandNotFound   = errors.New("CommandNotFoundError")
	ErrCommandIncomplete = errors.New("CommandIncompleteError")
)
