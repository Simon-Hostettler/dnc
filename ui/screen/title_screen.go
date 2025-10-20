package screen

import (
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/util"

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
	titleScreenWidth  = 40
	titleScreenHeight = 15
	inputWidth        = 18
	inputLimit        = 64
)

type TitleScreen struct {
	KeyMap util.KeyMap

	cursor     int
	characters *list.List
	editMode   bool
	nameInput  textinput.Model
}

func NewTitleScreen() *TitleScreen {
	ti := textinput.New()
	ti.Width = inputWidth
	ti.CharLimit = inputLimit
	ti.Placeholder = "Character Name"

	t := TitleScreen{
		KeyMap:     util.DefaultKeyMap(),
		cursor:     0,
		editMode:   false,
		nameInput:  ti,
		characters: list.NewListWithDefaults(),
	}
	return &t
}

func (t *TitleScreen) SetSummaries(s []models.CharacterSummary) {
	charRows := util.Map(s, func(sum models.CharacterSummary) list.Row {
		return list.NewCharacterRow(t.KeyMap, &sum)
	})
	t.characters.WithRows(charRows)
}

func (m *TitleScreen) Init() tea.Cmd {
	return command.LoadSummariesRequest
}

func (m *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// New character creation
	if m.editMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.KeyMap.Escape):
				m.editMode = false
				m.nameInput.Reset()
			case key.Matches(msg, m.KeyMap.Enter):
				name := m.nameInput.Value()
				m.nameInput.Reset()
				m.editMode = false
				cmd = command.CreateCharacterRequest(name)
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
		case command.FocusNextElementMsg:
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
		}
	}
	return m, cmd
}

func (m *TitleScreen) View() string {
	createField := util.RenderItem(m.cursor == 0, "Create new Character")

	separator := util.MakeHorizontalSeparator(titleScreenWidth/2, 1)

	chars := "\n" + m.characters.View()

	inputField := ""
	if m.editMode && m.cursor == 0 {
		inputField = "\n" + m.nameInput.View()
	}

	return util.DefaultBorderStyle.
		Width(titleScreenWidth).
		Height(titleScreenHeight).
		Render(lipgloss.PlaceVertical(titleScreenHeight, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, createField, inputField, separator, chars)))
}

// to fulfill FocusableModel interface
func (s *TitleScreen) Focus() {}

func (s *TitleScreen) Blur() {}
