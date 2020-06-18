package sqlmy

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"

	// TO BE OR NOT TO BE ...
	_ "github.com/go-sql-driver/mysql"
)

var tagName = "ddb"

func Struct2Where(raw interface{}) map[string]interface{} {
	return struct2map(raw, false)
}
func Struct2AssignList(raws ...interface{}) []map[string]interface{} {
	rst := make([]map[string]interface{}, 0, len(raws))
	for _, raw := range raws {
		rst = append(rst, struct2map(raw, true))
	}
	return rst
}

func Struct2Assign(raw interface{}) map[string]interface{} {
	return struct2map(raw, true)
}

func struct2map(raw interface{}, ignoreOpt bool) map[string]interface{} {
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

func tagSplitter(dbTag string) (key, opt string) {
	if dbTag == "" {
		return "", ""
	}
	i := strings.Index(dbTag, ",")
	if i == -1 {
		return dbTag, ""
	}
	return strings.TrimSpace(dbTag[:i]), strings.TrimSpace(dbTag[i+1:])
}

func MockDb() (DB, sqlmock.Sqlmock, error) {
	rawDb, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	return &MyDB{rawDB: rawDb}, mock, err
}

type TxError struct {
	CommitErr error
	RollError error
}

func (txError *TxError) Error() string {
	if txError.RollError != nil {
		return fmt.Sprintf("commit: %v, roll: %v", txError.CommitErr, txError.RollError)
	}

	if txError.CommitErr == nil {
		return "<nil>"
	}

	return txError.RollError.Error()
}

func CommitOrRollback(tx Tx) error {
	commitErr := tx.Commit()
	if commitErr == nil {
		return nil
	}

	var rollBackErr error
	if rollBackErr = tx.Rollback(); rollBackErr != nil {
		if rollBackErr == sql.ErrTxDone {
			rollBackErr = nil
		}
	}

	return &TxError{
		CommitErr: commitErr,
		RollError: rollBackErr,
	}
}
