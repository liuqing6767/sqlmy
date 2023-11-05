package sqlmy

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
)

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

	OrderBy *string `db:"_orderby"`
	Limit   []uint  `db:"_limit"`
}

func ExampleQuery() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectQuery(`SELECT \* FROM students`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "status"}).
				AddRow(1, "N1", 2).
				AddRow(2, "N2", 1),
		)

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	student, err := StudentCURD.Query(ctx, nil)
	fmt.Println(err)
	if student != nil {
		fmt.Println(student.ID, student.Name, student.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("there were unfulfilled expectations: %s\n", err)
	}

	// output: <nil>
	// 1 N1 2
}

func ExampleQuery1() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectQuery(`SELECT \* FROM students WHERE \(id=\?\)`).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "status"}).AddRow(1, "N1", 2),
		)

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	student, err := StudentCURD.Query(ctx, &StudentParam{
		ID: P(int64(1)),
	})
	fmt.Println(err)
	if student != nil {
		fmt.Println(student.ID, student.Name, student.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("there were unfulfilled expectations: %s\n", err)
	}

	// output: <nil>
	// 1 N1 2
}

func ExampleQuery2() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectQuery(`SELECT name FROM students WHERE \(id=\?\)`).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("N1"),
		)

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	student, err := StudentCURD.Query(ctx, &StudentParam{
		ID: P(int64(1)),
	}, WithSelectFileds("name"))
	fmt.Println(err)
	if student != nil {
		fmt.Println(student.ID, student.Name, student.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("there were unfulfilled expectations: %s\n", err)
	}

	// output: <nil>
	// 0 N1 0
}

func ExampleQueryList() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectQuery(`SELECT \* FROM students WHERE \(id=\?\)`).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "status"}).
				AddRow(1, "N1", 2).
				AddRow(2, "N2", 3),
		)

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	students, err := StudentCURD.QueryList(ctx, &StudentParam{
		ID: P(int64(1)),
	})
	fmt.Println(err)
	for _, student := range students {
		fmt.Println(student.ID, student.Name, student.Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("there were unfulfilled expectations: %s\n", err)
	}

	// output: <nil>
	// 1 N1 2
	// 2 N2 3
}

func ExampleInsert() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectExec(`INSERT INTO students \(id,name\) VALUES \(\?,\?\)`).
		WithArgs(1, "n1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	id, err := StudentCURD.Insert(ctx, &StudentParam{
		ID:   P(int64(1)),
		Name: P("n1"),
	})

	fmt.Println(id)
	fmt.Println(err)
	// output: 1
	// <nil>
}

func ExampleInsertIgnore() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectExec(`INSERT IGNORE INTO students \(id,name\) VALUES \(\?,\?\)`).
		WithArgs(1, "n1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	id, err := StudentCURD.Insert(ctx, &StudentParam{
		ID:   P(int64(1)),
		Name: P("n1"),
	}, WithInsertType(InsertTypeIgnoreInsert))

	fmt.Println(id)
	fmt.Println(err)
	// output: 1
	// <nil>
}

func ExampleInsertList() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectExec(`INSERT INTO students \(id,name\) VALUES \(\?,\?\),\(\?,\?\),\(\?,\?\)`).
		WithArgs(1, "n1", 2, "n2", 3, "n3").
		WillReturnResult(sqlmock.NewResult(3, 1))

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	id, err := StudentCURD.InsertList(ctx, []*StudentParam{
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
	})

	fmt.Println(id)
	fmt.Println(err)
	// output: 3
	// <nil>
}

func ExampleInsertList1() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectExec(`INSERT INTO students \(id,name\) VALUES \(\?,\?\),\(\?,\?\)`).
		WithArgs(1, "n1", 2, "n2").
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.
		ExpectExec(`INSERT INTO students \(id,name\) VALUES \(\?,\?\)`).
		WithArgs(3, "n3").
		WillReturnResult(sqlmock.NewResult(3, 1))

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	id, err := StudentCURD.InsertList(ctx, []*StudentParam{
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
	}, WithInsertBatchSize(2))

	fmt.Println(id)
	fmt.Println(err)
	// output: 3
	// <nil>
}

func ExampleDelete() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectExec(`DELETE FROM students WHERE \(id=\?\)`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	affectedRows, err := StudentCURD.Delete(ctx, &StudentParam{
		ID: P(int64(1)),
	})

	fmt.Println(affectedRows)
	fmt.Println(err)
	// output: 1
	// <nil>
}

func ExampleUpdate() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.
		ExpectExec(`UPDATE students SET name=\? WHERE \(id=\?\)`).
		WithArgs("n1", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	affectedRows, err := StudentCURD.Update(ctx, &StudentParam{
		ID: P(int64(1)),
	}, &StudentParam{
		Name: P("n1"),
	})

	fmt.Println(affectedRows)
	fmt.Println(err)
	// output: 1
	// <nil>
}

func ExampleQueryEmpty() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.ExpectQuery(`SELECT \* FROM students WHERE \(id=\?\)`).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	student, err := StudentCURD.Query(ctx, &StudentParam{
		ID: P(int64(1)),
	})
	fmt.Println(err)
	if student != nil {
		fmt.Println(student.ID, student.Name, student.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("there were unfulfilled expectations: %s\n", err)
	}

	// output: <nil>
}

func ExampleQueryBad() {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	mock.ExpectQuery(`SELECT \* FROM students WHERE \(id=\?\)`).
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	ctx, err := WithConn(context.Background(), func() (conn Conn, err error) { return db, nil })
	if err != nil {
		panic(err)
	}

	student, err := StudentCURD.Query(ctx, &StudentParam{
		ID: P(int64(1)),
	})
	fmt.Println(err)
	if student != nil {
		fmt.Println(student.ID, student.Name, student.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("there were unfulfilled expectations: %s\n", err)
	}

	// output: sql: connection is already closed
}
