package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
)

var (
	TopBarHeight = 6

	TopSeparatorWidth = 20

	ColHeight    = 25
	LeftColWidth = 30
	MidColWidth  = 28

	RightColWidth     = 38
	RightContentWidth = RightColWidth - 6

	LongColWidth   = 20
	ColWidth       = 16
	MediumColWidth = 12
	ShortColWidth  = 8
	TinyColWidth   = 3
)

type StatScreen struct {
	keymap             KeyMap
	character          *models.Character
	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	characterInfo *List
	abilities     *List
	skills        *List
	savingThrows  *List
	combatInfo    *List
	attacks       *List
	actions       *SimpleStringComponent
	bonusActions  *SimpleStringComponent
}

func NewStatScreen(keymap KeyMap, c *models.Character) *StatScreen {
	return &StatScreen{
		keymap:    keymap,
		character: c,
		characterInfo: NewListWithDefaults().
			WithRows(GetCharacterInfoRows(keymap, c)),
		abilities: NewListWithDefaults().
			WithRows(GetAbilityRows(keymap, c)),
		skills: NewListWithDefaults().
			WithTitle("Skills").
			WithRows(GetSkillRows(keymap, c)),
		savingThrows: NewListWithDefaults().
			WithTitle("Saving Throws").
			WithRows(GetSavingThrowRows(keymap, c)),
		combatInfo: NewListWithDefaults().
			WithTitle("Combat").
			WithRows(GetCombatInfoRows(keymap, c)),
		attacks: NewListWithDefaults().
			WithTitle("Attacks").
			WithRows(GetAttackRows(keymap, c)).
			WithAppender(),
		actions:      NewSimpleStringComponent(keymap, "Actions", &c.Actions),
		bonusActions: NewSimpleStringComponent(keymap, "Bonus Actions", &c.BonusActions),
	}
}

func (s *StatScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.characterInfo.Init())
	cmds = append(cmds, s.abilities.Init())
	cmds = append(cmds, s.skills.Init())
	cmds = append(cmds, s.savingThrows.Init())
	cmds = append(cmds, s.combatInfo.Init())
	cmds = append(cmds, s.attacks.Init())
	cmds = append(cmds, s.actions.Init())
	cmds = append(cmds, s.bonusActions.Init())

	s.lastFocusedElement = s.characterInfo
	s.focusOn(s.characterInfo)

	cmds = Filter(cmds, func(c tea.Cmd) bool { return c != nil })
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}

	return nil
}

