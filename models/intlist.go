package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// IntList is a []int that implements sql.Scanner and driver.Valuer for
// DuckDB LIST<INTEGER> columns.
type IntList []int

// Scan implements sql.Scanner. The DuckDB driver delivers LIST<INTEGER>
// values as []any whose elements are int32; (see repository/intlist_scan_db_test.go)
func (il *IntList) Scan(src any) error {
	if src == nil {
		*il = nil
		return nil
	}
	v, ok := src.([]any)
	if !ok {
		return fmt.Errorf("IntList.Scan: unsupported source type %T", src)
	}
	out := make([]int, 0, len(v))
	for _, e := range v {
		n, ok := e.(int32)
		if !ok {
			return fmt.Errorf("IntList.Scan: unsupported element type %T", e)
		}
		out = append(out, int(n))
	}
	*il = out
	return nil
}

// Value implements driver.Valuer. We return a bracketed literal; DuckDB
// implicitly casts the VARCHAR back to LIST<INTEGER> during parameter binding.
func (il IntList) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('[')
	for i, n := range il {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(n))
	}
	b.WriteByte(']')
	return b.String(), nil
}
