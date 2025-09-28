package screen

import (
	"fmt"
	"strconv"
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
	keymap             util.KeyMap
	character          *models.Character
	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	characterInfo *list.List
	abilities     *list.List
	skills        *list.List
	savingThrows  *list.List
	combatInfo    *list.List
	attacks       *list.List
	actions       *component.SimpleStringComponent
	bonusActions  *component.SimpleStringComponent
}

func NewStatScreen(keymap util.KeyMap, c *models.Character) *StatScreen {
	return &StatScreen{
		keymap:    keymap,
		character: c,
		characterInfo: list.NewListWithDefaults().
			WithRows(GetCharacterInfoRows(keymap, c)),
		abilities: list.NewListWithDefaults().
			WithRows(GetAbilityRows(keymap, c)),
		skills: list.NewListWithDefaults().
			WithTitle("Skills").
			WithRows(GetSkillRows(keymap, c)),
		savingThrows: list.NewListWithDefaults().
			WithTitle("Saving Throws").
			WithRows(GetSavingThrowRows(keymap, c)),
		combatInfo: list.NewListWithDefaults().
			WithTitle("Combat").
			WithRows(GetCombatInfoRows(keymap, c)),
		attacks: list.NewListWithDefaults().
			WithTitle("Attacks").
			WithRows(GetAttackRows(keymap, c)),
		actions:      component.NewSimpleStringComponent(keymap, "Actions", &c.Actions, false, false),
		bonusActions: component.NewSimpleStringComponent(keymap, "Bonus Actions", &c.BonusActions, false, false),
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

	cmds = util.Filter(cmds, func(c tea.Cmd) bool { return c != nil })
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}

	return nil
}

func (s *StatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if msg.Tag == "attack" {
			s.character.AddEmptyAttack()
			newRows := GetAttackRows(s.keymap, s.character)
			s.attacks.WithRows(newRows)
			cmd = editor.SwitchToEditorCmd(command.StatScreenIndex, s.character, newRows[len(newRows)-1].Editors())
		} else {
			_, cmd = s.focusedElement.Update(msg)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case editor.EditValueMsg:
		cmd = editor.SwitchToEditorCmd(command.StatScreenIndex, s.character, msg.Editors)
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

func (s *StatScreen) moveFocus(d command.Direction) tea.Cmd {
	var cmd tea.Cmd
	s.Blur()

	switch s.lastFocusedElement {
	case s.characterInfo:
		switch d {
		case command.DownDirection:
			s.focusOn(s.skills)
		case command.RightDirection:
			s.focusOn(s.abilities)
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		default:
			s.focusOn(s.characterInfo)
		}
	case s.abilities:
		switch d {
		case command.DownDirection:
			s.focusOn(s.actions)
		case command.LeftDirection:
			s.focusOn(s.characterInfo)
		default:
			s.focusOn(s.abilities)
		}
	case s.skills:
		switch d {
		case command.UpDirection:
			s.focusOn(s.characterInfo)
		case command.RightDirection:
			if s.skills.CursorPos() < s.skills.Size()/2 {
				s.focusOn(s.combatInfo)
			} else {
				s.focusOn(s.savingThrows)
			}
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		default:
			s.focusOn(s.skills)
		}
	case s.combatInfo:
		switch d {
		case command.UpDirection:
			s.focusOn(s.characterInfo)
		case command.RightDirection:
			s.focusOn(s.actions)
		case command.DownDirection:
			s.focusOn(s.savingThrows)
		case command.LeftDirection:
			s.focusOn(s.skills)
			s.skills.SetCursor(0)
		}
	case s.savingThrows:
		switch d {
		case command.UpDirection:
			s.focusOn(s.combatInfo)
		case command.RightDirection:
			s.focusOn(s.attacks)
		case command.LeftDirection:
			s.focusOn(s.skills)
			s.skills.SetCursor(s.skills.Size() / 2)
		default:
			s.focusOn(s.savingThrows)
		}
	case s.actions:
		switch d {
		case command.UpDirection:
			s.focusOn(s.abilities)
		case command.LeftDirection:
			s.focusOn(s.combatInfo)
		case command.DownDirection:
			s.focusOn(s.bonusActions)
		default:
			s.focusOn(s.actions)
		}
	case s.bonusActions:
		switch d {
		case command.UpDirection:
			s.focusOn(s.actions)
		case command.LeftDirection:
			s.focusOn(s.combatInfo)
		case command.DownDirection:
			s.focusOn(s.attacks)
		default:
			s.focusOn(s.bonusActions)
		}
	case s.attacks:
		switch d {
		case command.UpDirection:
			s.focusOn(s.bonusActions)
		case command.LeftDirection:
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

	topBarSeparator := util.MakeVerticalSeparator(TopBarHeight)

	topBar := util.DefaultBorderStyle.
		Height(TopBarHeight).
		Width(util.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			characterInfo,
			lipgloss.PlaceHorizontal(20, lipgloss.Center, topBarSeparator),
			abilities))

	leftColumn := util.DefaultBorderStyle.
		Height(ColHeight).
		Width(LeftColWidth).
		Render(s.skills.View())

	savingThrows := s.savingThrows.View()

	combatInfo := s.combatInfo.View()

	midBoxInnerSeparator := util.MakeHorizontalSeparator(MidColWidth-4, 1)

	midColumn := util.DefaultBorderStyle.
		Width(MidColWidth).
		Height(ColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, combatInfo, midBoxInnerSeparator, savingThrows))

	actions := s.RenderActions()

	attacks := s.attacks.View()

	rightBoxInnerSeparator := util.MakeHorizontalSeparator(RightContentWidth, 1)

	rightColumn := util.DefaultBorderStyle.
		Width(RightColWidth).
		Height(ColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, actions, rightBoxInnerSeparator, attacks))

	body := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn, midColumn, rightColumn)

	return lipgloss.JoinVertical(lipgloss.Center, topBar, body)
}

