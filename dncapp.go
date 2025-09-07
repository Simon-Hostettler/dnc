package main

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui"
)

const (
	defaultPadding int = 2
)

type DnCApp struct {
	pageCursor int
	pages      []tea.Model
	editMode   bool
	config     Config
	width      int
	height     int
}

func NewApp() (*DnCApp, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return &DnCApp{
		pageCursor: 1,
		pages: []tea.Model{
			ui.NewTitleScreen(config.CharacterDir),
		},
		editMode: false,
		config:   config,
	}, nil
}

func (a *DnCApp) GetCurrentPage() tea.Model {
	return a.pages[a.pageCursor-1]
}

func (a *DnCApp) Init() tea.Cmd {
	return nil
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
				a.pageCursor, _ = strconv.Atoi(msg.String())
			}
		}
		_, cmd = a.GetCurrentPage().Update(msg)
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case ui.EditMessage:
		if msg == "start" {
			a.editMode = true
		} else {
			a.editMode = false
		}
	default:
		_, cmd = a.GetCurrentPage().Update(msg)
	}

	return a, cmd
}

func (a *DnCApp) View() string {
	s := ui.AppTitleStyle.Render("DNC") + "\n\n"

	titleHeight := lipgloss.Height(s)

	pageContent := a.GetCurrentPage().View()

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
