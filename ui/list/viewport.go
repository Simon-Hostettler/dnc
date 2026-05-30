package list

// Scrolling window over the list's rows. Works on row indices, not lines rendered.
type viewport struct {
	enabled bool
	height  int
	offset  int // index of the first visible row
}

func (v *viewport) reset() {
	v.offset = 0
}

// [start, end) range of rows to render given total row count and reserved lines
func (v *viewport) window(total, reserved int) (int, int) {
	if !v.enabled {
		return 0, total
	}
	h := max(v.height-reserved, 0)
	return v.offset, min(total, v.offset+h)
}

func (v *viewport) scrollTo(cursor, reserved int) {
	if !v.enabled {
		return
	}
	h := max(v.height-reserved, 1)
	if cursor < v.offset {
		v.offset = cursor
	} else if cursor >= v.offset+h {
		v.offset = cursor - h + 1
	}
	if v.offset < 0 {
		v.offset = 0
	}
}
