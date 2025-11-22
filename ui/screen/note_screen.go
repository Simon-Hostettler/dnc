package screen

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	noteColHeight = 30
	noteColWidth  = styles.ScreenWidth - 8
)

type NoteScreen struct {
	keymap    util.KeyMap
	character *repository.CharacterAggregate

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	searchField *textinput.TextInput
	noteList    *list.List
}

func NewNoteScreen(k util.KeyMap, c *repository.CharacterAggregate) *NoteScreen {
	sf := ti.New()
	sf.Width = noteColWidth
	sf.CharLimit = noteColWidth
	sf.Placeholder = ""
	sf.Prompt = ""

	return &NoteScreen{
		keymap:      k,
		character:   c,
		searchField: textinput.New(sf),
	}
}

func (s *NoteScreen) Init() tea.Cmd {
	s.populateNotes()

	s.focusOn(s.searchField)
	s.lastFocusedElement = s.searchField

	return nil
}

func (s *NoteScreen) populateNotes() {
	if s.noteList == nil {
		s.noteList = list.NewList(s.keymap,
			list.LeftAlignedListStyle).
			WithFixedWidth(noteColWidth).
			WithViewport(noteColHeight - 2)
	}
	s.CreateNoteRows()
}

func (s *NoteScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "note") {
			note_id := s.character.AddEmptyNote()
			s.populateNotes()
			cmd = editor.SwitchToEditorCmd(
				s.getNoteRow(note_id).Editors(),
			)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case tea.KeyMsg:
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
		case *list.List:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		default:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			case key.Matches(msg, s.keymap.Up):
				cmd = s.moveFocus(command.UpDirection)
			case key.Matches(msg, s.keymap.Down):
				cmd = s.moveFocus(command.DownDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		}
	}
	return s, cmd
}

func (s *NoteScreen) View() string {
	topbar := s.RenderNoteScreenTopBar()
	renderedNotes := s.noteList.View()

	content := styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		Height(spellColHeight).
		Render(renderedNotes)

	return lipgloss.JoinVertical(lipgloss.Left, topbar, content)
}

func (s *NoteScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *NoteScreen) moveFocus(d command.Direction) tea.Cmd {
	var cmd tea.Cmd
	s.Blur()

	switch s.lastFocusedElement {
	case s.searchField:
		switch d {
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		case command.DownDirection:
			s.focusOn(s.noteList)
		default:
			s.focusOn(s.searchField)
		}
	case s.noteList:
		switch d {
		case command.UpDirection:
			s.focusOn(s.searchField)
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		default:
			s.focusOn(s.noteList)
		}
	}
	return cmd
}

func (s *NoteScreen) Focus() {
	s.focusOn(s.lastFocusedElement)
}

func (s *NoteScreen) Blur() {
	if s.focusedElement != nil {
		s.focusedElement.Blur()
		s.lastFocusedElement = s.focusedElement
	}

	s.focusedElement = nil
}

func (s *NoteScreen) CreateNoteRows() {
	rows := []list.Row{}
	for i := range s.character.Notes {
		note := &s.character.Notes[i]
		rows = append(rows, list.NewStructRow(s.keymap, note,
			renderNoteInfoRow,
			createNoteEditors(s.keymap, note),
		).WithDestructor(deleteNoteCallback(s, note)).
			WithReader(renderFullNoteInfo))
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, "note"))
	s.noteList.WithRows(rows)
}

func (s *NoteScreen) getNoteRow(id uuid.UUID) list.Row {
	for _, r := range s.noteList.Content() {
		switch r := r.(type) {
		case *list.StructRow[models.NoteTO]:
			if r.Value().ID == id {
				return r
			}
		}
	}
	return nil
}

func deleteNoteCallback(s *NoteScreen, n *models.NoteTO) func() tea.Cmd {
	return func() tea.Cmd {
		s.character.DeleteNote(n.ID)
		s.populateNotes()
		return command.WriteBackRequest
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
		Render(s.searchField.View())
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
