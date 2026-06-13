package util

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// builds KeyPressMsg for a binding's first key string.
func BindingToKeyPress(b key.Binding) tea.KeyPressMsg {
	keys := b.Keys()
	if len(keys) == 0 {
		return tea.KeyPressMsg{}
	}
	return keyStringToKeyPress(keys[0])
}

// parses key string into tea.KeyPressMessage, mirroring tea's rules
func keyStringToKeyPress(s string) tea.KeyPressMsg {
	var (
		mod  tea.KeyMod
		code rune
		text string
	)
	for part := range strings.SplitSeq(s, "+") {
		if m, ok := keyMods[part]; ok {
			mod |= m
			continue
		}
		if c, ok := keyNames[part]; ok {
			code = c
		} else if utf8.RuneCountInString(part) == 1 {
			code, _ = utf8.DecodeRuneInString(part)
		} else {
			code = tea.KeyExtended
			text = part
		}
	}

	// Printable keys carry their text, uppercased when shifted.
	if mod&^(tea.ModShift|tea.ModCapsLock) == 0 && text == "" && unicode.IsPrint(code) {
		if mod&(tea.ModShift|tea.ModCapsLock) != 0 {
			text = string(unicode.ToUpper(code))
		} else {
			text = string(code)
		}
	}

	return tea.KeyPressMsg{Mod: mod, Code: code, Text: text}
}

var keyMods = map[string]tea.KeyMod{
	"ctrl":       tea.ModCtrl,
	"alt":        tea.ModAlt,
	"shift":      tea.ModShift,
	"meta":       tea.ModMeta,
	"hyper":      tea.ModHyper,
	"super":      tea.ModSuper,
	"capslock":   tea.ModCapsLock,
	"scrolllock": tea.ModScrollLock,
	"numlock":    tea.ModNumLock,
}

var keyNames = map[string]rune{
	"enter":     tea.KeyEnter,
	"tab":       tea.KeyTab,
	"backspace": tea.KeyBackspace,
	"escape":    tea.KeyEscape,
	"esc":       tea.KeyEscape,
	"space":     tea.KeySpace,
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"begin":     tea.KeyBegin,
	"insert":    tea.KeyInsert,
	"delete":    tea.KeyDelete,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
}
