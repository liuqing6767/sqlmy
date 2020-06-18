package sqlmy

import (
	"context"
	"fmt"
	"time"
)

// TODO pool
func NewEvent(db DB, typ EventType, cost time.Duration, sql string, args []interface{}, err error, log Logger) *Event {
	return &Event{
		DB:   db,
		Type: typ,
		Sql:  sql,
		Args: args,
		Err:  err,
		Log:  log,
		Cost: cost,
	}
}

type Event struct {
	DB   DB
	Type EventType
	Sql  string
	Args []interface{}
	Err  error
	Log  Logger

	Cost time.Duration
}

type Handler func(ctx context.Context, ev *Event)

type EventType int8

const (
	EventPing EventType = iota
	EventStmt
	EventStmtClose
	EventTxBegin
	EventTxCommit
	EventTxRollback
	EventExec
	EventQueryRow
	EventQuery
)

var eventNames = [...]string{
	EventPing:       "Ping",
	EventStmt:       "stmt",
	EventStmtClose:  "StmtClose",
	EventTxBegin:    "TxBegin",
	EventTxCommit:   "TxCommit",
	EventTxRollback: "TxRollback",
	EventExec:       "Exec",
	EventQueryRow:   "QueryRow",
	EventQuery:      "Query",
}

type Hooks struct {
	Handlers [9][]Handler
}

func (hs *Hooks) Empty(et EventType) bool {
	return len(hs.Handlers[et]) == 0
}

func (hs *Hooks) SetHandler(he EventType, h Handler) {
	hs.Handlers[he] = []Handler{h}
}

func (hs *Hooks) AddHandler(he EventType, h Handler) {
	hs.Handlers[he] = append(hs.Handlers[he], h)
}

func (hs *Hooks) Append(hook *Hooks) {
	for typ, handlers := range hook.Handlers {
		hs.Handlers[typ] = append(hs.Handlers[typ], handlers...)
	}
}

func (hs *Hooks) Trigger(ctx context.Context, event *Event) {
	if event.Type < EventPing || event.Type > EventQuery {
		panic(fmt.Sprintf("bad event type: %d", event.Type))
	}

	for _, h := range hs.Handlers[event.Type] {
		h(ctx, event)
	}
}

func LogHook(ctx context.Context, ev *Event) {
	type logIDer interface {
		GetLogID() string
	}
	logID := ""
	if lctx, ok := ctx.(logIDer); ok {
		logID = lctx.GetLogID()
	}

	msg := fmt.Sprintf("logid[%s] event[%s] db[%s] sql[%s] args[%v] err[%v]", logID, eventNames[ev.Type],
		ev.DB.Name(), ev.Sql, ev.Args, ev.Err)
	ev.Log.Info(msg)
	if ev.Err != nil {
		ev.Log.Error(msg)
	}
}

func NewLogHooks() *Hooks {
	return &Hooks{
		Handlers: [9][]Handler{
			EventPing:       {LogHook},
			EventStmt:       {LogHook},
			EventStmtClose:  {LogHook},
			EventTxBegin:    {LogHook},
			EventTxCommit:   {LogHook},
			EventTxRollback: {LogHook},
			EventExec:       {LogHook},
			EventQueryRow:   {LogHook},
			EventQuery:      {LogHook},
		},
	}
}
