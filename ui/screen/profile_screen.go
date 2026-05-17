package screen

import (
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
	profileTopBarHeight = 6

	profileColHeight    = 27
	profileLeftColWidth = 32
	profileMidColWidth  = 40

	profileRightColWidth     = 28
	profileRightContentWidth = profileRightColWidth - 6

	profileLongColWidth  = 20
	profileShortColWidth = 8
)

type ProfileScreen struct {
	keymap util.KeyMap
	agg    *repository.CharacterAggregate
	FocusManager

	characterInfo       *list.List
	characterAppearance *list.List
	features            *list.List
	backstory           *component.SimpleTextComponent
	appearance          *component.SimpleTextComponent
	personality         *component.SimpleTextComponent

	featureRows *CollectionRows[models.FeatureTO]
}

func NewProfileScreen(km util.KeyMap, c *repository.CharacterAggregate) *ProfileScreen {
	s := &ProfileScreen{
		keymap:              km,
		agg:                 c,
		backstory:           component.NewSimpleTextComponent(km, "Backstory", &c.Character.Backstory, profileColHeight-6, profileMidColWidth-6),
		appearance:          component.NewSimpleTextComponent(km, "Appearance", &c.Character.Appearance, (profileColHeight-10)/2, profileRightColWidth-4),
		personality:         component.NewSimpleTextComponent(km, "Personality", &c.Character.Personality, (profileColHeight-10)/2, profileRightColWidth-4),
		characterInfo:       list.NewListWithDefaults(km),
		characterAppearance: list.NewListWithDefaults(km),
		features:            list.NewListWithDefaults(km).WithTitle("Features & Traits"),
	}
	s.featureRows = NewCollectionRows(km, s.features, "feature",
		func() []*models.FeatureTO { return util.Pointers(s.agg.Features) },
		func(f *models.FeatureTO) uuid.UUID { return f.ID },
		s.agg.AddEmptyFeature,
		s.agg.DeleteFeature,
		func(f *models.FeatureTO) *list.StructRow[models.FeatureTO] {
			return list.NewStructRow(s.keymap, f, RenderFeature, []editor.ValueEditor{
				editor.NewStringEditor(s.keymap, "Name", &f.Name),
				editor.NewTextEditor(s.keymap, "Description", &f.Description),
			}).WithReader(renderFullFeature)
		},
	)
	return s
}

func (s *ProfileScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.characterInfo.Init())
	cmds = append(cmds, s.characterAppearance.Init())
	cmds = append(cmds, s.features.Init())
	cmds = append(cmds, s.backstory.Init())
	cmds = append(cmds, s.appearance.Init())
	cmds = append(cmds, s.personality.Init())

	s.CreateCharacterInfoRows()
	s.CreateCharacterAppearanceRows()
	s.featureRows.Repopulate()

	s.wireFocusGraph()

	return tea.Batch(cmds...)
}

func (s *ProfileScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if msg.Tag == "feature" {
			cmd = s.featureRows.HandleAppend(msg.Tag)
		} else {
			_, cmd = s.focusedElement.Update(msg)
		}
	case command.FocusNextElementMsg:
		s.MoveFocus(msg.Direction)
	case tea.KeyPressMsg:
		cmd = RouteKey(s.focusedElement, msg, s.keymap, s.MoveFocus)
	}
	return s, cmd
}

func (s *ProfileScreen) wireFocusGraph() {
	s.Wire(FocusGraph{
		s.characterInfo: {
			command.DownDirection:  To(s.features),
			command.RightDirection: To(s.characterAppearance),
			command.LeftDirection:  Emit(command.ReturnFocusToParentCmd),
		},
		s.characterAppearance: {
			command.DownDirection: To(s.backstory),
			command.LeftDirection: To(s.characterInfo),
		},
		s.features: {
			command.UpDirection:    To(s.characterInfo),
			command.RightDirection: To(s.backstory),
			command.LeftDirection:  Emit(command.ReturnFocusToParentCmd),
		},
		s.backstory: {
			command.UpDirection:    To(s.characterInfo),
			command.RightDirection: To(s.personality),
			command.LeftDirection:  ToWith(s.features, func() { s.features.SetCursor(0) }),
		},
		s.personality: {
			command.UpDirection:   To(s.characterAppearance),
			command.LeftDirection: To(s.backstory),
			command.DownDirection: To(s.appearance),
		},
		s.appearance: {
			command.UpDirection:   To(s.personality),
			command.LeftDirection: To(s.backstory),
		},
	}, s.characterInfo)
}

