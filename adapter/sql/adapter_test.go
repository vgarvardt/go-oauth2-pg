package sql

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
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
}

func (l *memoryLogger) Printf(format string, v ...interface{}) {
	l.formats = append(l.formats, format)
	l.args = append(l.args, v)
}

func generateTableName() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}

func TestNewSQL(t *testing.T) {
	l := new(memoryLogger)

	conn, err := sql.Open("pgx", uri)
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, conn.Close())
	}()

	adapter := NewSQL(conn)
	tableName := generateTableName()

	store, err := pg.NewStore(adapter, pg.WithLogger(l), pg.WithTableName(tableName), pg.WithGCInterval(time.Second))
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, store.Close())
	}()

	runStoreTest(t, store, l)
}

func TestNewSQLx(t *testing.T) {
	l := new(memoryLogger)

	conn, err := sql.Open("pgx", uri)
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, conn.Close())
	}()

	adapter := NewSQLx(sqlx.NewDb(conn, ""))
	tableName := generateTableName()

	store, err := pg.NewStore(adapter, pg.WithLogger(l), pg.WithTableName(tableName), pg.WithGCInterval(time.Second))
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, store.Close())
	}()

	runStoreTest(t, store, l)
}

func runStoreTest(t *testing.T, store *pg.Store, l *memoryLogger) {
	runStoreCodeTest(t, store)
	runStoreAccessTest(t, store)
	runStoreRefreshTest(t, store)

	// sleep for a while just to wait for GC run for sure to ensure there were no errors there
	time.Sleep(3 * time.Second)

	assert.Equal(t, 0, len(l.formats))
}

func runStoreCodeTest(t *testing.T, store *pg.Store) {
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

func runStoreAccessTest(t *testing.T, store *pg.Store) {
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

func runStoreRefreshTest(t *testing.T, store *pg.Store) {
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
