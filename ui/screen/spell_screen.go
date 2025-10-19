package screen

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
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
	keymap              util.KeyMap
	CharacterRepository repository.CharacterRepository
	Context             context.Context

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	characterId   uuid.UUID
	spellAbility  *component.SimpleStringComponent
	spellSaveDC   *component.SimpleIntComponent
	spellAtkBonus *component.SimpleIntComponent
	spellList     *list.List
}

func NewSpellScreen(k util.KeyMap, cr repository.CharacterRepository, ctx context.Context, characterId uuid.UUID) *SpellScreen {
	s := SpellScreen{
		keymap:              k,
		CharacterRepository: cr,
		Context:             ctx,
		characterId:         characterId,
	}

	s.spellAbility = component.NewSimpleStringComponent(k, "Spellcasting Ability", "", s.persistCharStringField("spellcasting_ability"), true, true)
	s.spellSaveDC = component.NewSimpleIntComponent(k, "Spell Save DC", 0, s.persistCharIntField("spell_save_dc"), true, true)
	s.spellAtkBonus = component.NewSimpleIntComponent(k, "Spell Attack Bonus", 0, s.persistCharIntField("spell_attack_bonus"), true, true)
	return &s
}

func (s *SpellScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{s.reloadData()}
	cmds = append(cmds, s.spellAbility.Init())
	cmds = append(cmds, s.spellSaveDC.Init())
	cmds = append(cmds, s.spellAtkBonus.Init())
	cmds = util.DropNil(cmds)
	s.focusOn(s.spellAbility)
	s.lastFocusedElement = s.spellAbility

	return tea.Batch(cmds...)
}

func (s *SpellScreen) reloadData() tea.Cmd {
	return command.LoadCharacterCmd(s.CharacterRepository, s.Context, s.characterId)
}

func (s *SpellScreen) populateSpells(agg repository.CharacterAggregate) {
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
		rows = append(rows, s.GetSpellListByLevel(agg, i)...)
	}
	s.spellList.WithRows(rows[:len(rows)-1])
}

func (s *SpellScreen) populateSpellcasting(c models.CharacterTO) {
	s.spellAbility = component.NewSimpleStringComponent(s.keymap, "Spellcasting Ability", c.SpellcastingAbility, s.persistCharStringField("spellcasting_ability"), true, true)
	s.spellSaveDC = component.NewSimpleIntComponent(s.keymap, "Spell Save DC", c.SpellSaveDC, s.persistCharIntField("spell_save_dc"), true, true)
	s.spellAtkBonus = component.NewSimpleIntComponent(s.keymap, "Spell Attack Bonus", c.SpellAttackBonus, s.persistCharIntField("spell_attack_bonus"), true, true)
}

func (s *SpellScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.DataOpMsg:
		if msg.Op != command.DataSave {
			cmd = command.LoadSpellsCommand(s.CharacterRepository, s.Context, s.characterId)
		}
	case command.LoadCharacterMsg:
		s.populateSpellcasting(*msg.Character.Character)
		s.populateSpells(msg.Character)
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "spell:") {
			l, _ := strconv.Atoi(strings.Split(msg.Tag, ":")[1])
			cmd = command.DataOperationCommand(func() error {
				_, err := s.CharacterRepository.AddSpell(s.Context, s.characterId, models.SpellTO{Level: l})
				return err
			}, command.DataCreate)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case editor.EditValueMsg:
		cmd = editor.SwitchToEditorCmd(msg.Editors)
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

func (s *SpellScreen) GetSpellListByLevel(agg repository.CharacterAggregate, l int) []list.Row {
	rows := []list.Row{}
	rows = append(rows, list.NewStructRow(s.keymap,
		&SpellListHeader{l, agg.Character.SpellSlots[l], agg.Character.SpellSlotsUsed[l]},
		RenderSpellHeaderRow,
		[]editor.ValueEditor{
			editor.NewIntEditor(s.keymap, "Used Spell Slots", agg.Character.SpellSlotsUsed[l]),
			editor.NewIntEditor(s.keymap, "Max Spell Slots", agg.Character.SpellSlots[l]),
		}))
	rows = append(rows, list.NewSeparatorRow("─", spellColWidth-6))
	for _, spell := range spells {
		rows = append(rows, list.NewStructRow(s.keymap, spell,
			RenderSpellInfoRow,
			s.CreateSpellEditors(spell),
		).WithDestructor(s.DeleteSpellCallback(spell)).
			WithReader(RenderFullSpellInfo))
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, fmt.Sprintf("spell:%d", l)))
	rows = append(rows, list.NewSeparatorRow(" ", spellColWidth-6))
	return rows
}

