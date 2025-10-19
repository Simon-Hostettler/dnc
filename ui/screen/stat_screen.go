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

func NewStatScreen(keymap util.KeyMap, cr repository.CharacterRepository, ctx context.Context, id uuid.UUID) *StatScreen {
	return &StatScreen{
		keymap:              keymap,
		CharacterRepository: cr,
		Context:             ctx,
		characterID:         id,
		characterInfo:       list.NewListWithDefaults(),
		abilities:           list.NewListWithDefaults(),
		skills: list.NewListWithDefaults().
			WithTitle("Skills"),
		savingThrows: list.NewListWithDefaults().
			WithTitle("Saving Throws"),
		combatInfo: list.NewListWithDefaults().
			WithTitle("Combat"),
		attacks: list.NewListWithDefaults().
			WithTitle("Attacks"),
		actions: component.NewSimpleStringComponent(
			keymap,
			"Actions",
			"",
			func(a string) error {
				return nil
			},
			false,
			false,
		),
		bonusActions: component.NewSimpleStringComponent(
			keymap,
			"Bonus Actions",
			"",
			func(a string) error {
				return nil
			},
			false,
			false,
		),
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
	s.characterInfo.WithRows(s.GetCharacterInfoRows(c.Character))
	s.abilities.WithRows(s.GetAbilityRows(c))
	s.skills.WithRows(s.GetSkillRows(c))
	s.savingThrows.WithRows(s.GetSavingThrowRows(c))
	s.combatInfo.WithRows(s.GetCombatInfoRows(c.Character))
	s.attacks.WithRows(s.GetAttackRows(c.Attacks))

	s.actions = component.NewSimpleStringComponent(
		s.keymap,
		"Actions",
		c.Character.Actions,
		func(a string) error {
			c.Character.Actions = a
			return s.CharacterRepository.UpdateCharacter(s.Context, *c.Character)
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
			return s.CharacterRepository.UpdateCharacter(s.Context, *c.Character)
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
	case command.DataOpMsg:
		if msg.Op != command.DataSave {
			cmd = command.LoadCharacterCmd(s.CharacterRepository, s.Context, s.characterID)
		}
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

func (s *StatScreen) GetCharacterInfoRows(c *models.CharacterTO) []list.Row {
	rowCfg := list.LabeledStringRowConfig{JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0}

	rows := []list.Row{
		list.NewLabeledStringRow(s.keymap, "Name:", c.Name,
			editor.NewStringEditor(s.keymap, "Name", c.Name, s.persistCharStringField("name"))).
			WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Levels:", c.ClassLevels,
			editor.NewStringEditor(s.keymap, "Levels", c.ClassLevels, s.persistCharStringField("class_levels"))).
			WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Race:", c.Race,
			editor.NewStringEditor(s.keymap, "Race", c.Race, s.persistCharStringField("race"))).
			WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Alignment:", c.Alignment,
			editor.NewStringEditor(s.keymap, "Alignment", c.Alignment, s.persistCharStringField("alignment"))).
			WithConfig(rowCfg),
		list.NewLabeledIntRow(s.keymap, "Proficiency Bonus:", c.ProficiencyBonus,
			editor.NewIntEditor(s.keymap, "Proficiency Bonus", c.ProficiencyBonus, s.persistCharIntField("proficiency_bonus"))).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: false, LabelWidth: LongColWidth, ValueWidth: 0,
			}),
	}
	return rows
}

func (s *StatScreen) GetAbilityRows(agg repository.CharacterAggregate) []list.Row {
	scorePrinter := func(score int) string {
		mod := (score - 10) / 2
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), mod)
	}
	rowCfg := list.LabeledIntRowConfig{ValuePrinter: scorePrinter, JustifyValue: true, LabelWidth: ColWidth, ValueWidth: ShortColWidth}
	ab := *agg.Abilities

	rows := []list.Row{
		list.NewLabeledIntRow(s.keymap, "Strength:", ab.Strength,
			editor.NewIntEditor(s.keymap, "Strength", ab.Strength, s.persistAbilityIntField("strength"))).WithConfig(rowCfg),
		list.NewLabeledIntRow(s.keymap, "Constitution:", ab.Constitution,
			editor.NewIntEditor(s.keymap, "Constitution", ab.Constitution, s.persistAbilityIntField("constitution"))).WithConfig(rowCfg),
		list.NewLabeledIntRow(s.keymap, "Dexterity:", ab.Dexterity,
			editor.NewIntEditor(s.keymap, "Dexterity", ab.Dexterity, s.persistAbilityIntField("dexterity"))).WithConfig(rowCfg),
		list.NewLabeledIntRow(s.keymap, "Intelligence:", ab.Intelligence,
			editor.NewIntEditor(s.keymap, "Intelligence", ab.Intelligence, s.persistAbilityIntField("intelligence"))).WithConfig(rowCfg),
		list.NewLabeledIntRow(s.keymap, "Wisdom:", ab.Wisdom,
			editor.NewIntEditor(s.keymap, "Wisdom", ab.Wisdom, s.persistAbilityIntField("wisdom"))).WithConfig(rowCfg),
		list.NewLabeledIntRow(s.keymap, "Charisma:", ab.Charisma,
			editor.NewIntEditor(s.keymap, "Charisma", ab.Charisma, s.persistAbilityIntField("charisma"))).WithConfig(rowCfg),
	}
	return rows
}

