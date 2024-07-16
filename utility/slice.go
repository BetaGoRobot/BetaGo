package utility

import "math/rand"

func SampleSlice[T any](slices []T) T {
	return slices[rand.Intn(len(slices))]
}
