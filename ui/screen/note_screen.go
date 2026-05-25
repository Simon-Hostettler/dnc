package screen

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var (
	noteColHeight = 37
	noteColWidth  = styles.ScreenWidth - 10
)

type NoteScreen struct {
	keymap    util.KeyMap
	character *repository.CharacterAggregate
	FocusManager

	noteList *list.List

	noteRows *Collection[models.NoteTO]
}

func NewNoteScreen(k util.KeyMap, c *repository.CharacterAggregate) *NoteScreen {
	s := &NoteScreen{
		keymap:    k,
		character: c,
		noteList: list.NewList(k, list.LeftAlignedListStyle).
			WithFixedWidth(noteColWidth).
			WithViewport(noteColHeight - 2).
			WithSearch(),
	}
	s.noteRows = NewCollection(k, s.noteList,
		func() []*models.NoteTO { return util.Pointers(s.character.Notes) },
		func(n *models.NoteTO) uuid.UUID { return n.ID },
		s.character.AddEmptyNote,
		s.character.DeleteNote,
		func(note *models.NoteTO) *list.StructRow[models.NoteTO] {
			return list.NewStructRow(s.keymap, note, renderNoteInfoRow,
				createNoteEditors(s.keymap, note)).
				WithReader(renderFullNoteInfo).
				WithSearchText(noteSearchText)
		},
	)
	return s
}

func (s *NoteScreen) Init() tea.Cmd {
	s.noteRows.Repopulate()

	s.wireFocusGraph()

	return nil
}

func (s *NoteScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.FocusNextElementMsg:
		s.MoveFocus(msg.Direction)
	case tea.KeyPressMsg:
		cmd = RouteKey(s.focusedElement, msg, s.keymap, s.MoveFocus)
	}
	return s, cmd
}

func (s *NoteScreen) View() tea.View {
	content := styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		Height(noteColHeight).
		Render(s.noteList.View().Content)

	return tea.NewView(content)
}

func (s *NoteScreen) wireFocusGraph() {
	s.Wire(FocusGraph{
		s.noteList: {
			command.LeftDirection: Emit(command.ReturnFocusToParentCmd),
		},
	}, s.noteList)
}

func createNoteEditors(k util.KeyMap, note *models.NoteTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(k, "Title", &note.Title),
		editor.NewTextEditor(k, "Note", &note.Note),
	}
}

func renderNoteInfoRow(n *models.NoteTO) string {
	return n.Title
}

func noteSearchText(n *models.NoteTO) string {
	return n.Title + " " + n.Note
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
