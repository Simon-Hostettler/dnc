package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/screen"
	"hostettler.dev/dnc/ui/util"
)

var (
	defaultPadding = 2
	tabWidth       = 11
	tabHeight      = 3
)

type DnCApp struct {
	config    Config
	keymap    util.KeyMap
	width     int
	height    int
	character *models.Character

	selectedTab     *ScreenTab
	isScreenFocused bool
	statTab         *ScreenTab
	spellTab        *ScreenTab
	inventoryTab    *ScreenTab

	curScreenIdx       command.ScreenIndex
	prevScreenIdx      command.ScreenIndex
	screenInView       screen.FocusableModel
	titleScreen        *screen.TitleScreen
	editorScreen       *screen.EditorScreen
	statScreen         *screen.StatScreen
	spellScreen        *screen.SpellScreen
	confirmationScreen *screen.ConfirmationScreen
	inventoryScreen    *screen.InventoryScreen
	readerScreen       *screen.ReaderScreen
}

type ScreenTab struct {
	keymap      util.KeyMap
	name        string
	screenIndex command.ScreenIndex
	focus       bool
}

func NewApp() (*DnCApp, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	km := util.DefaultKeyMap()
	return &DnCApp{
		config:             config,
		keymap:             km,
		statTab:            NewScreenTab(km, "Stats", command.StatScreenIndex, false),
		spellTab:           NewScreenTab(km, "Spells", command.SpellScreenIndex, false),
		inventoryTab:       NewScreenTab(km, "Inventory", command.InventoryScreenIndex, false),
		titleScreen:        screen.NewTitleScreen(config.CharacterDir),
		editorScreen:       screen.NewEditorScreen(km, []editor.ValueEditor{}),
		confirmationScreen: screen.NewConfirmationScreen(km),
		readerScreen:       screen.NewReaderScreen(km),
	}, nil
}

func (a *DnCApp) Init() tea.Cmd {
	cmds := []tea.Cmd{}

	a.selectedTab = a.statTab

	if a.titleScreen != nil {
		cmds = append(cmds, a.titleScreen.Init())
		a.switchScreen(command.TitleScreenIndex)
	}
	if a.statScreen != nil {
		cmds = append(cmds, a.statScreen.Init())
	}
	if a.editorScreen != nil {
		cmds = append(cmds, a.editorScreen.Init())
	}
	if a.confirmationScreen != nil {
		cmds = append(cmds, a.confirmationScreen.Init())
	}
	if a.readerScreen != nil {
		cmds = append(cmds, a.readerScreen.Init())
	}
	cmds = util.DropNil(cmds)
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
				a.moveTab(command.DownDirection)
			case key.Matches(msg, a.keymap.Up):
				a.moveTab(command.UpDirection)
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
	case command.ReturnFocusToParentMsg:
		a.isScreenFocused = false
		a.screenInView.Blur()
		a.selectedTab.Focus()
	case command.SwitchScreenMsg:
		a.switchScreen(msg.Screen)
	case command.SelectCharacterMsg:
		if msg.Err == nil {
			a.character = msg.Character
			cmds := a.populateCharacterScreens()
			cmd = tea.Batch(cmds, command.SwitchScreenCmd(command.StatScreenIndex))
		}
	case editor.SwitchToEditorMsg:
		a.editorScreen.StartEdit(msg.Character, msg.Editors)
		cmd = command.SwitchScreenCmd(command.EditScreenIndex)
	case command.LaunchConfirmationDialogueMsg:
		a.confirmationScreen.LaunchConfirmation(msg.Callback)
		cmd = command.SwitchScreenCmd(command.ConfirmationScreenIndex)
	case command.LaunchReaderScreenMsg:
		a.readerScreen.StartRead(msg.Content)
		cmd = command.SwitchScreenCmd(command.ReaderScreenIndex)
	case command.SwitchToPrevScreenMsg:
		a.switchScreen(a.prevScreenIdx)
	default:
		_, cmd = a.screenInView.Update(msg)
	}

	return a, cmd
}

func (a *DnCApp) View() string {
	screenContent := a.screenInView.View()

	pageContent := screenContent
	if a.displayTabs() {
		tabs := lipgloss.JoinVertical(lipgloss.Center,
			a.statTab.View(),
			a.spellTab.View(),
			a.inventoryTab.View(),
		)
		pageContent = lipgloss.JoinHorizontal(lipgloss.Left, tabs, pageContent)
	}

	pageWidth := a.width - defaultPadding
	pageHeight := a.height - defaultPadding

	topPad := (pageHeight - lipgloss.Height(pageContent)) / 2
	leftPad := (pageWidth - lipgloss.Width(pageContent)) / 2

	s := util.NoBorderStyle.
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
	a.statScreen = screen.NewStatScreen(a.keymap, a.character)
	cmds = append(cmds, a.statScreen.Init())
	a.spellScreen = screen.NewSpellScreen(a.keymap, a.character)
	cmds = append(cmds, a.spellScreen.Init())
	a.inventoryScreen = screen.NewInventoryScreen(a.keymap, a.character)
	cmds = append(cmds, a.inventoryScreen.Init())

	cmds = util.DropNil(cmds)
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (a *DnCApp) switchScreen(idx command.ScreenIndex) {
	a.isScreenFocused = true
	a.selectedTab.Blur()
	if a.screenInView != nil {
		a.screenInView.Blur()
	}
	a.prevScreenIdx = a.curScreenIdx
	switch idx {
	case command.StatScreenIndex:
		a.screenInView = a.statScreen
	case command.EditScreenIndex:
		a.screenInView = a.editorScreen
	case command.TitleScreenIndex:
		a.screenInView = a.titleScreen
	case command.SpellScreenIndex:
		a.screenInView = a.spellScreen
	case command.ConfirmationScreenIndex:
		a.screenInView = a.confirmationScreen
	case command.InventoryScreenIndex:
		a.screenInView = a.inventoryScreen
	case command.ReaderScreenIndex:
		a.screenInView = a.readerScreen
	}
	a.curScreenIdx = idx
	a.screenInView.Focus()
}

func (a *DnCApp) displayTabs() bool {
	return a.screenInView != a.editorScreen &&
		a.screenInView != a.titleScreen &&
		a.screenInView != a.confirmationScreen &&
		a.screenInView != a.readerScreen
}

func (a *DnCApp) moveTab(d command.Direction) {
	a.selectedTab.Blur()
	switch a.selectedTab {
	case a.statTab:
		if d == command.DownDirection {
			a.selectedTab = a.spellTab
		}
	case a.spellTab:
		switch d {
		case command.UpDirection:
			a.selectedTab = a.statTab
		case command.DownDirection:
			a.selectedTab = a.inventoryTab

		}
	case a.inventoryTab:
		if d == command.UpDirection {
			a.selectedTab = a.spellTab
		}

	}
	a.selectedTab.Focus()
}

func (a *DnCApp) Blur() {
	a.statTab.Blur()
	a.spellTab.Blur()
	a.inventoryTab.Blur()
}

func NewScreenTab(keymap util.KeyMap, name string, idx command.ScreenIndex, focus bool) *ScreenTab {
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
			cmd = command.SwitchScreenCmd(s.screenIndex)
		}
	}
	return s, cmd
}

func (s *ScreenTab) View() string {
	name := s.name
	if s.focus {
		name = util.ItemStyleSelected.Render(name)
	} else {
		name = util.ItemStyleDefault.Render(name)
	}
	return util.DefaultBorderStyle.UnsetPadding().
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
