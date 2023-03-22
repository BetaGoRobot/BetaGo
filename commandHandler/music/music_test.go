package music

import (
	"context"
	"os"
	"runtime/pprof"
	"testing"
)

func TestSearchMusicByRobot(t *testing.T) {
	f, _ := os.OpenFile("/tmp/cpu.prof", os.O_RDWR|os.O_CREATE, 0o644)
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	for i := 0; i < 1000; i++ {
		SearchMusicByRobot(context.Background(), "", "", "", "命名")
	}
}
