package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var numEditorsVisible = 6

type EditorScreen struct {
	keymap   util.KeyMap
	cursor   int
	vpCursor int
	editors  []editor.ValueEditor
}

func NewEditorScreen(keymap util.KeyMap, editors []editor.ValueEditor) *EditorScreen {
	return &EditorScreen{keymap, 0, 0, editors}
}

func (s *EditorScreen) Init() tea.Cmd {
	return nil
}

func (s *EditorScreen) StartEdit(editors []editor.ValueEditor) {
	s.editors = editors
	if len(s.editors) > 0 {
		s.cursor = 0
		s.vpCursor = 0
		s.focusCurrentRow()
	}
}

func (s *EditorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// editor in focus
	if s.cursor >= 0 && s.cursor < len(s.editors) {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if key.Matches(msg, s.keymap.Escape) {
				cmd = command.SwitchToPrevScreenCmd
			} else {
				cmd = s.editors[s.cursor].Update(msg)
			}
		case command.FocusNextElementMsg:
			s.moveCursor(msg.Direction)
		default:
			cmd = s.editors[s.cursor].Update(msg)
		}
	} else if s.cursor == len(s.editors) { // save button in focus
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, s.keymap.Escape):
				cmd = command.SwitchToPrevScreenCmd
			case key.Matches(msg, s.keymap.Up):
				s.moveCursor(command.UpDirection)
			case key.Matches(msg, s.keymap.Down):
				s.moveCursor(command.DownDirection)
			case key.Matches(msg, s.keymap.Enter):
				cmds := []tea.Cmd{}
				for _, e := range s.editors {
					cmds = append(cmds, e.Save())
				}
				saveCmds := tea.Batch(cmds...)
				cmd = tea.Sequence(saveCmds, command.SwitchToPrevScreenCmd, command.WriteBackRequest)
			}
		}
	}
	return s, cmd
}

func (s *EditorScreen) moveCursor(dir command.Direction) {
	switch dir {
	case command.UpDirection:
		if s.cursor > 0 {
			s.blurCurrentRow()
			s.cursor--
			s.focusCurrentRow()
			if s.cursor < s.vpCursor {
				s.vpCursor = s.cursor
			}
		} else if s.cursor == 0 {
			s.blurCurrentRow()
			s.cursor = len(s.editors)
			s.vpCursor = max(0, s.cursor-numEditorsVisible)
		}
	case command.DownDirection:
		s.blurCurrentRow()
		if s.cursor == len(s.editors) {
			s.cursor = 0
			s.vpCursor = 0
		} else {
			s.cursor++
			if s.vpCursor+numEditorsVisible <= s.cursor && s.cursor <= len(s.editors) {
				s.vpCursor++
			}
		}
		s.focusCurrentRow()
	}
}

func (s *EditorScreen) View() string {
	rows := []string{}
	for _, e := range s.editors {
		rows = append(rows, styles.ForceWidth(e.View(), styles.SmallScreenWidth-8))
	}
	saveButton := styles.RenderItem(s.cursor == len(s.editors), "[ Save ]")

	horizontalSeparator := styles.MakeHorizontalSeparator(styles.SmallScreenWidth-8, 1)

	separated := []string{rows[s.vpCursor]}

	for _, row := range rows[s.vpCursor+1 : s.viewportEnd()] {
		separated = append(separated, horizontalSeparator, row)
	}

	separated = append(separated, horizontalSeparator, saveButton)

	return styles.DefaultBorderStyle.
		Width(styles.SmallScreenWidth).
		Render(lipgloss.JoinVertical(lipgloss.Center, separated...))
}

func (s *EditorScreen) viewportEnd() int {
	return min(len(s.editors), s.vpCursor+numEditorsVisible)
}

func (s *EditorScreen) blurCurrentRow() {
	if s.cursor != len(s.editors) {
		s.editors[s.cursor].Blur()
	}
}

func (s *EditorScreen) focusCurrentRow() {
	if s.cursor != len(s.editors) {
		s.editors[s.cursor].Focus()
	}
}

// to fulfill FocusableModel interface
func (s *EditorScreen) Focus() {}

func (s *EditorScreen) Blur() {}
