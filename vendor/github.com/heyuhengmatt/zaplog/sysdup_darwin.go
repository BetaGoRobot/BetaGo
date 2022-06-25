package zaplog

import "syscall"

func setDup(panicFd, stderrFd int) {
	syscall.Dup2(panicFd, stderrFd)
}
