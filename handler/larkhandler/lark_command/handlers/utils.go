package handlers

import "strings"

func parseArgs(args ...string) map[string]string {
	resMap := make(map[string]string)
	for _, arg := range args {
		if argKV := strings.Split(arg, "="); len(argKV) == 2 {
			resMap[strings.TrimLeft(argKV[0], "--")] = argKV[1]
		}
	}
	return resMap
}
