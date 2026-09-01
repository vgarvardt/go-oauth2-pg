package pg

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithClientStoreInitTableDisabled(t *testing.T) {
	store, err := NewClientStore(nil, WithClientStoreInitTableDisabled())
	require.NoError(t, err)
	assert.True(t, store.initTableDisabled)
}

func TestWithClientStoreTableName(t *testing.T) {
	randomName := time.Now().String()

	store, err := NewClientStore(nil, WithClientStoreTableName(randomName), WithClientStoreInitTableDisabled())
	require.NoError(t, err)
	assert.Equal(t, randomName, store.tableName)
}

func TestWithClientStoreLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	l := slog.New(slog.NewTextHandler(buf, nil))

	store, err := NewClientStore(nil, WithClientStoreLogger(l), WithClientStoreInitTableDisabled())
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
