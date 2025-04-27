package utility

import (
	"strconv"
	"time"
)

func EpoMil2DateStr(epoMil string) string {
	epoMilInt, _ := strconv.ParseInt(epoMil, 10, 64)
	return time.Unix(int64(epoMilInt)/1000, 0).UTC().Format("2006-01-02 15:04:05")
}