func (s *StatScreen) persistAbilityIntField(field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateAbilityFields(s.Context, s.characterID, map[string]interface{}{field: v})
	}
}

func (s *StatScreen) GetCombatInfoRows(c *models.CharacterTO) []list.Row {
	standardCfg := list.LabeledIntRowConfig{
		ValuePrinter: strconv.Itoa, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	dsConfig := list.LabeledIntRowConfig{
		ValuePrinter: DeathSaveSymbols, JustifyValue: true,
		LabelWidth: ColWidth, ValueWidth: TinyColWidth,
	}
	rows := []list.Row{
		list.NewLabeledIntRow(s.keymap, "AC", c.ArmorClass,
			editor.NewIntEditor(s.keymap, "AC", c.ArmorClass, s.persistCharIntField("armor_class"))).
			WithConfig(standardCfg),
		list.NewLabeledIntRow(s.keymap, "Initiative", c.Initiative,
			editor.NewIntEditor(s.keymap, "Initiative", c.Initiative, s.persistCharIntField("initiative"))).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: func(i int) string { return fmt.Sprintf("%+d", i) },
				JustifyValue: true, LabelWidth: ColWidth, ValueWidth: TinyColWidth,
			}),
		list.NewLabeledIntRow(s.keymap, "Speed", c.Speed,
			editor.NewIntEditor(s.keymap, "Speed", c.Speed, s.persistCharIntField("speed"))).
			WithConfig(standardCfg),
		list.NewStructRow(s.keymap, &HPInfo{&c.CurrHitPoints, &c.MaxHitPoints}, renderHPInfoRow,
			[]editor.ValueEditor{
				editor.NewIntEditor(s.keymap, "Current HP", c.CurrHitPoints, s.persistCharIntField("curr_hit_points")),
				editor.NewIntEditor(s.keymap, "Max HP", c.MaxHitPoints, s.persistCharIntField("max_hit_points")),
			}),
		list.NewStructRow(s.keymap, &HitDiceInfo{&c.UsedHitDice, &c.HitDice}, renderHitDiceInfoRow,
			[]editor.ValueEditor{
				editor.NewStringEditor(s.keymap, "Used Hit Dice", c.UsedHitDice, s.persistCharStringField("used_hit_dice")),
				editor.NewStringEditor(s.keymap, "Hit Dice", c.HitDice, s.persistCharStringField("hit_dice")),
			}),
		list.NewLabeledIntRow(s.keymap, "DS Successes", c.DeathSaveSuccesses,
			editor.NewIntEditor(s.keymap, "DS Successes", c.DeathSaveSuccesses, s.persistCharIntField("death_save_successes"))).
			WithConfig(dsConfig),
		list.NewLabeledIntRow(s.keymap, "DS Failures", c.DeathSaveFailures,
			editor.NewIntEditor(s.keymap, "DS Failures", c.DeathSaveFailures, s.persistCharIntField("death_save_failures"))).
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

func (s *StatScreen) GetAttackRows(attacks []models.AttackTO) []list.Row {
	rows := []list.Row{}
	for i := range attacks {
		a := attacks[i]
		row := list.NewStructRow(s.keymap, a, RenderAttack, []editor.ValueEditor{
			editor.NewStringEditor(s.keymap, "Name", a.Name, s.persistAttackStringField(a.ID, "name")),
			editor.NewIntEditor(s.keymap, "Bonus", a.Bonus, s.persistAttackIntField(a.ID, "bonus")),
			editor.NewStringEditor(s.keymap, "Damage", a.Damage, s.persistAttackStringField(a.ID, "damage")),
			editor.NewStringEditor(s.keymap, "Damage Type", a.DamageType, s.persistAttackStringField(a.ID, "damage_type")),
		})
		rows = append(rows, row)
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, "attack"))
	return rows
}

func (s *StatScreen) persistAttackIntField(id uuid.UUID, field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateAttackFields(s.Context, id, map[string]interface{}{field: v})
	}
}

func (s *StatScreen) persistAttackStringField(id uuid.UUID, field string) func(string) error {
	return func(v string) error {
		return s.CharacterRepository.UpdateAttackFields(s.Context, id, map[string]interface{}{field: v})
	}
}

func (s *StatScreen) GetSkillRows(agg repository.CharacterAggregate) []list.Row {
	rows := []list.Row{}

	for i := range agg.Skills {
		skill := &agg.Skills[i]

		row := list.NewStructRow(s.keymap, *skill, renderSkillInfoRow(*agg.Abilities, agg.Character.ProficiencyBonus),
			[]editor.ValueEditor{
				editor.NewEnumEditor(s.keymap, ProficiencySymbols, "Proficiency", skill.Proficiency, s.persistSkillIntField(skill.ID, "proficiency")),
				editor.NewIntEditor(s.keymap, "Custom Modifier", skill.CustomModifier, s.persistSkillIntField(skill.ID, "custom_modifier")),
			})
		rows = append(rows, row)
	}

	return rows
}