func GetCharacterInfoRows(k util.KeyMap, c *models.Character) []list.Row {
	rowCfg := list.LabeledStringRowConfig{JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0}
	rows := []list.Row{
		list.NewLabeledStringRow(k, "Name:", &c.Name,
			editor.NewStringEditor(k, "Name", &c.Name)).WithConfig(rowCfg),
		list.NewLabeledStringRow(k, "Levels:", &c.ClassLevels,
			editor.NewStringEditor(k, "Levels", &c.ClassLevels)).WithConfig(rowCfg),
		list.NewLabeledStringRow(k, "Race:", &c.Race,
			editor.NewStringEditor(k, "Race", &c.Race)).WithConfig(rowCfg),
		list.NewLabeledStringRow(k, "Alignment:", &c.Alignment,
			editor.NewStringEditor(k, "Alignment", &c.Alignment)).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Proficiency Bonus:", &c.ProficiencyBonus,
			editor.NewIntEditor(k, "Proficiency Bonus", &c.ProficiencyBonus)).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0,
			}),
	}
	return rows
}

func GetAbilityRows(k util.KeyMap, c *models.Character) []list.Row {
	scorePrinter := func(score int) string {
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), models.ToModifier(score))
	}
	rowCfg := list.LabeledIntRowConfig{ValuePrinter: scorePrinter, JustifyValue: true, LabelWidth: ColWidth, ValueWidth: ShortColWidth}
	rows := []list.Row{
		list.NewLabeledIntRow(k, "Strength:", &c.Abilities.Strength,
			editor.NewIntEditor(k, "Strength", &c.Abilities.Strength)).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Constitution:", &c.Abilities.Constitution,
			editor.NewIntEditor(k, "Constitution", &c.Abilities.Constitution)).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Dexterity:", &c.Abilities.Dexterity,
			editor.NewIntEditor(k, "Dexterity", &c.Abilities.Dexterity)).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Intelligence:", &c.Abilities.Intelligence,
			editor.NewIntEditor(k, "Intelligence", &c.Abilities.Intelligence)).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Wisdom:", &c.Abilities.Wisdom,
			editor.NewIntEditor(k, "Wisdom", &c.Abilities.Wisdom)).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Charisma:", &c.Abilities.Charisma,
			editor.NewIntEditor(k, "Charisma", &c.Abilities.Charisma)).WithConfig(rowCfg),
	}
	return rows
}

func GetCombatInfoRows(k util.KeyMap, c *models.Character) []list.Row {
	standardCfg := list.LabeledIntRowConfig{
		ValuePrinter: strconv.Itoa, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	dsConfig := list.LabeledIntRowConfig{
		ValuePrinter: DeathSaveSymbols, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	rows := []list.Row{
		list.NewLabeledIntRow(k, "AC", &c.ArmorClass,
			editor.NewIntEditor(k, "AC", &c.ArmorClass)).WithConfig(standardCfg),
		list.NewLabeledIntRow(k, "Initiative", &c.Initiative,
			editor.NewIntEditor(k, "Initiative", &c.Initiative)).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: true, LabelWidth: ColWidth, ValueWidth: TinyColWidth,
			}),
		list.NewLabeledIntRow(k, "Speed", &c.Speed,
			editor.NewIntEditor(k, "Speed", &c.Speed)).WithConfig(standardCfg),
		list.NewStructRow(k, &HPInfo{&c.CurrentHitPoints, &c.MaxHitPoints}, renderHPInfoRow,
			[]editor.ValueEditor{
				editor.NewIntEditor(k, "Current HP", &c.CurrentHitPoints),
				editor.NewIntEditor(k, "Max HP", &c.MaxHitPoints),
			}),
		list.NewStructRow(k, &HitDiceInfo{&c.UsedHitDice, &c.HitDice}, renderHitDiceInfoRow,
			[]editor.ValueEditor{
				editor.NewStringEditor(k, "Used Hit Dice", &c.UsedHitDice),
				editor.NewStringEditor(k, "Hit Dice", &c.HitDice),
			}),
		list.NewLabeledIntRow(k, "DS Successes", &c.DeathSaves.Successes,
			editor.NewIntEditor(k, "DS Successes", &c.DeathSaves.Successes)).WithConfig(dsConfig),
		list.NewLabeledIntRow(k, "DS Failures", &c.DeathSaves.Failures,
			editor.NewIntEditor(k, "DS Failures", &c.DeathSaves.Failures)).WithConfig(dsConfig),
	}
	return rows
}

