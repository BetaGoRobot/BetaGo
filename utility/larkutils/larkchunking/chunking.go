package larkchunking

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility/chunking"
)

var M *chunking.Management

func init() {
	M = chunking.NewManagement()
	M.StartBackgroundCleaner(context.Background())
}
