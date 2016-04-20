package sqlhook

import (
	"database/sql"
	"database/sql/driver"
)

func convertArgs(args []driver.Value) []interface{} {
	r := make([]interface{}, len(args))
	for i, arg := range args {
		r[i] = arg
	}
	return r
}

type stmt struct {
	driver.Stmt
	query string
	hooks *Hooks
}

func (s stmt) Close() error {
	return s.Stmt.Close()
}

func (s stmt) Exec(args []driver.Value) (driver.Result, error) {
	if err := s.hooks.Exec(s.query, convertArgs(args)...); err != nil {
		return nil, err
	}
	return s.Stmt.Exec(args)
}

func (s stmt) NumInput() int {
	return s.Stmt.NumInput()
}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	if err := s.hooks.Query(s.query, convertArgs(args)...); err != nil {
		return nil, err
	}
	return s.Stmt.Query(args)
}

type conn struct {
	driver.Conn
	hooks *Hooks
}

func (c conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.Conn.(driver.Queryer); ok {
		if err := c.hooks.Query(query, convertArgs(args)...); err != nil {
			return nil, err
		}
		return queryer.Query(query, args)
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.Conn.(driver.Execer); ok {
		if err := c.hooks.Exec(query, convertArgs(args)...); err != nil {
			return nil, err
		}
		return execer.Exec(query, args)
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
	return c.Conn.Begin()
}

// Hooks contains hook functions for instrumenting Query and Exec
type Hooks struct {
	Query func(query string, args ...interface{}) error
	Exec  func(query string, args ...interface{}) error
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
	noop := func(string, ...interface{}) error { return nil }

	if hooks.Exec == nil {
		hooks.Exec = noop
	}

	if hooks.Query == nil {
		hooks.Query = noop
	}

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

// Register will register the driver using sql.Register()
func Register(name string, driver Driver) {
	sql.Register(name, &driver)
}
