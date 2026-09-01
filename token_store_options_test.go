package pg

import (
	"bytes"
	"log/slog"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTokenStoreGCDisabled(t *testing.T) {
	store, err := NewTokenStore(t.Context(), nil, WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)
	assert.True(t, store.gcDisabled)
	assert.True(t, store.initTableDisabled)
}

func TestWithTokenStoreTableName(t *testing.T) {
	randomName := time.Now().String()

	store, err := NewTokenStore(t.Context(), nil, WithTokenStoreTableName(randomName), WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)
	assert.Equal(t, randomName, store.tableName)
}

func TestWithTokenStoreGCInterval(t *testing.T) {
	randomInterval := time.Duration(rand.Int63())

	store, err := NewTokenStore(t.Context(), nil, WithTokenStoreGCInterval(randomInterval), WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)
	assert.Equal(t, randomInterval, store.gcInterval)
}

func TestWithTokenStoreLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	l := slog.New(slog.NewTextHandler(buf, nil))

	store, err := NewTokenStore(t.Context(), nil, WithTokenStoreLogger(l), WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)

	store.logger.Info("log1", slog.Int("int", 1), slog.String("string", "hello"))
	store.logger.Error("log2", slog.Int("int", 12), slog.String("string", "22"))

	logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, logs, 2)

	assert.Contains(t, logs[0], "level=INFO")
	assert.Contains(t, logs[0], "msg=log1")
	assert.Contains(t, logs[0], "int=1")
	assert.Contains(t, logs[0], "string=hello")

	assert.Contains(t, logs[1], "level=ERROR")
	assert.Contains(t, logs[1], "msg=log2")
	assert.Contains(t, logs[1], "int=12")
	assert.Contains(t, logs[1], "string=22")
}
