package screen

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	profileColHeight    = 25
	profileLeftColWidth = 30
	profileMidColWidth  = 38

	profileRightColWidth     = 28
	profileRightContentWidth = profileRightColWidth - 6

	profileLongColWidth  = 20
	profileShortColWidth = 8
)

type ProfileScreen struct {
	keymap             util.KeyMap
	agg                *repository.CharacterAggregate
	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	characterInfo       *list.List
	characterAppearance *list.List
	features            *list.List
	backstory           *component.SimpleTextComponent
	appearance          *component.SimpleTextComponent
	personality         *component.SimpleTextComponent
}

func NewProfileScreen(km util.KeyMap, c *repository.CharacterAggregate) *ProfileScreen {
	s := &ProfileScreen{
		keymap:              km,
		agg:                 c,
		backstory:           component.NewSimpleTextComponent(km, "Backstory", &c.Character.Backstory, profileColHeight-4, profileMidColWidth-4),
		appearance:          component.NewSimpleTextComponent(km, "Appearance", &c.Character.Appearance, (profileColHeight-10)/2, profileRightColWidth-4),
		personality:         component.NewSimpleTextComponent(km, "Personality", &c.Character.Personality, (profileColHeight-10)/2, profileRightColWidth-4),
		characterInfo:       list.NewListWithDefaults(km),
		characterAppearance: list.NewListWithDefaults(km),
		features:            list.NewListWithDefaults(km).WithTitle("Features & Traits"),
	}
	return s
}

func (s *ProfileScreen) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, s.characterInfo.Init())
	cmds = append(cmds, s.characterAppearance.Init())
	cmds = append(cmds, s.features.Init())
	cmds = append(cmds, s.backstory.Init())
	cmds = append(cmds, s.appearance.Init())

	s.CreateCharacterInfoRows()
	s.CreateCharacterAppearanceRows()
	s.CreateFeatureRows()

	s.lastFocusedElement = s.characterInfo
	s.focusOn(s.characterInfo)

	return tea.Batch(cmds...)
}

func (s *ProfileScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if msg.Tag == "feature" {
			id := s.agg.AddEmptyFeature()
			s.CreateFeatureRows()
			cmd = editor.SwitchToEditorCmd(
				s.getFeatureRow(id).Editors(),
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

func (s *ProfileScreen) Focus() {
	s.focusOn(s.lastFocusedElement)
}

func (s *ProfileScreen) Blur() {
	// blur should be idempotent
	if s.focusedElement != nil {
		s.lastFocusedElement = s.focusedElement
	}
	s.focusedElement = nil
	s.characterInfo.Blur()
	s.characterAppearance.Blur()
	s.features.Blur()
	s.appearance.Blur()
	s.backstory.Blur()
	s.personality.Blur()
}

func (s *ProfileScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *ProfileScreen) moveFocus(d command.Direction) tea.Cmd {
	var cmd tea.Cmd
	s.Blur()

	switch s.lastFocusedElement {
	case s.characterInfo:
		switch d {
		case command.DownDirection:
			s.focusOn(s.features)
		case command.RightDirection:
			s.focusOn(s.characterAppearance)
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		default:
			s.focusOn(s.characterInfo)
		}
	case s.characterAppearance:
		switch d {
		case command.DownDirection:
			s.focusOn(s.backstory)
		case command.LeftDirection:
			s.focusOn(s.characterInfo)
		default:
			s.focusOn(s.characterAppearance)
		}
	case s.features:
		switch d {
		case command.UpDirection:
			s.focusOn(s.characterInfo)
		case command.RightDirection:
			s.focusOn(s.backstory)
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		default:
			s.focusOn(s.features)
		}
	case s.backstory:
		switch d {
		case command.UpDirection:
			s.focusOn(s.characterInfo)
		case command.RightDirection:
			s.focusOn(s.personality)
		case command.LeftDirection:
			s.focusOn(s.features)
			s.features.SetCursor(0)
		default:
			s.focusOn(s.backstory)
		}
	case s.personality:
		switch d {
		case command.UpDirection:
			s.focusOn(s.characterAppearance)
		case command.LeftDirection:
			s.focusOn(s.backstory)
		case command.DownDirection:
			s.focusOn(s.appearance)
		default:
			s.focusOn(s.personality)
		}
	case s.appearance:
		switch d {
		case command.UpDirection:
			s.focusOn(s.personality)
		case command.LeftDirection:
			s.focusOn(s.backstory)
		default:
			s.focusOn(s.appearance)
		}
	}
	return cmd
}

func (s *ProfileScreen) View() string {
	characterInfo := s.characterInfo.View()

	characterAppearance := s.characterAppearance.View()

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
		Render(s.features.View())

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

	return lipgloss.JoinVertical(lipgloss.Center, topBar, body)
}

func (s *ProfileScreen) RenderBackstory() string {
	title := lipgloss.NewStyle().Width(profileMidColWidth).Render(
		styles.RenderItem(s.backstory.InFocus(), "Backstory") + "\n",
	)
	body := styles.DefaultTextStyle.Width(profileMidColWidth).Render(s.backstory.View())

	return lipgloss.JoinVertical(lipgloss.Center, title, body)
}

func (s *ProfileScreen) RenderPersonality() string {
	title := lipgloss.NewStyle().Width(profileRightColWidth).Render(
		styles.RenderItem(s.personality.InFocus(), "Personality") + "\n",
	)
	body := styles.DefaultTextStyle.Width(profileRightColWidth).Render(s.personality.View())

	return lipgloss.JoinVertical(lipgloss.Center, title, body)
}

func (s *ProfileScreen) RenderAppearance() string {
	title := lipgloss.NewStyle().Width(profileRightColWidth).Render(
		styles.RenderItem(s.appearance.InFocus(), "Appearance") + "\n",
	)
	body := styles.DefaultTextStyle.Width(profileRightColWidth).Render(s.appearance.View())

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

func (s *ProfileScreen) CreateFeatureRows() {
	rows := []list.Row{}
	for i := range s.agg.Features {
		f := &s.agg.Features[i]
		row := list.NewStructRow(s.keymap, f, RenderFeature, []editor.ValueEditor{
			editor.NewStringEditor(s.keymap, "Name", &f.Name),
			editor.NewTextEditor(s.keymap, "Description", &f.Description),
		}).WithDestructor(s.deleteFeatureCallback(f)).
			WithReader(renderFullFeature)
		rows = append(rows, row)
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, "feature"))
	s.features.WithRows(rows)
}

func (s *ProfileScreen) deleteFeatureCallback(f *models.FeatureTO) func() tea.Cmd {
	return func() tea.Cmd {
		s.agg.DeleteFeature(f.ID)
		s.CreateFeatureRows()
		return command.WriteBackRequest
	}
}

func (s *ProfileScreen) getFeatureRow(id uuid.UUID) list.Row {
	for _, r := range s.features.Content() {
		switch r := r.(type) {
		case *list.StructRow[models.FeatureTO]:
			if r.Value().ID == id {
				return r
			}
		}
	}
	return nil
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
