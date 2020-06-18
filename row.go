package sqlmy

import (
	"database/sql"
)

var _ Row = &row{}

type row struct {
	rawRow *sql.Row
	err    error
}

func newRow(raw *sql.Row) *row {
	return &row{rawRow: raw}
}

func (r *row) Raw() *sql.Row {
	return r.rawRow
}

const (
	stdScan    = 0
	structScan = 1
)

// func rowScannerJudge(dest ...interface{}) int {
// 	if len(dest) != 1 {
// 		return stdScan
// 	}
//
// 	v := reflect.ValueOf(dest[0])
// 	for v.Kind() == reflect.Ptr {
// 		v = v.Elem()
// 	}
// 	if v.Kind() == reflect.Struct {
// 		return structScan
// 	}
//
// 	return stdScan
// }

// 对能力进行增强
func (r *row) Scan(dest ...interface{}) error {
	return r.rawRow.Scan(dest...)
}

var _ irows = &rows{}

type rows struct {
	rawRows *sql.Rows
}

func newRows(raw *sql.Rows) *rows {
	return &rows{rawRows: raw}
}

func (rs *rows) Raw() *sql.Rows {
	return rs.rawRows
}

func (rs *rows) Close() error {
	return rs.rawRows.Close()
}

func (rs *rows) ColumnTypes() ([]*sql.ColumnType, error) {
	return rs.rawRows.ColumnTypes()
}

func (rs *rows) Columns() ([]string, error) {
	return rs.rawRows.Columns()
}

func (rs *rows) Err() error {
	return rs.rawRows.Err()
}

func (rs *rows) Next() bool {
	return rs.rawRows.Next()
}

func (rs *rows) NextResultSet() bool {
	return rs.rawRows.NextResultSet()
}

func (rs *rows) Scan(dest ...interface{}) error {
	return rs.rawRows.Scan(dest...)
}

func (rs *rows) ScanStruct(dst interface{}) error {
	return Scan(rs, dst)
}
