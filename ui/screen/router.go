package screen

import (
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
)

type ScreenRouter struct {
	screens        map[command.ScreenIndex]FocusableModel
	modalIndexes   map[command.ScreenIndex]bool
	contentIdx     command.ScreenIndex
	modalStack     []command.ScreenIndex
	characterReady bool
	focused        bool
}

func NewScreenRouter() *ScreenRouter {
	return &ScreenRouter{
		screens:      map[command.ScreenIndex]FocusableModel{},
		modalIndexes: map[command.ScreenIndex]bool{},
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
