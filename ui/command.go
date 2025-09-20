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
	EditScreenIndex ScreenIndex = iota
	ScoreScreenIndex
)

type Direction int

const (
	UpDirection Direction = iota
	DownDirection
	LeftDirection
	RightDirection
)

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
	Direction Direction
}

type EditValueMsg struct {
	Editors []ValueEditor
}

type SwitchToEditorMsg struct {
	Originator ScreenIndex
	Character  *models.Character
	Editors    []ValueEditor
}

type AppendElementMsg struct{}

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

func EditValueCmd(editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return EditValueMsg{editors}
	}
}

func SwitchToEditorCmd(caller ScreenIndex, character *models.Character, editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return SwitchToEditorMsg{caller, character, editors}
	}
}

func SwitchScreenCmd(s ScreenIndex) func() tea.Msg {
	return func() tea.Msg {
		return SwitchScreenMsg{s}
	}
}

func ExitListCmd(d Direction) func() tea.Msg {
	return func() tea.Msg {
		return ExitTableMsg{d}
	}
}

func AppendElementCmd() tea.Msg {
	return AppendElementMsg{}
}
