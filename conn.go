package sqlmy

import (
	"context"
	"database/sql"
	"time"
)

var _ Conn = &conn{}

type conn struct {
	rawConn *sql.Conn
	db      *MyDB
	hooks   *Hooks
}

func (conn *conn) DB() DB {
	return conn.db
}

func (conn *conn) RawConn() *sql.Conn {
	return conn.rawConn
}

func (conn *conn) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	rawStmt, err := conn.rawConn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return newStmt(rawStmt, query, conn.db), nil
}

func (conn *conn) emptyHook(typ EventType) bool {
	return conn.db == nil || conn.db.emptyHook(typ)
}

func (conn *conn) triggerHook(ctx context.Context, typ EventType, cost time.Duration, sql string, args []interface{}, err error) {
	if conn.emptyHook(typ) {
		return
	}
	conn.db.triggerHook(ctx, typ, cost, sql, args, err)
}

func (conn *conn) withCost(f func(), typ EventType) (cost time.Duration) {
	if conn.emptyHook(typ) {
		f()
		return
	}
	now := time.Now()
	f()
	return time.Now().Sub(now)
}

func (conn *conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	var rawTx *sql.Tx
	var err error
	cost := conn.withCost(func() {
		rawTx, err = conn.rawConn.BeginTx(ctx, opts)
	}, EventTxBegin)
	conn.triggerHook(ctx, EventTxBegin, cost, "", nil, err)
	if err != nil {
		return nil, err
	}

	return newTx(rawTx, conn.db), nil
}

func (conn *conn) Close() error {
	return conn.rawConn.Close()
}

func (conn *conn) Ping(ctx context.Context) error {
	var err error
	cost := conn.withCost(func() {
		err = conn.rawConn.PingContext(ctx)
	}, EventPing)
	conn.triggerHook(ctx, EventPing, cost, "", nil, err)

	return err
}

func (conn *conn) ExecContext(ctx context.Context, query string, args ...interface{}) (rst sql.Result, err error) {
	cost := conn.withCost(func() {
		rst, err = conn.rawConn.ExecContext(ctx, query, args...)
	}, EventExec)
	conn.triggerHook(ctx, EventExec, cost, query, args, err)

	return rst, err
}

func (conn *conn) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	var rawRow *sql.Row
	cost := conn.withCost(func() {
		rawRow = conn.rawConn.QueryRowContext(ctx, query, args...)
	}, EventQueryRow)
	conn.triggerHook(ctx, EventQueryRow, cost, query, args, nil)
	return newRow(rawRow)
}

func (conn *conn) Raw(f func(driverConn interface{}) error) (err error) {
	return conn.rawConn.Raw(f)
}

func (conn *conn) ExecContextWithBuilder(ctx context.Context, builder SqlBuilder) (sql.Result, error) {
	query, args, err := builder.Compile()
	if err != nil {
		conn.triggerHook(ctx, EventExec, 0, query, args, err)
		return nil, err
	}
	return conn.ExecContext(ctx, query, args...)
}

func (conn *conn) QueryRowContextWithBuilder(ctx context.Context, builder SqlBuilder) Row {
	query, args, err := builder.Compile()
	if err != nil {
		conn.triggerHook(ctx, EventQueryRow, 0, query, args, err)
		return &row{err: err}
	}
	return conn.QueryRowContext(ctx, query, args...)
}

func (conn *conn) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	var rawRows *sql.Rows
	var err error
	cost := conn.withCost(func() {
		rawRows, err = conn.rawConn.QueryContext(ctx, query, args...)
	}, EventQuery)
	conn.triggerHook(ctx, EventQuery, cost, query, args, err)
	if err != nil {
		return nil, err
	}

	return newRows(rawRows), nil
}

func (conn *conn) QueryContextWithBuilder(ctx context.Context, builder SqlBuilder) (Rows, error) {
	query, args, err := builder.Compile()
	if err != nil {
		conn.triggerHook(ctx, EventQuery, 0, query, args, err)
		return nil, err
	}
	return conn.QueryContext(ctx, query, args...)
}

func (conn *conn) QueryContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error {
	rows, err := conn.QueryContextWithBuilder(ctx, builder)
	if err != nil {
		return err
	}

	return rows.ScanStruct(dst)
}

func (conn *conn) QueryRowContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error {
	return conn.QueryContextWithBuilderStruct(ctx, builder, dst)
}
