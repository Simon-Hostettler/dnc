package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type CharacterRow struct {
	keymap       KeyMap
	characterDir string
	name         string
}

func NewCharacterRow(name string, characterDir string, keymap KeyMap) *CharacterRow {
	return &CharacterRow{keymap, characterDir, name}
}

func (c *CharacterRow) Init() tea.Cmd {
	return nil
}

func (c *CharacterRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keymap.Select):
			return c, SelectCharacterAndSwitchScreenCommand(c.name)
		case key.Matches(msg, c.keymap.Delete):
			return c, DeleteCharacterFileCmd(c.characterDir, c.name)
		}

	}
	return c, nil
}

func (c *CharacterRow) View() string {
	return c.name
}

func (c *CharacterRow) Editors() []ValueEditor {
	return []ValueEditor{}
}
