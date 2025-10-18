package screen

import (
	"context"

	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/util"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
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
	KeyMap              util.KeyMap
	CharacterRepository repository.CharacterRepository
	Context             context.Context

	cursor     int
	characters *list.List
	editMode   bool
	nameInput  textinput.Model
}

func NewTitleScreen(km util.KeyMap, cr repository.CharacterRepository, ctx context.Context) *TitleScreen {
	ti := textinput.New()
	ti.Width = inputWidth
	ti.CharLimit = inputLimit
	ti.Placeholder = "Character Name"

	t := TitleScreen{
		KeyMap:              km,
		CharacterRepository: cr,
		Context:             ctx,
		cursor:              0,
		editMode:            false,
		nameInput:           ti,
		characters:          list.NewListWithDefaults(),
	}
	return &t
}

func (t *TitleScreen) UpdateFiles() {
	characters := t.listCharacters()
	charRows := util.Map(characters, func(s models.CharacterSummary) list.Row {
		return list.NewCharacterRow(
			t.KeyMap,
			s,
			func() error { return t.deleteCharacter(s.ID) },
		)
	})
	t.characters.WithRows(charRows)
}

func (m *TitleScreen) Init() tea.Cmd {
	return ReloadCharactersCmd(m)
}

func (m *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch t := msg.(type) {
	case command.DataOpMsg:
		if t.Success && t.Op != command.DataUpdate {
			return m, ReloadCharactersCmd(m)
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
				cmd = command.DataOperationCommand(
					func() error { return m.createCharacter(m.nameInput.Value()) },
					command.DataCreate)
				m.nameInput.Reset()
				m.editMode = false
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

func (m *TitleScreen) listCharacters() []models.CharacterSummary {
	chars, err := m.CharacterRepository.ListSummary(m.Context)
	if err != nil {
		return []models.CharacterSummary{}
	}
	return chars
}

func (m *TitleScreen) createCharacter(name string) error {
	_, err := m.CharacterRepository.CreateEmpty(m.Context, name)
	return err
}

func (m *TitleScreen) deleteCharacter(id uuid.UUID) error {
	err := m.CharacterRepository.Delete(m.Context, id)
	return err
}

// to fulfill FocusableModel interface
func (s *TitleScreen) Focus() {}

func (s *TitleScreen) Blur() {}
