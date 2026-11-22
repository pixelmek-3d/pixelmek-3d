package common

import (
	"math/rand/v2"
	"slices"
)

func Pop[T any](s []T, index int) (T, []T) {
	val := s[index]
	return val, slices.Delete(s, index, index+1)
}

func PopRandom[T any](s []T) (T, []T) {
	return Pop(s, rand.IntN(len(s)))
}
