package env

import (
	"os"
	"time"
)

// CheckPeriod  1
var CheckPeriod = GetEnvWithDefault("CHECK_PERIOD", "5")

// GithubSha  1
var GithubSha = GetEnvWithDefault("GITHUB_SHA", "")

// GitCommitMessage 1
var GitCommitMessage = GetEnvWithDefault("GIT_COMMIT_MESSAGE", "")

var OSS_EXPIRATION_TIME time.Duration

var (
	NETEASE_EMAIL    = os.Getenv("NETEASE_EMAIL")
	NETEASE_PASSWORD = os.Getenv("NETEASE_PASSWORD")
)

var (
	LarkAppID     = os.Getenv("LARK_CLIENT_ID")
	LarkAppSecret = os.Getenv("LARK_SECRET")
)

func init() {
	var err error
	OSS_EXPIRATION_TIME, err = time.ParseDuration(GetEnvWithDefault("OSS_EXPIRATION_TIME", "1h"))
	if err != nil {
		panic(err)
	}
}
