package sqlmy

import (
	"github.com/didi/gendry/builder"
)

var (
	_ SqlBuilder = &SelectBuilder{}
	_ SqlBuilder = &DeleteBuilder{}
	_ SqlBuilder = &UpdateBuilder{}
	_ SqlBuilder = &InsertBuilder{}
)

type SelectBuilder struct {
	Table       string
	Wheres      map[string]interface{}
	SelectField []string
}

func (s *SelectBuilder) From(table string) *SelectBuilder {
	s.Table = table
	return s
}

func (s *SelectBuilder) Where(wheres map[string]interface{}) *SelectBuilder {
	s.Wheres = wheres
	return s
}

func (s *SelectBuilder) Select(selectFields []string) *SelectBuilder {
	s.SelectField = selectFields
	return s
}

func (s *SelectBuilder) Compile() (sql string, args []interface{}, err error) {
	return builder.BuildSelect(s.Table, s.Wheres, s.SelectField)
}

func NewSelectBuilder(table string, wheres map[string]interface{}, selectField []string) *SelectBuilder {
	return &SelectBuilder{
		Table:       table,
		Wheres:      wheres,
		SelectField: selectField,
	}
}

type DeleteBuilder struct {
	Table  string
	Wheres map[string]interface{}
}

func (d *DeleteBuilder) Compile() (sql string, args []interface{}, err error) {
	return builder.BuildDelete(d.Table, d.Wheres)
}

func (d *DeleteBuilder) From(table string) *DeleteBuilder {
	d.Table = table
	return d
}

func (d *DeleteBuilder) Where(wheres map[string]interface{}) *DeleteBuilder {
	d.Wheres = wheres
	return d
}

func NewDeleteBuilder(table string, wheres map[string]interface{}) *DeleteBuilder {
	return &DeleteBuilder{
		Table:  table,
		Wheres: wheres,
	}
}

type UpdateBuilder struct {
	Table   string
	Wheres  map[string]interface{}
	Assigns map[string]interface{}
}

func (u *UpdateBuilder) Compile() (sql string, args []interface{}, err error) {
	return builder.BuildUpdate(u.Table, u.Wheres, u.Assigns)
}

func (u *UpdateBuilder) From(table string) *UpdateBuilder {
	u.Table = table
	return u
}

func (u *UpdateBuilder) Where(wheres map[string]interface{}) *UpdateBuilder {
	u.Wheres = wheres
	return u
}

func (u *UpdateBuilder) Assign(assign map[string]interface{}) *UpdateBuilder {
	u.Assigns = assign
	return u
}

func NewUpdateBuilder(table string, wheres map[string]interface{}, assigns map[string]interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		Table:   table,
		Wheres:  wheres,
		Assigns: assigns,
	}
}

type InsertBuilder struct {
	Table   string
	Assigns []map[string]interface{}

	typ int
}

func (i *InsertBuilder) Values(assigns []map[string]interface{}) *InsertBuilder {
	i.Assigns = assigns
	return i
}

func (i *InsertBuilder) InsertInto(table string) *InsertBuilder {
	i.Table = table
	i.typ = CommonInsert
	return i
}

func (i *InsertBuilder) InsertIgnoreInto(table string) *InsertBuilder {
	i.Table = table
	i.typ = IgnoreInsert
	return i
}

func (i *InsertBuilder) ReplaceInto(table string) *InsertBuilder {
	i.Table = table
	i.typ = ReplaceInsert
	return i
}

const (
	CommonInsert  = 0
	IgnoreInsert  = 1
	ReplaceInsert = 2
)

func (i *InsertBuilder) Compile() (sql string, args []interface{}, err error) {
	if i.typ == CommonInsert {
		return builder.BuildInsert(i.Table, i.Assigns)
	}
	if i.typ == IgnoreInsert {
		return builder.BuildInsertIgnore(i.Table, i.Assigns)
	}
	if i.typ == ReplaceInsert {
		return builder.BuildReplaceInsert(i.Table, i.Assigns)
	}

	return builder.BuildInsert(i.Table, i.Assigns)
}

func NewInsertBuilder(table string, assigns []map[string]interface{}) *InsertBuilder {
	return &InsertBuilder{
		Table:   table,
		Assigns: assigns,
		typ:     CommonInsert,
	}
}
func NewIgnoreInsertBuilder(table string, assigns []map[string]interface{}) *InsertBuilder {
	return &InsertBuilder{
		Table:   table,
		Assigns: assigns,
		typ:     IgnoreInsert,
	}
}
func NewReplaceInsertBuilder(table string, assigns []map[string]interface{}) *InsertBuilder {
	return &InsertBuilder{
		Table:   table,
		Assigns: assigns,
		typ:     ReplaceInsert,
	}
}
