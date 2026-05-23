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

	nodes []*editorNode
	save  *saveButton
	vpTop int
}

func NewEditorScreen(keymap util.KeyMap) *EditorScreen {
	return &EditorScreen{
		keymap: keymap,
		save:   &saveButton{keymap: keymap},
	}
}

func (s *EditorScreen) Init() tea.Cmd { return nil }

func (s *EditorScreen) StartEdit(editors []editor.ValueEditor) {
	s.nodes = make([]*editorNode, len(editors))
	for i, e := range editors {
		s.nodes[i] = &editorNode{editor: e}
	}
	s.save.onSave = s.buildSaveCmd()
	s.vpTop = 0

	initial := s.firstEnabled()
	if initial == nil {
		initial = FocusableModel(s.save)
	}
	s.Wire(s.graph(), initial)
	s.Focus()
	s.adjustViewport()
}

func (s *EditorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, s.keymap.Escape) && !util.IsLetterKey(msg) {
			return s, command.SwitchToPrevScreenCmd
		}
		if focused := s.Focused(); focused != nil {
			_, cmd := focused.Update(msg)
			return s, cmd
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
	separated = append(separated, horizontalSeparator, s.save.render())

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
	g[s.save] = map[command.Direction]FocusEdge{
		command.UpDirection:   ToCond(func() FocusableModel { return s.lastEnabled() }),
		command.DownDirection: ToCond(func() FocusableModel { return s.firstEnabled() }),
	}
	return g
}

func (s *EditorScreen) nextEnabled(from int) FocusableModel {
	for i := from + 1; i < len(s.nodes); i++ {
		if !s.nodes[i].Disabled() {
			return s.nodes[i]
		}
	}
	return s.save
}

func (s *EditorScreen) prevEnabled(from int) FocusableModel {
	for i := from - 1; i >= 0; i-- {
		if !s.nodes[i].Disabled() {
			return s.nodes[i]
		}
	}
	return s.save
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

func isDisabled(e editor.ValueEditor) bool {
	if d, ok := e.(interface{ Disabled() bool }); ok {
		return d.Disabled()
	}
	return false
}

type saveButton struct {
	keymap  util.KeyMap
	focused bool
	onSave  func() tea.Cmd
}

func (b *saveButton) Init() tea.Cmd { return nil }

func (b *saveButton) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return b, nil
	}
	switch {
	case key.Matches(keyMsg, b.keymap.Enter):
		if b.onSave != nil {
			return b, b.onSave()
		}
	case key.Matches(keyMsg, b.keymap.Up):
		return b, command.FocusNextElementCmd(command.UpDirection)
	case key.Matches(keyMsg, b.keymap.Down):
		return b, command.FocusNextElementCmd(command.DownDirection)
	}
	return b, nil
}

func (b *saveButton) View() tea.View { return tea.NewView(b.render()) }

func (b *saveButton) render() string { return styles.RenderItem(b.focused, "[ Save ]") }

func (b *saveButton) Focus() { b.focused = true }

func (b *saveButton) Blur() { b.focused = false }
