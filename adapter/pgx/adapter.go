package pgx

import (
	"github.com/jackc/pgx"
	"github.com/vgarvardt/go-oauth2-pg"
	pgxHelpers "github.com/vgarvardt/pgx-helpers"
)

type ConnPoolAdapter struct {
	conn *pgx.ConnPool
}

func NewConnPoolAdapter(conn *pgx.ConnPool) *ConnPoolAdapter {
	return &ConnPoolAdapter{conn}
}

type ConnAdapter struct {
	conn *pgx.Conn
}

func NewConnAdapter(conn *pgx.Conn) *ConnAdapter {
	return &ConnAdapter{conn}
}

func (a *ConnPoolAdapter) Exec(query string, params ...interface{}) error {
	_, err := a.conn.Exec(query, params...)
	return err
}

func (a *ConnPoolAdapter) SelectOne(dst interface{}, query string, params ...interface{}) error {
	row := a.conn.QueryRow(query, params...)
	if err := pgxHelpers.ScanStruct(row, dst); err != nil {
		if err == pgx.ErrNoRows {
			return pg.ErrNoRows
		}
		return err
	}

	return nil
}

func (a *ConnAdapter) Exec(query string, params ...interface{}) error {
	_, err := a.conn.Exec(query, params...)
	return err
}

func (a *ConnAdapter) SelectOne(dst interface{}, query string, params ...interface{}) error {
	row := a.conn.QueryRow(query, params...)
	if err := pgxHelpers.ScanStruct(row, dst); err != nil {
		if err == pgx.ErrNoRows {
			return pg.ErrNoRows
		}
		return err
	}

	return nil
}
