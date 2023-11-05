package internal

import (
	"database/sql"

	"github.com/didi/gendry/scanner"
)

func Scan(rs *sql.Rows, target interface{}) error {
	err := scanner.Scan(rs, target)
	_ = rs.Close()
	return err
}
