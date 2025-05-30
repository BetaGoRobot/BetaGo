package utility

import (
	"strconv"
	"time"
)

func EpoMil2DateStr(epoMil string) string {
	epoMilInt, _ := strconv.ParseInt(epoMil, 10, 64)
	return time.Unix(int64(epoMilInt)/1000, 0).UTC().Add(time.Hour * 8).Format("2006-01-02 15:04:05")
}

func EpoMicro2DateStr(epoMicro string) string {
	epoMilInt, _ := strconv.ParseInt(epoMicro, 10, 64)
	return time.Unix(int64(epoMilInt)/1000/1000, 0).UTC().Add(time.Hour * 8).Format("2006-01-02 15:04:05")
}
