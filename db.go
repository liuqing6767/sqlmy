package sqlmy

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"time"
)

var _ DB = &MyDB{}

type MyDB struct {
	DBName   string
	DSN      string
	Protocol string
	Config   *DBConfig

	StatusLogger Logger
	AccessLog    Logger

	Hooks *Hooks

	enableLog bool
	rawDB     *sql.DB
}

type DBConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifeTime time.Duration
}

func (db *MyDB) Name() string {
	return db.DBName
}

func (db *MyDB) SetHook(et EventType, h Handler) {
	if db.Hooks == nil {
		db.Hooks = &Hooks{}
	}

	db.Hooks.SetHandler(et, h)
}

func (db *MyDB) AddHook(et EventType, h Handler) {
	if db.Hooks == nil {
		db.Hooks = &Hooks{}
	}

	db.Hooks.AddHandler(et, h)
}

func (db *MyDB) EnableLog(enable bool) {
	db.enableLog = enable
	if enable {
		if db.AccessLog == nil {
			// TODO
			// db.AccessLog = NewStdoutLogger()
		}
	}
}

func (db *MyDB) Raw() *sql.DB {
	return db.rawDB
}

func (db *MyDB) Init() (err error) {
	if db.StatusLogger == nil {
		// if db.AccessLog == nil || db.StatusLogger == nil {
		return fmt.Errorf("Log must be set")
	}

	if db.enableLog {
		if db.Hooks == nil {
			db.Hooks = &Hooks{}
		}
		db.Hooks.Append(NewLogHooks())
	}

	if db.Protocol == "" {
		db.Protocol = "mysql"
	}

	// 因为权限问题而重试
	for i := 0; i < 3; i++ {
		db.rawDB, err = sql.Open(db.Protocol, db.DSN)
		if err != nil {
			db.StatusLogger.Warnf("%s open db try %d fail: %v", db.Name(), i+1, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		err = db.rawDB.PingContext(context.Background())
		if err != nil {
			db.StatusLogger.Warnf("%s open db try %d ping fail: %v", db.Name(), i+1, err)
			time.Sleep(500 * time.Millisecond)
		}
	}
	if err != nil {
		err = fmt.Errorf("%s open db fail: %v", db.Name(), err)
		db.StatusLogger.Error(err)
		return err
	}

	if config := db.Config; config != nil {
		if maxOpenConns := config.MaxOpenConns; maxOpenConns != 0 {
			db.SetMaxOpenConns(maxOpenConns)
		}
		if maxIdleConns := config.MaxIdleConns; maxIdleConns != 0 {
			db.SetMaxIdleConns(maxIdleConns)
		}
		if maxLifeTime := config.ConnMaxLifeTime; maxLifeTime != 0 {
			db.SetConnMaxLifetime(maxLifeTime)
		}
	}

	db.StatusLogger.Infof("conn db %s success, DSN: %v", db.Name(), db.DSN)
	return nil
}

func (db *MyDB) Start() error {
	return nil
}

func (db *MyDB) Stop(os.Signal) {
	if err := db.rawDB.Close(); err != nil {
		db.StatusLogger.Errorf("[%s][%s]close db fail: %v", err)
	}
}

func (db *MyDB) emptyHook(typ EventType) bool {
	return db.Hooks == nil || db.Hooks.Empty(typ)
}

func (db *MyDB) triggerHook(ctx context.Context, typ EventType, cost time.Duration, sql string, args []interface{}, err error) {
	if db.emptyHook(typ) {
		return
	}
	db.Hooks.Trigger(ctx, NewEvent(db, typ, cost, sql, args, err, db.AccessLog))
}

func (db *MyDB) withCost(f func(), typ EventType) (cost time.Duration) {
	if db.emptyHook(typ) {
		f()
		return
	}
	now := time.Now()
	f()
	return time.Now().Sub(now)
}

func (db *MyDB) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	var rawStmt *sql.Stmt
	var err error
	cost := db.withCost(func() {
		rawStmt, err = db.rawDB.PrepareContext(ctx, query)
	}, EventStmt)
	db.triggerHook(ctx, EventStmt, cost, query, nil, err)
	if err != nil {
		return nil, err
	}

	return newStmt(rawStmt, query, db), nil
}

func (db *MyDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	var rawTx *sql.Tx
	var err error
	cost := db.withCost(func() {
		rawTx, err = db.rawDB.BeginTx(ctx, opts)
	}, EventTxBegin)
	db.triggerHook(ctx, EventTxBegin, cost, "", nil, err)
	if err != nil {
		return nil, err
	}

	return newTx(rawTx, db), nil
}

func (db *MyDB) Close() error {
	return db.rawDB.Close()
}

func (db *MyDB) Ping(ctx context.Context) error {
	var err error
	cost := db.withCost(func() {
		err = db.rawDB.PingContext(ctx)
	}, EventPing)
	db.triggerHook(ctx, EventPing, cost, "", nil, err)

	return err
}

func (db *MyDB) ExecContext(ctx context.Context, query string, args ...interface{}) (rst sql.Result, err error) {
	cost := db.withCost(func() {
		rst, err = db.rawDB.ExecContext(ctx, query, args...)
	}, EventExec)
	db.triggerHook(ctx, EventExec, cost, query, args, err)
	return rst, err
}

func (db *MyDB) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	var rawRows *sql.Rows
	var err error
	cost := db.withCost(func() {
		rawRows, err = db.rawDB.QueryContext(ctx, query, args...)
	}, EventQuery)
	db.triggerHook(ctx, EventQuery, cost, query, args, err)
	if err != nil {
		return nil, err
	}

	return newRows(rawRows), nil
}