func (s *ProfileScreen) View() tea.View {
	characterInfo := s.characterInfo.View().Content

	characterAppearance := s.characterAppearance.View().Content

	topBarSeparator := styles.MakeVerticalSeparator(profileTopBarHeight)

	topBar := styles.DefaultBorderStyle.
		Height(profileTopBarHeight).
		Width(styles.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			characterInfo,
			lipgloss.PlaceHorizontal(20, lipgloss.Center, topBarSeparator),
			lipgloss.PlaceHorizontal(26, lipgloss.Left, characterAppearance)))

	leftColumn := styles.DefaultBorderStyle.
		Height(profileColHeight).
		Width(profileLeftColWidth).
		Render(s.features.View().Content)

	midColumn := styles.DefaultBorderStyle.
		Width(profileMidColWidth).
		Height(profileColHeight).
		Render(s.RenderBackstory())

	characterPersonality := s.RenderPersonality()
	appearance := s.RenderAppearance()

	rightBoxInnerSeparator := styles.MakeHorizontalSeparator(profileRightContentWidth, 1)

	rightColumn := styles.DefaultBorderStyle.
		Width(profileRightColWidth).
		Height(profileColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, characterPersonality, rightBoxInnerSeparator, appearance))

	body := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn, midColumn, rightColumn)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Center, topBar, body))
}

func (s *ProfileScreen) RenderBackstory() string {
	title := lipgloss.NewStyle().Width(profileMidColWidth).Render(
		styles.RenderItem(s.backstory.InFocus(), "Backstory") + "\n",
	)
	body := styles.DefaultTextStyle.Width(profileMidColWidth).Render(s.backstory.View().Content)

	return lipgloss.JoinVertical(lipgloss.Center, title, body)
}

func (s *ProfileScreen) RenderPersonality() string {
	title := lipgloss.NewStyle().Width(profileRightColWidth).Render(
		styles.RenderItem(s.personality.InFocus(), "Personality") + "\n",
	)
	body := styles.DefaultTextStyle.Width(profileRightColWidth).Render(s.personality.View().Content)

	return lipgloss.JoinVertical(lipgloss.Center, title, body)
}

func (s *ProfileScreen) RenderAppearance() string {
	title := lipgloss.NewStyle().Width(profileRightColWidth).Render(
		styles.RenderItem(s.appearance.InFocus(), "Appearance") + "\n",
	)
	body := styles.DefaultTextStyle.Width(profileRightColWidth).Render(s.appearance.View().Content)

	return lipgloss.JoinVertical(lipgloss.Center, title, body)
}

func (s *ProfileScreen) CreateCharacterInfoRows() {
	rowCfg := list.LabeledStringRowConfig{JustifyValue: false, LabelWidth: profileLongColWidth, ValueWidth: 0}
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
				JustifyValue: false, LabelWidth: profileLongColWidth, ValueWidth: 0,
			}),
	}
	s.characterInfo.WithRows(rows)
}

func (s *ProfileScreen) CreateCharacterAppearanceRows() {
	rowCfg := list.LabeledStringRowConfig{JustifyValue: false, LabelWidth: profileShortColWidth, ValueWidth: 0}
	rows := []list.Row{
		list.NewLabeledIntRow(s.keymap, "Age:", &s.agg.Character.Age,
			editor.NewIntEditor(s.keymap, "Age", &s.agg.Character.Age)).
			WithConfig(list.LabeledIntRowConfig{
				ValuePrinter: strconv.Itoa,
				JustifyValue: false, LabelWidth: profileShortColWidth, ValueWidth: 0,
			}),
		list.NewLabeledStringRow(s.keymap, "Height:", &s.agg.Character.Height,
			editor.NewStringEditor(s.keymap, "Height", &s.agg.Character.Height)).WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Weight:", &s.agg.Character.Weight,
			editor.NewStringEditor(s.keymap, "Weight", &s.agg.Character.Weight)).WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Eyes:", &s.agg.Character.Eyes,
			editor.NewStringEditor(s.keymap, "Eyes", &s.agg.Character.Eyes)).WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Skin:", &s.agg.Character.Skin,
			editor.NewStringEditor(s.keymap, "Skin", &s.agg.Character.Skin)).WithConfig(rowCfg),
		list.NewLabeledStringRow(s.keymap, "Hair:", &s.agg.Character.Hair,
			editor.NewStringEditor(s.keymap, "Hair", &s.agg.Character.Hair)).WithConfig(rowCfg),
	}
	s.characterAppearance.WithRows(rows)
}

// screen specific types + utility functions

func RenderFeature(f *models.FeatureTO) string {
	return f.Name
}

func renderFullFeature(f *models.FeatureTO) string {
	separator := styles.MakeHorizontalSeparator(styles.SmallScreenWidth-4, 1)
	content := strings.Join(
		[]string{
			f.Name,
			separator,
			f.Description,
		},
		"\n")
	return styles.DefaultTextStyle.
		AlignHorizontal(lipgloss.Left).
		Render(content)
}
