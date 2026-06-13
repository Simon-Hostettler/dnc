package screen

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/component"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var (
	spellColHeight       = 30
	spellColWidth        = styles.ScreenWidth - 10
	spellTopBarElemWidth = 28
)

type SpellScreen struct {
	keymap    util.KeyMap
	character *repository.CharacterAggregate
	FocusManager

	spellAbility  *component.SimpleComponent[string]
	spellSaveDC   *component.SimpleComponent[int]
	spellAtkBonus *component.SimpleComponent[int]
	spellList     *list.List

	perLevelRows [10]*Collection[models.SpellTO]
}

func NewSpellScreen(k util.KeyMap, c *repository.CharacterAggregate) *SpellScreen {
	s := &SpellScreen{
		keymap:        k,
		character:     c,
		spellAbility:  component.NewSimpleStringComponent(k, "Spellcasting Ability", &c.Character.SpellcastingAbility, true, true),
		spellSaveDC:   component.NewSimpleIntComponent(k, "Spell Save DC", &c.Character.SpellSaveDC, true, true),
		spellAtkBonus: component.NewSimpleIntComponent(k, "Spell Attack Bonus", &c.Character.SpellAttackBonus, true, true),
		spellList: list.NewList(k, list.ListStyles{
			Row:      styles.ItemStyleDefault.Align(lipgloss.Left),
			Selected: styles.ItemStyleSelected.Align(lipgloss.Left),
		}).
			WithFixedWidth(spellColWidth).
			WithViewport(spellColHeight - 2).
			WithSearch().
			WithSectionStyle(list.SectionStyle{
				HeaderSeparator: "─",
				SectionGap:      " ",
				SeparatorWidth:  spellColWidth - 6,
			}),
	}
	for i := range 10 {
		level := i
		s.perLevelRows[level] = NewCollection(k, s.spellList,
			func() []*models.SpellTO { return s.character.GetSpellsByLevel(level) },
			func(sp *models.SpellTO) uuid.UUID { return sp.ID },
			func() uuid.UUID { return s.character.AddEmptySpell(level) },
			s.character.DeleteSpell,
			func(sp *models.SpellTO) *list.StructRow[models.SpellTO] {
				return list.NewStructRow(s.keymap, sp, renderSpellInfoRow,
					s.createSpellEditors(sp)).
					WithReader(renderFullSpellInfo).
					WithSearchText(spellSearchText).
					WithCycleAction(toggleSpellPrepared)
			},
		).WithOnChange(s.populateSpells)
	}
	return s
}

func (s *SpellScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.spellAbility.Init())
	cmds = append(cmds, s.spellSaveDC.Init())
	cmds = append(cmds, s.spellAtkBonus.Init())
	s.populateSpells()
	s.wireFocusGraph()

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (s *SpellScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.FocusNextElementMsg:
		s.MoveFocus(msg.Direction)
	case tea.KeyPressMsg:
		cmd = s.RouteKey(msg, s.keymap)
	}
	return s, cmd
}

func (s *SpellScreen) View() tea.View {
	topbar := s.RenderSpellScreenTopBar()
	renderedSpells := s.spellList.View().Content

	content := styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		Height(spellColHeight).
		Render(renderedSpells)
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, topbar, content))
}

func (s *SpellScreen) wireFocusGraph() {
	s.Wire(FocusGraph{
		s.spellAbility: {
			command.RightDirection: To(s.spellSaveDC),
			command.LeftDirection:  Emit(command.ReturnFocusToParentCmd),
			command.DownDirection:  To(s.spellList),
		},
		s.spellSaveDC: {
			command.RightDirection: To(s.spellAtkBonus),
			command.LeftDirection:  To(s.spellAbility),
			command.DownDirection:  To(s.spellList),
		},
		s.spellAtkBonus: {
			command.LeftDirection: To(s.spellSaveDC),
			command.DownDirection: To(s.spellList),
		},
		s.spellList: {
			command.UpDirection:   To(s.spellAbility),
			command.LeftDirection: Emit(command.ReturnFocusToParentCmd),
		},
	}, s.spellAbility)
}

func (s *SpellScreen) populateSpells() {
	sections := make([]list.Section, 0, 10)
	for level := range 10 {
		sec := s.perLevelRows[level].Section()
		sec.Header = s.newSpellHeaderRow(level)
		sections = append(sections, sec)
	}
	s.spellList.WithSections(sections)
}

