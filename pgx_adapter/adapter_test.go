package pgx_adapter

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vgarvardt/go-oauth2-pg"
	"gopkg.in/oauth2.v3/models"
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
		level pgx.LogLevel
		msg   string
		data  map[string]interface{}
	}
}

func (l *memoryLogger) Printf(format string, v ...interface{}) {
	l.formats = append(l.formats, format)
	l.args = append(l.args, v)
}

func (l *memoryLogger) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	l.pgxLogs = append(l.pgxLogs, struct {
		level pgx.LogLevel
		msg   string
		data  map[string]interface{}
	}{level: level, msg: msg, data: data})
}

func generateTokenTableName() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}

func generateClientTableName() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func TestNewConn(t *testing.T) {
	l := new(memoryLogger)

	pgxConnConfig, err := pgx.ParseURI(uri)
	require.NoError(t, err)

	pgxConnConfig.Logger = l

	pgxConn, err := pgx.Connect(pgxConnConfig)
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, pgxConn.Close())
	}()

	adapter := NewConn(pgxConn)

	tokenStore, err := pg.NewTokenStore(
		adapter,
		pg.WithTokenStoreLogger(l),
		pg.WithTokenStoreTableName(generateTokenTableName()),
		pg.WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := pg.NewClientStore(
		adapter,
		pg.WithClientStoreLogger(l),
		pg.WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore, l)
	runClientStoreTest(t, clientStore)
}

func TestNewConnPool(t *testing.T) {
	l := new(memoryLogger)

	pgxConnConfig, err := pgx.ParseURI(uri)
	require.NoError(t, err)

	pgxConnConfig.Logger = l

	pgxPoolConfig := pgx.ConnPoolConfig{ConnConfig: pgxConnConfig}

	pgXConnPool, err := pgx.NewConnPool(pgxPoolConfig)
	require.NoError(t, err)

	defer pgXConnPool.Close()

	adapter := NewConnPool(pgXConnPool)

	tokenStore, err := pg.NewTokenStore(
		adapter,
		pg.WithTokenStoreLogger(l),
		pg.WithTokenStoreTableName(generateTokenTableName()),
		pg.WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, tokenStore.Close())
	}()

	clientStore, err := pg.NewClientStore(
		adapter,
		pg.WithClientStoreLogger(l),
		pg.WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore, l)
	runClientStoreTest(t, clientStore)
}

func runTokenStoreTest(t *testing.T, store *pg.TokenStore, l *memoryLogger) {
	runTokenStoreCodeTest(t, store)
	runTokenStoreAccessTest(t, store)
	runTokenStoreRefreshTest(t, store)

	// sleep for a while just to wait for GC run for sure to ensure there were no errors there
	time.Sleep(3 * time.Second)

	assert.Equal(t, 0, len(l.formats))
}

func runTokenStoreCodeTest(t *testing.T, store *pg.TokenStore) {
	code := fmt.Sprintf("code %s", time.Now().String())

	tokenCode := models.NewToken()
	tokenCode.SetCode(code)
	tokenCode.SetCodeCreateAt(time.Now())
	tokenCode.SetCodeExpiresIn(time.Minute)
	require.NoError(t, store.Create(tokenCode))

	token, err := store.GetByCode(code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetCode())

	require.NoError(t, store.RemoveByCode(code))

	_, err = store.GetByCode(code)
	assert.Equal(t, pg.ErrNoRows, err)
}

func runTokenStoreAccessTest(t *testing.T, store *pg.TokenStore) {
	code := fmt.Sprintf("access %s", time.Now().String())

	tokenCode := models.NewToken()
	tokenCode.SetAccess(code)
	tokenCode.SetAccessCreateAt(time.Now())
	tokenCode.SetAccessExpiresIn(time.Minute)
	require.NoError(t, store.Create(tokenCode))

	token, err := store.GetByAccess(code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetAccess())

	require.NoError(t, store.RemoveByAccess(code))

	_, err = store.GetByAccess(code)
	assert.Equal(t, pg.ErrNoRows, err)
}

func runTokenStoreRefreshTest(t *testing.T, store *pg.TokenStore) {
	code := fmt.Sprintf("refresh %s", time.Now().String())

	tokenCode := models.NewToken()
	tokenCode.SetRefresh(code)
	tokenCode.SetRefreshCreateAt(time.Now())
	tokenCode.SetRefreshExpiresIn(time.Minute)
	require.NoError(t, store.Create(tokenCode))

	token, err := store.GetByRefresh(code)
	require.NoError(t, err)
	assert.Equal(t, code, token.GetRefresh())

	require.NoError(t, store.RemoveByRefresh(code))

	_, err = store.GetByRefresh(code)
	assert.Equal(t, pg.ErrNoRows, err)
}

func runClientStoreTest(t *testing.T, store *pg.ClientStore) {
	originalClient := &models.Client{
		ID:     fmt.Sprintf("id %s", time.Now().String()),
		Secret: fmt.Sprintf("secret %s", time.Now().String()),
		Domain: fmt.Sprintf("domain %s", time.Now().String()),
		UserID: fmt.Sprintf("user id %s", time.Now().String()),
	}

	require.NoError(t, store.Create(originalClient))

	client, err := store.GetByID(originalClient.GetID())
	require.NoError(t, err)
	assert.Equal(t, originalClient.GetID(), client.GetID())
	assert.Equal(t, originalClient.GetSecret(), client.GetSecret())
	assert.Equal(t, originalClient.GetDomain(), client.GetDomain())
	assert.Equal(t, originalClient.GetUserID(), client.GetUserID())
}
