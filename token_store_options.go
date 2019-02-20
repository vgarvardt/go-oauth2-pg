package pg

import "time"

// TokenStoreOption is the configuration options type for token store
type TokenStoreOption func(s *TokenStore)

// WithTokenStoreTableName returns option that sets token store table name
func WithTokenStoreTableName(tableName string) TokenStoreOption {
	return func(s *TokenStore) {
		s.tableName = tableName
	}
}

// WithTokenStoreGCInterval returns option that sets token store garbage collection interval
func WithTokenStoreGCInterval(gcInterval time.Duration) TokenStoreOption {
	return func(s *TokenStore) {
		s.gcInterval = gcInterval
	}
}

// WithTokenStoreLogger returns option that sets token store logger implementation
func WithTokenStoreLogger(logger Logger) TokenStoreOption {
	return func(s *TokenStore) {
		s.logger = logger
	}
}

// WithTokenStoreGCDisabled returns option that disables token store garbage collection
func WithTokenStoreGCDisabled() TokenStoreOption {
	return func(s *TokenStore) {
		s.gcDisabled = true
	}
}

// WithTokenStoreInitTableDisabled returns option that disables table creation on token store instantiation
func WithTokenStoreInitTableDisabled() TokenStoreOption {
	return func(s *TokenStore) {
		s.initTableDisabled = true
	}
}
