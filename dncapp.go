package main

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jmoiron/sqlx"
	"hostettler.dev/dnc/db"
	"hostettler.dev/dnc/repository"
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
	config     Config
	keymap     util.KeyMap
	width      int
	height     int
	db         *sqlx.DB
	ctx        context.Context
	repository repository.CharacterRepository

	selectedTab     *ScreenTab
	isScreenFocused bool
	statTab         *ScreenTab
	spellTab        *ScreenTab
	inventoryTab    *ScreenTab

	character          *repository.CharacterAggregate
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

	handle, err := db.Open(config.DatabasePath)
	if err != nil {
		return nil, err
	}
	if err := db.MigrateUp(handle); err != nil {
		return nil, err
	}
	ctx := context.Background()
	repository := repository.NewDBCharacterRepository(handle)

	app := &DnCApp{
		config:             config,
		keymap:             km,
		db:                 handle,
		ctx:                ctx,
		repository:         repository,
		statTab:            NewScreenTab(km, "Stats", command.StatScreenIndex, false),
		spellTab:           NewScreenTab(km, "Spells", command.SpellScreenIndex, false),
		inventoryTab:       NewScreenTab(km, "Inventory", command.InventoryScreenIndex, false),
		titleScreen:        screen.NewTitleScreen(),
		editorScreen:       screen.NewEditorScreen(km, []editor.ValueEditor{}),
		confirmationScreen: screen.NewConfirmationScreen(km),
		readerScreen:       screen.NewReaderScreen(km),
	}

	return app, nil
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
			// Close DB before quitting
			if a.db != nil {
				_ = a.db.Close()
			}
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
	case command.LoadSummariesRequestMsg:
		cmd = command.LoadSummariesCommand(a.repository, a.ctx)
	case command.LoadSummariesMsg:
		a.titleScreen.SetSummaries(msg.Summaries)
	case command.WriteBackRequestMsg:
		cmd = command.WriteBackCmd(a.repository, a.ctx, a.character)
	case command.CreateCharacterRequestMsg:
		cmd = command.CreateCharacterCmd(a.repository, a.ctx, msg.Name)
	case command.CreateCharacterMsg:
		cmd = command.LoadSummariesCommand(a.repository, a.ctx)
	case command.DeleteCharacterRequestMsg:
		cmd = command.DeleteCharacterCmd(a.repository, a.ctx, msg.ID)
	case command.DeleteCharacterMsg:
		cmd = command.LoadSummariesCommand(a.repository, a.ctx)
	case command.LoadCharacterMsg:
		cmds := a.populateCharacterScreens(msg.Agg)
		cmd = tea.Sequence(cmds, command.SwitchScreenCmd(command.StatScreenIndex))
	case command.SelectCharacterMsg:
		cmd = command.LoadCharacterCmd(a.repository, a.ctx, msg.ID)
	case editor.SwitchToEditorMsg:
		a.editorScreen.StartEdit(msg.Editors)
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

func (a *DnCApp) populateCharacterScreens(agg *repository.CharacterAggregate) tea.Cmd {
	cmds := []tea.Cmd{}
	a.character = agg
	a.statScreen = screen.NewStatScreen(a.keymap, agg)
	cmds = append(cmds, a.statScreen.Init())
	a.spellScreen = screen.NewSpellScreen(a.keymap, agg)
	cmds = append(cmds, a.spellScreen.Init())
	a.inventoryScreen = screen.NewInventoryScreen(a.keymap, agg)
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
