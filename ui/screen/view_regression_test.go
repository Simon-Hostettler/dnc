package screen

import (
	"testing"

	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

var testID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func testAggregate() repository.CharacterAggregate {
	return repository.TestCharacter(testID)
}

func TestViewRegression(t *testing.T) {
	km := util.DefaultKeyMap()
	agg := testAggregate()

	t.Run("StatScreen", func(t *testing.T) {
		s := NewStatScreen(km, &agg)
		s.Init()
		util.AssertGolden(t, "stat_screen", s.View().Content)
	})

	t.Run("ProfileScreen", func(t *testing.T) {
		s := NewProfileScreen(km, &agg)
		s.Init()
		util.AssertGolden(t, "profile_screen", s.View().Content)
	})

	t.Run("SpellScreen", func(t *testing.T) {
		s := NewSpellScreen(km, &agg)
		s.Init()
		util.AssertGolden(t, "spell_screen", s.View().Content)
	})

	t.Run("InventoryScreen", func(t *testing.T) {
		s := NewInventoryScreen(km, &agg)
		s.Init()
		util.AssertGolden(t, "inventory_screen", s.View().Content)
	})

	t.Run("NoteScreen", func(t *testing.T) {
		s := NewNoteScreen(km, &agg)
		s.Init()
		util.AssertGolden(t, "note_screen", s.View().Content)
	})

	t.Run("TitleScreen", func(t *testing.T) {
		s := NewTitleScreen(km)
		s.SetSummaries([]models.CharacterSummary{
			{ID: testID, Name: "Bobby"},
		})
		util.AssertGolden(t, "title_screen", s.View().Content)
	})

	t.Run("ConfirmationScreen", func(t *testing.T) {
		s := NewConfirmationScreen(km)
		s.Init()
		util.AssertGolden(t, "confirmation_screen", s.View().Content)
	})

	t.Run("ReaderScreen", func(t *testing.T) {
		s := NewReaderScreen(km)
		s.Init()
		s.StartRead("Sample content for reader viewport testing.\nLine 2 of the reader.\nLine 3 of the reader.")
		util.AssertGolden(t, "reader_screen", s.View().Content)
	})

	t.Run("EditorScreen", func(t *testing.T) {
		strVal := "Bobby"
		intVal := 42
		editors := []editor.ValueEditor{
			editor.NewStringEditor(km, "Name", &strVal),
			editor.NewIntEditor(km, "Age", &intVal),
		}
		s := NewEditorScreen(km, editors)
		s.StartEdit(editors)
		util.AssertGolden(t, "editor_screen", s.View().Content)
	})

	t.Run("ScreenTab_unfocused", func(t *testing.T) {
		s := NewScreenTab(km, "Stats", command.StatScreenIndex, false)
		util.AssertGolden(t, "screen_tab_unfocused", s.View().Content)
	})

	t.Run("ScreenTab_focused", func(t *testing.T) {
		s := NewScreenTab(km, "Stats", command.StatScreenIndex, true)
		util.AssertGolden(t, "screen_tab_focused", s.View().Content)
	})
}
