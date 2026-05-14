package utils

import "math/rand/v2"

func RandFloat(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + rand.Float64()*(max-min)
}
