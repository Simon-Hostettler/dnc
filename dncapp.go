package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui"
)

const (
	defaultPadding int = 2
)

type DnCApp struct {
	config        Config
	keymap        ui.KeyMap
	width         int
	height        int
	character     *models.Character
	focusedScreen tea.Model
	titleScreen   *ui.TitleScreen
	scoreScreen   *ui.ScoreScreen
	editorScreen  *ui.EditorScreen
}

func NewApp() (*DnCApp, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return &DnCApp{
		config:       config,
		keymap:       ui.DefaultKeyMap(),
		titleScreen:  ui.NewTitleScreen(config.CharacterDir),
		editorScreen: ui.NewEditorScreen(ui.DefaultKeyMap(), []ui.ValueEditor{}),
	}, nil
}

func (a *DnCApp) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	if a.titleScreen != nil {
		cmds = append(cmds, a.titleScreen.Init())
		a.focusedScreen = a.titleScreen
	}
	if a.scoreScreen != nil {
		cmds = append(cmds, a.scoreScreen.Init())
	}
	if a.editorScreen != nil {
		cmds = append(cmds, a.editorScreen.Init())
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (a *DnCApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keymap.ForceQuit):
			return a, tea.Quit
		}
		_, cmd = a.focusedScreen.Update(msg)
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case ui.SwitchScreenMsg:
		switch msg.Screen {
		case ui.ScoreScreenIndex:
			a.focusedScreen = a.scoreScreen
		case ui.EditScreenIndex:
			a.focusedScreen = a.editorScreen
		}
	case ui.SelectCharacterAndSwitchScreenMsg:
		if msg.Err == nil {
			a.character = msg.Character
			a.scoreScreen = ui.NewScoreScreen(a.keymap, a.character)
			cmd = tea.Batch(a.scoreScreen.Init(), ui.SwitchScreenCmd(ui.ScoreScreenIndex))
		}
	case ui.SwitchToEditorMsg:
		a.editorScreen.StartEdit(msg.Originator, msg.Character, msg.Editors)
		cmd = ui.SwitchScreenCmd(ui.EditScreenIndex)
	default:
		_, cmd = a.focusedScreen.Update(msg)
	}

	return a, cmd
}

func (a *DnCApp) View() string {
	pageContent := a.focusedScreen.View()

	pageWidth := a.width - defaultPadding
	pageHeight := a.height - defaultPadding

	topPad := (pageHeight - lipgloss.Height(pageContent)) / 2
	leftPad := (pageWidth - lipgloss.Width(pageContent)) / 2

	s := ui.NoBorderStyle.
		UnsetAlign().
		Width(pageWidth).
		Height(pageHeight).
		PaddingLeft(leftPad).
		PaddingTop(topPad).
		Render(pageContent)

	return s
}
