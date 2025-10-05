package screen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/component"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/util"
)

var (
	spellColHeight = 30
	spellColWidth  = util.ScreenWidth - 8
)

type SpellScreen struct {
	keymap    util.KeyMap
	character *models.Character

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	spellAbility  *component.SimpleStringComponent
	spellSaveDC   *component.SimpleIntComponent
	spellAtkBonus *component.SimpleIntComponent
	spellList     *list.List
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

	s.populateSpells()

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (s *SpellScreen) populateSpells() {
	if s.spellList == nil {
		s.spellList = list.NewList(s.keymap,
			list.ListStyles{
				Row:      util.ItemStyleDefault.Align(lipgloss.Left),
				Selected: util.ItemStyleSelected.Align(lipgloss.Left),
			}).
			WithFixedWidth(spellColWidth).
			WithViewport(spellColHeight - 2)
	}
	rows := []list.Row{}
	for i := range 10 {
		rows = append(rows, s.GetSpellListByLevel(i)...)
	}
	s.spellList.WithRows(rows[:len(rows)-1])
}

func (s *SpellScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "spell:") {
			l, _ := strconv.Atoi(strings.Split(msg.Tag, ":")[1])
			spell_id := s.character.AddEmptySpell(l)
			s.populateSpells()
			cmd = editor.SwitchToEditorCmd(
				command.SpellScreenIndex,
				s.character,
				s.getSpellRow(spell_id).Editors(),
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
	renderedSpells := s.spellList.View()

	content := util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		Height(spellColHeight).
		Render(renderedSpells)
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
			s.focusOn(s.spellList)
		default:
			s.focusOn(s.spellAbility)
		}
	case s.spellSaveDC:
		switch d {
		case command.RightDirection:
			s.focusOn(s.spellAtkBonus)
		case command.LeftDirection:
			s.focusOn(s.spellAbility)
		case command.DownDirection:
			s.focusOn(s.spellList)
		default:
			s.focusOn(s.spellSaveDC)
		}
	case s.spellAtkBonus:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.spellSaveDC)
		case command.DownDirection:
			s.focusOn(s.spellList)
		default:
			s.focusOn(s.spellAtkBonus)
		}
	case s.spellList:
		switch d {
		case command.UpDirection:
			s.focusOn(s.spellAbility)
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		default:
			s.focusOn(s.spellList)
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
	rows = append(rows, list.NewSeparatorRow("─", spellColWidth-6))
	for _, spell := range spells {
		rows = append(rows, list.NewStructRow(s.keymap, spell,
			RenderSpellInfoRow,
			CreateSpellEditors(s.keymap, spell),
		).WithDestructor(command.SpellScreenIndex, DeleteSpellCallback(s, spell)))
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, fmt.Sprintf("spell:%d", l)))
	rows = append(rows, list.NewSeparatorRow(" ", spellColWidth-6))
	return rows
}

func (s *SpellScreen) getSpellRow(id uuid.UUID) list.Row {
	for _, r := range s.spellList.Content() {
		switch r := r.(type) {
		case *list.StructRow[models.Spell]:
			if r.Value().Id == id {
				return r
			}
		}
	}
	return nil
}

func DeleteSpellCallback(s *SpellScreen, sp *models.Spell) func() tea.Cmd {
	return func() tea.Cmd {
		s.character.DeleteSpell(sp.Id)
		s.populateSpells()
		return command.SaveToFileCmd(s.character)
	}
}

func CreateSpellEditors(k util.KeyMap, spell *models.Spell) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(k, "Name", &spell.Name),
		editor.NewStringEditor(k, "Damage", &spell.Damage),
		editor.NewStringEditor(k, "Casting Time", &spell.CastingTime),
		editor.NewStringEditor(k, "Range", &spell.Range),
		editor.NewStringEditor(k, "Duration", &spell.Duration),
		editor.NewStringEditor(k, "Components", &spell.Components),
		editor.NewStringEditor(k, "Description", &spell.Description),
	}
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
	values := []string{s.Name, s.Damage, s.Components, s.Range, s.CastingTime, s.Duration}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return strings.Join(values, " ∙ ")
}

func RenderSpellSlots(used int, max int) string {
	if max <= 0 {
		return "∅"
	}
	s := strings.Repeat("■", used)
	s += strings.Repeat("□", max-used)
	return util.DefaultTextStyle.Render(s)
}
