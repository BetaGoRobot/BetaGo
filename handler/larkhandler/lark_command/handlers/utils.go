package handlers

import "strings"

func parseArgs(args ...string) (argsMap map[string]string, input string) {
	argsMap = make(map[string]string)
	for idx, arg := range args {
		if strings.HasPrefix(arg, "--") {
			argKV := strings.Split(arg, "=")
			if len(argKV) > 1 {
				argsMap[strings.TrimPrefix(argKV[0], "--")] = argKV[1]
			} else {
				argsMap[strings.TrimPrefix(argKV[0], "--")] = ""
			}
		} else {
			input = strings.Join(args[idx:], " ")
			break
		}
	}
	return
}
