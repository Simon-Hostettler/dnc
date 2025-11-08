package screen

import (
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"

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

var logo = styles.DefaultTextStyle.Padding(1).Render(
	"______ _   _ _____\n" +
		"|  _  \\ \\ | /  __ \\\n" +
		"| | | |  \\| | /  \\/\n" +
		"| | | | . ` | |\n" +
		"| |/ /| |\\  | \\__/\\\n" +
		"|___/ \\_| \\_/\\____/",
)

var (
	titleScreenWidth  = 50
	titleScreenHeight = 12
	inputWidth        = 18
	inputLimit        = 64
)

type TitleScreen struct {
	KeyMap util.KeyMap

	characters *list.List
	nameInput  textinput.Model
}

func NewTitleScreen(km util.KeyMap) *TitleScreen {
	ti := textinput.New()
	ti.Width = inputWidth
	ti.CharLimit = inputLimit
	ti.Placeholder = "Character Name"

	t := TitleScreen{
		KeyMap:     km,
		nameInput:  ti,
		characters: list.NewListWithDefaults(km),
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
	if m.nameInput.Focused() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.KeyMap.Escape):
				m.nameInput.Reset()
				m.nameInput.Blur()
			case key.Matches(msg, m.KeyMap.Enter):
				name := m.nameInput.Value()
				m.nameInput.Reset()
				m.nameInput.Blur()
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
		default:
			_, cmd = m.characters.Update(msg)
		}
		return m, cmd
	} else {
		// Otherwise
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.KeyMap.Up):
				m.characters.SetCursor(m.characters.Size() - 1)
				m.characters.Focus()
			case key.Matches(msg, m.KeyMap.Down):
				m.characters.SetCursor(0)
				m.characters.Focus()
			case key.Matches(msg, m.KeyMap.Select):
				m.nameInput.Focus()
				cmd = textinput.Blink
			}
		}
	}
	return m, cmd
}

func (m *TitleScreen) View() string {
	createField := styles.RenderItem(!m.characters.InFocus(), "Create new Character")

	separator := styles.MakeHorizontalSeparator(titleScreenWidth/2, 1)

	chars := "\n" + m.characters.View()

	inputField := ""
	if m.nameInput.Focused() {
		inputField = "\n" + m.nameInput.View()
	}

	helperNotice := styles.GrayTextStyle.Render(
		"Press '" + m.KeyMap.ShowKeymap.Keys()[0] + "' to show key bindings",
	)

	return lipgloss.JoinVertical(lipgloss.Center,
		logo,
		styles.DefaultBorderStyle.
			Width(titleScreenWidth).
			Height(titleScreenHeight).
			Render(lipgloss.PlaceVertical(titleScreenHeight, lipgloss.Center,
				lipgloss.JoinVertical(lipgloss.Center, createField, inputField, separator, chars))),
		helperNotice)
}

// to fulfill FocusableModel interface
func (s *TitleScreen) Focus() {}

func (s *TitleScreen) Blur() {}
