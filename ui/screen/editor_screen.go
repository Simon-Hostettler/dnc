package screen

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var numEditorsVisible = 6

type EditorScreen struct {
	keymap util.KeyMap
	FocusManager

	// read-only state, owned by dncapp
	vimMode *util.VimMode

	nodes  []*editorNode
	onSave func() tea.Cmd
	vpTop  int
}

func NewEditorScreen(km util.KeyMap, vimMode *util.VimMode) *EditorScreen {
	return &EditorScreen{
		keymap:  km,
		vimMode: vimMode,
	}
}

func (s *EditorScreen) Init() tea.Cmd { return nil }

func (s *EditorScreen) StartEdit(editors []editor.ValueEditor) {
	s.nodes = make([]*editorNode, len(editors))
	for i, e := range editors {
		e.Reload()
		s.nodes[i] = &editorNode{editor: e}
	}
	s.onSave = s.buildSaveCmd()
	s.vpTop = 0

	s.Wire(s.graph(), s.firstEnabled())
	s.Focus()
	s.adjustViewport()
}

func (s *EditorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, s.keymap.Escape) && !util.IsLetterKey(msg) {
			if s.vimMode.InInsert() {
				return s, util.ExitInsertModeCmd()
			}
			return s, command.SwitchToPrevScreenCmd
		}
		if key.Matches(msg, s.keymap.Save) && !s.vimMode.InInsert() && s.onSave != nil {
			return s, s.onSave()
		}
		if focused := s.Focused(); focused != nil {
			if s.vimMode.InNormal() && capturesTextInput(focused) {
				switch {
				case key.Matches(msg, s.keymap.VimInsert):
					return s, util.EnterInsertModeCmd()
				case key.Matches(msg, s.keymap.Up):
					return s, command.FocusNextElementCmd(command.UpDirection)
				case key.Matches(msg, s.keymap.Down):
					return s, command.FocusNextElementCmd(command.DownDirection)
				default:
					return s, nil
				}
			} else {
				_, cmd := focused.Update(msg)
				return s, cmd
			}
		}
	case command.FocusNextElementMsg:
		cmd := s.MoveFocus(msg.Direction)
		s.adjustViewport()
		return s, cmd
	default:
		if focused := s.Focused(); focused != nil {
			_, cmd := focused.Update(msg)
			return s, cmd
		}
	}
	return s, nil
}

func (s *EditorScreen) View() tea.View {
	rows := []string{}
	for _, n := range s.nodes {
		rows = append(rows, styles.ForceWidth(n.editor.View(), styles.SmallScreenWidth-8))
	}

	horizontalSeparator := styles.MakeHorizontalSeparator(styles.SmallScreenWidth-8, 1)

	end := min(len(rows), s.vpTop+numEditorsVisible)
	separated := []string{}
	if s.vpTop < end {
		separated = append(separated, rows[s.vpTop])
		for _, row := range rows[s.vpTop+1 : end] {
			separated = append(separated, horizontalSeparator, row)
		}
	}

	footer := "\n" + styles.RenderKeyBinding(s.keymap.Save) + ": save"

	if s.vimMode.InNormal() && s.Focused() != nil && capturesTextInput(s.Focused()) {
		footer += " ∙ " + styles.RenderKeyBinding(s.keymap.VimInsert) + ": insert"
	}

	if s.vimMode.InInsert() {
		footer += " ∙ " + styles.RenderKeyBinding(s.keymap.Escape) + ": exit insert"
	} else {
		footer += " ∙ " + styles.RenderKeyBinding(s.keymap.Escape) + ": back"
	}

	separated = append(separated, "\n"+styles.GrayTextStyle.Render(footer))

	return tea.NewView(styles.DefaultBorderStyle.
		Width(styles.SmallScreenWidth).
		Render(lipgloss.JoinVertical(lipgloss.Center, separated...)))
}

func (s *EditorScreen) buildSaveCmd() func() tea.Cmd {
	return func() tea.Cmd {
		cmds := make([]tea.Cmd, 0, len(s.nodes))
		for _, n := range s.nodes {
			cmds = append(cmds, n.editor.Save())
		}
		return tea.Sequence(tea.Batch(cmds...), command.SwitchToPrevScreenCmd, command.WriteBackRequest)
	}
}

func (s *EditorScreen) graph() FocusGraph {
	g := FocusGraph{}
	for i, n := range s.nodes {
		g[n] = map[command.Direction]FocusEdge{
			command.UpDirection:   ToCond(func() FocusableModel { return s.prevEnabled(i) }),
			command.DownDirection: ToCond(func() FocusableModel { return s.nextEnabled(i) }),
		}
	}
	return g
}

func (s *EditorScreen) nextEnabled(from int) FocusableModel {
	for i := from + 1; i < len(s.nodes); i++ {
		if !s.nodes[i].Disabled() {
			return s.nodes[i]
		}
	}
	return s.firstEnabled()
}

func (s *EditorScreen) prevEnabled(from int) FocusableModel {
	for i := from - 1; i >= 0; i-- {
		if !s.nodes[i].Disabled() {
			return s.nodes[i]
		}
	}
	return s.lastEnabled()
}

func (s *EditorScreen) firstEnabled() FocusableModel {
	for _, n := range s.nodes {
		if !n.Disabled() {
			return n
		}
	}
	return nil
}

func (s *EditorScreen) lastEnabled() FocusableModel {
	for i := len(s.nodes) - 1; i >= 0; i-- {
		if !s.nodes[i].Disabled() {
			return s.nodes[i]
		}
	}
	return nil
}

func (s *EditorScreen) focusedIndex() int {
	focused := s.Focused()
	for i, n := range s.nodes {
		if n == focused {
			return i
		}
	}
	return -1
}

func capturesTextInput(m tea.Model) bool {
	if c, ok := m.(interface{ CapturesTextInput() bool }); ok {
		return c.CapturesTextInput()
	}
	return false
}

func (s *EditorScreen) adjustViewport() {
	idx := s.focusedIndex()
	switch {
	case idx < 0:
		s.vpTop = max(0, len(s.nodes)-numEditorsVisible)
	case idx < s.vpTop:
		s.vpTop = idx
	case idx >= s.vpTop+numEditorsVisible:
		s.vpTop = idx - numEditorsVisible + 1
	}
}

// adapts a ValueEditor to the FocusableModel interface
type editorNode struct {
	editor editor.ValueEditor
}

func (n *editorNode) Init() tea.Cmd { return nil }

func (n *editorNode) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return n, n.editor.Update(msg)
}

func (n *editorNode) View() tea.View { return tea.NewView(n.editor.View()) }

func (n *editorNode) Focus() { n.editor.Focus() }

func (n *editorNode) Blur() { n.editor.Blur() }

func (n *editorNode) Disabled() bool { return isDisabled(n.editor) }

func (n *editorNode) CapturesTextInput() bool { return n.editor.CapturesTextInput() }

func isDisabled(e editor.ValueEditor) bool {
	if d, ok := e.(interface{ Disabled() bool }); ok {
		return d.Disabled()
	}
	return false
}
