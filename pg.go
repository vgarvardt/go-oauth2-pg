package pg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/json-iterator/go"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

var ErrNoRows = errors.New("sql: no rows in result set")

// Adapter represents DB access layer interface for different PostgreSQL drivers
type Adapter interface {
	Exec(query string, args ...interface{}) error
	SelectOne(dst interface{}, query string, args ...interface{}) error
}

// Logger is the PostgreSQL store logger interface
type Logger interface {
	Printf(format string, v ...interface{})
}

// Store mysql token store
type Store struct {
	adapter   Adapter
	tableName string
	logger    Logger

	gcDisabled bool
	gcInterval time.Duration
	ticker     *time.Ticker

	initTableDisabled bool
}

// StoreItem data item
type StoreItem struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Code      string    `db:"code"`
	Access    string    `db:"access"`
	Refresh   string    `db:"refresh"`
	Data      []byte    `db:"data"`
}

// NewStore creates PostgreSQL store instance
func NewStore(adapter Adapter, options ...Option) (*Store, error) {
	store := &Store{
		adapter:    adapter,
		tableName:  "oauth2_token",
		logger:     log.New(os.Stderr, "[OAUTH2-PG-ERROR]", log.LstdFlags),
		gcInterval: 10 * time.Minute,
	}

	for _, o := range options {
		o(store)
	}

	var err error
	if !store.initTableDisabled {
		err = store.initTable()
	}

	if err != nil {
		return store, err
	}

	if !store.gcDisabled {
		store.ticker = time.NewTicker(store.gcInterval)
		go store.gc()
	}

	return store, err
}

// Close close the store
func (s *Store) Close() error {
	if !s.gcDisabled {
		s.ticker.Stop()
	}
	return nil
}

func (s *Store) gc() {
	for range s.ticker.C {
		s.clean()
	}
}

func (s *Store) initTable() error {
	return s.adapter.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %[1]s (
  id         BIGSERIAL   NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  code       TEXT        NOT NULL,
  access     TEXT        NOT NULL,
  refresh    TEXT        NOT NULL,
  data       JSONB       NOT NULL,
  CONSTRAINT %[1]s_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_%[1]s_expires_at ON %[1]s (expires_at);
CREATE INDEX IF NOT EXISTS idx_%[1]s_code ON %[1]s (code);
CREATE INDEX IF NOT EXISTS idx_%[1]s_access ON %[1]s (access);
CREATE INDEX IF NOT EXISTS idx_%[1]s_refresh ON %[1]s (refresh);
`, s.tableName))
}

func (s *Store) clean() {
	now := time.Now()
	err := s.adapter.Exec(fmt.Sprintf("DELETE FROM %s WHERE expires_at <= $1", s.tableName), now)
	if err != nil {
		s.logger.Printf("Error while cleaning out outdated entities: %+v", err)
	}
}

// Create create and store the new token information
func (s *Store) Create(info oauth2.TokenInfo) error {
	buf, err := jsoniter.Marshal(info)
	if err != nil {
		return err
	}

	item := &StoreItem{
		Data:      buf,
		CreatedAt: time.Now(),
	}

	if code := info.GetCode(); code != "" {
		item.Code = code
		item.ExpiresAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())
	} else {
		item.Access = info.GetAccess()
		item.ExpiresAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())

		if refresh := info.GetRefresh(); refresh != "" {
			item.Refresh = info.GetRefresh()
			item.ExpiresAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
		}
	}

	return s.adapter.Exec(
		fmt.Sprintf("INSERT INTO %s (created_at, expires_at, code, access, refresh, data) VALUES ($1, $2, $3, $4, $5, $6)", s.tableName),
		item.CreatedAt,
		item.ExpiresAt,
		item.Code,
		item.Access,
		item.Refresh,
		item.Data,
	)
}

// RemoveByCode delete the authorization code
func (s *Store) RemoveByCode(code string) error {
	err := s.adapter.Exec(fmt.Sprintf("DELETE FROM %s WHERE code = $1", s.tableName), code)
	if err == ErrNoRows {
		return nil
	}
	return err
}

// RemoveByAccess use the access token to delete the token information
func (s *Store) RemoveByAccess(access string) error {
	err := s.adapter.Exec(fmt.Sprintf("DELETE FROM %s WHERE access = $1", s.tableName), access)
	if err == ErrNoRows {
		return nil
	}
	return err
}

// RemoveByRefresh use the refresh token to delete the token information
func (s *Store) RemoveByRefresh(refresh string) error {
	err := s.adapter.Exec(fmt.Sprintf("DELETE FROM %s WHERE refresh = $1", s.tableName), refresh)
	if err == ErrNoRows {
		return nil
	}
	return err
}

func (s *Store) toTokenInfo(data []byte) (oauth2.TokenInfo, error) {
	var tm models.Token
	err := jsoniter.Unmarshal(data, &tm)
	return &tm, err
}

// GetByCode use the authorization code for token information data
func (s *Store) GetByCode(code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.adapter.SelectOne(&item, fmt.Sprintf("SELECT * FROM %s WHERE code = $1", s.tableName), code); err != nil {
		return nil, err
	}
	return s.toTokenInfo(item.Data)
}

// GetByAccess use the access token for token information data
func (s *Store) GetByAccess(access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.adapter.SelectOne(&item, fmt.Sprintf("SELECT * FROM %s WHERE access = $1", s.tableName), access); err != nil {
		return nil, err
	}
	return s.toTokenInfo(item.Data)
}

// GetByRefresh use the refresh token for token information data
func (s *Store) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.adapter.SelectOne(&item, fmt.Sprintf("SELECT * FROM %s WHERE refresh = $1", s.tableName), refresh); err != nil {
		return nil, err
	}
	return s.toTokenInfo(item.Data)
}
