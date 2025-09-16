package ui

import (
	"fmt"
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

type ExitTableMsg struct {
	Key tea.KeyType
}

func EnterEditModeCmd() tea.Msg {
	return EditMessage("start")
}

func ExitEditModeCmd() tea.Msg {
	return EditMessage("stop")
}

func DeleteCharacterFileCmd(characterDir string, name string) tea.Cmd {
	return func() tea.Msg {
		filename := fmt.Sprintf("%s.json", strings.ToLower(name))
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

func ExitListCmd(k tea.KeyType) func() tea.Msg {
	return func() tea.Msg {
		return ExitTableMsg{k}
	}
}
