package csy_number_util

import (
	"math/rand"
	"time"
)

func RandomBetween(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}
