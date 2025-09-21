package screen

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/component"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/util"
)

type SpellScreen struct {
	keymap    util.KeyMap
	character *models.Character

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	spellAbility  *component.SimpleStringComponent
	spellSaveDC   *component.SimpleIntComponent
	spellAtkBonus *component.SimpleIntComponent
}

func NewSpellScreen(k util.KeyMap, c *models.Character) *SpellScreen {
	return &SpellScreen{
		keymap:        k,
		character:     c,
		spellAbility:  component.NewSimpleStringComponent(k, "Spellcasting Ability", &c.Spells.SpellcastingAbility, true, true),
		spellSaveDC:   component.NewSimpleIntComponent(k, "Spell Save DC", &c.Spells.SpellSaveDC, true, true),
		spellAtkBonus: component.NewSimpleIntComponent(k, "Spell Attack Bonus", &c.Spells.SpellAttackBonus, true, true),
	}
}

func (s *SpellScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.spellAbility.Init())
	cmds = append(cmds, s.spellSaveDC.Init())
	cmds = append(cmds, s.spellAtkBonus.Init())
	cmds = util.DropNil(cmds)
	s.focusOn(s.spellAbility)
	s.lastFocusedElement = s.spellAbility
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (s *SpellScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		switch s.focusedElement {
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case editor.EditValueMsg:
		cmd = editor.SwitchToEditorCmd(command.SpellScreenIndex, s.character, msg.Editors)
	case tea.KeyMsg:
		switch s.focusedElement.(type) {
		case *list.List:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		default:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			case key.Matches(msg, s.keymap.Up):
				cmd = s.moveFocus(command.UpDirection)
			case key.Matches(msg, s.keymap.Down):
				cmd = s.moveFocus(command.DownDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		}
	}
	return s, cmd
}

func (s *SpellScreen) View() string {
	topbar := s.RenderSpellScreenTopBar()
	content := ""
	for i := range 9 {
		content += RenderSpellHeaderRow(s.character, i) + "\n"
	}
	content = util.DefaultBorderStyle.Width(util.ScreenWidth).Render(content)

	return lipgloss.JoinVertical(lipgloss.Center, topbar, content)
}

func (s *SpellScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *SpellScreen) moveFocus(d command.Direction) tea.Cmd {
	var cmd tea.Cmd
	s.Blur()

	switch s.lastFocusedElement {
	case s.spellAbility:
		switch d {
		case command.RightDirection:
			s.focusOn(s.spellSaveDC)
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		}
	case s.spellSaveDC:
		switch d {
		case command.RightDirection:
			s.focusOn(s.spellAtkBonus)
		case command.LeftDirection:
			s.focusOn(s.spellAbility)
		}
	case s.spellAtkBonus:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.spellSaveDC)
		}
	}
	return cmd
}

func (s *SpellScreen) Focus() {
	s.focusOn(s.lastFocusedElement)
}

func (s *SpellScreen) Blur() {
	if s.focusedElement != nil {
		s.focusedElement.Blur()
		s.lastFocusedElement = s.focusedElement
	}

	s.focusedElement = nil
}

func (s *SpellScreen) RenderSpellScreenTopBar() string {
	return util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			util.ForceWidth(s.spellAbility.View(), util.ScreenWidth/3),
			util.ForceWidth(s.spellSaveDC.View(), util.ScreenWidth/3),
			util.ForceWidth(s.spellAtkBonus.View(), util.ScreenWidth/3)))
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
