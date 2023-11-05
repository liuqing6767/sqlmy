package sqlmy

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/liuximu/sqlmy/internal"
)

type InsertType int

const (
	InsertTypeCommonInsert  InsertType = 0
	InsertTypeIgnoreInsert  InsertType = 1
	InsertTypeReplaceInsert InsertType = 2
)

type curdOpt func(*curdOption)

type curdOption struct {
	queryBuilder func(table string, fields []string, where any) (sql string, args []any, err error)
	fields       []string

	updateBuilder func(table string, where, assign any) (sql string, args []any, err error)

	insertBuilder func(table string, typ InsertType, datas ...any) (sql string, args []any, err error)
	insertType    InsertType
	batchSize     int

	deleteBuilder func(table string, where any) (sql string, args []any, err error)

	rowsScan func(rs *sql.Rows, target interface{}) error
}

func WithQueryBuilder(builder func(table string, fields []string, where any) (sql string, args []any, err error)) curdOpt {
	return func(co *curdOption) {
		co.queryBuilder = builder
	}
}

func WithSelectFileds(fields ...string) curdOpt {
	return func(co *curdOption) {
		co.fields = fields
	}
}

func WithUpdateBuilder(builder func(table string, where, assign any) (sql string, args []any, err error)) curdOpt {
	return func(co *curdOption) {
		co.updateBuilder = builder
	}
}

func WithInsertBuilder(builder func(table string, typ InsertType, datas ...any) (sql string, args []any, err error)) curdOpt {
	return func(co *curdOption) {
		co.insertBuilder = builder
	}
}

func WithInsertType(typ InsertType) curdOpt {
	return func(co *curdOption) {
		co.insertType = typ
	}
}

func WithInsertBatchSize(batchSize int) curdOpt {
	return func(co *curdOption) {
		co.batchSize = batchSize
	}
}

func WithDeleteBuilder(builder func(table string, where any) (sql string, args []any, err error)) curdOpt {
	return func(co *curdOption) {
		co.deleteBuilder = builder
	}
}

var defaultInsertInsert = func(table string, typ InsertType, datas ...any) (sql string, args []any, err error) {
	return internal.BuildInsert(table, int(typ), datas...)
}

var allFileds = []string{"*"}

func newCURDOption(opts ...curdOpt) *curdOption {
	option := &curdOption{
		queryBuilder: internal.BuildQuery,
		fields:       allFileds,

		insertBuilder: defaultInsertInsert,
		batchSize:     math.MaxInt,

		rowsScan: internal.Scan,

		deleteBuilder: internal.BuildDelete,
		updateBuilder: internal.BuildUpdate,
	}
	for _, opt := range opts {
		opt(option)
	}

	return option
}

type CURDer[Data, Param any] interface {
	Query(ctx context.Context, where *Param, opts ...curdOpt) (*Data, error)
	QueryList(ctx context.Context, where *Param, opts ...curdOpt) ([]*Data, error)

	Insert(ctx context.Context, data *Param, opts ...curdOpt) (lastInsertedID int64, err error)
	InsertList(ctx context.Context, datas []*Param, opts ...curdOpt) (lastInsertedID int64, err error)

	Update(ctx context.Context, where *Param, data *Param, opts ...curdOpt) (affectedRows int64, err error)

	Delete(ctx context.Context, where *Param, opts ...curdOpt) (affectedRows int64, err error)
}

type CURD[Data, Param any] struct {
	table string
}

func NewCURD[Data, Param any](tableName string) *CURD[Data, Param] {
	return &CURD[Data, Param]{
		table: tableName,
	}

}

func argsDeal(args []any) []any {
	return args
}

func costMs(begin time.Time) int64 {
	return time.Since(begin).Milliseconds()

}

func (curd *CURD[Data, Param]) Query(ctx context.Context, param *Param, opts ...curdOpt) (*Data, error) {
	list, err := curd.QueryList(ctx, param, opts...)
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (curd *CURD[Data, Param]) QueryList(ctx context.Context, param *Param, opts ...curdOpt) ([]*Data, error) {
	begin := time.Now()
	option := newCURDOption(opts...)

	query, args, err := option.queryBuilder(curd.table, option.fields, param)
	if err != nil {
		logger.Error(ctx, "cost[%d] [QueryBuild] sql[%s] args[%v] err[%v]", costMs(begin), query, argsDeal(args), err)
		return nil, err
	}

	rows, err := QueryContext(ctx, query, args...)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		logger.Error(ctx, "cost[%d] [QueryQuery] err[%v]", costMs(begin), err)
		return nil, err
	}

	var datas []*Data
	if rows != nil {
		datas = []*Data{}
		err = option.rowsScan(rows, &datas)

		if err != nil {
			logger.Error(ctx, "cost[%d] [QueryScan] err[%v]", costMs(begin), err)
			return nil, err
		}
	}

	logger.Info(ctx, "cost[%d] [QuerySucc] sql[%s] args[%v] len[%d]", costMs(begin), query, argsDeal(args), len(datas))
	return datas, nil
}