func (s *StatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case AppendElementMsg:
		switch s.focusedElement {
		case s.attacks:
			attack := models.Attack{}
			s.character.Attacks = append(s.character.Attacks, attack)
			newRows := GetAttackRows(s.keymap, s.character)
			s.attacks.WithRows(newRows)
			cmd = SwitchToEditorCmd(StatScreenIndex, s.character, newRows[len(newRows)-1].Editors())
		}
	case FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case EditValueMsg:
		cmd = SwitchToEditorCmd(StatScreenIndex, s.character, msg.Editors)
	case tea.KeyMsg:
		switch s.focusedElement.(type) {
		case *List:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(LeftDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		default:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(LeftDirection)
			case key.Matches(msg, s.keymap.Up):
				cmd = s.moveFocus(UpDirection)
			case key.Matches(msg, s.keymap.Down):
				cmd = s.moveFocus(DownDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		}
	}
	return s, cmd
}

func (s *StatScreen) Focus() {
	s.focusOn(s.lastFocusedElement)
}

func (s *StatScreen) Blur() {
	// blur should be idempotent
	if s.focusedElement != nil {
		s.lastFocusedElement = s.focusedElement
	}
	s.focusedElement = nil
	s.characterInfo.Blur()
	s.abilities.Blur()
	s.skills.Blur()
	s.savingThrows.Blur()
	s.combatInfo.Blur()
	s.attacks.Blur()
	s.actions.Blur()
	s.bonusActions.Blur()
}

func (s *StatScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *StatScreen) moveFocus(d Direction) tea.Cmd {
	var cmd tea.Cmd
	s.Blur()

	switch s.lastFocusedElement {
	case s.characterInfo:
		switch d {
		case DownDirection:
			s.focusOn(s.skills)
		case RightDirection:
			s.focusOn(s.abilities)
		case LeftDirection:
			cmd = ReturnFocusToParentCmd
		default:
			s.focusOn(s.characterInfo)
		}
	case s.abilities:
		switch d {
		case DownDirection:
			s.focusOn(s.actions)
		case LeftDirection:
			s.focusOn(s.characterInfo)
		default:
			s.focusOn(s.abilities)
		}
	case s.skills:
		switch d {
		case UpDirection:
			s.focusOn(s.characterInfo)
		case RightDirection:
			if s.skills.cursor < len(s.skills.content)/2 {
				s.focusOn(s.combatInfo)
			} else {
				s.focusOn(s.savingThrows)
			}
		case LeftDirection:
			cmd = ReturnFocusToParentCmd
		default:
			s.focusOn(s.skills)
		}
	case s.combatInfo:
		switch d {
		case UpDirection:
			s.focusOn(s.characterInfo)
		case RightDirection:
			s.focusOn(s.actions)
		case DownDirection:
			s.focusOn(s.savingThrows)
		case LeftDirection:
			s.focusOn(s.skills)
			s.skills.SetCursor(0)
		}
	case s.savingThrows:
		switch d {
		case UpDirection:
			s.focusOn(s.combatInfo)
		case RightDirection:
			s.focusOn(s.attacks)
		case LeftDirection:
			s.focusOn(s.skills)
			s.skills.SetCursor(len(s.skills.content) / 2)
		default:
			s.focusOn(s.savingThrows)
		}
	case s.actions:
		switch d {
		case UpDirection:
			s.focusOn(s.abilities)
		case LeftDirection:
			s.focusOn(s.combatInfo)
		case DownDirection:
			s.focusOn(s.bonusActions)
		default:
			s.focusOn(s.actions)
		}
	case s.bonusActions:
		switch d {
		case UpDirection:
			s.focusOn(s.actions)
		case LeftDirection:
			s.focusOn(s.combatInfo)
		case DownDirection:
			s.focusOn(s.attacks)
		default:
			s.focusOn(s.bonusActions)
		}
	case s.attacks:
		switch d {
		case UpDirection:
			s.focusOn(s.bonusActions)
		case LeftDirection:
			s.focusOn(s.savingThrows)
		default:
			s.focusOn(s.attacks)
		}
	}
	return cmd
}

func (s *StatScreen) View() string {
	characterInfo := s.characterInfo.View()

	abilities := s.abilities.View()

	topBarSeparator := MakeVerticalSeparator(TopBarHeight)

	topBar := DefaultBorderStyle.
		Height(TopBarHeight).
		Width(ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			characterInfo,
			lipgloss.PlaceHorizontal(20, lipgloss.Center, topBarSeparator),
			abilities))

	leftColumn := DefaultBorderStyle.
		Height(ColHeight).
		Width(LeftColWidth).
		Render(s.skills.View())

	savingThrows := s.savingThrows.View()

	combatInfo := s.combatInfo.View()

	midBoxInnerSeparator := MakeHorizontalSeparator(MidColWidth - 4)

	midColumn := DefaultBorderStyle.
		Width(MidColWidth).
		Height(ColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, combatInfo, midBoxInnerSeparator, savingThrows))

	actions := s.RenderActions()

	attacks := s.attacks.View()

	rightBoxInnerSeparator := MakeHorizontalSeparator(RightContentWidth)

	rightColumn := DefaultBorderStyle.
		Width(RightColWidth).
		Height(ColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, actions, rightBoxInnerSeparator, attacks))

	body := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn, midColumn, rightColumn)

	return lipgloss.JoinVertical(lipgloss.Center, topBar, body)
}

