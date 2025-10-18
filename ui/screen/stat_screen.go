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
	keymap              util.KeyMap
	CharacterRepository repository.CharacterRepository
	Context             context.Context

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	characterID   uuid.UUID
	characterInfo *list.List
	abilities     *list.List
	skills        *list.List
	savingThrows  *list.List
	combatInfo    *list.List
	attacks       *list.List
	actions       *component.SimpleStringComponent
	bonusActions  *component.SimpleStringComponent
}

func NewStatScreen(keymap util.KeyMap, id uuid.UUID) *StatScreen {
	return &StatScreen{
		keymap:        keymap,
		characterID:   id,
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
}

func (s *StatScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.characterInfo.Init())
	cmds = append(cmds, s.abilities.Init())
	cmds = append(cmds, s.skills.Init())
	cmds = append(cmds, s.savingThrows.Init())
	cmds = append(cmds, s.combatInfo.Init())
	cmds = append(cmds, s.attacks.Init())

	cmds = append(cmds, command.LoadCharacterCmd(s.CharacterRepository, s.Context, s.characterID))

	s.lastFocusedElement = s.characterInfo
	s.focusOn(s.characterInfo)

	cmds = util.Filter(cmds, func(c tea.Cmd) bool { return c != nil })
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}

	return nil
}

func (s *StatScreen) Populate(c repository.CharacterAggregate) tea.Cmd {
	var cmd tea.Cmd
	s.characterInfo.WithRows(s.GetCharacterInfoRows(s.keymap, c.Character))
	s.abilities.WithRows(s.GetAbilityRows(s.keymap, c))
	s.skills.WithRows(s.GetSkillRows(s.keymap, c.Skills))
	s.savingThrows.WithRows(GetSavingThrowRows(s.keymap, c.Character))
	s.combatInfo.WithRows(s.GetCombatInfoRows(s.keymap, c.Character))
	s.attacks.WithRows(s.GetAttackRows(s.keymap, c.Attacks))

	s.actions = component.NewSimpleStringComponent(
		s.keymap,
		"Actions",
		c.Character.Actions,
		func(a string) error {
			c.Character.Actions = a
			return s.CharacterRepository.UpdateCharacter(s.Context, c.Character)
		},
		false,
		false,
	)
	cmd = s.actions.Init()

	s.bonusActions = component.NewSimpleStringComponent(
		s.keymap,
		"Bonus Actions",
		c.Character.BonusActions,
		func(a string) error {
			c.Character.BonusActions = a
			return s.CharacterRepository.UpdateCharacter(s.Context, c.Character)
		},
		false,
		false,
	)
	cmd = tea.Batch(cmd, s.bonusActions.Init())

	return cmd
}

func (s *StatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.LoadCharacterMsg:
		s.Populate(msg.Character)
	case command.AppendElementMsg:
		if msg.Tag == "attack" {
			cmd = command.DataOperationCommand(func() error {
				_, err := s.CharacterRepository.AddAttack(s.Context, s.characterID, models.AttackTO{})
				return err
			}, command.DataCreate)
		} else {
			_, cmd = s.focusedElement.Update(msg)
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

func (s *StatScreen) GetCharacterInfoRows(k util.KeyMap, c models.CharacterTO) []list.Row {
	rowCfg := list.LabeledStringRowConfig{JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0}

	persistChar := func() error { return s.CharacterRepository.UpdateCharacter(s.Context, c) }

	rows := []list.Row{
		list.NewLabeledStringRow(k, "Name:", c.Name,
			editor.NewStringEditor(k, "Name", c.Name, editor.BindString(&c.Name, persistChar))).
			WithConfig(rowCfg),
		list.NewLabeledStringRow(k, "Levels:", c.ClassLevels,
			editor.NewStringEditor(k, "Levels", c.ClassLevels, editor.BindString(&c.ClassLevels, persistChar))).
			WithConfig(rowCfg),
		list.NewLabeledStringRow(k, "Race:", c.Race,
			editor.NewStringEditor(k, "Race", c.Race, editor.BindString(&c.Race, persistChar))).
			WithConfig(rowCfg),
		list.NewLabeledStringRow(k, "Alignment:", c.Alignment,
			editor.NewStringEditor(k, "Alignment", c.Alignment, editor.BindString(&c.Alignment, persistChar))).
			WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Proficiency Bonus:", c.ProficiencyBonus,
			editor.NewIntEditor(k, "Proficiency Bonus", c.ProficiencyBonus, editor.BindInt(&c.ProficiencyBonus, persistChar))).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0,
			}),
	}
	return rows
}

func (s *StatScreen) GetAbilityRows(k util.KeyMap, agg repository.CharacterAggregate) []list.Row {
	scorePrinter := func(score int) string {
		mod := (score - 10) / 2
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), mod)
	}
	rowCfg := list.LabeledIntRowConfig{ValuePrinter: scorePrinter, JustifyValue: true, LabelWidth: ColWidth, ValueWidth: ShortColWidth}
	ab := models.AbilitiesTO{}
	if agg.Abilities != nil {
		ab = *agg.Abilities
	}
	id := agg.Character.ID
	persistAb := func() error { return s.CharacterRepository.UpsertAbilities(s.Context, id, ab) }

	rows := []list.Row{
		list.NewLabeledIntRow(k, "Strength:", ab.Strength,
			editor.NewIntEditor(k, "Strength", ab.Strength, editor.BindInt(&ab.Strength, persistAb))).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Constitution:", ab.Constitution,
			editor.NewIntEditor(k, "Constitution", ab.Constitution, editor.BindInt(&ab.Constitution, persistAb))).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Dexterity:", ab.Dexterity,
			editor.NewIntEditor(k, "Dexterity", ab.Dexterity, editor.BindInt(&ab.Dexterity, persistAb))).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Intelligence:", ab.Intelligence,
			editor.NewIntEditor(k, "Intelligence", ab.Intelligence, editor.BindInt(&ab.Intelligence, persistAb))).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Wisdom:", ab.Wisdom,
			editor.NewIntEditor(k, "Wisdom", ab.Wisdom, editor.BindInt(&ab.Wisdom, persistAb))).WithConfig(rowCfg),
		list.NewLabeledIntRow(k, "Charisma:", ab.Charisma,
			editor.NewIntEditor(k, "Charisma", ab.Charisma, editor.BindInt(&ab.Charisma, persistAb))).WithConfig(rowCfg),
	}
	return rows
}

