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

var (
	spellColHeight = 30
	spellColWidth  = util.ScreenWidth/2 - 8
)

type SpellScreen struct {
	keymap    util.KeyMap
	character *models.Character

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel
	spellListIndex     int

	spellAbility  *component.SimpleStringComponent
	spellSaveDC   *component.SimpleIntComponent
	spellAtkBonus *component.SimpleIntComponent
	spellLists    map[int]*list.List
	colSplitIndex int
}

func NewSpellScreen(k util.KeyMap, c *models.Character) *SpellScreen {
	return &SpellScreen{
		keymap:        k,
		character:     c,
		spellAbility:  component.NewSimpleStringComponent(k, "Spellcasting Ability", &c.Spells.SpellcastingAbility, true, true),
		spellSaveDC:   component.NewSimpleIntComponent(k, "Spell Save DC", &c.Spells.SpellSaveDC, true, true),
		spellAtkBonus: component.NewSimpleIntComponent(k, "Spell Attack Bonus", &c.Spells.SpellAttackBonus, true, true),
		spellLists:    make(map[int]*list.List),
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

	for i := range 10 {
		s.spellLists[i] = list.NewList(s.keymap,
			list.ListStyles{
				Row:      util.ItemStyleDefault.Align(lipgloss.Left),
				Selected: util.ItemStyleSelected.Align(lipgloss.Left),
			}).
			WithAppender().
			WithRows(s.GetSpellListByLevel(i))
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (s *SpellScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		switch s.focusedElement.(type) {
		case *list.List:
			s.character.AddEmptySpell(s.spellListIndex)
			newSpellRows := s.GetSpellListByLevel(s.spellListIndex)
			s.spellLists[s.spellListIndex].WithRows(newSpellRows)
			cmd = editor.SwitchToEditorCmd(
				command.SpellScreenIndex,
				s.character,
				newSpellRows[len(newSpellRows)-1].Editors(),
			)
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
	renderedSpells := []string{}
	for i := range 10 {
		renderedSpells = append(renderedSpells, util.WithPadding(s.spellLists[i].View(), 0, 0, 0, 1))
	}

	columns := util.SplitIntoColumns(renderedSpells, spellColHeight-2)
	firstCol := columns[0]
	secondCol := []string{}
	if len(columns) > 1 {
		secondCol = columns[1]
	}
	s.colSplitIndex = len(firstCol)

	left := lipgloss.PlaceHorizontal(spellColWidth, lipgloss.Left, lipgloss.JoinVertical(lipgloss.Left, firstCol...))
	separator := lipgloss.PlaceHorizontal(8, lipgloss.Left, util.MakeVerticalSeparator(spellColHeight-2))
	right := lipgloss.PlaceHorizontal(spellColWidth, lipgloss.Left, lipgloss.JoinVertical(lipgloss.Left, secondCol...))

	content := util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		Height(spellColHeight).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				left,
				separator,
				right,
			))
	return lipgloss.JoinVertical(lipgloss.Left, topbar, content)
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
		case command.DownDirection:
			s.focusOn(s.spellLists[0])
			s.spellListIndex = 0
		}
	case s.spellSaveDC:
		switch d {
		case command.RightDirection:
			s.focusOn(s.spellAtkBonus)
		case command.LeftDirection:
			s.focusOn(s.spellAbility)
		case command.DownDirection:
			s.focusOn(s.spellLists[0])
			s.spellListIndex = 0
		}
	case s.spellAtkBonus:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.spellSaveDC)
		case command.DownDirection:
			s.focusOn(s.spellLists[0])
			s.spellListIndex = 0
		}
	default:
		switch d {
		case command.UpDirection:
			if s.spellListIndex == 0 {
				s.focusOn(s.spellAbility)
			} else {
				s.spellListIndex -= 1
				s.focusOn(s.spellLists[s.spellListIndex])
			}
		case command.DownDirection:
			if s.spellListIndex < 9 {
				s.spellListIndex += 1
				s.focusOn(s.spellLists[s.spellListIndex])
			}
		case command.RightDirection:
			if s.spellListIndex < s.colSplitIndex {
				s.spellListIndex = min(9, s.spellListIndex+s.colSplitIndex)
				s.focusOn(s.spellLists[s.spellListIndex])
			}
		case command.LeftDirection:
			if s.spellListIndex >= s.colSplitIndex {
				s.spellListIndex = max(0, s.spellListIndex-s.colSplitIndex)
				s.focusOn(s.spellLists[s.spellListIndex])
			}
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

func (s *SpellScreen) GetSpellListByLevel(l int) []list.Row {
	rows := []list.Row{}
	spells := s.character.GetSpellsByLevel(l)
	rows = append(rows, list.NewStructRow(s.keymap,
		&SpellListHeader{l, &s.character.Spells.SpellSlots[l], &s.character.Spells.SpellSlotsUsed[l]},
		RenderSpellHeaderRow,
		[]editor.ValueEditor{
			editor.NewIntEditor(s.keymap, "Used Spell Slots", &s.character.Spells.SpellSlotsUsed[l]),
			editor.NewIntEditor(s.keymap, "Max Spell Slots", &s.character.Spells.SpellSlots[l]),
		}))
	for _, spell := range spells {
		rows = append(rows, list.NewStructRow(s.keymap, spell,
			RenderSpellInfoRow,
			[]editor.ValueEditor{
				editor.NewStringEditor(s.keymap, "Name", &spell.Name),
				editor.NewStringEditor(s.keymap, "Casting Time", &spell.CastingTime),
				editor.NewStringEditor(s.keymap, "Range", &spell.Range),
				editor.NewStringEditor(s.keymap, "Duration", &spell.Duration),
				editor.NewStringEditor(s.keymap, "Components", &spell.Components),
				editor.NewStringEditor(s.keymap, "Description", &spell.Description),
			}))
	}
	return rows
}

func (s *SpellScreen) RenderSpellScreenTopBar() string {
	separator := util.GrayTextStyle.Width(8).Render(util.MakeVerticalSeparator(1))
	return util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			util.ForceWidth(s.spellAbility.View(), 28),
			separator,
			util.ForceWidth(s.spellSaveDC.View(), 28),
			separator,
			util.ForceWidth(s.spellAtkBonus.View(), 28)))
}

type SpellListHeader struct {
	level int
	slots *int
	used  *int
}

func RenderSpellHeaderRow(h *SpellListHeader) string {
	return fmt.Sprintf("Level %d ∙ %s", h.level,
		RenderSpellSlots(*h.used, *h.slots))
}

func RenderSpellInfoRow(s *models.Spell) string {
	return fmt.Sprintf("%s ∙ %s ∙ %s ∙ %s ∙ %s", s.Name, s.Components, s.Range, s.CastingTime, s.Duration)
}

func RenderSpellSlots(used int, max int) string {
	if max <= 0 {
		return "∅"
	}
	s := strings.Repeat("■", used)
	s += strings.Repeat("□", max-used)
	return util.DefaultTextStyle.Render(s)
}
