package csy_number_util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomBetween(min, max int) int {
	return rand.Intn(max-min+1) + min
}
func RandomFloatBetween(min, max float64) float64 {
	if min >= max {
		return min // 处理无效区间，或根据需求返回错误
	}
	return min + rand.Float64()*(max-min)
}
