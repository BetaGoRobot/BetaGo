package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	fmt.Println(
		RebuildAtMsg("@_user_1  123  @_user_2 123 123 123 @_user_3 123", []string{"@_user_1", "@_user_2", "@_user_3"}),
	)
}

func RebuildAtMsg(input string, substrings []string) []string {
	result := []string{}
	start := 0

	// Keep track of the positions to split
	splitPositions := []int{}

	// Iterate through the input to find all occurrences of substrings
	for _, sub := range substrings {
		start = 0
		for {
			pos := strings.Index(input[start:], sub)
			if pos == -1 {
				break
			}
			actualPos := start + pos
			splitPositions = append(splitPositions, actualPos, actualPos+len(sub))
			start = actualPos + len(sub)
		}
	}

	// Sort the positions to split
	sort.Slice(splitPositions, func(i, j int) bool { return splitPositions[i] < splitPositions[j] })

	// Remove duplicate positions
	uniquePositions := []int{}
	for i, pos := range splitPositions {
		if i == 0 || pos != splitPositions[i-1] {
			uniquePositions = append(uniquePositions, pos)
		}
	}

	// Add start and end of the string to the positions if not already present
	if uniquePositions[0] != 0 {
		uniquePositions = append([]int{0}, uniquePositions...)
	}
	if uniquePositions[len(uniquePositions)-1] != len(input) {
		uniquePositions = append(uniquePositions, len(input))
	}

	// Extract substrings based on split positions
	for i := 0; i < len(uniquePositions)-1; i++ {
		result = append(result, input[uniquePositions[i]:uniquePositions[i+1]])
	}

	return result
}