func (s *StatScreen) RenderActions() string {
	actionTitle := util.RenderItem(s.actions.InFocus(), "Actions") + "\n"

	actionBody := util.DefaultTextStyle.Width(RightContentWidth).Render(s.actions.View())

	separator := util.MakeHorizontalSeparator(RightContentWidth, 1)

	bonusActionTitle := util.RenderItem(s.bonusActions.InFocus(), "Bonus Actions") + "\n"

	bonusActionBody := util.DefaultTextStyle.Width(RightContentWidth).Render(s.bonusActions.View())

	return lipgloss.JoinVertical(lipgloss.Center, actionTitle, actionBody, separator, bonusActionTitle, bonusActionBody)
}

func GetAttackRows(k util.KeyMap, c *models.Character) []list.Row {
	rows := []list.Row{}
	for i := range c.Attacks {
		a := &c.Attacks[i]
		row := list.NewStructRow(k, a, RenderAttack, []editor.ValueEditor{
			editor.NewStringEditor(k, "Name", &a.Name),
			editor.NewIntEditor(k, "Bonus", &a.Bonus),
			editor.NewStringEditor(k, "Damage", &a.Damage),
			editor.NewStringEditor(k, "Damage Type", &a.DamageType),
		})
		rows = append(rows, row)
	}
	rows = append(rows, list.NewAppenderRow(k, "attack"))
	return rows
}

func GetSkillRows(k util.KeyMap, c *models.Character) []list.Row {
	rows := []list.Row{}

	for i := range c.Skills {
		skill := &c.Skills[i]
		row := list.NewStructRow(k, &SkillInfo{skill, &c.Abilities, &c.ProficiencyBonus}, renderSkillInfoRow,
			[]editor.ValueEditor{
				editor.NewEnumEditor(k, ProficiencySymbols, "Proficiency", &skill.Proficiency),
				editor.NewIntEditor(k, "Custom Modifier", &skill.CustomModifier),
			})
		rows = append(rows, row)
	}

	return rows
}

func GetSavingThrowRows(k util.KeyMap, c *models.Character) []list.Row {
	rows := []list.Row{}

	for i := range c.SavingThrows {
		saving := &c.SavingThrows[i]
		row := list.NewStructRow(k, &SavingThrowInfo{saving, &c.Abilities, &c.ProficiencyBonus}, renderSavingThrowInfoRow,
			[]editor.ValueEditor{editor.NewEnumEditor(k, ProficiencySymbols, "Proficiency", &saving.Proficiency)})
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
	return util.RenderEdgeBound(ColWidth-4, 7, "HP", strconv.Itoa(*hp.current)+"/"+strconv.Itoa(*hp.max))
}

type HitDiceInfo struct {
	current *string
	max     *string
}

func renderHitDiceInfoRow(hd *HitDiceInfo) string {
	return util.RenderEdgeBound(ShortColWidth, MediumColWidth, "Hit Dice", *hd.current+"/"+*hd.max)
}

type SavingThrowInfo struct {
	savingThrow *models.SavingThrow
	abilities   *models.Abilities
	profBonus   *int
}

func renderSavingThrowInfoRow(st *SavingThrowInfo) string {
	mod := st.savingThrow.ToModifier(*st.abilities, *st.profBonus)
	bullet := ProficiencySymbol(st.savingThrow.Proficiency)
	return util.RenderEdgeBound(ColWidth, TinyColWidth, bullet+" "+st.savingThrow.Ability, fmt.Sprintf("%+d", mod))
}

type SkillInfo struct {
	skill     *models.Skill
	abilities *models.Abilities
	profBonus *int
}

func renderSkillInfoRow(s *SkillInfo) string {
	mod := s.skill.ToModifier(*s.abilities, *s.profBonus)
	bullet := ProficiencySymbol(s.skill.Proficiency)
	return util.RenderEdgeBound(LongColWidth, TinyColWidth, bullet+" "+s.skill.Name, fmt.Sprintf("%+d", mod))
}

func RenderAttack(a *models.Attack) string {
	return fmt.Sprintf("%-11s %+3d %s (%s)", a.Name, a.Bonus, a.Damage, a.DamageType)
}

func DeathSaveSymbols(amount int) string {
	return strings.Repeat("●", amount) + strings.Repeat("○", 3-amount)
}

var ProficiencySymbols []editor.EnumMapping = []editor.EnumMapping{
	{Value: int(models.NoProficiency), Label: "○"},
	{Value: int(models.Proficient), Label: "◐"},
	{Value: int(models.Expertise), Label: "●"},
}

func ProficiencySymbol(p models.ProficiencyLevel) string {
	for _, m := range ProficiencySymbols {
		if int(p) == m.Value {
			return m.Label
		}
	}
	return ""
}
