package screen

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/util"
)

type ScreenRouter struct {
	screens        map[command.ScreenIndex]FocusableModel
	modalIndexes   map[command.ScreenIndex]bool
	contentOrder   []command.ScreenIndex
	contentIdx     command.ScreenIndex
	modalStack     []command.ScreenIndex
	characterReady bool
	focused        bool
}

func NewScreenRouter(order []command.ScreenIndex) *ScreenRouter {
	return &ScreenRouter{
		screens:      map[command.ScreenIndex]FocusableModel{},
		modalIndexes: map[command.ScreenIndex]bool{},
		contentOrder: order,
	}
}

// Adds a screen to the registry and returns its Init() cmd so the caller can batch it.
func (r *ScreenRouter) Register(idx command.ScreenIndex, m FocusableModel, isModal bool) tea.Cmd {
	r.screens[idx] = m
	if isModal {
		r.modalIndexes[idx] = true
	}
	return m.Init()
}

func (r *ScreenRouter) IsModal(idx command.ScreenIndex) bool {
	return r.modalIndexes[idx]
}

func (r *ScreenRouter) MarkCharacterReady()    { r.characterReady = true }
func (r *ScreenRouter) IsCharacterReady() bool { return r.characterReady }

func (r *ScreenRouter) SwitchContent(idx command.ScreenIndex) {
	if _, ok := r.screens[idx]; !ok || r.modalIndexes[idx] {
		return
	}
	if !r.characterReady && idx != command.TitleScreenIndex {
		return
	}
	r.blurActive()
	r.contentIdx = idx
	r.focusActive()
}

func (r *ScreenRouter) PushModal(idx command.ScreenIndex) {
	if !r.modalIndexes[idx] {
		return
	}
	r.blurActive()
	r.modalStack = append(r.modalStack, idx)
	r.focusActive()
}

func (r *ScreenRouter) PopModal() {
	if len(r.modalStack) == 0 {
		return
	}
	r.blurActive()
	r.modalStack = r.modalStack[:len(r.modalStack)-1]
	r.focusActive()
}

func (r *ScreenRouter) Active() FocusableModel { return r.screens[r.ActiveIndex()] }
func (r *ScreenRouter) ActiveIndex() command.ScreenIndex {
	if n := len(r.modalStack); n > 0 {
		return r.modalStack[n-1]
	}
	return r.contentIdx
}
func (r *ScreenRouter) ContentIndex() command.ScreenIndex { return r.contentIdx }
func (r *ScreenRouter) InModal() bool                     { return len(r.modalStack) > 0 }
func (r *ScreenRouter) IsFocused() bool                   { return r.focused }

func (r *ScreenRouter) Focus() {
	r.focused = true
	if m := r.Active(); m != nil {
		m.Focus()
	}
}

func (r *ScreenRouter) Blur() {
	if m := r.Active(); m != nil {
		m.Blur()
	}
	r.focused = false
}

func (r *ScreenRouter) blurActive() {
	if m := r.Active(); m != nil {
		m.Blur()
	}
}

func (r *ScreenRouter) focusActive() {
	if !r.focused {
		return
	}
	if m := r.Active(); m != nil {
		m.Focus()
	}
}

// nil if msg is not a navigation key
func (r *ScreenRouter) NavCmd(msg tea.KeyPressMsg, km util.KeyMap) tea.Cmd {
	if r.InModal() {
		return nil // don't steal keys from modals/editors
	}
	switch {
	case key.Matches(msg, km.Screen1):
		return r.jumpCmd(0)
	case key.Matches(msg, km.Screen2):
		return r.jumpCmd(1)
	case key.Matches(msg, km.Screen3):
		return r.jumpCmd(2)
	case key.Matches(msg, km.Screen4):
		return r.jumpCmd(3)
	case key.Matches(msg, km.Screen5):
		return r.jumpCmd(4)
	case key.Matches(msg, km.ScreenDown):
		return r.stepCmd(1)
	case key.Matches(msg, km.ScreenUp):
		return r.stepCmd(-1)
	default:
		return nil
	}
}

func (r *ScreenRouter) jumpCmd(i int) tea.Cmd {
	if i < 0 || i >= len(r.contentOrder) {
		return nil
	}
	return tea.Batch(command.SwitchScreenCmd(r.contentOrder[i]), command.FocusActiveScreenCmd)
}

func (r *ScreenRouter) stepCmd(delta int) tea.Cmd {
	n := len(r.contentOrder)
	cur := r.currentOrderPos()
	if n == 0 || cur < 0 {
		return nil
	}
	next := ((cur+delta)%n + n) % n
	return tea.Batch(command.SwitchScreenCmd(r.contentOrder[next]), command.FocusActiveScreenCmd)
}

func (r *ScreenRouter) currentOrderPos() int {
	for i, idx := range r.contentOrder {
		if idx == r.contentIdx {
			return i
		}
	}
	return -1
}
