package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
)

type SpellScreen struct {
	keymap    KeyMap
	character *models.Character
}

func NewSpellScreen(k KeyMap, c *models.Character) *SpellScreen {
	return &SpellScreen{k, c}
}

func (s *SpellScreen) Init() tea.Cmd {
	return nil
}

func (s *SpellScreen) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (s *SpellScreen) View() string {
	content := ""
	for i := range 9 {
		content += RenderSpellHeaderRow(s.character, i) + "\n"
	}
	return content
}

func (s *SpellScreen) Focus() {
}

func (s *SpellScreen) Blur() {
}

func RenderSpellHeaderRow(c *models.Character, level int) string {
	return fmt.Sprintf("%d • %s", level,
		RenderSpellSlots(c.Spells.SpellSlotsUsed[level], c.Spells.SpellSlots[level]))
}

func RenderSpellSlots(used int, max int) string {
	s := strings.Repeat("▣", used)
	s += strings.Repeat("□", max-used)
	return DefaultTextStyle.Render(s)
}
