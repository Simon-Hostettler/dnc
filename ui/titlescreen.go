package ui

import (
	"fmt"
	"os"
	"strings"

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
	}
	t.updateFiles()

	return &t
}

func (t *TitleScreen) updateFiles() {
	t.files = listCharacterFiles(t.characterDir)
	choices := []string{"Create new Character"}
	if len(t.files) > 0 {
		choices = append(choices, Map(t.files, PrettyFileName)...)
	}
	t.choices = choices
}

func listCharacterFiles(dir string) []string {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, entry.Name())
		}
	}
	return files
}

func (m *TitleScreen) Init() tea.Cmd {
	return nil
}

func (m *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.editMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.editMode = false
				m.nameInput.Reset()
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				c, err := models.NewCharacter(m.nameInput.Value())
				if err == nil {
					c.SaveToFile()
					m.updateFiles()
				}
				m.nameInput.Reset()
				m.editMode = false
			}

		}
		m.nameInput, cmd = m.nameInput.Update(msg)
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
					cmd = textinput.Blink
				default:

				}
				return m, nil
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		}
	}
	return m, cmd
}

func (m *TitleScreen) View() string {
	s := ""
	for i, choice := range m.choices {

		cursor := " "

		textInput := " "
		if m.cursor == i {
			cursor = ">"
			if m.cursor == 0 && m.editMode {
				textInput = m.nameInput.View()
			}
		}

		s += fmt.Sprintf("%s %s %s\n", cursor, choice, textInput)
	}
	return s
}
