package sqlmy

import (
	"context"
	"database/sql"
	"fmt"
)

var (
	_ Executor = &sql.DB{}
	_ Executor = &sql.Conn{}
	_ Executor = &sql.Tx{}

	_ Conn = &sql.DB{}
	_ Conn = &sql.Conn{}
)

var (
	ErrConnNotInit = fmt.Errorf("conn not init, call WithConn at first")
	ErrTxNotInit   = fmt.Errorf("tx not init, call WithConn at first")
)

type Executor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	// QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Conn is database/sql Conn or DB's execute function
type Conn interface {
	Executor

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type dbCtxKey int

var _dbCtxKey dbCtxKey

type dbContext struct {
	conn Conn
	tx   *sql.Tx

	// tx be open count
	openCount int
}

// WithConn will make sure context carray the same conn
// if context carray conn, do nothing, return old context and nil
// otherwis, try to get conn and create one new context
func WithConn(ctx context.Context, connFactory func() (conn Conn, err error)) (context.Context, error) {
	dbCtx := GetExecutor(ctx)
	if dbCtx != nil {
		return ctx, nil
	}

	conn, err := connFactory()
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, _dbCtxKey, &dbContext{
		conn: conn,
	}), nil
}

// GetExecutor will return tx at first if exist, or return conn, or return nil
func GetExecutor(ctx context.Context) Executor {
	hc, ok := ctx.Value(_dbCtxKey).(*dbContext)
	if !ok {
		return nil
	}

	if hc.tx != nil {
		return hc.tx
	}

	return hc.conn
}

func QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	executor := GetExecutor(ctx)
	if executor == nil {
		return nil, ErrConnNotInit
	}

	return executor.QueryContext(ctx, query, args...)
}

func ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	executor := GetExecutor(ctx)
	if executor == nil {
		return nil, ErrConnNotInit
	}

	return executor.ExecContext(ctx, query, args...)
}

type logKey int

var _logKey logKey

func WithLogID(ctx context.Context, logID string) context.Context {
	return context.WithValue(ctx, _logKey, logID)
}

func GetLogID(ctx context.Context) string {
	logID, _ := ctx.Value(_logKey).(string)
	return logID
}
