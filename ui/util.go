package ui

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
)

type ScreenIndex int

const (
	ScoreScreenIndex ScreenIndex = iota
)

type EditMessage string

type FileOpMsg struct {
	op      string
	success bool
}

type SelectCharacterAndSwitchScreenMsg struct {
	Character *models.Character
	Err       error
}

type SwitchScreenMsg struct {
	Screen ScreenIndex
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
		return FileOpMsg{"update", true}
	}
}

func SelectCharacterAndSwitchScreenCommand(name string) func() tea.Msg {
	return func() tea.Msg {
		c, err := models.LoadCharacterByName(name)
		if err != nil {
			return SelectCharacterAndSwitchScreenMsg{nil, err}
		}
		return SelectCharacterAndSwitchScreenMsg{c, nil}
	}
}

func SwitchScreenCmd(s ScreenIndex) func() tea.Msg {
	return func() tea.Msg {
		return SwitchScreenMsg{s}
	}
}
