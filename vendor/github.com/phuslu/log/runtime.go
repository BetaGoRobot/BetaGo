package log

import (
	"unsafe"
)

type funcInfo struct {
	_func unsafe.Pointer
	datap unsafe.Pointer //nolint:unused
}

//go:linkname findfunc runtime.findfunc
func findfunc(pc uintptr) funcInfo

//go:linkname funcInfoEntry runtime.funcInfo.entry
func funcInfoEntry(f funcInfo) uintptr

//go:linkname funcline1 runtime.funcline1
func funcline1(f funcInfo, targetpc uintptr, strict bool) (file string, line int32)

func pcFileLine(pc uintptr) (file string, line int32) {
	funcInfo := findfunc(pc)
	if funcInfo._func == nil {
		return
	}

	entry := funcInfoEntry(funcInfo)

	if pc > entry {
		// We store the pc of the start of the instruction following
		// the instruction in question (the call or the inline mark).
		// This is done for historical reasons, and to make FuncForPC
		// work correctly for entries in the result of runtime.Callers.
		pc--
	}

	return funcline1(funcInfo, pc, false)
}
