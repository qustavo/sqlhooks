package sqlhooks

import (
	"database/sql"
	"database/sql/driver"
	"strconv"
	"time"
)

func convertArgs(args []driver.Value) []interface{} {
	r := make([]interface{}, len(args))
	for i, arg := range args {
		r[i] = arg
	}
	return r
}

// Hooks contains hook functions for sql operations
// Returned func() will be executed after statements have completed
// ID will be the same within the same transaction
type Hooks struct {
	Exec     func(string, ...interface{}) func(error)
	Query    func(string, ...interface{}) func(error)
	Begin    func(id string)
	Commit   func(id string)
	Rollback func(id string)
}

func (h *Hooks) query(query string, args []driver.Value) func(error) {
	if hook := h.Query; hook != nil {
		fn := hook(query, convertArgs(args)...)
		if fn != nil {
			return fn
		}
	}
	return func(error) {}
}

func (h *Hooks) exec(query string, args []driver.Value) func(error) {
	if hook := h.Exec; hook != nil {
		fn := hook(query, convertArgs(args)...)
		if fn != nil {
			return fn
		}
	}
	return func(error) {}
}

type tx struct {
	driver.Tx
	hooks *Hooks
	id    string
}

func (t tx) Commit() error {
	if hook := t.hooks.Commit; hook != nil {
		hook(t.id)
	}

	return t.Tx.Commit()
}

func (t tx) Rollback() error {
	if hook := t.hooks.Rollback; hook != nil {
		hook(t.id)
	}

	return t.Tx.Rollback()
}

type stmt struct {
	driver.Stmt
	query string
	hooks *Hooks
}

func (s stmt) Close() error {
	return s.Stmt.Close()
}

func (s stmt) Exec(args []driver.Value) (res driver.Result, err error) {
	defer s.hooks.exec(s.query, args)(nil)
	return s.Stmt.Exec(args)
}

func (s stmt) NumInput() int {
	return s.Stmt.NumInput()
}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	defer s.hooks.query(s.query, args)(nil)
	return s.Stmt.Query(args)
}

type conn struct {
	driver.Conn
	hooks *Hooks
}

func (c conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.Conn.(driver.Queryer); ok {
		fn := c.hooks.query(query, args)
		rows, err := queryer.Query(query, args)
		fn(err)
		return rows, err
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.Conn.(driver.Execer); ok {
		fn := c.hooks.exec(query, args)
		res, err := execer.Exec(query, args)
		fn(err)
		return res, err
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Prepare(query string) (driver.Stmt, error) {
	_stmt, err := c.Conn.Prepare(query)
	return stmt{_stmt, query, c.hooks}, err
}

func (c conn) Close() error {
	return c.Conn.Close()
}

func (c conn) Begin() (driver.Tx, error) {
	_tx, err := c.Conn.Begin()
	id := strconv.FormatInt(time.Now().UnixNano(), 16)
	if hook := c.hooks.Begin; hook != nil {
		hook(id)
	}
	return tx{_tx, c.hooks, id}, err
}

// Driver it's a proxy for a specific sql driver
type Driver struct {
	driver driver.Driver
	name   string
	hooks  *Hooks
}

// NewDriver will create a Proxy Driver with defined Hooks
// name is the underlying driver name
func NewDriver(name string, hooks *Hooks) Driver {
	return Driver{name: name, hooks: hooks}
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
