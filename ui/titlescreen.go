package ui

import (
	"hostettler.dev/dnc/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TitleScreen struct {
	cursor       int
	choices      []string
	files        []string
	characterDir string
	editMode     bool
	nameInput    textinput.Model
}

func NewTitleScreen(character_dir string) *TitleScreen {
	ti := textinput.New()
	ti.Width = 20
	ti.CharLimit = 64
	ti.Placeholder = "Character Name"

	t := TitleScreen{
		characterDir: character_dir,
		editMode:     false,
		nameInput:    ti,
		choices:      []string{"Create new Character"},
	}
	return &t
}

func (t *TitleScreen) UpdateFiles() {
	t.files = ListCharacterFiles(t.characterDir)
	choices := []string{"Create new Character"}
	if len(t.files) > 0 {
		choices = append(choices, Map(t.files, PrettyFileName)...)
	}
	t.choices = choices
}

func (m *TitleScreen) Init() tea.Cmd {
	return UpdateFilesCmd(m)
}

func (m *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch t := msg.(type) {
	case FileOpMsg:
		if t.success && t.op != "update" {
			return m, UpdateFilesCmd(m)
		}
	}

	if m.editMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.editMode = false
				m.nameInput.Reset()
			case "enter":
				c, err := models.NewCharacter(m.nameInput.Value())
				m.nameInput.Reset()
				m.editMode = false
				if err == nil {
					cmd = tea.Batch(SaveToFileCmd(&c), ExitEditModeCmd)
				}
			default:
				m.nameInput, cmd = m.nameInput.Update(msg)
			}
		default:
			m.nameInput, cmd = m.nameInput.Update(msg)
		}

	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter", " ":
				switch m.cursor {
				case 0:
					m.editMode = true
					m.nameInput.Focus()
					cmd = tea.Batch(textinput.Blink, EnterEditModeCmd)
				default:
					charName := m.choices[m.cursor]
					cmd = SelectCharacterAndSwitchScreenCommand(charName)
				}
			case "x":
				if m.cursor != 0 {
					cmd = DeleteCharacterFileCmd(m.characterDir, m.files[m.cursor-1])
				}
			}
		}
	}
	return m, cmd
}

func (m *TitleScreen) View() string {
	s := ""

	s += RenderList(m.choices[0:1], m.cursor)
	if len(m.choices) > 1 {
		s += "\n"
		s += RenderList(m.choices[1:], m.cursor-1)
	}

	if m.editMode && m.cursor == 0 {
		s += "\n" + m.nameInput.View()
	}

	return s
}
