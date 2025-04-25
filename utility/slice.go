package utility

import "math/rand"

func SampleSlice[T any](slices []T) T {
	return slices[rand.Intn(len(slices))]
}

func InSlice[T comparable](slice []T, tgt T) bool {
	for _, item := range slice {
		if item == tgt {
			return true
		}
	}
	return false
}
