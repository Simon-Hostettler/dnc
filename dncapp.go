package main

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui"
)

const (
	defaultPadding int = 2
)

type DnCApp struct {
	page      tea.Model
	editMode  bool
	config    Config
	width     int
	height    int
	character *models.Character
}

func NewApp() (*DnCApp, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return &DnCApp{
		page:     ui.NewTitleScreen(config.CharacterDir),
		editMode: false,
		config:   config,
	}, nil
}

func (a *DnCApp) Init() tea.Cmd {
	return a.page.Init()
}

func (a *DnCApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
		switch msg.String() {
		case "1", "2", "3", "4", "5":
			if !a.editMode {
				p, _ := strconv.Atoi(msg.String())
				idx := ui.ScreenIndex(p - 1)
				return a, ui.SwitchScreenCmd(idx)
			}
		}
		_, cmd = a.page.Update(msg)
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case ui.EditMessage:
		if msg == "start" {
			a.editMode = true
		} else {
			a.editMode = false
		}
	case ui.SwitchScreenMsg:
		switch msg.Screen {
		case ui.ScoreScreenIndex:
			a.page = ui.NewScoreScreen(a.character)
			cmd = a.page.Init()
		}
	case ui.SelectCharacterAndSwitchScreenMsg:
		if msg.Err == nil {
			a.character = msg.Character
			cmd = ui.SwitchScreenCmd(ui.ScoreScreenIndex)
		}
	default:
		_, cmd = a.page.Update(msg)
	}

	return a, cmd
}

func (a *DnCApp) View() string {
	s := ui.AppTitleStyle.Render("DNC") + "\n\n"

	titleHeight := lipgloss.Height(s)

	pageContent := a.page.View()

	pageWidth := a.width - defaultPadding
	pageHeight := a.height - titleHeight - defaultPadding

	topPad := (pageHeight - lipgloss.Height(pageContent)) / 2
	leftPad := (pageWidth - lipgloss.Width(pageContent)) / 2

	s += ui.MainBorderStyle.
		Width(pageWidth).
		Height(pageHeight).
		PaddingLeft(leftPad).
		PaddingTop(topPad).
		Render(pageContent)

	return s

}
