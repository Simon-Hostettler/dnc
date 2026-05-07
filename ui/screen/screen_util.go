package screen

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/util"
)

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
