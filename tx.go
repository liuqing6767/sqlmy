package sqlmy

import (
	"context"
	"database/sql"
	"time"
)

var _ Tx = &tx{}

type tx struct {
	rawTx *sql.Tx
	db    *MyDB
}

func (t *tx) emptyHook(typ EventType) bool {
	return t.db == nil || t.db.emptyHook(typ)
}

func (t *tx) triggerHook(ctx context.Context, typ EventType, cost time.Duration, sql string, args []interface{}, err error) {
	if t.emptyHook(typ) {
		return
	}
	t.db.triggerHook(ctx, typ, cost, sql, args, err)
}

func (t *tx) withCost(f func(), typ EventType) (cost time.Duration) {
	if t.emptyHook(typ) {
		f()
		return
	}
	now := time.Now()
	f()
	return time.Now().Sub(now)
}

func newTx(raw *sql.Tx, db *MyDB) *tx {
	return &tx{
		rawTx: raw,
		db:    db,
	}
}

func (t *tx) Raw() *sql.Tx {
	return t.rawTx
}

func (t *tx) Commit() error {
	var err error
	cost := t.withCost(func() {
		err = t.rawTx.Commit()
	}, EventTxCommit)
	t.triggerHook(nil, EventTxCommit, cost, "", nil, err)
	return err
}

func (t *tx) Rollback() error {
	var err error
	cost := t.withCost(func() {
		err = t.rawTx.Rollback()
	}, EventTxRollback)
	t.triggerHook(nil, EventTxRollback, cost, "", nil, err)
	return err
}

func (t *tx) ExecContext(ctx context.Context, query string, args ...interface{}) (rst sql.Result, err error) {
	cost := t.withCost(func() {
		rst, err = t.rawTx.ExecContext(ctx, query, args...)
	}, EventExec)
	t.triggerHook(ctx, EventExec, cost, query, args, err)
	return
}

func (t *tx) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	var rawRows *sql.Rows
	var err error
	cost := t.withCost(func() {
		rawRows, err = t.rawTx.QueryContext(ctx, query, args...)
	}, EventQuery)
	t.triggerHook(ctx, EventQuery, cost, query, args, err)
	if err != nil {
		return nil, err
	}

	return newRows(rawRows), nil
}

func (t *tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	var rawRow *sql.Row
	cost := t.withCost(func() {
		rawRow = t.rawTx.QueryRowContext(ctx, query, args...)
	}, EventQueryRow)
	t.triggerHook(ctx, EventQueryRow, cost, query, args, nil)
	return newRow(rawRow)
}

func (t *tx) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	rawStmt, err := t.rawTx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return newStmt(rawStmt, query, t.db), nil
}

func (t *tx) StmtContext(ctx context.Context, stmt Stmt) Stmt {
	rawStmt := t.rawTx.StmtContext(ctx, stmt.Raw())
	return newStmt(rawStmt, "", t.db)

}

func (t *tx) ExecContextWithBuilder(ctx context.Context, builder SqlBuilder) (sql.Result, error) {
	query, args, err := builder.Compile()
	if err != nil {
		t.triggerHook(ctx, EventExec, 0, query, args, err)
		return nil, err
	}

	return t.ExecContext(ctx, query, args...)
}

func (t *tx) QueryContextWithBuilder(ctx context.Context, builder SqlBuilder) (Rows, error) {
	query, args, err := builder.Compile()
	if err != nil {
		t.triggerHook(ctx, EventQuery, 0, query, args, err)
		return nil, err
	}

	return t.QueryContext(ctx, query, args...)
}

func (t *tx) QueryRowContextWithBuilder(ctx context.Context, builder SqlBuilder) Row {
	query, args, err := builder.Compile()
	if err != nil {
		t.triggerHook(ctx, EventQueryRow, 0, query, args, err)
		return &row{err: err}
	}
	return t.QueryRowContext(ctx, query, args...)
}

func (t *tx) QueryContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error {
	rows, err := t.QueryContextWithBuilder(ctx, builder)
	if err != nil {
		return nil
	}

	return rows.ScanStruct(dst)
}

func (t *tx) QueryRowContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error {
	return t.QueryContextWithBuilderStruct(ctx, builder, dst)
}
