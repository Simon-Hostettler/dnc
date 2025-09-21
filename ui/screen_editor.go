package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
)

type EditorScreen struct {
	keymap     KeyMap
	prevScreen ScreenIndex
	character  *models.Character
	cursor     int
	editors    []ValueEditor
}

func NewEditorScreen(keymap KeyMap, editors []ValueEditor) *EditorScreen {
	return &EditorScreen{keymap, EditScreenIndex, nil, 0, editors}
}

func (s *EditorScreen) Init() tea.Cmd {
	return nil
}

func (s *EditorScreen) StartEdit(prevScreen ScreenIndex, c *models.Character, editors []ValueEditor) {
	s.prevScreen = prevScreen
	s.character = c
	s.editors = editors
	if len(s.editors) > 0 {
		s.cursor = 0
		s.editors[0].Focus()
	}
}

func (s *EditorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// editor in focus
	if s.cursor >= 0 && s.cursor < len(s.editors) {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, s.keymap.Up):
				if s.cursor > 0 {
					s.editors[s.cursor].Blur()
					s.cursor--
					s.editors[s.cursor].Focus()
				}
			case key.Matches(msg, s.keymap.Down, s.keymap.Enter):
				s.editors[s.cursor].Blur()
				s.cursor++
				if s.cursor < len(s.editors) {
					s.editors[s.cursor].Focus()
				}
			default:
				cmd = s.editors[s.cursor].Update(msg)
			}
		}
	} else if s.cursor == len(s.editors) { // save button in focus
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, s.keymap.Up):
				s.cursor--
				s.editors[s.cursor].Focus()
			case key.Matches(msg, s.keymap.Enter):
				for _, e := range s.editors {
					e.Save()
				}
				cmd = tea.Batch(SaveToFileCmd(s.character), SwitchScreenCmd(s.prevScreen))
			}
		}
	}
	return s, cmd
}

func (s *EditorScreen) View() string {
	rows := []string{}
	for _, e := range s.editors {
		rows = append(rows, ForceWidth(e.View(), SmallScreenWidth-8))
	}
	saveButton := RenderItem(s.cursor == len(s.editors), "[ Save ]")
	rows = append(rows, saveButton)

	horizontalSeparator := MakeHorizontalSeparator(SmallScreenWidth - 8)

	separated := []string{rows[0]}

	for _, row := range rows[1:] {
		separated = append(separated, horizontalSeparator, row)
	}

	return DefaultBorderStyle.
		Width(SmallScreenWidth).
		Render(lipgloss.JoinVertical(lipgloss.Center, separated...))
}

// to fulfill FocusableModel interface
func (s *EditorScreen) Focus() {}

func (s *EditorScreen) Blur() {}
