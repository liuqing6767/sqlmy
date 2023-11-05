package sqlmy

import (
	"context"
	"database/sql"
)

// TxExec exec do in a transaction
// nested call TxExec is ok, only one transaction will be open
func TxExec(ctx context.Context, do func(dbCtx context.Context) error, opts ...*sql.TxOptions) error {
	err := openTx(ctx, opts...)
	if err != nil {
		return err
	}

	err = do(ctx)
	if err1 := closeTx(ctx, err == nil); err1 != nil {
		logger.Error(ctx, "close tx fail: err: %v", err1)
	}

	return err
}

func openTx(ctx context.Context, opts ...*sql.TxOptions) error {
	hc, ok := ctx.Value(_dbCtxKey).(*dbContext)
	if !ok {
		return ErrConnNotInit
	}

	if hc.tx != nil {
		hc.openCount++
		return nil
	}

	var opt *sql.TxOptions
	if len(opts) == 1 {
		opt = opts[0]
	}
	tx, err := hc.conn.BeginTx(ctx, opt)
	if err != nil {
		return err
	}

	hc.openCount++
	hc.tx = tx
	return nil
}

func closeTx(ctx context.Context, succ bool) error {
	hc, ok := ctx.Value(_dbCtxKey).(*dbContext)
	if !ok {
		return ErrConnNotInit
	}

	if hc.tx == nil {
		return ErrTxNotInit
	}

	hc.openCount--

	if hc.openCount != 0 {
		return nil
	}

	if succ {
		return hc.tx.Commit()
	}

	return hc.tx.Rollback()
}
