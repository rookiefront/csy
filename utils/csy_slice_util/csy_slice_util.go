package csy_slice_util

import (
	"math"
	"reflect"
)

func SliceInclude[T any](src []T, target T) bool {
	for _, element := range src {
		if reflect.DeepEqual(target, element) {
			return true
		}
	}
	return false
}
func SliceUnique[T any](arr []T) (newArr []T) {
	newArr = make([]T, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if reflect.DeepEqual(arr[i], arr[j]) {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

func SliceChunk[T any](s []T, size int) [][]T {
	if size < 1 {
		return [][]T{}
	}
	length := len(s)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var n [][]T
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		n = append(n, s[i*size:end])
		i++
	}
	return n
}

func SliceReverse[T any](slice []T) []T {
	reversed := make([]T, len(slice))
	for i, j := len(slice)-1, 0; i >= 0; i, j = i-1, j+1 {
		reversed[j] = slice[i]
	}
	return reversed
}
