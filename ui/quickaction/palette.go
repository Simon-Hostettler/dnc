package quickaction

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

const paletteWidth = 40

type Palette struct {
	keymap      util.KeyMap
	registry    *Registry
	input       textinput.Model
	suggestions []Action
	cursor      int
	active      bool
	errMsg      string
	resultMsg   string
	agg         *repository.CharacterAggregate
}

func NewPalette(km util.KeyMap, registry *Registry) *Palette {
	ti := textinput.New()
	ti.Prompt = ": "
	ti.CharLimit = 64
	ti.SetWidth(paletteWidth - 6)
	return &Palette{
		keymap:   km,
		registry: registry,
		input:    ti,
	}
}

func (p *Palette) SetCharacter(agg *repository.CharacterAggregate) {
	p.agg = agg
}

func (p *Palette) Active() bool {
	return p.active
}

func (p *Palette) Open() {
	p.active = true
	p.input.Reset()
	p.input.Focus()
	p.errMsg = ""
	p.resultMsg = ""
	p.cursor = 0
	p.suggestions = p.registry.All()
}

func (p *Palette) Close() {
	p.active = false
	p.input.Blur()
}

func (p *Palette) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, p.keymap.Escape):
			p.Close()
			return nil
		case key.Matches(msg, p.keymap.Enter):
			return p.execute()
		case key.Matches(msg, p.keymap.Up):
			if p.cursor > 0 {
				p.cursor--
			}
			return nil
		case key.Matches(msg, p.keymap.Down):
			if p.cursor < len(p.suggestions)-1 {
				p.cursor++
			}
			return nil
		case key.Matches(msg, p.keymap.Cycle):
			p.autocomplete()
			return nil
		}
	}

	p.input, _ = p.input.Update(msg)
	p.errMsg = ""
	p.resultMsg = ""
	p.updateSuggestions()
	return nil
}

func (p *Palette) execute() tea.Cmd {
	action, args, found := p.registry.Parse(p.input.Value())
	if !found {
		p.errMsg = "unknown action"
		return nil
	}
	if p.agg == nil {
		p.errMsg = "no character loaded"
		return nil
	}
	res := action.Execute(p.agg, args)
	if res.ErrMsg != "" {
		p.errMsg = res.ErrMsg
		return nil
	}
	if res.Result != "" {
		p.resultMsg = res.Result
		return nil
	}
	p.Close()
	return res.Cmd
}

func (p *Palette) autocomplete() {
	if len(p.suggestions) == 0 {
		return
	}
	if p.cursor >= len(p.suggestions) {
		p.cursor = 0
	}
	selected := p.suggestions[p.cursor]
	value := selected.Name()
	if selected.ArgHint() != "" {
		value += " "
	}
	p.input.SetValue(value)
	p.input.CursorEnd()
}

func (p *Palette) updateSuggestions() {
	val := p.input.Value()
	parts := strings.SplitN(val, " ", 2)
	p.suggestions = p.registry.Match(parts[0])
	if p.cursor >= len(p.suggestions) {
		p.cursor = max(0, len(p.suggestions)-1)
	}
}

var (
	paletteBorder = styles.DefaultBorderStyle.
			Align(lipgloss.Left).
			Padding(1, 2).
			Width(paletteWidth)

	suggestionStyle = styles.GrayTextStyle
	selectedStyle   = lipgloss.NewStyle().Foreground(styles.HighlightColor)
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	hintStyle       = styles.GrayTextStyle
	resultStyle     = lipgloss.NewStyle().Foreground(styles.TextColor)
)

func (p *Palette) View() string {
	var lines []string

	lines = append(lines, p.input.View())

	for i, s := range p.suggestions {
		if i == p.cursor {
			lines = append(lines, selectedStyle.Render("▸ "+s.Name())+hintLabel(s))
		} else {
			lines = append(lines, suggestionStyle.Render("  "+s.Name())+hintLabel(s))
		}
	}

	if p.resultMsg != "" {
		lines = append(lines, resultStyle.Render(p.resultMsg))
	}

	if p.errMsg != "" {
		lines = append(lines, errorStyle.Render(p.errMsg))
	}

	content := strings.Join(lines, "\n")
	return paletteBorder.Render(content)
}

func hintLabel(a Action) string {
	if a.ArgHint() == "" {
		return ""
	}
	return " " + hintStyle.Render(a.ArgHint())
}
