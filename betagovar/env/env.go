package env

// CheckPeriod  1
var CheckPeriod = GetEnvWithDefault("CHECK_PERIOD", "15")

// GithubSha  1
var GithubSha = GetEnvWithDefault("GITHUB_SHA", "")

// GitCommitMessage 1
var GitCommitMessage = GetEnvWithDefault("GIT_COMMIT_MESSAGE", "")
