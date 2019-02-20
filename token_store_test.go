package pg

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type queryCall struct {
	query string
	args  []interface{}
}

type mockAdapter struct {
	execCalls      []queryCall
	selectOneCalls []queryCall

	execCallback   func(query string, args ...interface{}) error
	selectCallback func(dst interface{}, query string, args ...interface{}) error
}

func (a *mockAdapter) Exec(query string, args ...interface{}) error {
	a.execCalls = append(a.execCalls, queryCall{query: query, args: args})

	if a.execCallback != nil {
		return a.execCallback(query, args...)
	}

	return nil
}

func (a *mockAdapter) SelectOne(dst interface{}, query string, args ...interface{}) error {
	a.selectOneCalls = append(a.selectOneCalls, queryCall{query: query, args: args})

	if a.selectCallback != nil {
		return a.selectCallback(dst, query, args...)
	}

	return nil
}

func Test_initTable(t *testing.T) {
	adapter := new(mockAdapter)

	store, err := NewTokenStore(adapter, WithTokenStoreGCDisabled())
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, store.Close())
	}()

	assert.Equal(t, 1, len(adapter.execCalls))
	assert.Equal(t, 0, len(adapter.selectOneCalls))

	// new line character is the character at position 0
	assert.Equal(t, 1, strings.Index(adapter.execCalls[0].query, "CREATE TABLE IF NOT EXISTS"))
}

func Test_gc(t *testing.T) {
	adapter := new(mockAdapter)

	store, err := NewTokenStore(adapter, WithTokenStoreInitTableDisabled(), WithTokenStoreGCInterval(time.Second))
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, store.Close())
	}()

	time.Sleep(5 * time.Second)

	// in 5 seconds we should have 4-5 gc calls
	assert.True(t, 3 < len(adapter.execCalls))
	assert.True(t, 5 >= len(adapter.execCalls))
	assert.Equal(t, 0, len(adapter.selectOneCalls))

	for i := range adapter.execCalls {
		assert.Equal(t, 0, strings.Index(adapter.execCalls[i].query, "DELETE FROM"))
	}
}
