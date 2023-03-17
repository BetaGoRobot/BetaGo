package zaplog

import "syscall"

func setDup(panicFd, stderrFd int) {
	syscall.Dup3(panicFd, stderrFd, 0)
}
