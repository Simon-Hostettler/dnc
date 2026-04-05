package list

import (
	"testing"

	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

func TestViewRegression(t *testing.T) {
	km := util.DefaultKeyMap()

	t.Run("LabeledIntRow", func(t *testing.T) {
		val := 42
		r := NewLabeledIntRow(km, "Hit Points", &val, editor.NewIntEditor(km, "Hit Points", &val))
		util.AssertGolden(t, "labeled_int_row", r.View().Content)
	})

	t.Run("LabeledStringRow", func(t *testing.T) {
		val := "Bobby"
		r := NewLabeledStringRow(km, "Name", &val, editor.NewStringEditor(km, "Name", &val))
		util.AssertGolden(t, "labeled_string_row", r.View().Content)
	})

	t.Run("CharacterRow", func(t *testing.T) {
		id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		summary := models.CharacterSummary{ID: id, Name: "Bobby the Wizard"}
		r := NewCharacterRow(km, &summary)
		util.AssertGolden(t, "character_row", r.View().Content)
	})

	t.Run("AppenderRow", func(t *testing.T) {
		r := NewAppenderRow(km, "item")
		util.AssertGolden(t, "appender_row", r.View().Content)
	})

	t.Run("SeparatorRow", func(t *testing.T) {
		r := NewSeparatorRow("-", 20)
		util.AssertGolden(t, "separator_row", r.View().Content)
	})

	t.Run("List_with_rows", func(t *testing.T) {
		val1 := 10
		val2 := 20
		str := "Wizard"
		l := NewListWithDefaults(km).WithTitle("Stats").WithRows([]Row{
			NewLabeledIntRow(km, "Strength", &val1, editor.NewIntEditor(km, "Strength", &val1)),
			NewSeparatorRow("─", 30),
			NewLabeledIntRow(km, "Dexterity", &val2, editor.NewIntEditor(km, "Dexterity", &val2)),
			NewLabeledStringRow(km, "Class", &str, editor.NewStringEditor(km, "Class", &str)),
			NewAppenderRow(km, "stat"),
		})
		l.Init()
		util.AssertGolden(t, "list_with_rows", l.View().Content)
	})
}
