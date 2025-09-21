package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui"
)

var (
	defaultPadding = 2
	tabWidth       = 10
	tabHeight      = 3
)

type DnCApp struct {
	config    Config
	keymap    ui.KeyMap
	width     int
	height    int
	character *models.Character

	selectedTab     *ScreenTab
	isScreenFocused bool
	statTab         *ScreenTab
	spellTab        *ScreenTab

	screenInView ui.FocusableModel
	titleScreen  *ui.TitleScreen
	editorScreen *ui.EditorScreen
	statScreen   *ui.StatScreen
	spellScreen  *ui.SpellScreen
}

type ScreenTab struct {
	keymap      ui.KeyMap
	name        string
	screenIndex ui.ScreenIndex
	focus       bool
}

func NewApp() (*DnCApp, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	km := ui.DefaultKeyMap()
	return &DnCApp{
		config:       config,
		keymap:       km,
		statTab:      NewScreenTab(km, "Stats", ui.StatScreenIndex, false),
		spellTab:     NewScreenTab(km, "Spells", ui.SpellScreenIndex, false),
		titleScreen:  ui.NewTitleScreen(config.CharacterDir),
		editorScreen: ui.NewEditorScreen(km, []ui.ValueEditor{}),
	}, nil
}

func (a *DnCApp) Init() tea.Cmd {
	cmds := []tea.Cmd{}

	a.selectedTab = a.statTab

	if a.titleScreen != nil {
		cmds = append(cmds, a.titleScreen.Init())
		a.switchScreen(ui.TitleScreenIndex)
	}
	if a.statScreen != nil {
		cmds = append(cmds, a.statScreen.Init())
	}
	if a.editorScreen != nil {
		cmds = append(cmds, a.editorScreen.Init())
	}
	cmds = ui.DropNil(cmds)
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (a *DnCApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, a.keymap.ForceQuit) {
			return a, tea.Quit
		}
		if a.isScreenFocused {
			_, cmd = a.screenInView.Update(msg)
		} else {
			switch {
			case key.Matches(msg, a.keymap.Down):
				a.moveTab(ui.DownDirection)
			case key.Matches(msg, a.keymap.Up):
				a.moveTab(ui.UpDirection)
			case key.Matches(msg, a.keymap.Right):
				a.isScreenFocused = true
				a.screenInView.Focus()
				a.selectedTab.Blur()
			default:
				_, cmd = a.selectedTab.Update(msg)
			}
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case ui.ReturnFocusToParentMsg:
		a.isScreenFocused = false
		a.screenInView.Blur()
		a.selectedTab.Focus()
	case ui.SwitchScreenMsg:
		a.switchScreen(msg.Screen)
	case ui.SelectCharacterAndSwitchScreenMsg:
		if msg.Err == nil {
			a.character = msg.Character
			cmds := a.populateCharacterScreens()
			cmd = tea.Batch(cmds, ui.SwitchScreenCmd(ui.StatScreenIndex))
		}
	case ui.SwitchToEditorMsg:
		a.editorScreen.StartEdit(msg.Originator, msg.Character, msg.Editors)
		cmd = ui.SwitchScreenCmd(ui.EditScreenIndex)
	default:
		_, cmd = a.screenInView.Update(msg)
	}

	return a, cmd
}

func (a *DnCApp) View() string {
	screenContent := a.screenInView.View()

	pageContent := screenContent
	if a.displayTabs() {
		tabs := lipgloss.JoinVertical(lipgloss.Center, a.statTab.View(), a.spellTab.View())
		pageContent = lipgloss.JoinHorizontal(lipgloss.Left, tabs, pageContent)
	}

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

func (a *DnCApp) populateCharacterScreens() tea.Cmd {
	cmds := []tea.Cmd{}
	a.statScreen = ui.NewStatScreen(a.keymap, a.character)
	cmds = append(cmds, a.statScreen.Init())
	a.spellScreen = ui.NewSpellScreen(a.keymap, a.character)
	cmds = append(cmds, a.spellScreen.Init())

	cmds = ui.DropNil(cmds)
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (a *DnCApp) switchScreen(idx ui.ScreenIndex) {
	a.isScreenFocused = true
	a.selectedTab.Blur()
	switch idx {
	case ui.StatScreenIndex:
		a.screenInView = a.statScreen
	case ui.EditScreenIndex:
		a.screenInView = a.editorScreen
	case ui.TitleScreenIndex:
		a.screenInView = a.titleScreen
	case ui.SpellScreenIndex:
		a.screenInView = a.spellScreen
	}
	a.screenInView.Focus()
}

func (a *DnCApp) displayTabs() bool {
	return a.screenInView != a.editorScreen && a.screenInView != a.titleScreen
}

func (a *DnCApp) moveTab(d ui.Direction) {
	a.selectedTab.Blur()
	switch a.selectedTab {
	case a.statTab:
		if d == ui.DownDirection {
			a.selectedTab = a.spellTab
		}
	case a.spellTab:
		if d == ui.UpDirection {
			a.selectedTab = a.statTab
		}
	}
	a.selectedTab.Focus()
}

func (a *DnCApp) Blur() {
	a.statTab.Blur()
}

func NewScreenTab(keymap ui.KeyMap, name string, idx ui.ScreenIndex, focus bool) *ScreenTab {
	return &ScreenTab{keymap, name, idx, focus}
}

func (s *ScreenTab) Init() tea.Cmd {
	return nil
}

func (s *ScreenTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, s.keymap.Enter) {
			cmd = ui.SwitchScreenCmd(s.screenIndex)
		}
	}
	return s, cmd
}

func (s *ScreenTab) View() string {
	name := s.name
	if s.focus {
		name = ui.ItemStyleSelected.Render(name)
	} else {
		name = ui.ItemStyleDefault.Render(name)
	}
	return ui.DefaultBorderStyle.UnsetPadding().
		AlignVertical(lipgloss.Center).
		Width(tabWidth).
		Height(tabHeight).
		Render(name)
}

func (s *ScreenTab) Focus() {
	s.focus = true
}

func (s *ScreenTab) Blur() {
	s.focus = false
}
