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
	if s.hooks.Exec == nil {
		return s.Stmt.Exec(args)
	}

	return s.hooks.Exec(
		func() (driver.Result, error) {
			return s.Stmt.Exec(args)
		},
		s.query,
		convertArgs(args)...,
	)
}

func (s stmt) NumInput() int {
	return s.Stmt.NumInput()
}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	if s.hooks.Query == nil {
		return s.Stmt.Query(args)
	}

	return s.hooks.Query(
		func() (driver.Rows, error) {
			return s.Stmt.Query(args)
		},
		s.query,
		convertArgs(args)...,
	)
}

type conn struct {
	driver.Conn
	hooks *Hooks
}

func (c conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.Conn.(driver.Queryer); ok {
		if c.hooks.Query == nil {
			return queryer.Query(query, args)
		}

		return c.hooks.Query(
			func() (driver.Rows, error) {
				return queryer.Query(query, args)
			},
			query,
			convertArgs(args)...,
		)
	}

	// Not implemented by underlying driver
	return nil, driver.ErrSkip
}

func (c conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.Conn.(driver.Execer); ok {
		if c.hooks.Exec == nil {
			return execer.Exec(query, args)
		}

		return c.hooks.Exec(
			func() (driver.Result, error) {
				return execer.Exec(query, args)
			},
			query,
			convertArgs(args)...,
		)
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

type ExecFn func() (driver.Result, error)
type QueryFn func() (driver.Rows, error)

// Hooks contains hook functions for instrumenting Query and Exec
type Hooks struct {
	Exec  func(ExecFn, string, ...interface{}) (driver.Result, error)
	Query func(QueryFn, string, ...interface{}) (driver.Rows, error)
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

// Register will register the driver using sql.Register()
func Register(name string, driver Driver) {
	sql.Register(name, &driver)
}