func GetCharacterInfoRows(k KeyMap, c *models.Character) []Row {
	rowCfg := LabeledStringRowConfig{false, LongColWidth, 0}
	rows := []Row{
		NewLabeledStringRow(k, "Name:", &c.Name,
			NewStringEditor(k, "Name", &c.Name)).WithConfig(rowCfg),
		NewLabeledStringRow(k, "Levels:", &c.ClassLevels,
			NewStringEditor(k, "Levels", &c.ClassLevels)).WithConfig(rowCfg),
		NewLabeledStringRow(k, "Race:", &c.Race,
			NewStringEditor(k, "Race", &c.Race)).WithConfig(rowCfg),
		NewLabeledStringRow(k, "Alignment:", &c.Alignment,
			NewStringEditor(k, "Alignment", &c.Alignment)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Proficiency Bonus:", &c.ProficiencyBonus,
			NewIntEditor(k, "Proficiency Bonus", &c.ProficiencyBonus)).
			WithConfig(LabeledIntRowConfig{func(i int) string { return fmt.Sprintf("%+d", i) }, false, LongColWidth, 0}),
	}
	return rows
}

func GetAbilityRows(k KeyMap, c *models.Character) []Row {
	scorePrinter := func(score int) string {
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), models.ToModifier(score))
	}
	rowCfg := LabeledIntRowConfig{scorePrinter, true, ColWidth, ShortColWidth}
	rows := []Row{
		NewLabeledIntRow(k, "Strength:", &c.Abilities.Strength,
			NewIntEditor(k, "Strength", &c.Abilities.Strength)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Constitution:", &c.Abilities.Constitution,
			NewIntEditor(k, "Constitution", &c.Abilities.Constitution)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Dexterity:", &c.Abilities.Dexterity,
			NewIntEditor(k, "Dexterity", &c.Abilities.Dexterity)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Intelligence:", &c.Abilities.Intelligence,
			NewIntEditor(k, "Intelligence", &c.Abilities.Intelligence)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Wisdom:", &c.Abilities.Wisdom,
			NewIntEditor(k, "Wisdom", &c.Abilities.Wisdom)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Charisma:", &c.Abilities.Charisma,
			NewIntEditor(k, "Charisma", &c.Abilities.Charisma)).WithConfig(rowCfg),
	}
	return rows
}

func GetCombatInfoRows(k KeyMap, c *models.Character) []Row {
	standardCfg := LabeledIntRowConfig{strconv.Itoa, true, ColWidth, TinyColWidth}
	dsConfig := LabeledIntRowConfig{DeathSaveSymbols, true, ColWidth, TinyColWidth}
	rows := []Row{
		NewLabeledIntRow(k, "AC", &c.ArmorClass,
			NewIntEditor(k, "AC", &c.ArmorClass)).WithConfig(standardCfg),
		NewLabeledIntRow(k, "Initiative", &c.Initiative,
			NewIntEditor(k, "Initiative", &c.Initiative)).
			WithConfig(LabeledIntRowConfig{func(i int) string { return fmt.Sprintf("%+d", i) }, true, ColWidth, TinyColWidth}),
		NewLabeledIntRow(k, "Speed", &c.Speed,
			NewIntEditor(k, "Speed", &c.Speed)).WithConfig(standardCfg),
		NewStructRow(k, &HPInfo{&c.CurrentHitPoints, &c.MaxHitPoints}, renderHPInfoRow,
			[]ValueEditor{
				NewIntEditor(k, "Current HP", &c.CurrentHitPoints),
				NewIntEditor(k, "Max HP", &c.MaxHitPoints),
			}),
		NewStructRow(k, &HitDiceInfo{&c.UsedHitDice, &c.HitDice}, renderHitDiceInfoRow,
			[]ValueEditor{
				NewStringEditor(k, "Used Hit Dice", &c.UsedHitDice),
				NewStringEditor(k, "Hit Dice", &c.HitDice),
			}),
		NewLabeledIntRow(k, "DS Successes", &c.DeathSaves.Successes,
			NewIntEditor(k, "DS Successes", &c.DeathSaves.Successes)).WithConfig(dsConfig),
		NewLabeledIntRow(k, "DS Failures", &c.DeathSaves.Failures,
			NewIntEditor(k, "DS Failures", &c.DeathSaves.Failures)).WithConfig(dsConfig),
	}
	return rows
}

func (s *StatScreen) RenderActions() string {
	actionTitle := RenderItem(s.actions.InFocus(), "Actions")

	actionBody := DefaultTextStyle.Width(RightContentWidth).Render(s.actions.View())

	separator := MakeHorizontalSeparator(RightContentWidth)

	bonusActionTitle := RenderItem(s.bonusActions.InFocus(), "Bonus Actions")

	bonusActionBody := DefaultTextStyle.Width(RightContentWidth).Render(s.bonusActions.View())

	return lipgloss.JoinVertical(lipgloss.Center, actionTitle, actionBody, separator, bonusActionTitle, bonusActionBody)
}

