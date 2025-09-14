package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up        key.Binding `json:"up"`
	Down      key.Binding `json:"down"`
	Left      key.Binding `json:"left"`
	Right     key.Binding `json:"right"`
	Select    key.Binding `json:"select"`
	Enter     key.Binding `json:"enter"`
	Escape    key.Binding `json:"escape"`
	Delete    key.Binding `json:"delete"`
	ForceQuit key.Binding `json:"force_quit"`
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:        key.NewBinding(key.WithKeys("k", "up")),
		Down:      key.NewBinding(key.WithKeys("j", "down")),
		Left:      key.NewBinding(key.WithKeys("h", "left")),
		Right:     key.NewBinding(key.WithKeys("l", "right")),
		Select:    key.NewBinding(key.WithKeys(" ", "enter")),
		Enter:     key.NewBinding(key.WithKeys("enter")),
		Escape:    key.NewBinding(key.WithKeys("esc")),
		Delete:    key.NewBinding(key.WithKeys("x", "del")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}

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
	return strings.Title(fileName)
}

func RenderEdgeBound(w1 int, w2 int, str1 string, str2 string) string {
	format := fmt.Sprintf("%%-%ds%%%ds", w1, w2)
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

func ListCharacterFiles(dir string) []string {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, entry.Name())
		}
	}
	return files
}
