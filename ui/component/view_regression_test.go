package component

import (
	"testing"

	"hostettler.dev/dnc/util"
)

func TestViewRegression(t *testing.T) {
	km := util.DefaultKeyMap()

	t.Run("SimpleStringComponent", func(t *testing.T) {
		val := "Chaotic Evil"
		c := NewSimpleStringComponent(km, "Alignment", &val, true, true)
		util.AssertGolden(t, "simple_string_component", c.View().Content)
	})

	t.Run("SimpleIntComponent", func(t *testing.T) {
		val := 17
		c := NewSimpleIntComponent(km, "Armor Class", &val, true, true)
		util.AssertGolden(t, "simple_int_component", c.View().Content)
	})

	t.Run("SimpleTextComponent", func(t *testing.T) {
		val := "A disheveled gnome wizard with wild eyes and ink-stained fingers."
		c := NewSimpleTextComponent(km, "Backstory", &val, 5, 30)
		util.AssertGolden(t, "simple_text_component", c.View().Content)
	})
}
