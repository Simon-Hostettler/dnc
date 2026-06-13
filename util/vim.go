package util

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type VimLayer int

const (
	VimNormal VimLayer = iota
	VimInsert
)

type VimMode struct {
	Km      KeyMap
	Enabled bool
	Layer   VimLayer
}

func (v *VimMode) TranslateVimBindings(msg tea.KeyPressMsg) tea.KeyPressMsg {
	switch {
	case key.Matches(msg, v.Km.VimUp):
		return BindingToKeyPress(v.Km.Up)
	case key.Matches(msg, v.Km.VimDown):
		return BindingToKeyPress(v.Km.Down)
	case key.Matches(msg, v.Km.VimLeft):
		return BindingToKeyPress(v.Km.Left)
	case key.Matches(msg, v.Km.VimRight):
		return BindingToKeyPress(v.Km.Right)
	case key.Matches(msg, v.Km.VimScreenUp):
		return BindingToKeyPress(v.Km.ScreenUp)
	case key.Matches(msg, v.Km.VimScreenDown):
		return BindingToKeyPress(v.Km.ScreenDown)
	default:
		return msg
	}
}

func (v *VimMode) InNormal() bool {
	return v.Enabled && v.Layer == VimNormal
}

func (v *VimMode) InInsert() bool {
	return v.Enabled && v.Layer == VimInsert
}

type VimModeChangeMsg struct {
	Layer VimLayer
}

func EnterInsertModeCmd() tea.Cmd {
	return func() tea.Msg {
		return VimModeChangeMsg{Layer: VimInsert}
	}
}

func ExitInsertModeCmd() tea.Cmd {
	return func() tea.Msg {
		return VimModeChangeMsg{Layer: VimNormal}
	}
}
