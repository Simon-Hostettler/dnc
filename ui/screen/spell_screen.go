package screen

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/component"
	"hostettler.dev/dnc/ui/util"
)

type SpellScreen struct {
	keymap    util.KeyMap
	character *models.Character

	focusedElement FocusableModel

	spellAbility  *component.SimpleStringComponent
	spellSaveDC   *component.SimpleIntComponent
	spellAtkBonus *component.SimpleIntComponent
}

func NewSpellScreen(k util.KeyMap, c *models.Character) *SpellScreen {
	return &SpellScreen{
		keymap:        k,
		character:     c,
		spellAbility:  component.NewSimpleStringComponent(k, "Spellcasting Ability", &c.Spells.SpellcastingAbility, true),
		spellSaveDC:   component.NewSimpleIntComponent(k, "Spell Save DC", &c.Spells.SpellSaveDC, true),
		spellAtkBonus: component.NewSimpleIntComponent(k, "Spell Attack Bonus", &c.Spells.SpellAttackBonus, true),
	}
}

func (s *SpellScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.spellAbility.Init())
	cmds = append(cmds, s.spellSaveDC.Init())
	cmds = append(cmds, s.spellAtkBonus.Init())
	cmds = util.DropNil(cmds)
	s.focusOn(s.spellAbility)
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
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
	return util.DefaultBorderStyle.Width(util.ScreenWidth).Render(content)
}

func (s *SpellScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *SpellScreen) Focus() {
}

func (s *SpellScreen) Blur() {
}

func RenderSpellScreenTopBar() {
}

func RenderSpellHeaderRow(c *models.Character, level int) string {
	return fmt.Sprintf("%d ∙ %s", level,
		RenderSpellSlots(c.Spells.SpellSlotsUsed[level], c.Spells.SpellSlots[level]))
}

func RenderSpellSlots(used int, max int) string {
	s := strings.Repeat("▣", used)
	s += strings.Repeat("□", max-used)
	return util.DefaultTextStyle.Render(s)
}
