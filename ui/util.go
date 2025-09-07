package ui

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
)

type EditMessage string

type FileOpMsg struct {
	op      string
	success bool
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

func EnterEditModeCmd() tea.Msg {
	return EditMessage("start")
}

func ExitEditModeCmd() tea.Msg {
	return EditMessage("stop")
}

func DeleteCharacterFileCmd(characterDir string, filename string) tea.Cmd {
	return func() tea.Msg {
		err := os.Remove(filepath.Join(characterDir, filename))
		if err == nil {
			return FileOpMsg{"delete", true}
		} else {
			return FileOpMsg{"delete", false}
		}
	}
}

func SaveToFileCmd(c *models.Character) func() tea.Msg {
	return func() tea.Msg {
		err := c.SaveToFile()
		if err == nil {
			return FileOpMsg{"save", true}
		} else {
			return FileOpMsg{"save", false}
		}
	}
}

func UpdateFilesCmd(t *TitleScreen) func() tea.Msg {
	return func() tea.Msg {
		t.UpdateFiles()
		return FileOpMsg{"save", true}
	}

}
