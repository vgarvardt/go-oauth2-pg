package pg

// Logger is the PostgreSQL store logger interface
type Logger interface {
	Printf(format string, v ...interface{})
}
