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

func (f *FocusManager) moveFocus(d command.Direction) tea.Cmd {
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

func RouteKey(focused FocusableModel,
	msg tea.KeyPressMsg,
	km util.KeyMap,
	moveFocus func(command.Direction) tea.Cmd,
) tea.Cmd {
	var cmd tea.Cmd
	switch focused.(type) {
	case *list.List:
		switch {
		case key.Matches(msg, km.Right):
			cmd = moveFocus(command.RightDirection)
		case key.Matches(msg, km.Left):
			cmd = moveFocus(command.LeftDirection)
		default:
			_, cmd = focused.Update(msg)
		}
	default:
		switch {
		case key.Matches(msg, km.Right):
			cmd = moveFocus(command.RightDirection)
		case key.Matches(msg, km.Left):
			cmd = moveFocus(command.LeftDirection)
		case key.Matches(msg, km.Up):
			cmd = moveFocus(command.UpDirection)
		case key.Matches(msg, km.Down):
			cmd = moveFocus(command.DownDirection)
		default:
			_, cmd = focused.Update(msg)
		}
	}
	return cmd
}
