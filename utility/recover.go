package utility

import (
	"fmt"
	"runtime/debug"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
)

// CollectPanic is the function to collect panic
func CollectPanic() {
	if err := recover(); err != nil {
		SendEmail("Panic-Collected!", fmt.Sprintf("%v", string(debug.Stack())))
		SendMessageWithTitle(betagovar.TestChanID, "", "", fmt.Sprintf("`%v`", string(debug.Stack())), fmt.Sprintf("%s Panic-Collected! ErrorMsg: `%v`", emoji.Warning.String(), err))
		SugerLogger.Errorf("=====Panic====== %s", string(debug.Stack()))
	}
}