func (s *StatScreen) GetCombatInfoRows(k util.KeyMap, c models.CharacterTO) []list.Row {
	persistChar := func() error { return s.CharacterRepository.UpdateCharacter(s.Context, c) }

	standardCfg := list.LabeledIntRowConfig{
		ValuePrinter: strconv.Itoa, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	dsConfig := list.LabeledIntRowConfig{
		ValuePrinter: DeathSaveSymbols, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	rows := []list.Row{
		list.NewLabeledIntRow(k, "AC", c.ArmorClass,
			editor.NewIntEditor(k, "AC", c.ArmorClass, editor.BindInt(&c.ArmorClass, persistChar))).
			WithConfig(standardCfg),
		list.NewLabeledIntRow(k, "Initiative", c.Initiative,
			editor.NewIntEditor(k, "Initiative", c.Initiative, editor.BindInt(&c.Initiative, persistChar))).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: true, LabelWidth: ColWidth, ValueWidth: TinyColWidth,
			}),
		list.NewLabeledIntRow(k, "Speed", c.Speed,
			editor.NewIntEditor(k, "Speed", c.Speed, editor.BindInt(&c.Speed, persistChar))).
			WithConfig(standardCfg),
		list.NewStructRow(k, &HPInfo{&c.CurrHitPoints, &c.MaxHitPoints}, renderHPInfoRow,
			[]editor.ValueEditor{
				editor.NewIntEditor(k, "Current HP", c.CurrHitPoints, editor.BindInt(&c.CurrHitPoints, persistChar)),
				editor.NewIntEditor(k, "Max HP", c.MaxHitPoints, editor.BindInt(&c.MaxHitPoints, persistChar)),
			}),
		list.NewStructRow(k, &HitDiceInfo{&c.UsedHitDice, &c.HitDice}, renderHitDiceInfoRow,
			[]editor.ValueEditor{
				editor.NewStringEditor(k, "Used Hit Dice", c.UsedHitDice, editor.BindString(&c.UsedHitDice, persistChar)),
				editor.NewStringEditor(k, "Hit Dice", c.HitDice, editor.BindString(&c.HitDice, persistChar)),
			}),
		list.NewLabeledIntRow(k, "DS Successes", c.DeathSaveSuccesses,
			editor.NewIntEditor(k, "DS Successes", c.DeathSaveSuccesses, editor.BindInt(&c.DeathSaveSuccesses, persistChar))).
			WithConfig(dsConfig),
		list.NewLabeledIntRow(k, "DS Failures", c.DeathSaveFailures,
			editor.NewIntEditor(k, "DS Failures", c.DeathSaveFailures, editor.BindInt(&c.DeathSaveFailures, persistChar))).
			WithConfig(dsConfig),
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

func (s *StatScreen) GetAttackRows(k util.KeyMap, attacks []models.AttackTO) []list.Row {
	rows := []list.Row{}
	for i := range attacks {
		a := &attacks[i]
		persistAttack := func() error { return s.CharacterRepository.UpdateAttack(s.Context, *a) }
		row := list.NewStructRow(k, a, RenderAttack, []editor.ValueEditor{
			editor.NewStringEditor(k, "Name", a.Name, editor.BindString(&a.Name, persistAttack)),
			editor.NewIntEditor(k, "Bonus", a.Bonus, editor.BindInt(&a.Bonus, persistAttack)),
			editor.NewStringEditor(k, "Damage", a.Damage, editor.BindString(&a.Damage, persistAttack)),
			editor.NewStringEditor(k, "Damage Type", a.DamageType, editor.BindString(&a.DamageType, persistAttack)),
		})
		rows = append(rows, row)
	}
	rows = append(rows, list.NewAppenderRow(k, "attack"))
	return rows
}

func (s *StatScreen) GetSkillRows(k util.KeyMap, c repository.CharacterAggregate) []list.Row {
	rows := []list.Row{}

	for i := range sk {
		skill := sk[i]
		row := list.NewStructRow(k, &SkillInfo{skill, *c.Abilities, c.Character.ProficiencyBonus}, renderSkillInfoRow,
			[]editor.ValueEditor{
				editor.NewEnumEditor(k, ProficiencySymbols, "Proficiency", &skill.Proficiency),
				editor.NewIntEditor(k, "Custom Modifier", &skill.CustomModifier),
			})
		rows = append(rows, row)
	}

	return rows
}

func GetSavingThrowRows(k util.KeyMap, c models.CharacterTO) []list.Row {
	rows := []list.Row{}

	for i := range c.SavingThrows {
		saving := &c.SavingThrows[i]
		row := list.NewStructRow(k, &SavingThrowInfo{saving, c.Abilities, c.ProficiencyBonus}, renderSavingThrowInfoRow,
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
	skill     models.CharacterSkillDetailTO
	abilities models.AbilitiesTO
	profBonus int
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
