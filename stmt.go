package sqlmy

import (
	"context"
	"database/sql"
	"time"
)

var _ Stmt = &stmt{}

type stmt struct {
	rawStmt *sql.Stmt
	db      *MyDB
	query   string
	args    []interface{}
}

func newStmt(raw *sql.Stmt, sql string, db *MyDB) *stmt {
	return &stmt{
		rawStmt: raw,
		query:   sql,
		db:      db,
	}
}

func (s *stmt) emptyHook(typ EventType) bool {
	return s.db == nil || s.db.emptyHook(typ)
}

func (s *stmt) triggerHook(ctx context.Context, typ EventType, cost time.Duration, err error) {
	if s.emptyHook(typ) {
		return
	}
	s.db.triggerHook(ctx, typ, cost, s.query, s.args, err)
}

func (s *stmt) withCost(f func(), typ EventType) (cost time.Duration) {
	if s.emptyHook(typ) {
		f()
		return
	}
	now := time.Now()
	f()
	return time.Now().Sub(now)
}

func (s *stmt) Raw() *sql.Stmt {
	return s.rawStmt
}

func (s *stmt) Close() error {
	var err error
	cost := s.withCost(func() {
		err = s.rawStmt.Close()
	}, EventStmtClose)
	s.triggerHook(nil, EventStmtClose, cost, err)
	return err
}

func (s *stmt) ExecContext(ctx context.Context, args ...interface{}) (rst sql.Result, err error) {
	s.args = args
	cost := s.withCost(func() {
		rst, err = s.rawStmt.ExecContext(ctx, args...)
	}, EventExec)
	s.triggerHook(ctx, EventExec, cost, err)

	return rst, err
}

func (s *stmt) QueryContext(ctx context.Context, args ...interface{}) (Rows, error) {
	s.args = args
	var rawRows *sql.Rows
	var err error
	cost := s.withCost(func() {
		rawRows, err = s.rawStmt.QueryContext(ctx, args...)
	}, EventQuery)
	s.triggerHook(ctx, EventQuery, cost, err)
	if err != nil {
		return nil, err
	}

	return &rows{
		rawRows: rawRows,
	}, err
}

func (s *stmt) QueryRowContext(ctx context.Context, args ...interface{}) Row {
	s.args = args
	var rawRow *sql.Row
	cost := s.withCost(func() {
		rawRow = s.rawStmt.QueryRowContext(ctx, args...)
	}, EventQueryRow)
	s.triggerHook(ctx, EventQueryRow, cost, nil)
	return &row{
		rawRow: rawRow,
	}
}
