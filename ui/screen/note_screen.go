package screen

import (
	"strings"

	"charm.land/bubbles/v2/key"
	ti "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/ui/textinput"
	"hostettler.dev/dnc/util"
)

var (
	noteColHeight = 32
	noteColWidth  = styles.ScreenWidth - 10
)

type NoteScreen struct {
	keymap    util.KeyMap
	character *repository.CharacterAggregate
	FocusManager

	searchField *textinput.TextInput
	noteList    *list.List

	noteRows *CollectionRows[models.NoteTO]
}

func NewNoteScreen(k util.KeyMap, c *repository.CharacterAggregate) *NoteScreen {
	sf := ti.New()
	sf.SetWidth(noteColWidth)
	sf.CharLimit = noteColWidth
	sf.Placeholder = ""
	sf.Prompt = ""

	s := &NoteScreen{
		keymap:      k,
		character:   c,
		searchField: textinput.New(sf),
		noteList: list.NewList(k, list.LeftAlignedListStyle).
			WithFixedWidth(noteColWidth).
			WithViewport(noteColHeight - 2),
	}
	s.noteRows = NewCollectionRows(k, s.noteList, "note",
		func() []*models.NoteTO { return util.Pointers(s.character.Notes) },
		func(n *models.NoteTO) uuid.UUID { return n.ID },
		s.character.AddEmptyNote,
		s.character.DeleteNote,
		func(note *models.NoteTO) *list.StructRow[models.NoteTO] {
			return list.NewStructRow(s.keymap, note, renderNoteInfoRow,
				createNoteEditors(s.keymap, note)).
				WithReader(renderFullNoteInfo)
		},
	)
	return s
}

func (s *NoteScreen) Init() tea.Cmd {
	s.noteRows.Repopulate()

	s.lastFocusedElement = s.searchField
	s.wireFocusGraph()

	return nil
}

func (s *NoteScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "note") {
			cmd = s.noteRows.HandleAppend(msg.Tag)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case tea.KeyPressMsg:
		switch s.focusedElement.(type) {
		case *textinput.TextInput:
			switch {
			case key.Matches(msg, s.keymap.Down):
				cmd = s.moveFocus(command.DownDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
				term := s.searchField.Value()
				s.noteList.Filter(searchFilter(term))
			}
		default:
			cmd = RouteKey(s.focusedElement, msg, s.keymap, s.moveFocus)
		}
	}
	return s, cmd
}

func (s *NoteScreen) View() tea.View {
	topbar := s.RenderNoteScreenTopBar()
	renderedNotes := s.noteList.View().Content

	content := styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		Height(noteColHeight).
		Render(renderedNotes)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, topbar, content))
}

func (s *NoteScreen) wireFocusGraph() {
	s.focusGraph = FocusGraph{
		s.searchField: {
			command.LeftDirection: Emit(command.ReturnFocusToParentCmd),
			command.DownDirection: To(s.noteList),
		},
		s.noteList: {
			command.UpDirection:   To(s.searchField),
			command.LeftDirection: Emit(command.ReturnFocusToParentCmd),
		},
	}
}

func createNoteEditors(k util.KeyMap, note *models.NoteTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(k, "Title", &note.Title),
		editor.NewTextEditor(k, "Note", &note.Note),
	}
}

func (s *NoteScreen) RenderNoteScreenTopBar() string {
	return styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		AlignHorizontal(lipgloss.Left).
		Render(s.searchField.View().Content)
}

func renderNoteInfoRow(n *models.NoteTO) string {
	return n.Title
}

func searchFilter(term string) func(list.Row) bool {
	normalized := strings.ToLower(strings.TrimSpace(term))
	if normalized == "" {
		return func(r list.Row) bool { return true }
	}
	return func(r list.Row) bool {
		if rr, ok := r.(*list.StructRow[models.NoteTO]); ok {
			n := rr.Value()
			title := strings.ToLower(n.Title)
			body := strings.ToLower(n.Note)
			return strings.Contains(title, normalized) || strings.Contains(body, normalized)
		}
		return true
	}
}

func renderFullNoteInfo(n *models.NoteTO) string {
	separator := styles.MakeHorizontalSeparator(styles.SmallScreenWidth-4, 1)
	content := strings.Join(
		[]string{
			n.Title,
			separator,
			n.Note,
		},
		"\n")
	return styles.DefaultTextStyle.
		AlignHorizontal(lipgloss.Left).
		Render(content)
}
