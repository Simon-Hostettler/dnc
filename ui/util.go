package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type EditMessage string

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func PrettyFileName(file string) string {
	baseFile := strings.Split(file, "/")[0]
	fileName := strings.TrimSuffix(baseFile, ".json")
	return fileName
}

func EnterEditMode() tea.Msg {
	return EditMessage("start")
}

func ExitEditMode() tea.Msg {
	return EditMessage("stop")
}
