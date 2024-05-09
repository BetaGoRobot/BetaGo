package env

import "os"

// CheckPeriod  1
var CheckPeriod = GetEnvWithDefault("CHECK_PERIOD", "5")

// GithubSha  1
var GithubSha = GetEnvWithDefault("GITHUB_SHA", "")

// GitCommitMessage 1
var GitCommitMessage = GetEnvWithDefault("GIT_COMMIT_MESSAGE", "")

var (
	NETEASE_EMAIL    = os.Getenv("NETEASE_EMAIL")
	NETEASE_PASSWORD = os.Getenv("NETEASE_PASSWORD")
)
