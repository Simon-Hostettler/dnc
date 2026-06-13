package screen

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/util"
)

type FocusEdge func() (target FocusableModel, cmd tea.Cmd)

type FocusGraph map[FocusableModel]map[command.Direction]FocusEdge

// FocusManager holds the focus state shared by every screen and is meant to be
// embedded: its fields and methods promote into the embedding screen.
type FocusManager struct {
	focusGraph         FocusGraph
	lastFocusedElement FocusableModel
	focusedElement     FocusableModel
}

func (f *FocusManager) Focus() {
	f.focusOn(f.lastFocusedElement)
}

func (f *FocusManager) focusOn(m FocusableModel) {
	f.focusedElement = m
	m.Focus()
}

// Blur is idempotent: only the focused element can be active, so blurring it
// alone is sufficient.
func (f *FocusManager) Blur() {
	if f.focusedElement != nil {
		f.focusedElement.Blur()
		f.lastFocusedElement = f.focusedElement
	}
	f.focusedElement = nil
}

func (f *FocusManager) MoveFocus(d command.Direction) tea.Cmd {
	edge, ok := f.focusGraph[f.focusedElement][d]
	if !ok {
		return nil
	}
	target, cmd := edge()
	if target != nil {
		f.Blur()
		f.focusOn(target)
	}
	return cmd
}

func (f *FocusManager) Focused() FocusableModel { return f.focusedElement }

// Wire installs the focus graph and the element that Focus() should resume on
func (f *FocusManager) Wire(g FocusGraph, initial FocusableModel) {
	f.focusGraph = g
	f.lastFocusedElement = initial
}

func To(m FocusableModel) FocusEdge {
	return func() (FocusableModel, tea.Cmd) { return m, nil }
}

func ToCond(pick func() FocusableModel) FocusEdge {
	return func() (FocusableModel, tea.Cmd) { return pick(), nil }
}

func ToWith(m FocusableModel, sideEffect func()) FocusEdge {
	return func() (FocusableModel, tea.Cmd) {
		sideEffect()
		return m, nil
	}
}

func Emit(cmd tea.Cmd) FocusEdge {
	return func() (FocusableModel, tea.Cmd) { return nil, cmd }
}

func (f *FocusManager) RouteKey(
	msg tea.KeyPressMsg,
	km util.KeyMap,
) tea.Cmd {
	var cmd tea.Cmd
	switch fe := f.focusedElement.(type) {
	case *list.List:
		switch {
		case !fe.SearchInputFocused() && key.Matches(msg, km.Right):
			cmd = f.MoveFocus(command.RightDirection)
		case !fe.SearchInputFocused() && key.Matches(msg, km.Left):
			cmd = f.MoveFocus(command.LeftDirection)
		default:
			_, cmd = fe.Update(msg)
		}
	default:
		switch {
		case key.Matches(msg, km.Right):
			cmd = f.MoveFocus(command.RightDirection)
		case key.Matches(msg, km.Left):
			cmd = f.MoveFocus(command.LeftDirection)
		case key.Matches(msg, km.Up):
			cmd = f.MoveFocus(command.UpDirection)
		case key.Matches(msg, km.Down):
			cmd = f.MoveFocus(command.DownDirection)
		default:
			_, cmd = f.focusedElement.Update(msg)
		}
	}
	return cmd
}
