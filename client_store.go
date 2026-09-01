package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
)

var _ oauth2.ClientStore = (*ClientStore)(nil)

// ClientStore PostgreSQL client store
type ClientStore struct {
	db        *sql.DB
	tableName string
	logger    *slog.Logger

	initTableDisabled bool

	querySelectByID string
	queryCreate     string
}

// ClientStoreItem data item
type ClientStoreItem struct {
	ID     string
	Secret string
	Domain string
	Data   []byte
}

// NewClientStore creates PostgreSQL store instance
func NewClientStore(ctx context.Context, db *sql.DB, options ...ClientStoreOption) (*ClientStore, error) {
	store := &ClientStore{
		db:        db,
		tableName: "oauth2_clients",
		logger:    slog.New(slog.DiscardHandler),
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

	store.querySelectByID = fmt.Sprintf(`SELECT "id", "secret", "domain", "data" FROM "%s" WHERE "id" = $1`, store.tableName)
	store.queryCreate = fmt.Sprintf(`INSERT INTO "%s" ("id", "secret", "domain", "data") VALUES ($1, $2, $3, $4)`, store.tableName)

	return store, err
}

func (s *ClientStore) initTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %[1]s (
	"id"     TEXT  NOT NULL,
	"secret" TEXT  NOT NULL,
	"domain" TEXT  NOT NULL,
	"data"   JSONB NOT NULL,
	CONSTRAINT %[1]s_pkey PRIMARY KEY (id)
);
`, s.tableName))
	return err
}

func (s *ClientStore) toClientInfo(data []byte) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := json.Unmarshal(data, &cm)
	return &cm, err
}

// GetByID retrieves and returns client information by id
func (s *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var item ClientStoreItem
	if err := s.db.QueryRowContext(ctx, s.querySelectByID, id).Scan(
		&item.ID, &item.Secret, &item.Domain, &item.Data,
	); err != nil {
		return nil, err
	}

	return s.toClientInfo(item.Data)
}

// Create creates and stores the new client information
func (s *ClientStore) Create(ctx context.Context, info oauth2.ClientInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, s.queryCreate, info.GetID(), info.GetSecret(), info.GetDomain(), data)
	return err
}
