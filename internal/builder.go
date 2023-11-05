package internal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/didi/gendry/builder"
)

func BuildQuery(table string, fields []string, where any) (sql string, args []any, err error) {
	wheres := struct2Where(TagName, where)
	return builder.BuildSelect(table, wheres, fields)
}

func BuildDelete(table string, where any) (sql string, args []any, err error) {
	wheres := struct2Where(TagName, where)
	return builder.BuildDelete(table, wheres)
}

func BuildUpdate(table string, where, assign any) (sql string, args []any, err error) {
	wheres := struct2Where(TagName, where)
	assigns := struct2Assign(TagName, assign)
	return builder.BuildUpdate(table, wheres, assigns)
}

const (
	CommonInsert  = 0
	IgnoreInsert  = 1
	ReplaceInsert = 2
)

func BuildInsert(table string, typ int, datas ...any) (sql string, args []any, err error) {
	assigns := struct2AssignList(TagName, datas...)
	if typ == CommonInsert {
		return builder.BuildInsert(table, assigns)
	}
	if typ == IgnoreInsert {
		return builder.BuildInsertIgnore(table, assigns)
	}
	if typ == ReplaceInsert {
		return builder.BuildReplaceInsert(table, assigns)
	}

	err = fmt.Errorf("bad type: %d", typ)
	return
}

func tagSplitter(dbTag string) (key, opt string) {
	if dbTag == "" {
		return "", "="
	}
	i := strings.Index(dbTag, ",")
	if i == -1 {
		return dbTag, "="
	}

	opt = strings.TrimSpace(dbTag[i+1:])
	if opt == "" {
		opt = "="
	}
	return strings.TrimSpace(dbTag[:i]), opt
}

func struct2Where(tagName string, raw interface{}) map[string]interface{} {
	return struct2map(tagName, raw, false)
}
func struct2AssignList(tagName string, raws ...interface{}) []map[string]interface{} {
	rst := make([]map[string]interface{}, 0, len(raws))
	for _, raw := range raws {
		rst = append(rst, struct2map(tagName, raw, true))
	}
	return rst
}

func struct2Assign(tagName string, raw interface{}) map[string]interface{} {
	return struct2map(tagName, raw, true)
}

func struct2map(tagName string, raw interface{}, ignoreOpt bool) map[string]interface{} {
	rst := map[string]interface{}{}
	if raw == nil {
		return rst
	}

	structType := reflect.TypeOf(raw)
	if kind := structType.Kind(); kind == reflect.Ptr || kind == reflect.Interface {
		structType = structType.Elem()
	}

	structVal := reflect.ValueOf(raw)
	if structVal.IsZero() {
		return rst
	}
	if structVal.Kind() == reflect.Ptr {
		structVal = structVal.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < structVal.NumField(); i++ {
		valField := structVal.Field(i)

		valFieldKind := valField.Kind()
		if valFieldKind == reflect.Ptr {
			if valField.IsNil() {
				continue
			}

			valField = valField.Elem()
		}

		if valFieldKind == reflect.Slice {
			if valField.IsZero() {
				continue
			}
		}

		typeField := structType.Field(i)
		dbTag := typeField.Tag.Get(tagName)
		if dbTag == "-" {
			continue
		}
		key, opt := tagSplitter(dbTag)
		if key == "" {
			key = typeField.Name
		}
		if ignoreOpt {
			rst[key] = valField.Interface()
		} else {
			if opt == "" || opt == "=" {
				rst[key] = valField.Interface()
			} else {
				rst[key+" "+opt] = valField.Interface()
			}
		}
	}

	return rst
}
