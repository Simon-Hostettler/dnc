package screen

import (
	"fmt"
	"strconv"

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
	statTopBarHeight = 6

	statColHeight    = 27
	statLeftColWidth = 32
	statMidColWidth  = 30

	statRightColWidth     = 40
	statRightContentWidth = statRightColWidth - 8

	statLongColWidth   = 20
	statColWidth       = 16
	statMediumColWidth = 12
	statShortColWidth  = 8
	statTinyColWidth   = 3

	statActionHeight = 5
)

type StatScreen struct {
	keymap util.KeyMap
	agg    *repository.CharacterAggregate
	FocusManager

	characterInfo *list.List
	abilities     *list.List
	skills        *list.List
	savingThrows  *list.List
	combatInfo    *list.List
	attacks       *list.List
	actions       *component.SimpleTextComponent
	bonusActions  *component.SimpleTextComponent

	attackRows *Collection[models.AttackTO]
}

func NewStatScreen(km util.KeyMap, c *repository.CharacterAggregate) *StatScreen {
	s := &StatScreen{
		keymap:        km,
		agg:           c,
		actions:       component.NewSimpleTextComponent(km, "Actions", &c.Character.Actions, statActionHeight, statRightContentWidth),
		bonusActions:  component.NewSimpleTextComponent(km, "Bonus Actions", &c.Character.BonusActions, statActionHeight, statRightContentWidth),
		characterInfo: list.NewListWithDefaults(km),
		abilities:     list.NewListWithDefaults(km),
		skills: list.NewListWithDefaults(km).
			WithTitle("Skills"),
		savingThrows: list.NewListWithDefaults(km).
			WithTitle("Saving Throws"),
		combatInfo: list.NewListWithDefaults(km).
			WithTitle("Combat"),
		attacks: list.NewListWithDefaults(km).
			WithTitle("Attacks").WithViewport(4),
	}
	s.attackRows = NewCollection(km, s.attacks,
		func() []*models.AttackTO { return util.Pointers(s.agg.Attacks) },
		func(a *models.AttackTO) uuid.UUID { return a.ID },
		s.agg.AddEmptyAttack,
		s.agg.DeleteAttack,
		func(a *models.AttackTO) *list.StructRow[models.AttackTO] {
			return list.NewStructRow(s.keymap, a, RenderAttack, []editor.ValueEditor{
				editor.NewStringEditor(s.keymap, "Name", &a.Name),
				editor.NewIntEditor(s.keymap, "Bonus", &a.Bonus),
				editor.NewStringEditor(s.keymap, "Damage", &a.Damage),
				editor.NewStringEditor(s.keymap, "Damage Type", &a.DamageType),
			})
		},
	)
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
	s.attackRows.Repopulate()
	s.CreateSavingThrowRows()

	s.wireFocusGraph()

	return tea.Batch(cmds...)
}

func (s *StatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.FocusNextElementMsg:
		s.MoveFocus(msg.Direction)
	case tea.KeyPressMsg:
		cmd = RouteKey(s.focusedElement, msg, s.keymap, s.MoveFocus)
	}
	return s, cmd
}

func (s *StatScreen) wireFocusGraph() {
	s.Wire(FocusGraph{
		s.characterInfo: {
			command.DownDirection:  To(s.skills),
			command.RightDirection: To(s.abilities),
			command.LeftDirection:  Emit(command.ReturnFocusToParentCmd),
		},
		s.abilities: {
			command.DownDirection: To(s.actions),
			command.LeftDirection: To(s.characterInfo),
		},
		s.skills: {
			command.UpDirection:   To(s.characterInfo),
			command.LeftDirection: Emit(command.ReturnFocusToParentCmd),
			command.RightDirection: ToCond(func() FocusableModel {
				if s.skills.CursorPos() < s.skills.Size()/2 {
					return s.combatInfo
				}
				return s.savingThrows
			}),
		},
		s.combatInfo: {
			command.UpDirection:    To(s.characterInfo),
			command.RightDirection: To(s.actions),
			command.DownDirection:  To(s.savingThrows),
			command.LeftDirection:  ToWith(s.skills, func() { s.skills.SetCursor(0) }),
		},
		s.savingThrows: {
			command.UpDirection:    To(s.combatInfo),
			command.RightDirection: To(s.attacks),
			command.LeftDirection:  ToWith(s.skills, func() { s.skills.SetCursor(s.skills.Size() / 2) }),
		},
		s.actions: {
			command.UpDirection:   To(s.abilities),
			command.LeftDirection: To(s.combatInfo),
			command.DownDirection: To(s.bonusActions),
		},
		s.bonusActions: {
			command.UpDirection:   To(s.actions),
			command.LeftDirection: To(s.combatInfo),
			command.DownDirection: To(s.attacks),
		},
		s.attacks: {
			command.UpDirection:   To(s.bonusActions),
			command.LeftDirection: To(s.savingThrows),
		},
	}, s.characterInfo)
}

