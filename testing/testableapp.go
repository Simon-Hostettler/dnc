// Package testing provides testable app factory functions
package testing

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui"
)

// TestableAppConfig holds configuration for creating a testable app
type TestableAppConfig struct {
	CharacterDir string `json:"character_dir"`
}

// TestableScreenTab represents a screen tab in the testable app
type TestableScreenTab struct {
	keymap      ui.KeyMap
	name        string
	screenIndex ui.ScreenIndex
	focus       bool
}

// TestableDnCApp is a testable version of the main DnC application
type TestableDnCApp struct {
	config    TestableAppConfig
	keymap    ui.KeyMap
	width     int
	height    int
	character *models.Character

	selectedTab     *TestableScreenTab
	isScreenFocused bool
	scoreTab        *TestableScreenTab

	screenInView ui.FocusableModel
	titleScreen  *ui.TitleScreen
	scoreScreen  *ui.ScoreScreen
	editorScreen *ui.EditorScreen
}

var (
	defaultPadding = 2
	tabWidth       = 10
	tabHeight      = 3
)

// NewTestableApp creates a new testable app instance
func NewTestableApp(characterDir string) (*TestableDnCApp, error) {
	config := TestableAppConfig{CharacterDir: characterDir}
	km := ui.DefaultKeyMap()
	return &TestableDnCApp{
		config:       config,
		keymap:       km,
		scoreTab:     NewTestableScreenTab(km, "Stats", ui.ScoreScreenIndex, false),
		titleScreen:  ui.NewTitleScreen(characterDir),
		editorScreen: ui.NewEditorScreen(km, []ui.ValueEditor{}),
	}, nil
}

// NewTestableScreenTab creates a new testable screen tab
func NewTestableScreenTab(keymap ui.KeyMap, name string, idx ui.ScreenIndex, focus bool) *TestableScreenTab {
	return &TestableScreenTab{keymap, name, idx, focus}
}

// Implement tea.Model interface

func (a *TestableDnCApp) Init() tea.Cmd {
	cmds := []tea.Cmd{}

	a.selectedTab = a.scoreTab

	if a.titleScreen != nil {
		cmds = append(cmds, a.titleScreen.Init())
		a.switchScreen(ui.TitleScreenIndex)
	}
	if a.scoreScreen != nil {
		cmds = append(cmds, a.scoreScreen.Init())
	}
	if a.editorScreen != nil {
		cmds = append(cmds, a.editorScreen.Init())
	}
	cmds = ui.Filter(cmds, func(c tea.Cmd) bool { return c != nil })
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (a *TestableDnCApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if msg.Err != nil {
			// Handle error
			break
		}
		a.character = msg.Character
		a.scoreScreen = ui.NewScoreScreen(a.keymap, a.character)
		a.switchScreen(ui.ScoreScreenIndex)
	case ui.SwitchToEditorMsg:
		a.editorScreen.StartEdit(msg.Originator, msg.Character, msg.Editors)
		a.switchScreen(ui.EditScreenIndex)
	}
	return a, cmd
}

func (a *TestableDnCApp) View() string {
	var view string
	
	if a.screenInView != nil {
		view = a.screenInView.View()
	}
	
	if a.displayTabs() {
		tabs := a.selectedTab.View()
		view = lipgloss.JoinVertical(lipgloss.Left, tabs, view)
	}
	
	return view
}

// Helper methods

func (a *TestableDnCApp) switchScreen(idx ui.ScreenIndex) {
	a.isScreenFocused = true
	a.selectedTab.Blur()
	switch idx {
	case ui.ScoreScreenIndex:
		a.screenInView = a.scoreScreen
	case ui.EditScreenIndex:
		a.screenInView = a.editorScreen
	case ui.TitleScreenIndex:
		a.screenInView = a.titleScreen
	}
	if a.screenInView != nil {
		a.screenInView.Focus()
	}
}

func (a *TestableDnCApp) displayTabs() bool {
	return a.screenInView != a.editorScreen && a.screenInView != a.titleScreen
}

func (a *TestableDnCApp) moveTab(direction ui.Direction) {
	switch a.selectedTab {
	case a.scoreTab:
		return
	}
}

// Implement TestableApp interface

func (a *TestableDnCApp) GetCurrentScreenType() string {
	switch a.screenInView {
	case a.titleScreen:
		return "title"
	case a.scoreScreen:
		return "score"
	case a.editorScreen:
		return "editor"
	default:
		if a.titleScreen != nil && a.screenInView == nil {
			return "title"
		}
		return "unknown"
	}
}

func (a *TestableDnCApp) GetCharacter() *models.Character {
	return a.character
}

// TestableScreenTab methods

func (s *TestableScreenTab) Init() tea.Cmd {
	return nil
}

func (s *TestableScreenTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, s.keymap.Enter) {
			cmd = ui.SwitchScreenCmd(s.screenIndex)
		}
	}
	return s, cmd
}

func (s *TestableScreenTab) View() string {
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

func (s *TestableScreenTab) Focus() {
	s.focus = true
}

func (s *TestableScreenTab) Blur() {
	s.focus = false
}