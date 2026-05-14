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

	spellRows *CollectionRows[models.SpellTO]
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
			WithViewport(spellColHeight - 2),
	}
	s.spellRows = NewCustomCollectionRows(s.spellList,
		func(sp *models.SpellTO) uuid.UUID { return sp.ID },
		func(tag string) uuid.UUID {
			l, _ := strconv.Atoi(strings.Split(tag, ":")[1])
			return s.character.AddEmptySpell(l)
		},
		s.character.DeleteSpell,
	)
	s.spellRows.Repopulate = s.populateSpells
	return s
}

func (s *SpellScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.spellAbility.Init())
	cmds = append(cmds, s.spellSaveDC.Init())
	cmds = append(cmds, s.spellAtkBonus.Init())
	cmds = util.DropNil(cmds)
	s.lastFocusedElement = s.spellAbility

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
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "spell:") {
			cmd = s.spellRows.HandleAppend(msg.Tag)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case tea.KeyPressMsg:
		cmd = RouteKey(s.focusedElement, msg, s.keymap, s.moveFocus)
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
	s.focusGraph = FocusGraph{
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
	}
}

func (s *SpellScreen) populateSpells() {
	rows := []list.Row{}
	for i := range 10 {
		rows = append(rows, s.getSpellListByLevel(i)...)
	}
	s.spellList.WithRows(rows[:len(rows)-1]) // drop last separator row
}

func (s *SpellScreen) getSpellListByLevel(l int) []list.Row {
	rows := []list.Row{}
	spells := s.character.GetSpellsByLevel(l)
	rows = append(rows, s.newSpellHeaderRow(l))
	rows = append(rows, list.NewSeparatorRow("─", spellColWidth-6))
	for _, spell := range spells {
		rows = append(rows, list.NewStructRow(s.keymap, spell,
			renderSpellInfoRow,
			s.createSpellEditors(spell),
		).WithDestructor(s.spellRows.DeleteCallback(spell.ID)).
			WithReader(renderFullSpellInfo))
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, fmt.Sprintf("spell:%d", l)))
	rows = append(rows, list.NewSeparatorRow(" ", spellColWidth-6))
	return rows
}

func (s *SpellScreen) createSpellEditors(spell *models.SpellTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(s.keymap, "Name", &spell.Name),
		editor.NewStringEditor(s.keymap, "School", &spell.School),
		editor.NewEnumEditor(s.keymap, models.PreparedSymbols, "Prepared", &spell.Prepared),
		editor.NewEnumEditor(s.keymap, models.SpellSourceStrings, "Spell Source", &spell.SpellSource),
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

func renderSpellInfoRow(s *models.SpellTO) string {
	values := []string{s.Name, s.Damage, s.Components, s.Range, s.CastingTime, s.Duration, s.School, models.SpellSourceSymbols[s.SpellSource].Label}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return styles.PrettyBoolCircle(util.I2b(s.Prepared)) + " " + strings.Join(values, " ∙ ")
}

func renderFullSpellInfo(s *models.SpellTO) string {
	separator := styles.MakeHorizontalSeparator(styles.SmallScreenWidth-4, 1)
	content := strings.Join(
		[]string{
			s.Name + " ∙  Level: " + strconv.Itoa(s.Level),
			separator,
			"School: " + s.School,
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
	return styles.DefaultTextStyle.
		AlignHorizontal(lipgloss.Left).
		Render(content)
}