func (s *StatScreen) View() tea.View {
	characterInfo := s.characterInfo.View().Content

	abilities := s.abilities.View().Content

	topBarSeparator := styles.MakeVerticalSeparator(statTopBarHeight)

	topBar := styles.DefaultBorderStyle.
		Height(statTopBarHeight).
		Width(styles.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			characterInfo,
			lipgloss.PlaceHorizontal(20, lipgloss.Center, topBarSeparator),
			abilities))

	leftColumn := styles.DefaultBorderStyle.
		Height(statColHeight).
		Width(statLeftColWidth).
		Render(s.skills.View().Content)

	savingThrows := s.savingThrows.View().Content

	combatInfo := s.combatInfo.View().Content

	midBoxInnerSeparator := styles.MakeHorizontalSeparator(statMidColWidth-6, 1)

	midColumn := styles.DefaultBorderStyle.
		Width(statMidColWidth).
		Height(statColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, combatInfo, midBoxInnerSeparator, savingThrows))

	actions := s.RenderActions()

	attacks := s.attacks.View().Content

	rightBoxInnerSeparator := styles.MakeHorizontalSeparator(statRightContentWidth, 1)

	rightColumn := styles.DefaultBorderStyle.
		Width(statRightColWidth).
		Height(statColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, actions, rightBoxInnerSeparator, attacks))

	body := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn, midColumn, rightColumn)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Center, topBar, body))
}

func (s *StatScreen) RenderActions() string {
	actionTitle := styles.RenderItem(s.actions.InFocus(), "Actions") + "\n"
	actionBody := styles.DefaultTextStyle.Width(statRightContentWidth).Render(s.actions.View().Content)

	separator := styles.MakeHorizontalSeparator(statRightContentWidth, 1)

	bonusActionTitle := styles.RenderItem(s.bonusActions.InFocus(), "Bonus Actions") + "\n"
	bonusActionBody := styles.DefaultTextStyle.Width(statRightContentWidth).Render(s.bonusActions.View().Content)

	return lipgloss.JoinVertical(lipgloss.Center, actionTitle, actionBody, separator, bonusActionTitle, bonusActionBody)
}

func (s *StatScreen) CreateCharacterInfoRows() {
	rowCfg := list.LabeledStringRowConfig{JustifyValue: false, LabelWidth: statLongColWidth, ValueWidth: 0}
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
				ValuePrinter: styles.WithSign,
				JustifyValue: false, LabelWidth: statLongColWidth, ValueWidth: 0,
			}),
	}
	s.characterInfo.WithRows(rows)
}

func (s *StatScreen) CreateAbilityRows() {
	scorePrinter := func(score int) string {
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), models.ToModifier(score, 0, 0))
	}
	rowCfg := list.LabeledIntRowConfig{ValuePrinter: scorePrinter, JustifyValue: true, LabelWidth: statColWidth, ValueWidth: statShortColWidth}
	newAbilityRow := func(field *int, name string) list.Row {
		return list.NewLabeledIntRow(s.keymap, name+":", field,
			editor.NewIntEditor(s.keymap, name, field)).WithConfig(rowCfg)
	}
	rows := []list.Row{
		newAbilityRow(&s.agg.Abilities.Strength, "Strength"),
		newAbilityRow(&s.agg.Abilities.Constitution, "Constitution"),
		newAbilityRow(&s.agg.Abilities.Dexterity, "Dexterity"),
		newAbilityRow(&s.agg.Abilities.Intelligence, "Intelligence"),
		newAbilityRow(&s.agg.Abilities.Wisdom, "Wisdom"),
		newAbilityRow(&s.agg.Abilities.Charisma, "Charisma"),
	}
	s.abilities.WithRows(rows)
}

