package pg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientStore_initTable(t *testing.T) {
	adapter := new(mockAdapter)

	_, err := NewClientStore(adapter)
	require.NoError(t, err)

	assert.Equal(t, 1, len(adapter.execCalls))
	assert.Equal(t, 0, len(adapter.selectOneCalls))

	// new line character is the character at position 0
	assert.Equal(t, 1, strings.Index(adapter.execCalls[0].query, "CREATE TABLE IF NOT EXISTS"))
}
