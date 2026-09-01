package pg

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-oauth2/oauth2/v4/models"
	_ "github.com/lib/pq" // register driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTokenStore_initTable(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		mock.ExpectClose()
		assert.NoError(t, mockDB.Close())
	})

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 1))

	store, err := NewTokenStore(t.Context(), mockDB, WithTokenStoreGCDisabled())
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})
}

func TestTokenStore_gc(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		mock.ExpectClose()
		assert.NoError(t, mockDB.Close())
	})

	mock.ExpectExec("DELETE FROM").WillReturnResult(sqlmock.NewResult(0, 1))

	store, err := NewTokenStore(t.Context(), mockDB, WithTokenStoreInitTableDisabled(), WithTokenStoreGCInterval(time.Second))
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	time.Sleep(5 * time.Second)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func generateTokenTableName() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}

func generateClientTableName() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func TestSQL(t *testing.T) {
	conn, err := sql.Open("postgres", uri)
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, conn.Close())
	})

	tokenStore, err := NewTokenStore(
		t.Context(),
		conn,
		WithTokenStoreTableName(generateTokenTableName()),
		WithTokenStoreGCInterval(time.Second),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, tokenStore.Close())
	})

	clientStore, err := NewClientStore(
		t.Context(),
		conn,
		WithClientStoreTableName(generateClientTableName()),
	)
	require.NoError(t, err)

	runTokenStoreTest(t, tokenStore)
	runClientStoreTest(t, clientStore)
}

func runTokenStoreTest(t *testing.T, store *TokenStore) {
	runTokenStoreCodeTest(t, store)
	runTokenStoreAccessTest(t, store)
	runTokenStoreRefreshTest(t, store)

	// sleep for a while just to wait for GC run for sure to ensure there were no errors there
	time.Sleep(3 * time.Second)
}

func runTokenStoreCodeTest(t *testing.T, store *TokenStore) {
	code := fmt.Sprintf("code %s", time.Now().String())
	ctx := t.Context()

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
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func runTokenStoreAccessTest(t *testing.T, store *TokenStore) {
	code := fmt.Sprintf("access %s", time.Now().String())
	ctx := t.Context()

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
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func runTokenStoreRefreshTest(t *testing.T, store *TokenStore) {
	code := fmt.Sprintf("refresh %s", time.Now().String())
	ctx := t.Context()

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
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func runClientStoreTest(t *testing.T, store *ClientStore) {
	originalClient := &models.Client{
		ID:     fmt.Sprintf("id %s", time.Now().String()),
		Secret: fmt.Sprintf("secret %s", time.Now().String()),
		Domain: fmt.Sprintf("domain %s", time.Now().String()),
		UserID: fmt.Sprintf("user id %s", time.Now().String()),
	}
	ctx := context.Background()

	require.NoError(t, store.Create(t.Context(), originalClient))

	client, err := store.GetByID(ctx, originalClient.GetID())
	require.NoError(t, err)
	assert.Equal(t, originalClient.GetID(), client.GetID())
	assert.Equal(t, originalClient.GetSecret(), client.GetSecret())
	assert.Equal(t, originalClient.GetDomain(), client.GetDomain())
	assert.Equal(t, originalClient.GetUserID(), client.GetUserID())
}