func (s *StatScreen) persistSkillIntField(id uuid.UUID, field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateSkillFields(s.Context, id, map[string]interface{}{field: v})
	}
}

func (s *StatScreen) GetSavingThrowRows(agg repository.CharacterAggregate) []list.Row {
	renderer := renderSavingThrowInfoRow(*agg.Abilities, agg.Character.ProficiencyBonus)
	return []list.Row{
		list.NewStructRow(s.keymap, SavingThrowInfo{models.Proficiency(agg.SavingThrows.StrengthProficiency), "strength"}, renderer,
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, ProficiencySymbols, "Proficiency", agg.SavingThrows.StrengthProficiency, s.persistSavingThrowIntField("strength_proficiency"))}),
		list.NewStructRow(s.keymap, SavingThrowInfo{models.Proficiency(agg.SavingThrows.DexterityProficiency), "dexterity"}, renderer,
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, ProficiencySymbols, "Proficiency", agg.SavingThrows.DexterityProficiency, s.persistSavingThrowIntField("dexterity_proficiency"))}),
		list.NewStructRow(s.keymap, SavingThrowInfo{models.Proficiency(agg.SavingThrows.ConstitutionProficiency), "constitution"}, renderer,
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, ProficiencySymbols, "Proficiency", agg.SavingThrows.ConstitutionProficiency, s.persistSavingThrowIntField("constitution_proficiency"))}),
		list.NewStructRow(s.keymap, SavingThrowInfo{models.Proficiency(agg.SavingThrows.IntelligenceProficiency), "intelligence"}, renderer,
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, ProficiencySymbols, "Proficiency", agg.SavingThrows.IntelligenceProficiency, s.persistSavingThrowIntField("intelligence_proficiency"))}),
		list.NewStructRow(s.keymap, SavingThrowInfo{models.Proficiency(agg.SavingThrows.WisdomProficiency), "wisdom"}, renderer,
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, ProficiencySymbols, "Proficiency", agg.SavingThrows.WisdomProficiency, s.persistSavingThrowIntField("wisdom_proficiency"))}),
		list.NewStructRow(s.keymap, SavingThrowInfo{models.Proficiency(agg.SavingThrows.CharismaProficiency), "charisma"}, renderer,
			[]editor.ValueEditor{editor.NewEnumEditor(s.keymap, ProficiencySymbols, "Proficiency", agg.SavingThrows.CharismaProficiency, s.persistSavingThrowIntField("charisma_proficiency"))}),
	}
}

func (s *StatScreen) persistSavingThrowIntField(field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateSavingThrowFields(s.Context, s.characterID, map[string]interface{}{field: v})
	}
}

// screen specific types + utility functions

func (s *StatScreen) persistCharIntField(field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateCharacterFields(s.Context, s.characterID, map[string]interface{}{field: v})
	}
}

func (s *StatScreen) persistCharStringField(field string) func(string) error {
	return func(v string) error {
		return s.CharacterRepository.UpdateCharacterFields(s.Context, s.characterID, map[string]interface{}{field: v})
	}
}

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
	prof    models.Proficiency
	ability string
}

func renderSavingThrowInfoRow(a models.AbilitiesTO, profBonus int) func(SavingThrowInfo) string {
	return func(s SavingThrowInfo) string {
		mod := toModifier(a.ToScoreByName(strings.ToLower(s.ability)), profBonus)
		bullet := ProficiencySymbol(models.Proficiency(s.prof))
		return util.RenderEdgeBound(LongColWidth, TinyColWidth, bullet+" "+s.ability, fmt.Sprintf("%+d", mod))
	}
}

type SkillInfo struct {
	skill     *models.CharacterSkillDetailTO
	abilities *models.AbilitiesTO
	profBonus *int
}

func renderSkillInfoRow(a models.AbilitiesTO, profBonus int) func(models.CharacterSkillDetailTO) string {
	return func(s models.CharacterSkillDetailTO) string {
		mod := s.ToModifier(a.ToScoreByName(strings.ToLower(s.SkillAbility)), profBonus)
		bullet := ProficiencySymbol(models.Proficiency(s.Proficiency))
		return util.RenderEdgeBound(LongColWidth, TinyColWidth, bullet+" "+s.SkillName, fmt.Sprintf("%+d", mod))
	}
}

func RenderAttack(a models.AttackTO) string {
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

func ProficiencySymbol(p models.Proficiency) string {
	for _, m := range ProficiencySymbols {
		if int(p) == m.Value {
			return m.Label
		}
	}
	return ""
}

func toModifier(score int, bonus int) int {
	return (score-10)/2 + bonus
}
