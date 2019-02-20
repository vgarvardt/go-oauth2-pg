package pgx_adapter

import (
	"github.com/jackc/pgx"
	"github.com/vgarvardt/go-oauth2-pg"
	pgxHelpers "github.com/vgarvardt/pgx-helpers"
)

// ConnPool is the adapter type for PGx connection pool connection type
type ConnPool struct {
	conn *pgx.ConnPool
}

// NewConnPool instantiates PGx connection pool adapter
func NewConnPool(conn *pgx.ConnPool) *ConnPool {
	return &ConnPool{conn}
}

// Conn is the adapter type for PGx connection connection type
type Conn struct {
	conn *pgx.Conn
}

// NewConn instantiates PGx connection adapter
func NewConn(conn *pgx.Conn) *Conn {
	return &Conn{conn}
}

// Exec runs a query and returns an error if any
func (a *ConnPool) Exec(query string, args ...interface{}) error {
	_, err := a.conn.Exec(query, args...)
	return err
}

// SelectOne runs a select query and scans the object into a struct or returns an error
func (a *ConnPool) SelectOne(dst interface{}, query string, args ...interface{}) error {
	row := a.conn.QueryRow(query, args...)
	if err := pgxHelpers.ScanStruct(row, dst); err != nil {
		if err == pgx.ErrNoRows {
			return pg.ErrNoRows
		}
		return err
	}

	return nil
}

// Exec runs a query and returns an error if any
func (a *Conn) Exec(query string, args ...interface{}) error {
	_, err := a.conn.Exec(query, args...)
	return err
}

// SelectOne runs a select query and scans the object into a struct or returns an error
func (a *Conn) SelectOne(dst interface{}, query string, args ...interface{}) error {
	row := a.conn.QueryRow(query, args...)
	if err := pgxHelpers.ScanStruct(row, dst); err != nil {
		if err == pgx.ErrNoRows {
			return pg.ErrNoRows
		}
		return err
	}

	return nil
}
