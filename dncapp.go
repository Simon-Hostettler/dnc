package main

import (
	"context"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jmoiron/sqlx"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/db"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/quickaction"
	"hostettler.dev/dnc/ui/screen"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var defaultPadding = 2

type DnCApp struct {
	screen.FocusManager

	config     util.Config
	keymap     util.KeyMap
	width      int
	height     int
	db         *sqlx.DB
	ctx        context.Context
	cancel     context.CancelFunc
	cleanup    func()
	repository repository.CharacterRepository

	statTab      *screen.ScreenTab
	profileTab   *screen.ScreenTab
	spellTab     *screen.ScreenTab
	inventoryTab *screen.ScreenTab
	noteTab      *screen.ScreenTab

	character          *repository.CharacterAggregate
	router             *screen.ScreenRouter
	titleScreen        *screen.TitleScreen
	editorScreen       *screen.EditorScreen
	confirmationScreen *screen.ConfirmationScreen
	readerScreen       *screen.ReaderScreen
	palette            *quickaction.Palette
}

func NewApp(cfg util.Config, cleanup func()) (*DnCApp, error) {
	km := cfg.KeyMap

	handle, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	if err := db.MigrateUp(handle); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	repo := repository.NewDBCharacterRepository(handle)

	if cfg.Demo {
		if id, err := repo.CreateEmpty(ctx, "Bobby"); err != nil {
			slog.Warn("demo mode: failed to create test character", "error", err)
		} else {
			agg := repository.TestCharacter(id)
			if err := repo.Update(ctx, &agg); err != nil {
				slog.Warn("demo mode: failed to populate test character", "error", err)
			}
		}
	}

	app := &DnCApp{
		config:             cfg,
		keymap:             km,
		db:                 handle,
		ctx:                ctx,
		cancel:             cancel,
		cleanup:            cleanup,
		repository:         repo,
		statTab:            screen.NewScreenTab(km, "Stats", command.StatScreenIndex, false),
		profileTab:         screen.NewScreenTab(km, "Profile", command.ProfileScreenIndex, false),
		spellTab:           screen.NewScreenTab(km, "Spells", command.SpellScreenIndex, false),
		inventoryTab:       screen.NewScreenTab(km, "Inventory", command.InventoryScreenIndex, false),
		noteTab:            screen.NewScreenTab(km, "Notes", command.NoteScreenIndex, false),
		titleScreen:        screen.NewTitleScreen(km),
		editorScreen:       screen.NewEditorScreen(km, []editor.ValueEditor{}),
		confirmationScreen: screen.NewConfirmationScreen(km),
		readerScreen:       screen.NewReaderScreen(km),
		palette:            quickaction.NewPalette(km, quickaction.NewRegistry()),
		router:             screen.NewScreenRouter(),
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
	a.wireTabFocusGraph()

	cmds := []tea.Cmd{
		a.router.Register(command.TitleScreenIndex, a.titleScreen, false),
		a.router.Register(command.EditScreenIndex, a.editorScreen, true),
		a.router.Register(command.ConfirmationScreenIndex, a.confirmationScreen, true),
		a.router.Register(command.ReaderScreenIndex, a.readerScreen, true),
	}

	a.router.SwitchContent(command.TitleScreenIndex)
	a.router.Focus()
	return tea.Batch(cmds...)
}

func (a *DnCApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, a.keymap.ForceQuit):
			return a, tea.Quit
		case a.palette.Active():
			cmd = a.palette.Update(msg)
		case key.Matches(msg, a.keymap.QuickAction) && a.router.IsCharacterReady():
			a.palette.Open()
		case key.Matches(msg, a.keymap.Screen1):
			cmd = command.SwitchScreenCmd(command.StatScreenIndex)
		case key.Matches(msg, a.keymap.Screen2):
			cmd = command.SwitchScreenCmd(command.ProfileScreenIndex)
		case key.Matches(msg, a.keymap.Screen3):
			cmd = command.SwitchScreenCmd(command.SpellScreenIndex)
		case key.Matches(msg, a.keymap.Screen4):
			cmd = command.SwitchScreenCmd(command.InventoryScreenIndex)
		case key.Matches(msg, a.keymap.ShowKeymap):
			cmd = command.LaunchReaderScreenCmd(a.renderKeymap())
		default:
			if a.router.IsFocused() {
				_, cmd = a.router.Active().Update(msg)
			} else {
				cmd = screen.RouteKey(a.Focused(), msg, a.keymap, a.MoveFocus)
			}
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case command.ReturnFocusToParentMsg:
		a.router.Blur()
		a.Focus()
	case command.FocusActiveScreenMsg:
		a.Blur()
		a.router.Focus()
	case command.SwitchScreenMsg:
		if a.router.IsModal(msg.Screen) {
			a.router.PushModal(msg.Screen)
		} else {
			a.router.SwitchContent(msg.Screen)
		}
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
		a.router.PopModal()
	default:
		_, cmd = a.router.Active().Update(msg)
	}

	return a, cmd
}