func (s *StatScreen) CreateCombatInfoRows() {
	standardCfg := list.LabeledIntRowConfig{
		ValuePrinter: strconv.Itoa, JustifyValue: true,
		LabelWidth: statColWidth, ValueWidth: statTinyColWidth,
	}
	dsConfig := list.LabeledIntRowConfig{
		ValuePrinter: RenderDeathSaves, JustifyValue: true,
		LabelWidth: statColWidth, ValueWidth: statTinyColWidth,
	}
	rows := []list.Row{
		list.NewLabeledIntRow(s.keymap, "AC", &s.agg.Character.ArmorClass,
			editor.NewIntEditor(s.keymap, "AC", &s.agg.Character.ArmorClass)).WithConfig(standardCfg),
		list.NewLabeledIntRow(s.keymap, "Initiative", &s.agg.Character.Initiative,
			editor.NewIntEditor(s.keymap, "Initiative", &s.agg.Character.Initiative)).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: true, LabelWidth: statColWidth, ValueWidth: statTinyColWidth,
			}),
		list.NewLabeledIntRow(s.keymap, "Speed", &s.agg.Character.Speed,
			editor.NewIntEditor(s.keymap, "Speed", &s.agg.Character.Speed)).WithConfig(standardCfg),
		list.NewStructRow(s.keymap, &HPInfo{&s.agg.Character.CurrHitPoints, &s.agg.Character.MaxHitPoints, &s.agg.Character.TempHitPoints}, renderHPInfoRow,
			[]editor.ValueEditor{
				editor.NewIntEditor(s.keymap, "Current HP", &s.agg.Character.CurrHitPoints),
				editor.NewIntEditor(s.keymap, "Max HP", &s.agg.Character.MaxHitPoints),
				editor.NewIntEditor(s.keymap, "Temp HP", &s.agg.Character.TempHitPoints),
			}),
		list.NewStructRow(s.keymap, &HitDiceInfo{&s.agg.Character.UsedHitDice, &s.agg.Character.HitDice}, renderHitDiceInfoRow,
			[]editor.ValueEditor{
				editor.NewStringEditor(s.keymap, "Used Hit Dice", &s.agg.Character.UsedHitDice),
				editor.NewStringEditor(s.keymap, "Hit Dice", &s.agg.Character.HitDice),
			}),
		list.NewLabeledIntRow(s.keymap, "DS Successes", &s.agg.Character.DeathSaveSuccesses,
			editor.NewEnumEditor(s.keymap, styles.DeathSaveSymbols, "DS Successes", &s.agg.Character.DeathSaveSuccesses)).
			WithConfig(dsConfig).WithCycleAction(cycleDeathSaves),
		list.NewLabeledIntRow(s.keymap, "DS Failures", &s.agg.Character.DeathSaveFailures,
			editor.NewEnumEditor(s.keymap, styles.DeathSaveSymbols, "DS Failures", &s.agg.Character.DeathSaveFailures)).
			WithConfig(dsConfig).WithCycleAction(cycleDeathSaves),
	}
	s.combatInfo.WithRows(rows)
}

func (s *StatScreen) CreateSkillRows() {
	rows := []list.Row{}

	for i := range s.agg.Skills {
		skill := &s.agg.Skills[i]
		row := list.NewStructRow(s.keymap, &SkillInfo{skill, s.agg.Abilities, &s.agg.Character.ProficiencyBonus}, renderSkillInfoRow,
			[]editor.ValueEditor{
				editor.NewEnumEditor(s.keymap, styles.ProficiencySymbols, "Proficiency", &skill.Proficiency),
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
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, styles.ProficiencySymbols, "Proficiency", field)})
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
	temp    *int
}

func renderHPInfoRow(hp *HPInfo) string {
	tmp := ""
	if hp.temp != nil && *hp.temp > 0 {
		tmp = fmt.Sprintf("(+%d)", *hp.temp)
	}
	return styles.RenderEdgeBound(statColWidth-4, 7, "HP", strconv.Itoa(*hp.current)+tmp+"/"+strconv.Itoa(*hp.max))
}

type HitDiceInfo struct {
	current *string
	max     *string
}

func renderHitDiceInfoRow(hd *HitDiceInfo) string {
	return styles.RenderEdgeBound(statShortColWidth, statMediumColWidth, "Hit Dice", *hd.current+"/"+*hd.max)
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

		bullet := styles.ToSymbol(proficiency)
		return styles.RenderEdgeBound(statLongColWidth, statTinyColWidth, bullet+" "+s.ability, fmt.Sprintf("%+d", mod))
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
		*s.profBonus) + s.skill.CustomModifier
	bullet := styles.ToSymbol(proficiency)
	return styles.RenderEdgeBound(statLongColWidth, statTinyColWidth, bullet+" "+s.skill.SkillName, fmt.Sprintf("%+d", mod))
}

func RenderAttack(a *models.AttackTO) string {
	return fmt.Sprintf("%-11s %+3d %s (%s)", a.Name, a.Bonus, a.Damage, a.DamageType)
}

func RenderDeathSaves(amount int) string {
	amount = util.Clamp(amount, 0, 3)
	return styles.DeathSaveSymbols[amount].Label
}

func cycleDeathSaves(saves *int) tea.Cmd {
	*saves = (*saves + 1) % 4
	return command.WriteBackRequest
}
