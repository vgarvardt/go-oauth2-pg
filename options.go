package pg

type Option func(s *Store)

func WithTableName(tableName string) Option {
	return func(s *Store) {
		s.tableName = tableName
	}
}

func WithGCInterval(gcInterval int) Option {
	return func(s *Store) {
		s.gcInterval = gcInterval
	}
}

func WithLogger(logger Logger) Option {
	return func(s *Store) {
		s.logger = logger
	}
}
