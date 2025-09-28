package command

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
	StatScreenIndex
	TitleScreenIndex
	SpellScreenIndex
)

type Direction int

const (
	UpDirection Direction = iota
	DownDirection
	LeftDirection
	RightDirection
)

type FileOperation int

const (
	FileDelete = iota
	FileUpdate
	FileCreate
	FileSave
)

type FileOpMsg struct {
	Op      FileOperation
	Success bool
}

type SelectCharacterMsg struct {
	Character *models.Character
	Err       error
}

type SwitchScreenMsg struct {
	Screen ScreenIndex
}

type FocusNextElementMsg struct {
	Direction Direction
}

type ReturnFocusToParentMsg struct{}

type AppendElementMsg struct {
	Tag string
}

func DeleteCharacterFileCmd(characterDir string, name string) tea.Cmd {
	return func() tea.Msg {
		filename := fmt.Sprintf("%s.json", strings.ToLower(name))
		err := os.Remove(filepath.Join(characterDir, filename))
		if err == nil {
			return FileOpMsg{FileDelete, true}
		} else {
			return FileOpMsg{FileDelete, false}
		}
	}
}

func SaveToFileCmd(c *models.Character) func() tea.Msg {
	return func() tea.Msg {
		err := c.SaveToFile()
		if err == nil {
			return FileOpMsg{FileSave, true}
		} else {
			return FileOpMsg{FileSave, false}
		}
	}
}

func SelectCharacterCmd(name string) func() tea.Msg {
	return func() tea.Msg {
		c, err := models.LoadCharacterByName(name)
		if err != nil {
			return SelectCharacterMsg{nil, err}
		}
		return SelectCharacterMsg{c, nil}
	}
}

func SwitchScreenCmd(s ScreenIndex) func() tea.Msg {
	return func() tea.Msg {
		return SwitchScreenMsg{s}
	}
}

/*
Use to switch focus to other element on same screen.
For switching to element in parent, use ReturnFocusToParentCmd
*/
func FocusNextElementCmd(d Direction) func() tea.Msg {
	return func() tea.Msg {
		return FocusNextElementMsg{d}
	}
}

func AppendElementCmd(tag string) func() tea.Msg {
	return func() tea.Msg { return AppendElementMsg{tag} }
}

func ReturnFocusToParentCmd() tea.Msg {
	return ReturnFocusToParentMsg{}
}
