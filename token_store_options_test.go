package pg

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTokenStoreGCDisabled(t *testing.T) {
	store, err := NewTokenStore(nil, WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)
	assert.True(t, store.gcDisabled)
	assert.True(t, store.initTableDisabled)
}

func TestWithTokenStoreTableName(t *testing.T) {
	randomName := time.Now().String()

	store, err := NewTokenStore(nil, WithTokenStoreTableName(randomName), WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)
	assert.Equal(t, randomName, store.tableName)
}

func TestWithTokenStoreGCInterval(t *testing.T) {
	randomInterval := time.Duration(rand.Int63())

	store, err := NewTokenStore(nil, WithTokenStoreGCInterval(randomInterval), WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)
	assert.Equal(t, randomInterval, store.gcInterval)
}

func TestWithTokenStoreLogger(t *testing.T) {
	l := new(memoryLogger)

	store, err := NewTokenStore(nil, WithTokenStoreLogger(l), WithTokenStoreGCDisabled(), WithTokenStoreInitTableDisabled())
	require.NoError(t, err)

	store.logger.Printf("log1", 1, "2", "333")
	store.logger.Printf("log2", 12, "22")

	require.Equal(t, 2, len(l.formats))
	require.Equal(t, 2, len(l.args))

	assert.Equal(t, "log1", l.formats[0])
	assert.Equal(t, "log2", l.formats[1])

	require.Equal(t, 3, len(l.args[0]))
	require.Equal(t, 2, len(l.args[1]))

	assert.Equal(t, 1, l.args[0][0])
	assert.Equal(t, "2", l.args[0][1])
	assert.Equal(t, "333", l.args[0][2])

	assert.Equal(t, 12, l.args[1][0])
	assert.Equal(t, "22", l.args[1][1])
}
