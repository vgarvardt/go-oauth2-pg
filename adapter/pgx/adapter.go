package pgx

import (
	"github.com/jackc/pgx"
	"github.com/vgarvardt/go-oauth2-pg"
	pgxHelpers "github.com/vgarvardt/pgx-helpers"
)

// ConnPoolAdapter is the adapter type for PGx connection pool connection type
type ConnPoolAdapter struct {
	conn *pgx.ConnPool
}

// NewConnPoolAdapter instantiates PGx connection pool adapter
func NewConnPoolAdapter(conn *pgx.ConnPool) *ConnPoolAdapter {
	return &ConnPoolAdapter{conn}
}

// ConnPoolAdapter is the adapter type for PGx connection connection type
type ConnAdapter struct {
	conn *pgx.Conn
}

// NewConnAdapter instantiates PGx connection adapter
func NewConnAdapter(conn *pgx.Conn) *ConnAdapter {
	return &ConnAdapter{conn}
}

// Exec runs a query and returns an error if any
func (a *ConnPoolAdapter) Exec(query string, args ...interface{}) error {
	_, err := a.conn.Exec(query, args...)
	return err
}

// SelectOne runs a select query and scans the object into a struct or returns an error
func (a *ConnPoolAdapter) SelectOne(dst interface{}, query string, args ...interface{}) error {
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
func (a *ConnAdapter) Exec(query string, args ...interface{}) error {
	_, err := a.conn.Exec(query, args...)
	return err
}

// SelectOne runs a select query and scans the object into a struct or returns an error
func (a *ConnAdapter) SelectOne(dst interface{}, query string, args ...interface{}) error {
	row := a.conn.QueryRow(query, args...)
	if err := pgxHelpers.ScanStruct(row, dst); err != nil {
		if err == pgx.ErrNoRows {
			return pg.ErrNoRows
		}
		return err
	}

	return nil
}