func (s *SpellScreen) DeleteSpellCallback(sp models.SpellTO) func() tea.Cmd {
	return func() tea.Cmd {
		return command.DataOperationCommand(func() error { return s.CharacterRepository.DeleteSpell(s.Context, sp.ID) }, command.DataDelete)
	}
}

func (s *SpellScreen) CreateSpellEditors(spell models.SpellTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(s.keymap, "Name", spell.Name, s.persistSpellStringField(spell.ID, "name")),
		editor.NewBooleanEditor(s.keymap, "Prepared", spell.Prepared, s.persistSpellBoolField(spell.ID, "prepared")),
		editor.NewStringEditor(s.keymap, "Damage", spell.Damage, s.persistSpellStringField(spell.ID, "damage")),
		editor.NewStringEditor(s.keymap, "Casting Time", spell.CastingTime, s.persistSpellStringField(spell.ID, "casting_time")),
		editor.NewStringEditor(s.keymap, "Range", spell.Range, s.persistSpellStringField(spell.ID, "range")),
		editor.NewStringEditor(s.keymap, "Duration", spell.Duration, s.persistSpellStringField(spell.ID, "duration")),
		editor.NewStringEditor(s.keymap, "Components", spell.Components, s.persistSpellStringField(spell.ID, "components")),
		editor.NewStringEditor(s.keymap, "Description", spell.Description, s.persistSpellStringField(spell.ID, "description")),
	}
}

func (s *SpellScreen) persistCharIntField(field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateCharacterFields(s.Context, s.characterId, map[string]interface{}{field: v})
	}
}

func (s *SpellScreen) persistCharStringField(field string) func(string) error {
	return func(v string) error {
		return s.CharacterRepository.UpdateCharacterFields(s.Context, s.characterId, map[string]interface{}{field: v})
	}
}

func (s *SpellScreen) persistSpellStringField(id uuid.UUID, field string) func(string) error {
	return func(v string) error {
		return s.CharacterRepository.UpdateSpellFields(s.Context, id, map[string]interface{}{field: v})
	}
}

func (s *SpellScreen) persistSpellBoolField(id uuid.UUID, field string) func(bool) error {
	return func(v bool) error {
		return s.CharacterRepository.UpdateSpellFields(s.Context, id, map[string]interface{}{field: util.B2i(v)})
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
	slots int
	used  int
}

func RenderSpellHeaderRow(h *SpellListHeader) string {
	return fmt.Sprintf("Level %d ∙ %s", h.level,
		RenderSpellSlots(h.used, h.slots))
}

func RenderSpellInfoRow(s models.SpellTO) string {
	values := []string{s.Name, s.Damage, s.Components, s.Range, s.CastingTime, s.Duration}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return util.PrettyBoolCircle(util.I2b(s.Prepared)) + " " + strings.Join(values, " ∙ ")
}

func RenderFullSpellInfo(s models.SpellTO) string {
	separator := util.MakeHorizontalSeparator(util.SmallScreenWidth-4, 1)
	content := strings.Join(
		[]string{
			s.Name + " ∙  Level: " + strconv.Itoa(s.Level),
			separator,
			"Components: " + s.Components,
			separator,
			"Range: " + s.Range,
			separator,
			"Damage: " + s.Damage,
			separator,
			"Casting time: " + s.CastingTime,
			separator,
			"Duration: " + s.Duration,
			separator,
			s.Description,
		},
		"\n")
	return util.DefaultTextStyle.
		AlignHorizontal(lipgloss.Left).
		Render(content)
}

func RenderSpellSlots(used int, max int) string {
	if max <= 0 {
		return "∅"
	}
	s := strings.Repeat("■", used)
	s += strings.Repeat("□", max-used)
	return util.DefaultTextStyle.Render(s)
}
