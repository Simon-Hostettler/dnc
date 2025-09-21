package ui

import (
	"hostettler.dev/dnc/models"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusableModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Focus()
	Blur()
}

var (
	screenWidth  = 60
	screenHeight = 15
)

type TitleScreen struct {
	KeyMap KeyMap

	cursor       int
	characters   *List
	files        []string
	characterDir string
	editMode     bool
	nameInput    textinput.Model
}

func NewTitleScreen(character_dir string) *TitleScreen {
	ti := textinput.New()
	ti.Width = 18
	ti.CharLimit = 64
	ti.Placeholder = "Character Name"

	t := TitleScreen{
		KeyMap:       DefaultKeyMap(),
		cursor:       0,
		characterDir: character_dir,
		editMode:     false,
		nameInput:    ti,
		characters:   NewListWithDefaults(),
	}
	return &t
}

func (t *TitleScreen) UpdateFiles() {
	t.files = ListCharacterFiles(t.characterDir)
	charRows := Map(t.files, func(s string) Row { return NewCharacterRow(PrettyFileName(s), t.characterDir, t.KeyMap) })
	t.characters.WithRows(charRows)
}

func (m *TitleScreen) Init() tea.Cmd {
	return UpdateFilesCmd(m)
}

func (m *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch t := msg.(type) {
	case FileOpMsg:
		if t.success && t.op != FileUpdate {
			return m, UpdateFilesCmd(m)
		}
	}

	// New character creation
	if m.editMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.KeyMap.Escape):
				m.editMode = false
				m.nameInput.Reset()
			case key.Matches(msg, m.KeyMap.Enter):
				c, err := models.NewCharacter(m.nameInput.Value())
				m.nameInput.Reset()
				m.editMode = false
				if err == nil {
					cmd = SaveToFileCmd(&c)
				}
			default:
				m.nameInput, cmd = m.nameInput.Update(msg)
			}
		default:
			m.nameInput, cmd = m.nameInput.Update(msg)
		}
		return m, cmd
	}

	// Character selection
	if m.characters.InFocus() {
		switch msg.(type) {
		case FocusNextElementMsg:
			m.characters.Blur()
			m.cursor = 0
			return m, nil
		default:
			_, cmd = m.characters.Update(msg)
		}
		return m, cmd
	}

	// Otherwise
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Down):
			if m.cursor == 0 {
				m.cursor++
			}
			m.characters.Focus()
		case key.Matches(msg, m.KeyMap.Select):
			if m.cursor == 0 {
				m.editMode = true
				m.nameInput.Focus()
				cmd = textinput.Blink
			}
		case key.Matches(msg, m.KeyMap.Delete):
			if m.cursor != 0 {
				cmd = DeleteCharacterFileCmd(m.characterDir, m.files[m.cursor-1])
			}
		}
	}
	return m, cmd
}

func (m *TitleScreen) View() string {
	createField := RenderItem(m.cursor == 0, "Create new Character")

	separator := MakeHorizontalSeparator(screenWidth - 8)

	chars := "\n" + m.characters.View()

	inputField := ""
	if m.editMode && m.cursor == 0 {
		inputField = "\n" + m.nameInput.View()
	}

	return DefaultBorderStyle.
		Width(screenWidth).
		Height(screenHeight).
		Render(lipgloss.PlaceVertical(screenHeight, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, createField, inputField, separator, chars)))
}

// to fulfill FocusableModel interface
func (s *TitleScreen) Focus() {}

func (s *TitleScreen) Blur() {}
