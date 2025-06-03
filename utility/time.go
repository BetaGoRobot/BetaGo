package utility

import "time"

func UTCPlus8Loc() *time.Location {
	return time.FixedZone("UTC+8", 8*60*60)
}

func UTCPlus8Time() time.Time {
	return time.Now().In(UTCPlus8Loc())
}
