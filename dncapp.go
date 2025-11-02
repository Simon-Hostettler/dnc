package main

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jmoiron/sqlx"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/db"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/screen"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var defaultPadding = 2

type DnCApp struct {
	config     util.Config
	keymap     util.KeyMap
	width      int
	height     int
	db         *sqlx.DB
	ctx        context.Context
	cancel     context.CancelFunc
	cleanup    func()
	repository repository.CharacterRepository

	selectedTab     *screen.ScreenTab
	isScreenFocused bool
	statTab         *screen.ScreenTab
	spellTab        *screen.ScreenTab
	inventoryTab    *screen.ScreenTab

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

func NewApp(cfg util.Config, cleanup func()) (*DnCApp, error) {
	km := util.DefaultKeyMap()

	handle, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	if err := db.MigrateUp(handle); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	repository := repository.NewDBCharacterRepository(handle)

	app := &DnCApp{
		config:             cfg,
		keymap:             km,
		db:                 handle,
		ctx:                ctx,
		cancel:             cancel,
		cleanup:            cleanup,
		repository:         repository,
		statTab:            screen.NewScreenTab(km, "Stats", command.StatScreenIndex, false),
		spellTab:           screen.NewScreenTab(km, "Spells", command.SpellScreenIndex, false),
		inventoryTab:       screen.NewScreenTab(km, "Inventory", command.InventoryScreenIndex, false),
		titleScreen:        screen.NewTitleScreen(km),
		editorScreen:       screen.NewEditorScreen(km, []editor.ValueEditor{}),
		confirmationScreen: screen.NewConfirmationScreen(km),
		readerScreen:       screen.NewReaderScreen(km),
	}

	return app, nil
}

func (a *DnCApp) Close() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
	if a.cleanup != nil {
		a.cleanup()
	}
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
	case command.LoadSummariesRequestMsg:
		cmd = repository.LoadSummariesCommand(a.repository, a.ctx)
	case repository.LoadSummariesMsg:
		a.titleScreen.SetSummaries(msg.Summaries)
	case command.WriteBackRequestMsg:
		cmd = repository.WriteBackCmd(a.repository, a.ctx, a.character)
	case command.CreateCharacterRequestMsg:
		cmd = repository.CreateCharacterCmd(a.repository, a.ctx, msg.Name)
	case repository.CreateCharacterMsg:
		cmd = repository.LoadSummariesCommand(a.repository, a.ctx)
	case command.DeleteCharacterRequestMsg:
		cmd = repository.DeleteCharacterCmd(a.repository, a.ctx, msg.ID)
	case repository.DeleteCharacterMsg:
		cmd = repository.LoadSummariesCommand(a.repository, a.ctx)
	case repository.LoadCharacterMsg:
		cmds := a.populateCharacterScreens(msg.Agg)
		cmd = tea.Sequence(cmds, command.SwitchScreenCmd(command.StatScreenIndex))
	case command.SelectCharacterMsg:
		cmd = repository.LoadCharacterCmd(a.repository, a.ctx, msg.ID)
	case editor.EditValueMsg:
		cmd = editor.SwitchToEditorCmd(msg.Editors)
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

	s := styles.NoBorderStyle.
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
