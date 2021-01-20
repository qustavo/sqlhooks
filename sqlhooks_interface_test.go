package sqlhooks

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var interfaceTestCases = []struct {
	name               string
	expectedInterfaces []interface{}
}{
	{"Basic", []interface{}{(*driver.Conn)(nil)}},
	{"Execer", []interface{}{(*driver.Execer)(nil)}},
	{"ExecerContext", []interface{}{(*driver.ExecerContext)(nil)}},
	{"Queryer", []interface{}{(*driver.QueryerContext)(nil)}},
	{"QueryerContext", []interface{}{(*driver.QueryerContext)(nil)}},
	{"ExecerQueryerContext", []interface{}{
		(*driver.ExecerContext)(nil),
		(*driver.QueryerContext)(nil)}},
}

type fakeDriver struct{}

func (d *fakeDriver) Open(dsn string) (driver.Conn, error) {
	switch dsn {
	case "Basic":
		return &struct{ *FakeConnBasic }{}, nil
	case "Execer":
		return &struct {
			*FakeConnBasic
			*FakeConnExecer
		}{}, nil
	case "ExecerContext":
		return &struct {
			*FakeConnBasic
			*FakeConnExecerContext
		}{}, nil
	case "Queryer":
		return &struct {
			*FakeConnBasic
			*FakeConnQueryer
		}{}, nil
	case "QueryerContext":
		return &struct {
			*FakeConnBasic
			*FakeConnQueryerContext
		}{}, nil
	case "ExecerQueryerContext":
		return &struct {
			*FakeConnBasic
			*FakeConnExecerContext
			*FakeConnQueryerContext
		}{}, nil
	case "ExecerQueryerContextSessionResetter":
		return &struct {
			*FakeConnBasic
			*FakeConnExecer
			*FakeConnQueryer
			*FakeConnSessionResetter
		}{}, nil
	case "NonConnBeginTx":
		return &FakeConnUnsupported{}, nil
	}

	return nil, errors.New("Fake driver not implemented")
}

// Conn implements a database/sql.driver.Conn
type FakeConnBasic struct{}

func (*FakeConnBasic) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("Not implemented")
}
func (*FakeConnBasic) Close() error {
	return errors.New("Not implemented")
}
func (*FakeConnBasic) Begin() (driver.Tx, error) {
	return nil, errors.New("Not implemented")
}
func (*FakeConnBasic) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return nil, errors.New("Not implemented")
}

type FakeConnExecer struct{}

func (*FakeConnExecer) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, errors.New("Not implemented")
}

type FakeConnExecerContext struct{}

func (*FakeConnExecerContext) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("Not implemented")
}

type FakeConnQueryer struct{}

func (*FakeConnQueryer) Query(query string, args []driver.Value) (driver.Rows, error) {
	return nil, errors.New("Not implemented")
}

type FakeConnQueryerContext struct{}

func (*FakeConnQueryerContext) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("Not implemented")
}

type FakeConnSessionResetter struct{}

func (*FakeConnSessionResetter) ResetSession(ctx context.Context) error {
	return errors.New("Not implemented")
}

// FakeConnUnsupported implements a database/sql.driver.Conn but doesn't implement
// driver.ConnBeginTx.
type FakeConnUnsupported struct{}

func (*FakeConnUnsupported) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("Not implemented")
}
func (*FakeConnUnsupported) Close() error {
	return errors.New("Not implemented")
}
func (*FakeConnUnsupported) Begin() (driver.Tx, error) {
	return nil, errors.New("Not implemented")
}

func TestInterfaces(t *testing.T) {
	drv := Wrap(&fakeDriver{}, &testHooks{})

	for _, c := range interfaceTestCases {
		conn, err := drv.Open(c.name)
		require.NoErrorf(t, err, "Driver name %s", c.name)

		for _, i := range c.expectedInterfaces {
			assert.Implements(t, i, conn)
		}
	}
}

func TestUnsupportedDrivers(t *testing.T) {
	drv := Wrap(&fakeDriver{}, &testHooks{})
	_, err := drv.Open("NonConnBeginTx")
	require.EqualError(t, err, "driver must implement driver.ConnBeginTx")
}