func GetAttackRows(k KeyMap, c *models.Character) []Row {
	rows := []Row{}
	for i := range c.Attacks {
		a := &c.Attacks[i]
		row := NewStructRow(k, a, RenderAttack, []ValueEditor{
			NewStringEditor(k, "Name", &a.Name),
			NewIntEditor(k, "Bonus", &a.Bonus),
			NewStringEditor(k, "Damage", &a.Damage),
			NewStringEditor(k, "Damage Type", &a.DamageType),
		})
		rows = append(rows, row)
	}
	return rows
}

func GetSkillRows(k KeyMap, c *models.Character) []Row {
	rows := []Row{}

	for i := range c.Skills {
		skill := &c.Skills[i]
		row := NewStructRow(k, &SkillInfo{skill, &c.Abilities, &c.ProficiencyBonus}, renderSkillInfoRow,
			[]ValueEditor{
				NewEnumEditor(k, ProficiencySymbols, "Proficiency", &skill.Proficiency),
				NewIntEditor(k, "Custom Modifier", &skill.CustomModifier),
			})
		rows = append(rows, row)
	}

	return rows
}

func GetSavingThrowRows(k KeyMap, c *models.Character) []Row {
	rows := []Row{}

	for i := range c.SavingThrows {
		saving := &c.SavingThrows[i]
		row := NewStructRow(k, &SavingThrowInfo{saving, &c.Abilities, &c.ProficiencyBonus}, renderSavingThrowInfoRow,
			[]ValueEditor{NewEnumEditor(k, ProficiencySymbols, "Proficiency", &saving.Proficiency)})
		rows = append(rows, row)
	}

	return rows
}

// screen specific types + utility functions

type HPInfo struct {
	current *int
	max     *int
}

func renderHPInfoRow(hp *HPInfo) string {
	return RenderEdgeBound(ColWidth-4, 7, "HP", strconv.Itoa(*hp.current)+"/"+strconv.Itoa(*hp.max))
}

type HitDiceInfo struct {
	current *string
	max     *string
}

func renderHitDiceInfoRow(hd *HitDiceInfo) string {
	return RenderEdgeBound(ShortColWidth, MediumColWidth, "Hit Dice", *hd.current+"/"+*hd.max)
}

type SavingThrowInfo struct {
	savingThrow *models.SavingThrow
	abilities   *models.Abilities
	profBonus   *int
}

func renderSavingThrowInfoRow(st *SavingThrowInfo) string {
	mod := st.savingThrow.ToModifier(*st.abilities, *st.profBonus)
	bullet := ProficiencySymbol(st.savingThrow.Proficiency)
	return RenderEdgeBound(ColWidth, TinyColWidth, bullet+" "+st.savingThrow.Ability, fmt.Sprintf("%+d", mod))
}

type SkillInfo struct {
	skill     *models.Skill
	abilities *models.Abilities
	profBonus *int
}

func renderSkillInfoRow(s *SkillInfo) string {
	mod := s.skill.ToModifier(*s.abilities, *s.profBonus)
	bullet := ProficiencySymbol(s.skill.Proficiency)
	return RenderEdgeBound(LongColWidth, TinyColWidth, bullet+" "+s.skill.Name, fmt.Sprintf("%+d", mod))
}

func RenderAttack(a *models.Attack) string {
	return fmt.Sprintf("%-11s %+3d %s (%s)", a.Name, a.Bonus, a.Damage, a.DamageType)
}

func DeathSaveSymbols(amount int) string {
	return strings.Repeat("●", amount) + strings.Repeat("○", 3-amount)
}

var ProficiencySymbols []EnumMapping = []EnumMapping{
	{int(models.NoProficiency), "○"},
	{int(models.Proficient), "◐"},
	{int(models.Expertise), "●"},
}

func ProficiencySymbol(p models.ProficiencyLevel) string {
	for _, m := range ProficiencySymbols {
		if int(p) == m.Value {
			return m.Label
		}
	}
	return ""
}
