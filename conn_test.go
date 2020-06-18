package sqlmy

import (
	"context"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type UserForTest struct {
	ID   int64  `ddb:"id"`
	Name string `ddb:"name"`
}

var (
	selectBuilder    = NewSelectBuilder("user", map[string]interface{}{"id": 1}, nil)
	selectBuilderBad = NewSelectBuilder("", map[string]interface{}{"id": 1}, nil)
	insertBuilder    = NewInsertBuilder("user", []map[string]interface{}{
		{"id": 1, "name": "Liu Qing"},
	})
	insertBuilderBad = NewInsertBuilder("", []map[string]interface{}{
		{"id": 1, "name": "Liu Qing"},
	})
)
var (
	selectSql = "SELECT (.+) FROM user WHERE .id=.."
	insertSql = "INSERT INTO user .id,name. VALUES ..,.."
)
var User1 = UserForTest{
	ID:   1,
	Name: "Liu Qing",
}

func TestConn_QueryRowContextWithBuilderStruct(t *testing.T) {
	tests := []struct {
		name       string
		doMock     func(mock sqlmock.Sqlmock)
		sqlBuilder SqlBuilder
		wantErr    bool
		wantData   UserForTest
	}{
		{
			name: "case1",
			doMock: func(mock sqlmock.Sqlmock) {
				columns := []string{"id", "name"}
				rows := sqlmock.NewRows(columns).AddRow(1, "Liu Qing")
				mock.ExpectQuery(selectSql).WillReturnRows(rows)
			},
			sqlBuilder: selectBuilder,
			wantErr:    false,
			wantData:   User1,
		},
		{
			name: "case2",
			doMock: func(mock sqlmock.Sqlmock) {
				columns := []string{"id", "name"}
				rows := sqlmock.NewRows(columns).AddRow(1, "Liu Qing")
				mock.ExpectQuery(selectSql).WillReturnRows(rows)
			},
			sqlBuilder: selectBuilderBad,
			wantErr:    true,
		},
	}

	for _, cas := range tests {
		t.Run(cas.name, func(t *testing.T) {
			db, mock, err := MockDb()
			if err != nil {
				t.Fatal(err)
			}

			conn, err := db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			cas.doMock(mock)
			u := &UserForTest{}
			err = conn.QueryRowContextWithBuilderStruct(context.Background(), cas.sqlBuilder, u)
			if cas.wantErr {
				if err == nil {
					t.Errorf("want err, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if *u != cas.wantData {
				t.Errorf("want %v, got %v", cas.wantData, *u)
			}
		})
	}
}

func TestConn_QueryContextWithBuilderStruct(t *testing.T) {
	tests := []struct {
		name       string
		doMock     func(mock sqlmock.Sqlmock)
		sqlBuilder SqlBuilder
		wantErr    bool
		wantData   []UserForTest
	}{
		{
			name: "case1",
			doMock: func(mock sqlmock.Sqlmock) {
				columns := []string{"id", "name"}
				rows := sqlmock.NewRows(columns).AddRow(1, "Liu Qing").AddRow(2, "Liu Qing2")
				mock.ExpectQuery(selectSql).WillReturnRows(rows)
			},
			sqlBuilder: selectBuilder,
			wantErr:    false,
			wantData:   []UserForTest{User1, UserForTest{2, "Liu Qing2"}},
		},
		{
			name: "case2",
			doMock: func(mock sqlmock.Sqlmock) {
				columns := []string{"id", "name"}
				rows := sqlmock.NewRows(columns).AddRow(1, "Liu Qing").AddRow(2, "Liu Qing2")
				mock.ExpectQuery(selectSql).WillReturnRows(rows)
			},
			sqlBuilder: selectBuilderBad,
			wantErr:    true,
		},
	}

	for _, cas := range tests {
		t.Run(cas.name, func(t *testing.T) {
			db, mock, err := MockDb()
			if err != nil {
				t.Fatal(err)
			}

			conn, err := db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			cas.doMock(mock)
			us := []UserForTest{}
			err = conn.QueryContextWithBuilderStruct(context.Background(), cas.sqlBuilder, &us)
			if cas.wantErr {
				if err == nil {
					t.Errorf("want err, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(us, cas.wantData) {
				t.Errorf("want %v, got %v", cas.wantData, us)
			}
		})
	}
}

func TestConn_ExecContextWithBuilderStruct(t *testing.T) {
	tests := []struct {
		name           string
		doMock         func(mock sqlmock.Sqlmock)
		sqlBuilder     SqlBuilder
		wantErr        bool
		wantAffectRows int64
	}{
		{
			name: "case1",
			doMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(insertSql).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			sqlBuilder:     insertBuilder,
			wantErr:        false,
			wantAffectRows: 1,
		},
		{
			name:       "case2",
			doMock:     func(mock sqlmock.Sqlmock) {},
			sqlBuilder: insertBuilderBad,
			wantErr:    true,
		},
	}

	for _, cas := range tests {
		t.Run(cas.name, func(t *testing.T) {
			db, mock, err := MockDb()
			if err != nil {
				t.Fatal(err)
			}

			conn, err := db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			cas.doMock(mock)
			rs, err := conn.ExecContextWithBuilder(context.Background(), cas.sqlBuilder)
			if cas.wantErr {
				if err == nil {
					t.Errorf("want err, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			got, err := rs.RowsAffected()
			if err != nil {
				t.Error(err)
				return
			}

			if got != cas.wantAffectRows {
				t.Errorf("want %v, got %v", cas.wantAffectRows, got)
			}
		})
	}
}
