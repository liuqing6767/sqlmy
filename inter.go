package sqlmy

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

type SqlBuilder interface {
	Compile() (sql string, args []interface{}, err error)
}

// curdStmtDeprecated 这些被官方废弃，将直接丢弃
// type curdStmtDeprecated interface {
// 	Exec(args ...interface{}) (sql.Result, error)
// 	Query(args ...interface{}) (Rows, error)
// 	QueryRow(args ...interface{}) Scanner
// }
type stdCURDStmt interface {
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, args ...interface{}) Row
}

// curdDeprecated 这些被官方废弃，将直接丢弃
// type curdDeprecated interface {
// 	Exec(query string, args ...interface{}) (sql.Result, error)
// 	Query(query string, args ...interface{}) (Rows, error)
// 	QueryRow(query string, args ...interface{}) Scanner
// }

// 官方库
type stdCURD interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

type curdWithBuilder interface {
	ExecContextWithBuilder(ctx context.Context, builder SqlBuilder) (sql.Result, error)
	QueryContextWithBuilder(ctx context.Context, builder SqlBuilder) (Rows, error)
	QueryRowContextWithBuilder(ctx context.Context, builder SqlBuilder) Row
}

type curdWithBuilderStruct interface {
	QueryContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error
	QueryRowContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error
}

type stdConnMeta interface {
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
}

type QueryExecer interface {
	stdCURD
	curdWithBuilder
	curdWithBuilderStruct
}

type stdConn interface {
	// PrepareContext(ctx context.Context, query string) (Stmt, error)
	// BeginTx(ctx context.Context, opts TxOptions) (Tx, error)
	// Close() error
	stdConnMeta

	// PingContext(ctx context.Context) error
	driver.Pinger

	// ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	// QueryRowContext(ctx context.Context, query string, args ...interface{}) Scanner
	stdCURD

	Raw(f func(driverConn interface{}) error) (err error)
}

type Conn interface {
	stdConn
	curdWithBuilder
	curdWithBuilderStruct

	DB() DB

	RawConn() *sql.Conn
}

type stdDB interface {
	// Prepare(query string) (Stmt, error)
	// Begin() (Tx, error)

	// PrepareContext(ctx context.Context, query string) (Stmt, error)
	// BeginTx(ctx context.Context, opts TxOptions) (Tx, error)
	// Close() error
	stdConnMeta

	// Ping() error
	// PingContext(ctx context.Context) error
	driver.Pinger

	// Exec(query string, args ...interface{}) (sql.Result, error)
	// ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// Query(query string, args ...interface{}) (Rows, error)
	// QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	// QueryRow(query string, args ...interface{}) Scanner
	// QueryRowContext(ctx context.Context, query string, args ...interface{}) Scanner
	// curdDeprecated
	stdCURD

	Conn(ctx context.Context) (Conn, error)

	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
	Driver() driver.Driver
}

type DB interface {
	Name() string

	stdDB
	curdWithBuilder
	curdWithBuilderStruct

	Raw() *sql.DB
}

type Stmt interface {
	Close() error

	stdCURDStmt
	// curdStmtDeprecated

	Raw() *sql.Stmt
}

type stdTx interface {
	// Commit() error
	// Rollback() error
	driver.Tx

	// Exec(query string, args ...interface{}) (sql.Result, error)
	// ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// Query(query string, args ...interface{}) (Rows, error)
	// QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	// QueryRow(query string, args ...interface{}) Scanner
	// QueryRowContext(ctx context.Context, query string, args ...interface{}) Scanner
	// curdDeprecated
	stdCURD

	// Prepare(query string) (Stmt, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	// Stmt(stmt Stmt) Stmt
	StmtContext(ctx context.Context, stmt Stmt) Stmt
}

type Tx interface {
	stdTx
	curdWithBuilder
	curdWithBuilderStruct

	Raw() *sql.Tx
}

type irows interface {
	Close() error
	ColumnTypes() ([]*sql.ColumnType, error)
	Columns() ([]string, error)
	Err() error
	Next() bool
	NextResultSet() bool

	scanner
}

type Rows interface {
	irows
	scannerStruct

	Raw() *sql.Rows
}

type Row interface {
	scanner

	// scannerStruct

	Raw() *sql.Row
}

type scanner interface {
	Scan(dest ...interface{}) error
}

type scannerStruct interface {
	ScanStruct(dst interface{}) error
}

type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Error(...interface{})
	Info(...interface{})
}
