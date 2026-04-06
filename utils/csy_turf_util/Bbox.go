package csy_turf_util

import "math"

type TJsBbox struct {
	XMin float64
	XMax float64
	YMin float64
	YMax float64
}

func Bbox(x []float64, y []float64) TJsBbox {
	var xMin, xMax, yMin, yMax = math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64
	for i := range x {
		if x[i] < xMin {
			xMin = x[i]
		}
		if x[i] > xMax {
			xMax = x[i]
		}
		if y[i] < yMin {
			yMin = y[i]
		}
		if y[i] > yMax {
			yMax = y[i]
		}
	}
	return TJsBbox{
		XMin: xMin,
		XMax: xMax,
		YMin: yMin,
		YMax: yMax,
	}
}
