package util

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap struct {
	Up        key.Binding `json:"up"`
	Down      key.Binding `json:"down"`
	Left      key.Binding `json:"left"`
	Right     key.Binding `json:"right"`
	Select    key.Binding `json:"select"`
	Edit      key.Binding `json:"edit"`
	Enter     key.Binding `json:"enter"`
	Escape    key.Binding `json:"escape"`
	Delete    key.Binding `json:"delete"`
	ForceQuit key.Binding `json:"force_quit"`
	Show      key.Binding `json:"show"`
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:        key.NewBinding(key.WithKeys("up")),
		Down:      key.NewBinding(key.WithKeys("down")),
		Left:      key.NewBinding(key.WithKeys("left")),
		Right:     key.NewBinding(key.WithKeys("right")),
		Select:    key.NewBinding(key.WithKeys(" ", "enter")),
		Edit:      key.NewBinding(key.WithKeys("e")),
		Enter:     key.NewBinding(key.WithKeys("enter")),
		Escape:    key.NewBinding(key.WithKeys("esc", "q")),
		Delete:    key.NewBinding(key.WithKeys("x", "del")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
		Show:      key.NewBinding(key.WithKeys(" ")),
	}
}

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

func PrettyFileName(file string) string {
	baseFile := strings.Split(file, "/")[0]
	fileName := strings.TrimSuffix(baseFile, ".json")
	return strings.Title(fileName)
}

func RenderEdgeBound(w1 int, w2 int, str1 string, str2 string) string {
	format := fmt.Sprintf("%%-%ds%%%ds", w1, w2)
	return fmt.Sprintf(format, str1, str2)
}

func RenderLeftBound(w1 int, str1 string, str2 string) string {
	format := fmt.Sprintf("%%-%ds %%s", w1)
	return fmt.Sprintf(format, str1, str2)
}

func ForceLineBreaks(s string, w int) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		b.WriteRune(r)
		if (i+1)%w == 0 && i != len(runes)-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

func SplitIntoColumns(elements []string, maxHeight int) [][]string {
	var (
		columns       [][]string
		currentColumn []string
		currentHeight int
	)
	for _, elem := range elements {
		elemHeight := lipgloss.Height(elem)
		if currentHeight+elemHeight > maxHeight {
			columns = append(columns, currentColumn)
			currentColumn = []string{}
			currentHeight = 0
		}
		currentColumn = append(currentColumn, elem)
		currentHeight += elemHeight
	}
	if len(currentColumn) > 0 {
		columns = append(columns, currentColumn)
	}
	return columns
}

func B2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func I2b(i int) bool {
	if i == 0 {
		return false
	}
	return true
}
