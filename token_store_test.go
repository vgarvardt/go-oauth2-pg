package pg

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pgAdapter "github.com/vgarvardt/go-pg-adapter"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"
	"github.com/vgarvardt/go-pg-adapter/sqladapter"
)

var uri string

func TestMain(m *testing.M) {
	uri = os.Getenv("PG_URI")
	if uri == "" {
		fmt.Println("Env variable PG_URI is required to run the tests")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type memoryLogger struct {
	formats []string
	args    [][]interface{}

	pgxLogs []struct {
		ctx   context.Context
		level pgx.LogLevel
		msg   string
		data  map[string]interface{}
	}
}

func (l *memoryLogger) Printf(format string, v ...interface{}) {
	l.formats = append(l.formats, format)
	l.args = append(l.args, v)
}

func (l *memoryLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	l.pgxLogs = append(l.pgxLogs, struct {
		ctx   context.Context
		level pgx.LogLevel
		msg   string
		data  map[string]interface{}
	}{ctx: ctx, level: level, msg: msg, data: data})
}

type mockAdapter struct {
	mock.Mock
}

func (m *mockAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	mArgs := m.Called(ctx, query, args)
	return mArgs.Error(0)
}

func (m *mockAdapter) SelectOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	mArgs := m.Called(ctx, dst, query, args)
	return mArgs.Error(0)
}

func TestTokenStore_initTable(t *testing.T) {
	adapter := new(mockAdapter)

	adapter.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		query := args.Get(1).(string)
		// new line character is the character at position 0
		assert.Equal(t, 1, strings.Index(query, "CREATE TABLE IF NOT EXISTS"))
	})

	store, err := NewTokenStore(adapter, WithTokenStoreGCDisabled())
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, store.Close())
	}()
}

func TestTokenStore_gc(t *testing.T) {
	adapter := new(mockAdapter)

	var execCalls int
	adapter.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		execCalls++

		query := args.Get(1).(string)
		// new line character is the character at position 0
		assert.Equal(t, 0, strings.Index(query, "DELETE FROM"))
	})

	store, err := NewTokenStore(adapter, WithTokenStoreInitTableDisabled(), WithTokenStoreGCInterval(time.Second))
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, store.Close())
	}()

	time.Sleep(5 * time.Second)

	// in 5 seconds we should have 4-5 gc calls
	assert.True(t, 3 < execCalls)
	assert.True(t, 5 >= execCalls)
}

func generateTokenTableName() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}

func generateClientTableName() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func TestPGXConn(t *testing.T) {
	l := new(memoryLogger)

	pgxConnConfig, err := pgx.ParseConfig(uri)
	require.NoError(t, err)

	pgxConnConfig.Logger = l

	pgxConn, err := pgx.ConnectConfig(context.Background(), pgxConnConfig)
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, pgxConn.Close(context.Background()))
	}()

	adapter := pgx4adapter.NewConn(pgxConn)

	tokenStore, err := NewTokenStore(
		adapter,
		WithTokenStoreLogger(l),
		WithTokenStoreTableName(generateTokenTableName()),
		WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := NewClientStore(
		adapter,
		WithClientStoreLogger(l),
		WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore, l)
	runClientStoreTest(t, clientStore)
}

func TestPGXConnPool(t *testing.T) {
	l := new(memoryLogger)

	pgxPoolConnConfig, err := pgxpool.ParseConfig(uri)
	require.NoError(t, err)

	pgxPoolConnConfig.ConnConfig.Logger = l

	pgXConnPool, err := pgxpool.ConnectConfig(context.Background(), pgxPoolConnConfig)
	require.NoError(t, err)

	defer pgXConnPool.Close()

	adapter := pgx4adapter.NewPool(pgXConnPool)

	tokenStore, err := NewTokenStore(
		adapter,
		WithTokenStoreLogger(l),
		WithTokenStoreTableName(generateTokenTableName()),
		WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := NewClientStore(
		adapter,
		WithClientStoreLogger(l),
		WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore, l)
	runClientStoreTest(t, clientStore)
}

func TestSQL(t *testing.T) {
	l := new(memoryLogger)

	conn, err := sql.Open("pgx", uri)
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, conn.Close())
	}()

	adapter := sqladapter.New(conn)

	tokenStore, err := NewTokenStore(
		adapter,
		WithTokenStoreLogger(l),
		WithTokenStoreTableName(generateTokenTableName()),
		WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := NewClientStore(
		adapter,
		WithClientStoreLogger(l),
		WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore, l)
	runClientStoreTest(t, clientStore)
}