func (db *MyDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	var rawRow *sql.Row
	cost := db.withCost(func() {
		rawRow = db.rawDB.QueryRowContext(ctx, query, args...)
	}, EventQueryRow)
	db.triggerHook(ctx, EventQueryRow, cost, query, args, nil)
	return newRow(rawRow)
}

func (db *MyDB) Conn(ctx context.Context) (Conn, error) {
	rawConn, err := db.rawDB.Conn(ctx)
	if err != nil {
		return nil, err
	}

	return &conn{
		rawConn: rawConn,
		db:      db,
	}, nil
}

func (db *MyDB) SetConnMaxLifetime(d time.Duration) {
	if db.StatusLogger != nil {
		db.StatusLogger.Warnf("%s SetConnMaxLifetime: %v", db.Name(), d)
	}
	db.rawDB.SetConnMaxLifetime(d)
}

func (db *MyDB) SetMaxIdleConns(n int) {
	if db.StatusLogger != nil {
		db.StatusLogger.Warnf("%s SetMaxIdleConns: %v", db.Name(), n)
	}
	db.rawDB.SetMaxIdleConns(n)
}

func (db *MyDB) SetMaxOpenConns(n int) {
	if db.StatusLogger != nil {
		db.StatusLogger.Warnf("%s SetMaxOpenConns: %v", db.Name(), n)
	}
	db.rawDB.SetMaxOpenConns(n)
}

func (db *MyDB) Stats() sql.DBStats {
	return db.rawDB.Stats()
}

func (db *MyDB) Driver() driver.Driver {
	return db.rawDB.Driver()
}

func (db *MyDB) ExecContextWithBuilder(ctx context.Context, builder SqlBuilder) (sql.Result, error) {
	query, args, err := builder.Compile()
	if err != nil {
		db.triggerHook(ctx, EventExec, 0, query, args, err)
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

func (db *MyDB) QueryContextWithBuilder(ctx context.Context, builder SqlBuilder) (Rows, error) {
	query, args, err := builder.Compile()
	if err != nil {
		db.triggerHook(ctx, EventQuery, 0, query, args, err)
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

func (db *MyDB) QueryRowContextWithBuilder(ctx context.Context, builder SqlBuilder) Row {
	query, args, err := builder.Compile()
	if err != nil {
		db.triggerHook(ctx, EventQueryRow, 0, query, args, err)
		return &row{err: err}
	}
	return db.QueryRowContext(ctx, query, args)
}

func (db *MyDB) QueryContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error {
	rows, err := db.QueryContextWithBuilder(ctx, builder)
	if err != nil {
		return err
	}

	return rows.ScanStruct(dst)
}

func (db *MyDB) QueryRowContextWithBuilderStruct(ctx context.Context, builder SqlBuilder, dst interface{}) error {
	return db.QueryContextWithBuilderStruct(ctx, builder, dst)
}