func (curd *CURD[Data, Param]) Insert(ctx context.Context, data *Param, opts ...curdOpt) (lastInsertedID int64, err error) {
	return curd.InsertList(ctx, []*Param{data}, opts...)
}

func (curd *CURD[Data, Param]) InsertList(ctx context.Context, datas []*Param, opts ...curdOpt) (lastInsertedID int64, err error) {
	if len(datas) == 0 { // what are U doing...
		return
	}

	begin := time.Now()
	option := newCURDOption(opts...)

	var rst sql.Result
	for i := 0; i <= len(datas)/option.batchSize; i++ {
		a := i * option.batchSize
		b := (i + 1) * option.batchSize
		if b > len(datas) {
			b = len(datas)
		}

		tmp := make([]any, 0, b-a)
		for i := a; i < b; i++ {
			tmp = append(tmp, datas[i])
		}
		query, args, err := option.insertBuilder(curd.table, option.insertType, tmp...)
		if err != nil {
			logger.Error(ctx, "cost[%d] [UpdateBuild] [i] sql[%s] args[%v] err[%v]", costMs(begin), i, query, argsDeal(args), err)
			return 0, err
		}

		rst, err = ExecContext(ctx, query, args...)
		if err != nil {
			logger.Error(ctx, "cost[%d] [UpdateExec] [i] sql[%s] args[%v] err[%v]", costMs(begin), i, query, argsDeal(args), err)
			return 0, err
		}

		logger.Info(ctx, "cost[%d] [QuerySucc] [i] sql[%s] args[%v] len[%d]", costMs(begin), query, argsDeal(args), len(datas))
	}

	id, err := rst.LastInsertId()
	if err != nil {
		logger.Error(ctx, "cost[%d] [UpdateLastInsertID] last_id[%d] err[%v]", costMs(begin), id, err)
		return 0, err
	}

	return id, nil
}

func (curd *CURD[Data, Param]) Update(ctx context.Context, where *Param, assign *Param, opts ...curdOpt) (affectedRows int64, err error) {
	begin := time.Now()
	option := newCURDOption(opts...)

	query, args, err := option.updateBuilder(curd.table, where, assign)
	if err != nil {
		logger.Error(ctx, "cost[%d] [UpdateBuild] sql[%s] args[%v] err[%v]", costMs(begin), query, argsDeal(args), err)
		return 0, err
	}

	rst, err := ExecContext(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "cost[%d] [UpdateExec] err[%v]", costMs(begin), err)
		return 0, err
	}

	affectedRows, err = rst.RowsAffected()
	if err != nil {
		logger.Error(ctx, "cost[%d] [UpdateRowsAffected] err[%v]", costMs(begin), err)
		return 0, err
	}

	logger.Info(ctx, "cost[%d] [UpdateSucc] sql[%s] args[%v] rows[%d]", costMs(begin), query, argsDeal(args), affectedRows)
	return affectedRows, nil

}

func (curd *CURD[Data, Param]) Delete(ctx context.Context, where *Param, opts ...curdOpt) (affectedRows int64, err error) {
	begin := time.Now()
	option := newCURDOption(opts...)

	query, args, err := option.deleteBuilder(curd.table, where)
	if err != nil {
		logger.Error(ctx, "cost[%d] [DeleteBuild] sql[%s] args[%v] err[%v]", costMs(begin), query, argsDeal(args), err)
		return 0, err
	}

	rst, err := ExecContext(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "cost[%d] [DeleteExec] err[%v]", costMs(begin), err)
		return 0, err
	}

	affectedRows, err = rst.RowsAffected()
	if err != nil {
		logger.Error(ctx, "cost[%d] [DeleteRowsAffected] err[%v]", costMs(begin), err)
		return 0, err
	}

	logger.Info(ctx, "cost[%d] [DeleteSucc] sql[%s] args[%v] rows[%d]", costMs(begin), query, argsDeal(args), affectedRows)
	return affectedRows, nil
}
