package dao

import (
	"bytes"
	"html/template"
)

type DaoTemplate struct {
	PkgName  string
	MidName  string
	LongName string
}

func (dt *DaoTemplate) Gen() ([]byte, error) {
	var buf bytes.Buffer
	err := dbTemplate.Execute(&buf, dt)
	return buf.Bytes(), err
}

var dbTemplate = template.Must(template.New("").Parse(dbTemplateContent))

var dbTemplateContent = `// AUTO GEN BY  dao, Modify it as you want
package {{.PkgName}}

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/liuximu/sqlmy"
)

var {{.MidName}}Singleton sqlmy.DB

func Set{{.LongName}}({{.MidName}} sqlmy.DB) {
	{{.MidName}}Singleton = {{.MidName}}
}

func {{.LongName}}() sqlmy.DB {
	return {{.MidName}}Singleton
}

func Mock{{.LongName}}() sqlmock.Sqlmock {
	db, mock, err := sqlmyMockDb()
	if err != nil {
		panic(err)
	}

	{{.MidName}}Singleton = db
	return mock
}

func New{{.LongName}}Tx(ctx context.Context) (sqlmy.Tx, error) {
	return {{.LongName}}().BeginTx(ctx, nil)
}`
