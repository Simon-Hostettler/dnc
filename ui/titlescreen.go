package ui

import (
	"fmt"
	"os"
	"path/filepath"
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
	files := listCharacterFiles(character_dir)
	choices := []string{"Create new Character"}
	if len(files) > 0 {
		choices = append(choices, Map(files, PrettyFileName)...)
	}

	ti := textinput.New()
	ti.Width = 20
	ti.CharLimit = 64
	ti.Placeholder = "Character Name"

	return &TitleScreen{
		choices:      choices,
		files:        files,
		characterDir: character_dir,
		editMode:     false,
		nameInput:    ti,
	}
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
			case "enter":
				c := models.Character{
					Name:     m.nameInput.Value(),
					SaveFile: filepath.Join(m.characterDir, strings.ToLower(m.nameInput.Value())+".json"),
				}
				c.SaveToFile()
				return m, tea.Quit
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
		if m.cursor == i {
			cursor = ">"
		}
		textInput := " "
		if m.cursor == 0 && m.editMode {
			textInput = m.nameInput.View()
		}
		s += fmt.Sprintf("%s %s %s\n", cursor, choice, textInput)
	}
	return s
}
