package internal

import "github.com/didi/gendry/scanner"

var TagName = "db"

func SetTagName(name string) {
	TagName = name
	scanner.SetTagName(name)
}

func init() {
	SetTagName(TagName)
}