func TestNewX(t *testing.T) {
	l := new(memoryLogger)

	conn, err := sql.Open("pgx", uri)
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, conn.Close())
	}()

	adapter := sqladapter.NewX(sqlx.NewDb(conn, ""))

	tokenStore, err := NewTokenStore(
		adapter,
		WithTokenStoreLogger(l),
		WithTokenStoreTableName(generateTokenTableName()),
		WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := NewClientStore(
		adapter,
		WithClientStoreLogger(l),
		WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore, l)
	runClientStoreTest(t, clientStore)
}

func runTokenStoreTest(t *testing.T, store *TokenStore, l *memoryLogger) {
	runTokenStoreCodeTest(t, store)
	runTokenStoreAccessTest(t, store)
	runTokenStoreRefreshTest(t, store)

	// sleep for a while just to wait for GC run for sure to ensure there were no errors there
	time.Sleep(3 * time.Second)

	assert.Equal(t, 0, len(l.formats))
}

func runTokenStoreCodeTest(t *testing.T, store *TokenStore) {
	code := fmt.Sprintf("code %s", time.Now().String())
	ctx := context.Background()

	tokenCode := models.NewToken()
	tokenCode.SetCode(code)
	tokenCode.SetCodeCreateAt(time.Now())
	tokenCode.SetCodeExpiresIn(time.Minute)
	require.NoError(t, store.Create(ctx, tokenCode))

	token, err := store.GetByCode(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetCode())

	require.NoError(t, store.RemoveByCode(ctx, code))

	_, err = store.GetByCode(ctx, code)
	assert.Equal(t, pgAdapter.ErrNoRows, err)
}

func runTokenStoreAccessTest(t *testing.T, store *TokenStore) {
	code := fmt.Sprintf("access %s", time.Now().String())
	ctx := context.Background()

	tokenCode := models.NewToken()
	tokenCode.SetAccess(code)
	tokenCode.SetAccessCreateAt(time.Now())
	tokenCode.SetAccessExpiresIn(time.Minute)
	require.NoError(t, store.Create(ctx, tokenCode))

	token, err := store.GetByAccess(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetAccess())

	require.NoError(t, store.RemoveByAccess(ctx, code))

	_, err = store.GetByAccess(ctx, code)
	assert.Equal(t, pgAdapter.ErrNoRows, err)
}

func runTokenStoreRefreshTest(t *testing.T, store *TokenStore) {
	code := fmt.Sprintf("refresh %s", time.Now().String())
	ctx := context.Background()

	tokenCode := models.NewToken()
	tokenCode.SetRefresh(code)
	tokenCode.SetRefreshCreateAt(time.Now())
	tokenCode.SetRefreshExpiresIn(time.Minute)
	require.NoError(t, store.Create(ctx, tokenCode))

	token, err := store.GetByRefresh(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetRefresh())

	require.NoError(t, store.RemoveByRefresh(ctx, code))

	_, err = store.GetByRefresh(ctx, code)
	assert.Equal(t, pgAdapter.ErrNoRows, err)
}

func runClientStoreTest(t *testing.T, store *ClientStore) {
	originalClient := &models.Client{
		ID:     fmt.Sprintf("id %s", time.Now().String()),
		Secret: fmt.Sprintf("secret %s", time.Now().String()),
		Domain: fmt.Sprintf("domain %s", time.Now().String()),
		UserID: fmt.Sprintf("user id %s", time.Now().String()),
	}
	ctx := context.Background()

	require.NoError(t, store.Create(originalClient))

	client, err := store.GetByID(ctx, originalClient.GetID())
	require.NoError(t, err)
	assert.Equal(t, originalClient.GetID(), client.GetID())
	assert.Equal(t, originalClient.GetSecret(), client.GetSecret())
	assert.Equal(t, originalClient.GetDomain(), client.GetDomain())
	assert.Equal(t, originalClient.GetUserID(), client.GetUserID())
}
