package handlers

import "strings"

func parseArgs(args ...string) (argsMap map[string]string, input string) {
	argsMap = make(map[string]string)
	for idx, arg := range args {
		if argKV := strings.Split(arg, "="); len(argKV) == 2 {
			argsMap[strings.TrimLeft(argKV[0], "--")] = argKV[1]
		} else {
			input = strings.Join(args[idx:], " ")
		}
	}
	return
}
