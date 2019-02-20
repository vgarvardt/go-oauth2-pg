package sql

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/vgarvardt/go-oauth2-pg"
)

// Adapter is the adapter type for sqlx.DB connection type
type Adapter struct {
	conn *sqlx.DB
}

// NewSQL instantiates sqlx.DB connection adapter from sql.DB connection
func NewSQL(conn *sql.DB) *Adapter {
	// The driverName of the original database is required for named query support - we do not use it here
	return &Adapter{sqlx.NewDb(conn, "")}
}

// NewSQLx instantiates sqlx.DB connection adapter
func NewSQLx(conn *sqlx.DB) *Adapter {
	return &Adapter{conn}
}

// Exec runs a query and returns an error if any
func (a *Adapter) Exec(query string, args ...interface{}) error {
	_, err := a.conn.Exec(query, args...)
	return err
}

// SelectOne runs a select query and scans the object into a struct or returns an error
func (a *Adapter) SelectOne(dst interface{}, query string, args ...interface{}) error {
	if err := a.conn.Get(dst, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return pg.ErrNoRows
		}
		return err
	}

	return nil
}
