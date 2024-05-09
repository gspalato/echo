package utils

import "os"

func GetenvOr(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func Map[I any, O any](list []I, f func(I, int) O) []O {
	mapped := make([]O, len(list))
	for i, v := range list {
		mapped[i] = f(v, i)
	}
	return mapped
}

func Sum[I any, O int | int64 | float32 | float64](list []I, selector func(v I) O) O {
	var sum O
	for _, v := range list {
		sum += selector(v)
	}
	return sum
}