func (s *SpellScreen) createSpellEditors(spell *models.SpellTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(s.keymap, "Name", &spell.Name),
		editor.NewStringEditor(s.keymap, "School", &spell.School),
		editor.NewEnumEditor(s.keymap, styles.BinarySymbols, "Prepared", &spell.Prepared),
		editor.NewEnumEditor(s.keymap, styles.BinarySymbols, "Concentration", &spell.Concentration),
		editor.NewEnumEditor(s.keymap, styles.BinarySymbols, "Ritual", &spell.Ritual),
		editor.NewEnumEditor(s.keymap, styles.SpellSourceStrings, "Spell Source", &spell.SpellSource),
		editor.NewStringEditor(s.keymap, "Damage", &spell.Damage),
		editor.NewStringEditor(s.keymap, "Casting Time", &spell.CastingTime),
		editor.NewStringEditor(s.keymap, "Range", &spell.Range),
		editor.NewStringEditor(s.keymap, "Duration", &spell.Duration),
		editor.NewStringEditor(s.keymap, "Components", &spell.Components),
		editor.NewTextEditor(s.keymap, "Description", &spell.Description),
	}
}

func (s *SpellScreen) RenderSpellScreenTopBar() string {
	separator := styles.GrayTextStyle.Width(8).Render(styles.MakeVerticalSeparator(1))
	return styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			styles.ForceWidth(s.spellAbility.View().Content, spellTopBarElemWidth),
			separator,
			styles.ForceWidth(s.spellSaveDC.View().Content, spellTopBarElemWidth),
			separator,
			styles.ForceWidth(s.spellAtkBonus.View().Content, spellTopBarElemWidth)))
}

type SpellListHeader struct {
	level int
	slots *int
	used  *int
}

func (s *SpellScreen) newSpellHeaderRow(l int) *list.StructRow[SpellListHeader] {
	return list.NewStructRow(s.keymap,
		&SpellListHeader{l, &s.character.Character.SpellSlots[l], &s.character.Character.SpellSlotsUsed[l]},
		renderSpellHeaderRow,
		[]editor.ValueEditor{
			editor.NewIntEditor(s.keymap, "Used Spell Slots", &s.character.Character.SpellSlotsUsed[l]),
			editor.NewIntEditor(s.keymap, "Max Spell Slots", &s.character.Character.SpellSlots[l]),
		}).WithCycleAction(cycleSpellSlots)
}

func cycleSpellSlots(h *SpellListHeader) tea.Cmd {
	if *h.slots <= 0 {
		return nil
	}
	*h.used = (*h.used + 1) % (*h.slots + 1)
	return command.WriteBackRequest
}

func renderSpellHeaderRow(h *SpellListHeader) string {
	return fmt.Sprintf("Level %d ∙ %s", h.level,
		styles.PrettySpellSlots(*h.used, *h.slots))
}

func spellSearchText(s *models.SpellTO) string {
	return s.Name + " " + s.School + " " + s.Description
}

func toggleSpellPrepared(spell *models.SpellTO) tea.Cmd {
	spell.Prepared = 1 - spell.Prepared
	return command.WriteBackRequest
}

func renderSpellInfoRow(s *models.SpellTO) string {
	values := []string{s.Name, s.Damage, s.Components, s.Range, s.CastingTime, s.Duration, s.School, styles.SpellSourceSymbols[s.SpellSource].Label}
	if util.I2b(s.Concentration) {
		values = append(values, "C")
	}
	if util.I2b(s.Ritual) {
		values = append(values, "R")
	}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return styles.PrettyBoolCircle(util.I2b(s.Prepared)) + " " + strings.Join(values, " ∙ ")
}

func renderFullSpellInfo(s *models.SpellTO) string {
	innerWidth := styles.SmallScreenWidth - 4
	colWidth := innerWidth / 2
	separator := styles.MakeHorizontalSeparator(innerWidth, 0)

	title := s.Name + " ∙ Level " + strconv.Itoa(s.Level)
	if s.School != "" {
		title += " ∙ " + s.School
	}
	if util.I2b(s.Concentration) {
		title += " [C]"
	}
	if util.I2b(s.Ritual) {
		title += " [R]"
	}
	titleRow := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Left).Render(title)

	pairs := []struct{ label, value string }{
		{"Casting time", s.CastingTime},
		{"Range", s.Range},
		{"Duration", s.Duration},
		{"Components", s.Components},
		{"Damage", s.Damage},
	}
	pairs = util.Filter(pairs, func(p struct{ label, value string }) bool { return p.value != "" })

	var gridRows []string
	for i := 0; i < len(pairs); i += 2 {
		left := styles.ForceWidth(pairs[i].label+": "+pairs[i].value, colWidth)
		var right string
		if i+1 < len(pairs) {
			right = styles.ForceWidth(pairs[i+1].label+": "+pairs[i+1].value, colWidth)
		}
		gridRows = append(gridRows, lipgloss.JoinHorizontal(lipgloss.Top, left, right))
	}
	grid := strings.Join(gridRows, "\n")

	sections := []string{titleRow, separator}
	if grid != "" {
		sections = append(sections, grid, separator)
	}
	sections = append(sections, s.Description)

	return styles.DefaultTextStyle.
		AlignHorizontal(lipgloss.Left).
		Render(strings.Join(sections, "\n"))
}
