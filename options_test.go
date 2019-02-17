package pg

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithGCDisabled(t *testing.T) {
	store := NewStore(nil, WithGCDisabled())
	assert.True(t, store.gcDisabled)
}

func TestWithTableName(t *testing.T) {
	randomName := time.Now().String()

	store := NewStore(nil, WithTableName(randomName), WithGCDisabled())
	assert.Equal(t, randomName, store.tableName)
}

func TestWithGCInterval(t *testing.T) {
	randomInterval := rand.Int()

	store := NewStore(nil, WithGCInterval(randomInterval), WithGCDisabled())
	assert.Equal(t, randomInterval, store.gcInterval)
}

type memoryLogger struct {
	formats []string
	args    [][]interface{}
}

func (l *memoryLogger) Printf(format string, v ...interface{}) {
	l.formats = append(l.formats, format)
	l.args = append(l.args, v)
}

func TestWithLogger(t *testing.T) {
	l := new(memoryLogger)

	store := NewStore(nil, WithLogger(l), WithGCDisabled())

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
