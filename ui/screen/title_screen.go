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
	screenWidth  = 60
	screenHeight = 15
	inputWidth   = 18
	inputLimit   = 64
)

type TitleScreen struct {
	KeyMap util.KeyMap

	cursor       int
	characters   *list.List
	files        []string
	characterDir string
	editMode     bool
	nameInput    textinput.Model
}

func NewTitleScreen(character_dir string) *TitleScreen {
	ti := textinput.New()
	ti.Width = inputWidth
	ti.CharLimit = inputLimit
	ti.Placeholder = "Character Name"

	t := TitleScreen{
		KeyMap:       util.DefaultKeyMap(),
		cursor:       0,
		characterDir: character_dir,
		editMode:     false,
		nameInput:    ti,
		characters:   list.NewListWithDefaults(),
	}
	return &t
}

func (t *TitleScreen) UpdateFiles() {
	t.files = util.ListCharacterFiles(t.characterDir)
	charRows := util.Map(t.files, func(s string) list.Row { return list.NewCharacterRow(util.PrettyFileName(s), t.characterDir, t.KeyMap) })
	t.characters.WithRows(charRows)
}

func (m *TitleScreen) Init() tea.Cmd {
	return UpdateFilesCmd(m)
}

func (m *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch t := msg.(type) {
	case command.FileOpMsg:
		if t.Success && t.Op != command.FileUpdate {
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
					cmd = command.SaveToFileCmd(&c)
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

	separator := util.MakeHorizontalSeparator(screenWidth/2, 1)

	chars := "\n" + m.characters.View()

	inputField := ""
	if m.editMode && m.cursor == 0 {
		inputField = "\n" + m.nameInput.View()
	}

	return util.DefaultBorderStyle.
		Width(screenWidth).
		Height(screenHeight).
		Render(lipgloss.PlaceVertical(screenHeight, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, createField, inputField, separator, chars)))
}

// to fulfill FocusableModel interface
func (s *TitleScreen) Focus() {}

func (s *TitleScreen) Blur() {}
