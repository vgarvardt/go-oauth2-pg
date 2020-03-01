package pg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientStore_initTable(t *testing.T) {
	adapter := new(mockAdapter)

	adapter.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		query := args.Get(1).(string)
		// new line character is the character at position 0
		assert.Equal(t, 1, strings.Index(query, "CREATE TABLE IF NOT EXISTS"))
	})

	_, err := NewClientStore(adapter)
	require.NoError(t, err)

	adapter.AssertExpectations(t)
}
