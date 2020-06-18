package dao

import (
	"bytes"
	"html/template"
	"strings"
)

type TableTemplate struct {
	TableName      string
	TableNameLower string
	DBTableName    string
	FuncParamName  string

	Flag string

	Pkgs    map[string]Pkg
	Columns []ColumnTemplate

	PkgName      string
	WithJsonFlag bool
}

type Pkg struct {
	Name string
}

type ColumnTemplate struct {
	Name      string
	DBName    string
	Type      string
	ParamType string

	Flag         string // dirty
	WithJsonFlag bool
}

func (tt *TableTemplate) Gen() ([]byte, error) {
	// 去掉复数
	tt.TableName = strings.TrimRight(tt.TableName, "s")
	tt.TableNameLower = strings.TrimRight(tt.TableNameLower, "s") + "Table"

	var buf bytes.Buffer
	err := daoTemplate.Execute(&buf, tt)
	return buf.Bytes(), err
}

var daoTemplate = template.Must(template.New("").Parse(templa))

var templa = `// AUTO GEN BY dao, Modify it as you want
package {{.PkgName}}

import (
	"context"
	"database/sql"
{{range .Pkgs}}    "{{ .Name }}" {{ end }}
)

import (
	"github.com/liuximu/sqlmy"
)

const {{.TableNameLower}} = "{{.DBTableName}}"

type {{.TableName}} struct {

{{range $column := .Columns}} 
	{{$column.Name}} {{$column.Type}}    {{.Flag}}ddb:"{{$column.DBName}}"{{if .WithJsonFlag }} json:"{{$column.DBName}}"{{end}}{{.Flag}} {{end}}
}

type {{.TableName}}Param struct {
{{range $column := .Columns}} 
	{{$column.Name}} {{$column.ParamType}}    {{.Flag}}ddb:"{{$column.DBName}}"{{if .WithJsonFlag}} json:"{{$column.DBName}}"{{end}}{{.Flag}} {{end}}
}

type {{.TableName}}List []*{{.TableName}}
type {{.TableName}}ParamList []*{{.TableName}}Param

func ({{.FuncParamName}} *{{.TableName}}) Query(ctx context.Context, db sqlmy.QueryExecer, where *{{.TableName}}Param, columns []string) error {
	return db.QueryRowContextWithBuilderStruct(ctx, sqlmy.NewSelectBuilder({{.TableNameLower}}, sqlmy.Struct2Where(where), columns), {{.FuncParamName}})
}

func ({{.FuncParamName}}p *{{.TableName}}Param) Create(ctx context.Context, db sqlmy.QueryExecer) (sql.Result, error) {
	return db.ExecContextWithBuilder(ctx, sqlmy.NewInsertBuilder({{.TableNameLower}}, sqlmy.Struct2AssignList({{.FuncParamName}}p)))
}

func ({{.FuncParamName}}p *{{.TableName}}Param) Update(ctx context.Context, db sqlmy.QueryExecer, where *{{.TableName}}Param) (sql.Result, error) {
	return db.ExecContextWithBuilder(ctx, sqlmy.NewUpdateBuilder({{.TableNameLower}}, sqlmy.Struct2Where(where), sqlmy.Struct2Assign({{.FuncParamName}}p)))
}

func ({{.FuncParamName}}p *{{.TableName}}Param) Delete(ctx context.Context, db sqlmy.QueryExecer) (sql.Result, error) {
	return db.ExecContextWithBuilder(ctx, sqlmy.NewDeleteBuilder({{.TableNameLower}}, sqlmy.Struct2Where({{.FuncParamName}}p)))
}

func ({{.FuncParamName}}l *{{.TableName}}List) Query(ctx context.Context, db sqlmy.QueryExecer, where *{{.TableName}}Param, columns []string) error {
	return db.QueryContextWithBuilderStruct(ctx, sqlmy.NewSelectBuilder({{.TableNameLower}}, sqlmy.Struct2Where(where), columns), {{.FuncParamName}}l)
}

func ({{.FuncParamName}}pl {{.TableName}}ParamList) Create(ctx context.Context, db sqlmy.QueryExecer) (sql.Result, error) {
	_{{.FuncParamName}}pl := make([]interface{}, len({{.FuncParamName}}pl))
	for i, one := range {{.FuncParamName}}pl {
		_{{.FuncParamName}}pl[i] = one
	}
	return db.ExecContextWithBuilder(ctx, sqlmy.NewInsertBuilder({{.TableNameLower}}, sqlmy.Struct2AssignList(_{{.FuncParamName}}pl...)))
}`
