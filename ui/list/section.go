package list

// Logical group of rows within a List: an optional header
// row, zero or more item rows, and an optional appender.
//
// List renders separators (configured via SectionStyle) between the
// header and the items
type Section struct {
	Header   Row
	Items    []Row
	Appender *AppenderRow
}

// controls how section separators are rendered
type SectionStyle struct {
	HeaderSeparator string // symbol between Header and Items; empty -> none
	SectionGap      string // symbol between consecutive sections; empty -> none
	SeparatorWidth  int    // width of the auto-separators; <=0 -> derive from fixedWidth
}
