package screen

import (
	"fmt"
	"strconv"

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
	agg                *repository.CharacterAggregate
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

func NewStatScreen(keymap util.KeyMap, c *repository.CharacterAggregate) *StatScreen {
	s := &StatScreen{
		keymap:        keymap,
		agg:           c,
		actions:       component.NewSimpleStringComponent(keymap, "Actions", &c.Character.Actions, false, false),
		bonusActions:  component.NewSimpleStringComponent(keymap, "Bonus Actions", &c.Character.BonusActions, false, false),
		characterInfo: list.NewListWithDefaults(),
		abilities:     list.NewListWithDefaults(),
		skills: list.NewListWithDefaults().
			WithTitle("Skills"),
		savingThrows: list.NewListWithDefaults().
			WithTitle("Saving Throws"),
		combatInfo: list.NewListWithDefaults().
			WithTitle("Combat"),
		attacks: list.NewListWithDefaults().
			WithTitle("Attacks"),
	}
	return s
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

	s.CreateCharacterInfoRows()
	s.CreateAbilityRows()
	s.CreateSkillRows()
	s.CreateCombatInfoRows()
	s.CreateAttackRows()
	s.CreateSavingThrowRows()

	s.lastFocusedElement = s.characterInfo
	s.focusOn(s.characterInfo)

	return tea.Batch(cmds...)
}

func (s *StatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if msg.Tag == "attack" {
			id := s.agg.AddEmptyAttack()
			s.CreateAttackRows()
			cmd = editor.SwitchToEditorCmd(
				s.getAttackRow(id).Editors(),
			)
		} else {
			_, cmd = s.focusedElement.Update(msg)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
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

func (s *StatScreen) RenderActions() string {
	actionTitle := util.RenderItem(s.actions.InFocus(), "Actions") + "\n"
	actionBody := util.DefaultTextStyle.Width(RightContentWidth).Render(s.actions.View())

	separator := util.MakeHorizontalSeparator(RightContentWidth, 1)

	bonusActionTitle := util.RenderItem(s.bonusActions.InFocus(), "Bonus Actions") + "\n"
	bonusActionBody := util.DefaultTextStyle.Width(RightContentWidth).Render(s.bonusActions.View())

	return lipgloss.JoinVertical(lipgloss.Center, actionTitle, actionBody, separator, bonusActionTitle, bonusActionBody)
}

func (s *StatScreen) CreateCharacterInfoRows() {
	rowCfg := list.LabeledStringRowConfig{JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0}
	rows := []list.Row{
		list.NewLabeledStringRow(s.keymap, "Name:", &s.agg.Character.Name,
			editor.NewStringEditor(s.keymap, "Name", &s.agg.Character.Name)).WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Levels:", &s.agg.Character.ClassLevels,
			editor.NewStringEditor(s.keymap, "Levels", &s.agg.Character.ClassLevels)).WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Race:", &s.agg.Character.Race,
			editor.NewStringEditor(s.keymap, "Race", &s.agg.Character.Race)).WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Alignment:", &s.agg.Character.Alignment,
			editor.NewStringEditor(s.keymap, "Alignment", &s.agg.Character.Alignment)).WithConfig(rowCfg),
		list.NewLabeledIntRow(s.keymap, "Proficiency Bonus:", &s.agg.Character.ProficiencyBonus,
			editor.NewIntEditor(s.keymap, "Proficiency Bonus", &s.agg.Character.ProficiencyBonus)).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: util.WithSign,
				JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0,
			}),
	}
	s.characterInfo.WithRows(rows)
}

func (s *StatScreen) CreateAbilityRows() {
	scorePrinter := func(score int) string {
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), models.ToModifier(score, 0, 0))
	}
	rowCfg := list.LabeledIntRowConfig{ValuePrinter: scorePrinter, JustifyValue: true, LabelWidth: ColWidth, ValueWidth: ShortColWidth}
	newAbilityRow := func(field *int, name string) list.Row {
		return list.NewLabeledIntRow(s.keymap, name+":", field,
			editor.NewIntEditor(s.keymap, name, field)).WithConfig(rowCfg)
	}
	rows := []list.Row{
		newAbilityRow(&s.agg.Abilities.Strength, "Strength"),
		newAbilityRow(&s.agg.Abilities.Strength, "Constitution"),
		newAbilityRow(&s.agg.Abilities.Strength, "Dexterity"),
		newAbilityRow(&s.agg.Abilities.Strength, "Intelligence"),
		newAbilityRow(&s.agg.Abilities.Strength, "Wisdom"),
		newAbilityRow(&s.agg.Abilities.Strength, "Charisma"),
	}
	s.abilities.WithRows(rows)
}

