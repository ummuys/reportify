package config

type RepDBConfig struct {
	Addr                  string
	MinConn               int32
	MaxConn               int32
	MaxConnLifetime       int // SECOND
	MaxConnLifetimeJitter int // SECOND
	MaxConnIdleTime       int // SECOND
	HealthCheckPeriod     int // SECOND
}
