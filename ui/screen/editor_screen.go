package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/util"
)

type EditorScreen struct {
	keymap    util.KeyMap
	character *models.Character
	cursor    int
	editors   []editor.ValueEditor
}

func NewEditorScreen(keymap util.KeyMap, editors []editor.ValueEditor) *EditorScreen {
	return &EditorScreen{keymap, nil, 0, editors}
}

func (s *EditorScreen) Init() tea.Cmd {
	return nil
}

func (s *EditorScreen) StartEdit(c *models.Character, editors []editor.ValueEditor) {
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
				} else if s.cursor == 0 {
					s.editors[s.cursor].Blur()
					s.cursor = len(s.editors)
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
			case key.Matches(msg, s.keymap.Down):
				s.cursor = 0
				s.editors[s.cursor].Focus()
			case key.Matches(msg, s.keymap.Enter):
				for _, e := range s.editors {
					e.Save()
				}
				cmd = tea.Batch(command.SaveToFileCmd(s.character), command.SwitchToPrevScreenCmd)
			}
		}
	}
	return s, cmd
}

func (s *EditorScreen) View() string {
	rows := []string{}
	for _, e := range s.editors {
		rows = append(rows, util.ForceWidth(e.View(), util.SmallScreenWidth-8))
	}
	saveButton := util.RenderItem(s.cursor == len(s.editors), "[ Save ]")
	rows = append(rows, saveButton)

	horizontalSeparator := util.MakeHorizontalSeparator(util.SmallScreenWidth-8, 1)

	separated := []string{rows[0]}

	for _, row := range rows[1:] {
		separated = append(separated, horizontalSeparator, row)
	}

	return util.DefaultBorderStyle.
		Width(util.SmallScreenWidth).
		Render(lipgloss.JoinVertical(lipgloss.Center, separated...))
}

// to fulfill FocusableModel interface
func (s *EditorScreen) Focus() {}

func (s *EditorScreen) Blur() {}
