package sqlmy

import (
	gscanner "github.com/didi/gendry/scanner"
)

func Scan(rs Rows, target interface{}) error {
	err := gscanner.Scan(rs, target)
	_ = rs.Close()
	return err
}

func SetTagName(name string) {
	gscanner.SetTagName(name)
}
