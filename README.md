`sqlmy` is a tiny experience of my bussiness develop where need access relation database.

# yet another style
normally, `dao` layer use as access database, but there are so many styles. here offer yet another style(maybe easiest):
- define one struct name A similar as table name for one table as carry of query's result
- define one struct name as AParam as carry of where and insert data
- define one CURD sigleton for accessing database

for example, we have one student table:
```
CREATE TABLE students (
    id bigint(64),
    name varchar(64),
    status tinyint(2)
)
```

we can define two structs:
```
var StudentCURD = NewCURD[Student, StudentParam]("students")

type Student struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Status int    `db:"status"`
}

type StudentParam struct {
	ID     *int64  `db:"id"`
	IDs    []int64 `db:"id,in"`
	Name   *string `db:"name"`
	Status *int    `db:"status"`
	StatusRange []int    `db:"status,in"`

	OrderBy *string `db:"_orderby"`
	Limit   []uint  `db:"_limit"`
}
```

then define one CURD sigleton:
```
var StudentCURD = NewCURD[Student, StudentParam]("students")
```

we can do most normal things:
```
// QueryOne
// SELECT * FROM students
student, err := StudentCURD.Query(ctx, nil)
// SELECT name FROM students WHERE id = 1
student, err := StudentCURD.Query(ctx,  &StudentParam{ 
        ID: P(int64(1)),
    }, 
    WithSelectFileds("name"),
)

// Query List
// SELECT * FROM students WHERE status IN(1, 2) limit 0, 1
students, err := StudentCURD.Query(ctx, &StudentParam{ 
        StatusRange: []int{1,2}, 
        Limit: []unit{0, 2}, 
    }, 
    WithSelectFileds("name"),
)

// Update
// UPDATE students SET name=`n1` WHERE id=1
affectedRows, err := StudentCURD.Update(ctx, &StudentParam{
		ID: P(int64(1)),
	}, &StudentParam{
		Name: P("n1"),
	},
)

// Delete
// DELETE FROM students WHERE id=1
affectedRows, err := StudentCURD.Delete(ctx, &StudentParam{
		ID: P(int64(1)),
	},
)

// Insert
// INSERT INTO students (id,name) VALUES (1, "n1")
lastInsertedID, err := StudentCURD.Insert(ctx, &StudentParam{
		ID:   P(int64(1)),
		Name: P("n1"),
	},
)
// INSERT IGNORE INTO students (id,name) VALUES (1, "n1")
lastInsertedID, err := StudentCURD.Insert(ctx, &StudentParam{
		ID:   P(int64(1)),
		Name: P("n1"),
	},
    WithInsertType(InsertTypeIgnoreInsert),
)

// Insert List
// INSERT INTO students (id,name) VALUES (1, "n1"), (2, "n2"), (3, "n3")
lastInsertedID, err := StudentCURD.InsertList(ctx, []*StudentParam{
		{
			ID:   P(int64(1)),
			Name: P("n1"),
		},
		{
			ID:   P(int64(2)),
			Name: P("n2"),
		},
		{
			ID:   P(int64(3)),
			Name: P("n3"),
		},
	},
)
// INSERT INTO students (id,name) VALUES (1, "n1"), (2, "n2")
// INSERT INTO students (id,name) VALUES (3, "n3")
lastInsertedID, err := StudentCURD.InsertList(ctx, []*StudentParam{
		{
			ID:   P(int64(1)),
			Name: P("n1"),
		},
		{
			ID:   P(int64(2)),
			Name: P("n2"),
		},
		{
			ID:   P(int64(3)),
			Name: P("n3"),
		},
	},
    WithInsertBatchSize(2),
)
```

# API
- Context with Conn
	- WithConn(ctx context.Context, connFactory func() (conn Conn, err error)) (context.Context, error) 
	- GetExecutor(ctx context.Context) Executor 
	- QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) 
	- ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) 
- CURD Function
	- NewCURD[Data, Param any](tableName string) *CURD[Data, Param]
	- Query(ctx context.Context, param *Param, opts ...curdOpt) (*Data, error)
	- QueryList(ctx context.Context, param *Param, opts ...curdOpt) ([]*Data, error)
	- Insert(ctx context.Context, data *Param, opts ...curdOpt) (lastInsertedID int64, err error)
	- InsertList(ctx context.Context, datas []*Param, opts ...curdOpt) (lastInsertedID int64, err error)
	- Update(ctx context.Context, where *Param, assign *Param, opts ...curdOpt) (affectedRows int64, err error)
	- Delete(ctx context.Context, where *Param, opts ...curdOpt) (affectedRows int64, err error)
	- WithQueryBuilder(builder func(table string, fields []string, where any) (sql string, args []any, err error)) curdOpt
- CURD option
	- WithSelectFileds(fields ...string) curdOpt
	- WithUpdateBuilder(builder func(table string, where, assign any) (sql string, args []any, err error)) curdOpt
	- WithInsertBuilder(builder func(table string, typ InsertType, datas ...any) (sql string, args []any, err error)) curdOpt
	- WithInsertType(typ InsertType) curdOpt
	- WithInsertBatchSize(batchSize int) curdOpt
	- WithDeleteBuilder(builder func(table string, where any) (sql string, args []any, err error)) curdOpt
- Context with Log
	- WithLogID(ctx context.Context, logID string) context.Context 
	- GetLogID(ctx context.Context) string 
- Log
	- SetLogger(log Logger)
	- DumbLogger
	- StdLogger
- Others
	- P[V any](v V) *V