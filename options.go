package pg

import "time"

// Option is teh configuration options type for store
type Option func(s *Store)

// WithTableName returns option that sets store table name
func WithTableName(tableName string) Option {
	return func(s *Store) {
		s.tableName = tableName
	}
}

// WithGCInterval returns option that sets store garbage collection interval
func WithGCInterval(gcInterval time.Duration) Option {
	return func(s *Store) {
		s.gcInterval = gcInterval
	}
}

// WithLogger returns option that sets store logger implementation
func WithLogger(logger Logger) Option {
	return func(s *Store) {
		s.logger = logger
	}
}

// WithGCDisabled returns option that disables store garbage collection
func WithGCDisabled() Option {
	return func(s *Store) {
		s.gcDisabled = true
	}
}

// WithInitTableDisabled returns option that disables table creation on storage instantiation
func WithInitTableDisabled() Option {
	return func(s *Store) {
		s.initTableDisabled = true
	}
}
