package sqlmy

import (
	"context"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDB_QueryRowContextWithBuilderStruct(t *testing.T) {
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

			cas.doMock(mock)
			u := &UserForTest{}
			err = db.QueryRowContextWithBuilderStruct(context.Background(), cas.sqlBuilder, u)
			if cas.wantErr {
				if err == nil {
					t.Errorf("want err, got nil")
				}
				return
			}

			if *u != cas.wantData {
				t.Errorf("want %v, got %v", cas.wantData, *u)
			}
		})
	}
}

func TestDB_QueryContextWithBuilderStruct(t *testing.T) {
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

			cas.doMock(mock)
			us := []UserForTest{}
			err = db.QueryContextWithBuilderStruct(context.Background(), cas.sqlBuilder, &us)
			if cas.wantErr {
				if err == nil {
					t.Errorf("want err, got nil")
				}
				return
			}

			if !reflect.DeepEqual(us, cas.wantData) {
				t.Errorf("want %v, got %v", cas.wantData, us)
			}
		})
	}
}

func TestDB_ExecContextWithBuilderStruct(t *testing.T) {
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

			cas.doMock(mock)
			rs, err := db.ExecContextWithBuilder(context.Background(), cas.sqlBuilder)
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
