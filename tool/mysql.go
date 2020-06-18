package dao

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func exit(output io.Writer, val string) int {
	_, _ = fmt.Fprintf(output, val+"\n")
	os.Exit(2)
	return 2
}

type MysqlDao struct {
	Table  string
	DSN    string
	DB     string
	Output io.Writer

	PkgName      string
	WithJsonFlag bool
}

func (md *MysqlDao) Create() int {
	tables, err := md.getColumns()
	if err != nil {
		return exit(md.Output, err.Error())
	}

	if err := md.createDaoFile(); err != nil {
		return exit(md.Output, err.Error())
	}

	for _, table := range tables {
		table.PkgName = md.PkgName
		table.WithJsonFlag = md.WithJsonFlag

		if err := table.Create(); err != nil {
			return exit(md.Output, err.Error())
		}
	}

	return 0
}

func (md *MysqlDao) getColumns() (map[string]*Table, error) {
	ps := []interface{}{md.DB}
	q := `select TABLE_NAME, COLUMN_NAME, DATA_TYPE, COLUMN_TYPE from  information_schema.COLUMNS WHERE TABLE_SCHEMA = ?`
	if md.Table != "" {
		q = fmt.Sprintf("%s AND TABLE_NAME = ?", q)
		ps = append(ps, md.Table)
	}

	db, err := sql.Open("mysql", md.DSN)
	if err != nil {
		return nil, fmt.Errorf("dsn: %s, err: %s", md.DSN, err.Error())
	}
	rows, err := db.Query(q, ps...)
	if err != nil {
		return nil, fmt.Errorf("err: %v, sql: %s, params: [%v]", err, q, ps)
	}
	defer rows.Close()

	tables := map[string]*Table{}
	for rows.Next() {
		column := column{}
		tableName := ""
		if err := rows.Scan(&tableName, &column.name, &column.typ, &column.columnType); err != nil {
			return nil, err
		}
		table := tables[tableName]
		if table == nil {
			table = &Table{
				tableName: tableName,
			}
			tables[tableName] = table
		}

		table.columns = append(table.columns, column)
	}

	return tables, nil
}

func (md *MysqlDao) createDaoFile() error {
	_, midName, longName := underlineConvert(md.DB)
	fileName := strings.ToLower(midName) + ".go"
	if fileExists(fileName) {
		fmt.Printf("file %s existed, skipped\n", fileName)
		return nil
	}

	tt := DaoTemplate{
		PkgName:  md.PkgName,
		MidName:  midName,
		LongName: longName,
	}
	bs, err := tt.Gen()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fileName, bs, 0755); err != nil {
		return err
	}

	if err := exec.Command("gofmt", "-w", fileName).Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("SUCC: db", md.DB, "file: ", fileName)

	return nil
}

type Table struct {
	tableName string
	columns   []column

	PkgName      string
	WithJsonFlag bool
}

type column struct {
	name       string
	typ        string
	columnType string
}

func (table *Table) Create() error {
	fileName := table.tableName + ".go"
	if fileExists(fileName) {
		fmt.Printf("file %s existed, table %s skipped\n", fileName, table.tableName)
		return nil
	}

	tt := NewTableTemplate(table)
	bs, err := tt.Gen()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fileName, bs, 0755); err != nil {
		return err
	}

	if err := exec.Command("gofmt", "-w", fileName).Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("SUCC: table ", table.tableName, "file: ", fileName)

	return nil
}

func NewTableTemplate(table *Table) *TableTemplate {
	tt := &TableTemplate{
		DBTableName: table.tableName,
		Flag:        "`",
		Pkgs:        map[string]Pkg{},

		PkgName:      table.PkgName,
		WithJsonFlag: table.WithJsonFlag,
	}
	tt.FuncParamName, tt.TableNameLower, tt.TableName = underlineConvert(table.tableName)
	for _, col := range table.columns {
		_, _, name := underlineConvert(col.name)
		ft, fpt, pkg := col.Type()
		tt.Columns = append(tt.Columns, ColumnTemplate{
			Name:         name,
			DBName:       col.name,
			Type:         ft,
			ParamType:    fpt,
			Flag:         "`",
			WithJsonFlag: table.WithJsonFlag,
		})
		if pkg != "" {
			tt.Pkgs[pkg] = Pkg{pkg}
		}
	}

	return tt
}

// https://dev.mysql.com/doc/refman/8.0/en/integer-types.html
var typeMapping = map[string]*struct {
	basic   string
	pkg     string
	isArray bool
}{
	// Numeric Data Types
	"TINYINT":   {"int8", "", false},
	"SMALLINT":  {"int16", "", false},
	"MEDIUMINT": {"int32", "", false},
	"INT":       {"int32", "", false},
	"BIGINT":    {"int64", "", false},
	"DECIMAL":   {"float64", "", false},
	"NUMERIC":   {"float64", "", false},
	"FLOAT":     {"float64", "", false},
	"DOUBLE":    {"float64", "", false},
	"BIT":       {"[]byte", "", true}, // TODO

	// Date and Time Data Types
	"DATE":      {"time.Time", "time", false},
	"TIME":      {"time.Time", "time", false},
	"DATETIME":  {"time.Time", "time", false},
	"TIMESTAMP": {"time.Time", "time", false},
	"YEAR":      {"time.Time", "time", false},

	// String Data Type Syntax
	"CHAR":      {"string", "", false},
	"VARCHAR":   {"string", "", false},
	"BINARY":    {"[]byte", "", true},
	"VARBINARY": {"[]byte", "", true},
	"BLOB":      {"[]byte", "", true},
	"TEXT":      {"[]byte", "", true},
	"ENUM":      {"string", "", false},
	"SET":       {"string", "", false}, // TODO
}

func (c column) Type() (ft, fpt, pkg string) {
	typ, ok := typeMapping[strings.ToUpper(c.typ)]
	if !ok {
		return "interface{}", "interface{}", ""
	}

	ft = typ.basic
	if strings.HasSuffix(c.columnType, "unsigned") {
		ft = "u" + ft
	}
	fpt = ft
	if !typ.isArray {
		fpt = "*" + fpt
	}
	pkg = typ.pkg
	return
}

// is_a_good_name => IsAGoodName
func underlineConvert(val string) (short, mid, log string) {
	if val == "" {
		return
	}

	underlineIndex := []int{}

	for i, c := range val {
		if c == '_' {
			underlineIndex = append(underlineIndex, i)
		}
	}

	if len(underlineIndex) == 0 {
		return val[:1], first2Lower(val), first2Upper(val)
	}

	srs, lrs := "", ""
	prev := 0
	for _, i := range underlineIndex {
		if len(val) > i && prev != i {
			span := val[prev:i]
			srs += first2Lower(span[:1])
			lrs += first2Upper(span)
		}
		prev = i + 1
	}
	if prev < len(val) {
		span := val[prev:]
		srs += first2Lower(span[:1])
		lrs += first2Upper(span)
	}
	if srs == "" {
		srs = ""
	}

	return srs, first2Lower(lrs), lrs
}

func first2Lower(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArry := []rune(str)
	if strArry[0] >= 65 && strArry[0] <= 90 {
		strArry[0] += 32
	}
	return string(strArry)
}

func first2Upper(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArry := []rune(str)
	if strArry[0] >= 97 && strArry[0] <= 122 {
		strArry[0] -= 32
	}
	return string(strArry)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