func (s *StatScreen) CreateCombatInfoRows() {
	standardCfg := list.LabeledIntRowConfig{
		ValuePrinter: strconv.Itoa, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	dsConfig := list.LabeledIntRowConfig{
		ValuePrinter: util.PrettyDeathSaves, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	rows := []list.Row{
		list.NewLabeledIntRow(s.keymap, "AC", &s.agg.Character.ArmorClass,
			editor.NewIntEditor(s.keymap, "AC", &s.agg.Character.ArmorClass)).WithConfig(standardCfg),
		list.NewLabeledIntRow(s.keymap, "Initiative", &s.agg.Character.Initiative,
			editor.NewIntEditor(s.keymap, "Initiative", &s.agg.Character.Initiative)).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: true, LabelWidth: ColWidth, ValueWidth: TinyColWidth,
			}),
		list.NewLabeledIntRow(s.keymap, "Speed", &s.agg.Character.Speed,
			editor.NewIntEditor(s.keymap, "Speed", &s.agg.Character.Speed)).WithConfig(standardCfg),
		list.NewStructRow(s.keymap, &HPInfo{&s.agg.Character.CurrHitPoints, &s.agg.Character.MaxHitPoints}, renderHPInfoRow,
			[]editor.ValueEditor{
				editor.NewIntEditor(s.keymap, "Current HP", &s.agg.Character.CurrHitPoints),
				editor.NewIntEditor(s.keymap, "Max HP", &s.agg.Character.MaxHitPoints),
			}),
		list.NewStructRow(s.keymap, &HitDiceInfo{&s.agg.Character.UsedHitDice, &s.agg.Character.HitDice}, renderHitDiceInfoRow,
			[]editor.ValueEditor{
				editor.NewStringEditor(s.keymap, "Used Hit Dice", &s.agg.Character.UsedHitDice),
				editor.NewStringEditor(s.keymap, "Hit Dice", &s.agg.Character.HitDice),
			}),
		list.NewLabeledIntRow(s.keymap, "DS Successes", &s.agg.Character.DeathSaveSuccesses,
			editor.NewIntEditor(s.keymap, "DS Successes", &s.agg.Character.DeathSaveSuccesses)).WithConfig(dsConfig),
		list.NewLabeledIntRow(s.keymap, "DS Failures", &s.agg.Character.DeathSaveFailures,
			editor.NewIntEditor(s.keymap, "DS Failures", &s.agg.Character.DeathSaveFailures)).WithConfig(dsConfig),
	}
	s.combatInfo.WithRows(rows)
}

func (s *StatScreen) CreateAttackRows() {
	rows := []list.Row{}
	for i := range s.agg.Attacks {
		a := &s.agg.Attacks[i]
		row := list.NewStructRow(s.keymap, a, RenderAttack, []editor.ValueEditor{
			editor.NewStringEditor(s.keymap, "Name", &a.Name),
			editor.NewIntEditor(s.keymap, "Bonus", &a.Bonus),
			editor.NewStringEditor(s.keymap, "Damage", &a.Damage),
			editor.NewStringEditor(s.keymap, "Damage Type", &a.DamageType),
		})
		rows = append(rows, row)
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, "attack"))
	s.attacks.WithRows(rows)
}

func (s *StatScreen) getAttackRow(id uuid.UUID) list.Row {
	for _, r := range s.attacks.Content() {
		switch r := r.(type) {
		case *list.StructRow[models.AttackTO]:
			if r.Value().ID == id {
				return r
			}
		}
	}
	return nil
}

func (s *StatScreen) CreateSkillRows() {
	rows := []list.Row{}

	for i := range s.agg.Skills {
		skill := &s.agg.Skills[i]
		row := list.NewStructRow(s.keymap, &SkillInfo{skill, s.agg.Abilities, &s.agg.Character.ProficiencyBonus}, renderSkillInfoRow,
			[]editor.ValueEditor{
				editor.NewEnumEditor(s.keymap, models.ProficiencySymbols, "Proficiency", &skill.Proficiency),
				editor.NewIntEditor(s.keymap, "Custom Modifier", &skill.CustomModifier),
			})
		rows = append(rows, row)
	}

	s.skills.WithRows(rows)
}

func (s *StatScreen) CreateSavingThrowRows() {
	renderer := renderSavingThrowInfoRow(s.agg.Abilities, s.agg.Character.ProficiencyBonus)
	newSavingThrowRow := func(field *int, name string) list.Row {
		return list.NewStructRow(s.keymap, &SavingThrowInfo{field, name}, renderer,
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, models.ProficiencySymbols, "Proficiency", field)})
	}
	s.savingThrows.WithRows([]list.Row{
		newSavingThrowRow(&s.agg.SavingThrows.StrengthProficiency, "Strength"),
		newSavingThrowRow(&s.agg.SavingThrows.DexterityProficiency, "Dexterity"),
		newSavingThrowRow(&s.agg.SavingThrows.ConstitutionProficiency, "Constitution"),
		newSavingThrowRow(&s.agg.SavingThrows.IntelligenceProficiency, "Intelligence"),
		newSavingThrowRow(&s.agg.SavingThrows.WisdomProficiency, "Wisdom"),
		newSavingThrowRow(&s.agg.SavingThrows.CharismaProficiency, "Charisma"),
	})
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
	proficiency *int
	ability     string
}

func renderSavingThrowInfoRow(a *models.AbilitiesTO, profBonus int) func(*SavingThrowInfo) string {
	return func(s *SavingThrowInfo) string {
		proficiency := models.Proficiency(*s.proficiency)
		mod := models.ToModifier(a.ToScoreByName(
			s.ability),
			proficiency,
			profBonus)

		bullet := proficiency.ToSymbol()
		return util.RenderEdgeBound(LongColWidth, TinyColWidth, bullet+" "+s.ability, fmt.Sprintf("%+d", mod))
	}
}

type SkillInfo struct {
	skill     *models.CharacterSkillDetailTO
	abilities *models.AbilitiesTO
	profBonus *int
}

func renderSkillInfoRow(s *SkillInfo) string {
	proficiency := models.Proficiency(s.skill.Proficiency)
	mod := models.ToModifier(
		s.abilities.ToScoreByName(s.skill.SkillAbility),
		proficiency,
		*s.profBonus)
	bullet := proficiency.ToSymbol()
	return util.RenderEdgeBound(LongColWidth, TinyColWidth, bullet+" "+s.skill.SkillName, fmt.Sprintf("%+d", mod))
}

func RenderAttack(a *models.AttackTO) string {
	return fmt.Sprintf("%-11s %+3d %s (%s)", a.Name, a.Bonus, a.Damage, a.DamageType)
}