func (a *DnCApp) View() tea.View {
	screenContent := a.router.Active().View().Content

	pageContent := screenContent
	if a.displayTabs() {
		tabs := lipgloss.JoinVertical(lipgloss.Center,
			a.statTab.View().Content,
			a.profileTab.View().Content,
			a.spellTab.View().Content,
			a.inventoryTab.View().Content,
			a.noteTab.View().Content,
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

	if a.palette.Active() {
		paletteView := a.palette.View()
		cx := (pageWidth - lipgloss.Width(paletteView)) / 2
		cy := (pageHeight - lipgloss.Height(paletteView)) / 2
		bg := lipgloss.NewLayer(s)
		fg := lipgloss.NewLayer(paletteView).X(cx).Y(cy).Z(1)
		s = lipgloss.NewCompositor(bg, fg).Render()
	}

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (a *DnCApp) populateCharacterScreens(agg *repository.CharacterAggregate) tea.Cmd {
	a.character = agg

	cmds := []tea.Cmd{
		a.router.Register(command.StatScreenIndex, screen.NewStatScreen(a.keymap, agg), false),
		a.router.Register(command.ProfileScreenIndex, screen.NewProfileScreen(a.keymap, agg), false),
		a.router.Register(command.SpellScreenIndex, screen.NewSpellScreen(a.keymap, agg), false),
		a.router.Register(command.InventoryScreenIndex, screen.NewInventoryScreen(a.keymap, agg), false),
		a.router.Register(command.NoteScreenIndex, screen.NewNoteScreen(a.keymap, agg), false),
	}

	a.palette.SetCharacter(agg)
	a.router.MarkCharacterReady()

	return tea.Batch(cmds...)
}

func (a *DnCApp) displayTabs() bool {
	return !a.router.InModal() && a.router.ContentIndex() != command.TitleScreenIndex
}

func (a *DnCApp) wireTabFocusGraph() {
	a.Wire(screen.FocusGraph{
		a.statTab: {
			command.DownDirection:  screen.To(a.profileTab),
			command.RightDirection: screen.Emit(command.FocusActiveScreenCmd),
		},
		a.profileTab: {
			command.UpDirection:    screen.To(a.statTab),
			command.DownDirection:  screen.To(a.spellTab),
			command.RightDirection: screen.Emit(command.FocusActiveScreenCmd),
		},
		a.spellTab: {
			command.UpDirection:    screen.To(a.profileTab),
			command.DownDirection:  screen.To(a.inventoryTab),
			command.RightDirection: screen.Emit(command.FocusActiveScreenCmd),
		},
		a.inventoryTab: {
			command.UpDirection:    screen.To(a.spellTab),
			command.DownDirection:  screen.To(a.noteTab),
			command.RightDirection: screen.Emit(command.FocusActiveScreenCmd),
		},
		a.noteTab: {
			command.UpDirection:    screen.To(a.inventoryTab),
			command.RightDirection: screen.Emit(command.FocusActiveScreenCmd),
		},
	}, a.statTab)
}

func (a *DnCApp) renderKeymap() string {
	return lipgloss.Place(
		styles.SmallScreenWidth,
		screen.ReaderHeight,
		lipgloss.Center,
		lipgloss.Center,
		styles.DefaultTextStyle.
			Render(util.PrettyPrintKeymap(a.keymap)),
	)
}
