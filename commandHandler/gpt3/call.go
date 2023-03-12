package gpt3

import (
	"time"
)

type Options struct {
	// Debug is used to output debug message
	Debug bool
	// Timeout is used to end http request after timeout duration
	Timeout time.Duration
	// ApiKey is used to authoration
	ApiKey string
}

func Post() {
	// httptool.PostWithParams()
}
