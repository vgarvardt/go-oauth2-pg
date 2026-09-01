package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
)

var _ oauth2.TokenStore = (*TokenStore)(nil)

// TokenStore PostgreSQL token store
type TokenStore struct {
	db        *sql.DB
	tableName string
	logger    *slog.Logger

	gcDisabled bool
	gcInterval time.Duration
	ticker     *time.Ticker

	initTableDisabled bool

	queryDeleteExpired   string
	queryCreate          string
	queryDeleteByCode    string
	queryDeleteByAccess  string
	queryDeleteByRefresh string
	querySelectByCode    string
	querySelectByAccess  string
	querySelectByRefresh string
}

// TokenStoreItem data item
type TokenStoreItem struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Code      string    `db:"code"`
	Access    string    `db:"access"`
	Refresh   string    `db:"refresh"`
	Data      []byte    `db:"data"`
}

// NewTokenStore creates PostgreSQL store instance
func NewTokenStore(ctx context.Context, db *sql.DB, options ...TokenStoreOption) (*TokenStore, error) {
	store := &TokenStore{
		db:         db,
		tableName:  "oauth2_tokens",
		logger:     slog.New(slog.DiscardHandler),
		gcInterval: 10 * time.Minute,
	}

	for _, o := range options {
		o(store)
	}

	var err error
	if !store.initTableDisabled {
		err = store.initTable(ctx)
	}

	if err != nil {
		return store, err
	}

	if !store.gcDisabled {
		store.ticker = time.NewTicker(store.gcInterval)
		go store.gc()
	}

	store.queryDeleteExpired = fmt.Sprintf("DELETE FROM %s WHERE expires_at <= $1", store.tableName)
	store.queryCreate = fmt.Sprintf("INSERT INTO %s (created_at, expires_at, code, access, refresh, data) VALUES ($1, $2, $3, $4, $5, $6)", store.tableName)
	store.queryDeleteByCode = fmt.Sprintf("DELETE FROM %s WHERE code = $1", store.tableName)
	store.queryDeleteByAccess = fmt.Sprintf("DELETE FROM %s WHERE access = $1", store.tableName)
	store.queryDeleteByRefresh = fmt.Sprintf("DELETE FROM %s WHERE refresh = $1", store.tableName)
	store.querySelectByCode = fmt.Sprintf("SELECT id, created_at, expires_at, code, access, refresh, data FROM %s WHERE code = $1", store.tableName)
	store.querySelectByAccess = fmt.Sprintf("SELECT id, created_at, expires_at, code, access, refresh, data FROM %s WHERE access = $1", store.tableName)
	store.querySelectByRefresh = fmt.Sprintf("SELECT id, created_at, expires_at, code, access, refresh, data FROM %s WHERE refresh = $1", store.tableName)

	return store, err
}

// Close closes the store
func (s *TokenStore) Close() error {
	if !s.gcDisabled {
		s.ticker.Stop()
	}
	return nil
}

func (s *TokenStore) gc() {
	for range s.ticker.C {
		s.clean()
	}
}

func (s *TokenStore) initTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
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
	return err
}

func (s *TokenStore) clean() {
	now := time.Now()
	_, err := s.db.ExecContext(context.Background(), s.queryDeleteExpired, now)
	if err != nil {
		s.logger.Error("Error while cleaning out outdated entities", slog.Any("error", err))
	}
}

// Create creates and stores the new token information
func (s *TokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	buf, err := json.Marshal(info)
	if err != nil {
		return err
	}

	item := &TokenStoreItem{
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

	_, err = s.db.ExecContext(ctx, s.queryCreate, item.CreatedAt, item.ExpiresAt, item.Code, item.Access, item.Refresh, item.Data)
	return err
}

// RemoveByCode deletes the authorization code
func (s *TokenStore) RemoveByCode(ctx context.Context, code string) error {
	_, err := s.db.ExecContext(ctx, s.queryDeleteByCode, code)
	return err
}

// RemoveByAccess uses the access token to delete the token information
func (s *TokenStore) RemoveByAccess(ctx context.Context, access string) error {
	_, err := s.db.ExecContext(ctx, s.queryDeleteByAccess, access)
	return err
}

// RemoveByRefresh uses the refresh token to delete the token information
func (s *TokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	_, err := s.db.ExecContext(ctx, s.queryDeleteByRefresh, refresh)
	return err
}

func (s *TokenStore) toTokenInfo(data []byte) (oauth2.TokenInfo, error) {
	var tm models.Token
	err := json.Unmarshal(data, &tm)
	return &tm, err
}

// GetByCode uses the authorization code for token information data
func (s *TokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	return s.getByQuery(ctx, s.querySelectByCode, code)
}

// GetByAccess uses the access token for token information data
func (s *TokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	return s.getByQuery(ctx, s.querySelectByAccess, access)
}

// GetByRefresh uses the refresh token for token information data
func (s *TokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	return s.getByQuery(ctx, s.querySelectByRefresh, refresh)
}

func (s *TokenStore) getByQuery(ctx context.Context, query string, args ...any) (oauth2.TokenInfo, error) {
	var item TokenStoreItem
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&item.ID,
		&item.CreatedAt,
		&item.ExpiresAt,
		&item.Code,
		&item.Access,
		&item.Refresh,
		&item.Data,
	); err != nil {
		return nil, err
	}

	return s.toTokenInfo(item.Data)
}
