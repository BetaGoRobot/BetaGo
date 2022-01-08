package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetCurrentTime(t *testing.T) {
	GetCurrentTime()
}

func TestSomething(t *testing.T) {
	sampleSlice := []string{"abc", "def"}
	sd := strings.Join(sampleSlice, "%20")
	fmt.Println(sd)
}
