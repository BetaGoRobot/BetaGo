package consts

import "errors"

var (
	ErrStageSkip         = errors.New("Stage Skip")
	ErrStageError        = errors.New("Stage Error")
	ErrStageWarn         = errors.New("Stage Warn")
	ErrArgsIncompelete   = errors.New("ArgsIncompeleteError")
	ErrCommandNotFound   = errors.New("CommandNotFoundError")
	ErrCommandIncomplete = errors.New("CommandIncompleteError")
)
