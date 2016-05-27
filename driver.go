package sqlhooks

import (
	"database/sql"
	"database/sql/driver"
	"strconv"
	"time"
)

func driverToInterface(args []driver.Value) []interface{} {
	r := make([]interface{}, len(args))
	for i, arg := range args {
		r[i] = arg
	}
	return r
}

func interfaceToDriver(args []interface{}) []driver.Value {
	r := make([]driver.Value, len(args))
	for i, arg := range args {
		r[i] = arg
	}
	return r
}

type Context struct {
	id    string
	Error error
	Query string
	Args  []interface{}
}

func NewContext() *Context {
	now := time.Now().UnixNano()
	return &Context{id: strconv.FormatInt(now, 10)}
}

func (ctx *Context) GetID() string {
	return ctx.id
}

type Beginner interface {
	BeforeBegin(c *Context) error
	AfterBegin(c *Context) error
}

type Commiter interface {
	BeforeCommit(c *Context) error
	AfterCommit(c *Context) error
}

type Rollbacker interface {
	BeforeRollback(c *Context) error
	AfterRollback(c *Context) error
}

type Stmter interface {
	BeforePrepare(c *Context) error
	AfterPrepare(c *Context) error

	BeforeStmtQuery(c *Context) error
	AfterStmtQuery(c *Context) error

	BeforeStmtExec(c *Context) error
	AfterStmtExec(c *Context) error
}

type Queryer interface {
	BeforeQuery(c *Context) error
	AfterQuery(c *Context) error
}

type Execer interface {
	BeforeExec(c *Context) error
	AfterExec(c *Context) error
}

type tx struct {
	driver.Tx
	hooks interface{}
	ctx   *Context
}

func (t tx) Commit() error {
	var ctx *Context

	if v, ok := t.hooks.(Commiter); ok {
		ctx = NewContext()
		if err := v.BeforeCommit(ctx); err != nil {
			return err
		}
	}

	err := t.Tx.Commit()

	if v, ok := t.hooks.(Commiter); ok {
		ctx.Error = err
		err = v.AfterCommit(ctx)
	}

	return err
}

func (t tx) Rollback() error {
	var ctx *Context

	if v, ok := t.hooks.(Rollbacker); ok {
		ctx = NewContext()
		if err := v.BeforeRollback(ctx); err != nil {
			return err
		}
	}

	err := t.Tx.Rollback()

	if v, ok := t.hooks.(Rollbacker); ok {
		ctx.Error = err
		err = v.AfterRollback(ctx)
	}

	return err
}

type stmt struct {
	driver.Stmt
	hooks interface{}
	ctx   *Context
}

func (s stmt) Close() error {
	return s.Stmt.Close()
}

func (s stmt) Exec(args []driver.Value) (res driver.Result, err error) {
	if t, ok := s.hooks.(Stmter); ok {
		s.ctx.Args = driverToInterface(args)
		if err := t.BeforeStmtExec(s.ctx); err != nil {
			return nil, err
		}
		args = interfaceToDriver(s.ctx.Args)
	}

	return s.Stmt.Exec(args)
}

func (s stmt) NumInput() int {
	return s.Stmt.NumInput()
}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	if t, ok := s.hooks.(Stmter); ok {
		s.ctx.Args = driverToInterface(args)
		if err := t.BeforeStmtQuery(s.ctx); err != nil {
			return nil, err
		}
		args = interfaceToDriver(s.ctx.Args)
	}

	rows, err := s.Stmt.Query(args)

	if t, ok := s.hooks.(Stmter); ok {
		s.ctx.Error = err
		err = t.AfterStmtQuery(s.ctx)
	}

	return rows, err
}

type conn struct {
	driver.Conn
	hooks interface{}
}

func (c conn) Prepare(query string) (driver.Stmt, error) {
	var ctx *Context

	if t, ok := c.hooks.(Stmter); ok {
		ctx = NewContext()
		ctx.Query = query

		if err := t.BeforePrepare(ctx); err != nil {
			return nil, err
		}

		query = ctx.Query
	}

	_stmt, err := c.Conn.Prepare(query)

	if t, ok := c.hooks.(Stmter); ok {
		err = t.AfterPrepare(ctx)
	}

	return stmt{_stmt, c.hooks, ctx}, err
}

func (c conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.Conn.(driver.Queryer); ok {
		var ctx *Context
		if t, ok := c.hooks.(Queryer); ok {
			ctx = NewContext()
			ctx.Query = query
			ctx.Args = driverToInterface(args)

			if err := t.BeforeQuery(ctx); err != nil {
				return nil, err
			}

			query = ctx.Query
			args = interfaceToDriver(ctx.Args)
		}

		rows, err := queryer.Query(query, args)

		if t, ok := c.hooks.(Queryer); ok {
			ctx.Error = err
			err = t.AfterQuery(ctx)
		}

		return rows, err
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.Conn.(driver.Execer); ok {
		var ctx *Context
		if t, ok := c.hooks.(Execer); ok {
			ctx = NewContext()
			ctx.Query = query
			ctx.Args = driverToInterface(args)

			if err := t.BeforeExec(ctx); err != nil {
				return nil, err
			}

			query = ctx.Query
			args = interfaceToDriver(ctx.Args)

		}

		res, err := execer.Exec(query, args)

		if t, ok := c.hooks.(Execer); ok {
			ctx.Error = err
			err = t.AfterExec(ctx)
		}

		return res, err
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Close() error {
	return c.Conn.Close()
}

func (c conn) Begin() (driver.Tx, error) {
	var ctx *Context

	if t, ok := c.hooks.(Beginner); ok {
		ctx = NewContext()

		if err := t.BeforeBegin(ctx); err != nil {
			return nil, err
		}
	}

	_tx, err := c.Conn.Begin()

	if t, ok := c.hooks.(Beginner); ok {
		ctx.Error = err
		err = t.AfterBegin(ctx)
	}

	return tx{_tx, c.hooks, ctx}, err
}

// Driver it's a proxy for a specific sql driver
type Driver struct {
	driver driver.Driver
	name   string
	hooks  interface{}
}

// NewDriver will create a Proxy Driver with defined Hooks
// name is the underlying driver name
func NewDriver(name string, hooks interface{}) *Driver {
	return &Driver{name: name, hooks: hooks}
}

// Open returns a new connection to the database, using the underlying specified driver
func (d *Driver) Open(dsn string) (driver.Conn, error) {
	if d.driver == nil {
		// Get Driver by Opening a new connection
		db, err := sql.Open(d.name, dsn)
		if err != nil {
			return nil, err
		}
		if err := db.Close(); err != nil {
			return nil, err
		}
		d.driver = db.Driver()
	}

	_conn, err := d.driver.Open(dsn)
	return conn{_conn, d.hooks}, err
}
