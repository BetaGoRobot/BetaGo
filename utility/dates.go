package utility

import (
	"strconv"
	"time"
)

func EpoMil2DateStr(epoMil string) string {
	epoMilInt, _ := strconv.ParseInt(epoMil, 10, 64)
	return time.Unix(int64(epoMilInt)/1000, 0).In(UTCPlus8Loc()).Format("2006-01-02 15:04:05")
}

func EpoMicro2DateStr(epoMicro string) string {
	epoMilInt, _ := strconv.ParseInt(epoMicro, 10, 64)
	return time.Unix(int64(epoMilInt)/1000/1000, 0).In(UTCPlus8Loc()).Format("2006-01-02 15:04:05")
}

func MustInt(str string) int64 {
	i, _ := strconv.ParseInt(str, 10, 64)
	return i
}

func Epo2DateZoneMil(millSecs int64, location *time.Location) string {
	return Epo2DateZoneStr(millSecs/1000, location)
}

func Epo2DateZoneSec(secs int64, location *time.Location) string {
	return Epo2DateZoneStr(secs, location)
}

func Epo2DateZoneMicro(microSecs int64, location *time.Location) string {
	return Epo2DateZoneStr(microSecs/1000/1000, location)
}

func Epo2DateZoneStr(secs int64, location *time.Location) string {
	return time.Unix(secs, 0).In(location).Format(time.RFC3339)
}
