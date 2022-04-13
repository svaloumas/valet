package relational

// DBOptions represents config parameters for the relational DB clients.
type DBOptions struct {
	ConnectionMaxLifetime int
	MaxIdleConnections    int
	MaxOpenConnections    int
}
