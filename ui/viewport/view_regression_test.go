package viewport

import (
	"testing"

	"hostettler.dev/dnc/util"
)

func TestViewRegression(t *testing.T) {
	km := util.DefaultKeyMap()

	t.Run("Viewport", func(t *testing.T) {
		v := NewViewport(km, 5, 30)
		setupViewport(v, "Line 1 of the viewport content\nLine 2 of the viewport content\nLine 3 of the viewport content\nLine 4 of the viewport content\nLine 5 of the viewport content\nLine 6 of the viewport content\nLine 7 of the viewport content")
		util.AssertGolden(t, "viewport", v.View().Content)
	})
}
