package util

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func Filter[T any](ts []T, fn func(T) bool) []T {
	var result []T
	for _, t := range ts {
		if fn(t) {
			result = append(result, t)
		}
	}
	return result
}

type Nilable interface {
	~*int | ~*string | ~[]int | ~map[string]int | ~func() | tea.Cmd
}

func DropNil[T Nilable](ts []T) []T {
	return Filter(ts, func(t T) bool { return t != nil })
}

func B2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func I2b(i int) bool {
	return i != 0
}

func Clamp(i int, l int, h int) int {
	return min(h, max(l, i))
}
