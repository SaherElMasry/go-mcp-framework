// framework/auth/database.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DatabaseProvider authenticates to databases
type DatabaseProvider struct {
	*BaseProvider
	connections map[string]*sql.DB
}

// DatabaseConfig holds database provider configuration
type DatabaseConfig struct {
	Driver   string `yaml:"driver" json:"driver"` // postgres, mysql, sqlite
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`

	// Connection pool settings
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// NewDatabaseProvider creates a new database provider
func NewDatabaseProvider(name string) *DatabaseProvider {
	return &DatabaseProvider{
		BaseProvider: NewBaseProvider(name),
		connections:  make(map[string]*sql.DB),
	}
}

// GetResource returns a database connection
func (p *DatabaseProvider) GetResource(ctx context.Context, resourceID string) (Resource, error) {
	// Check if we already have a connection
	if db, exists := p.connections[resourceID]; exists {
		// Ping to verify connection is alive
		if err := db.PingContext(ctx); err == nil {
			return &DatabaseResource{db: db, resourceID: resourceID}, nil
		}
		// Connection is dead, close and remove it
		db.Close()
		delete(p.connections, resourceID)
	}

	// Get resource config
	config, err := p.GetResourceConfig(resourceID)
	if err != nil {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Parse database config
	dbConfig, err := parseDatabaseConfig(config.Config)
	if err != nil {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Build connection string based on driver
	connStr, err := buildConnectionString(dbConfig)
	if err != nil {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Open database connection
	db, err := sql.Open(dbConfig.Driver, connStr)
	if err != nil {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Store connection for reuse
	p.connections[resourceID] = db

	return &DatabaseResource{db: db, resourceID: resourceID}, nil
}

// Validate checks if we can connect to the database
func (p *DatabaseProvider) Validate(ctx context.Context) error {
	// Try to get at least one resource to validate credentials
	for _, resourceID := range p.ListResources() {
		_, err := p.GetResource(ctx, resourceID)
		return err // Return after first attempt
	}
	return nil
}

// Close closes all database connections
func (p *DatabaseProvider) Close() error {
	var errs []error
	for id, db := range p.connections {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("resource %q: %w", id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}

	return nil
}

// DatabaseResource wraps a database connection
type DatabaseResource struct {
	db         *sql.DB
	resourceID string
}

func (r *DatabaseResource) Close() error {
	// Don't close here - provider manages connection lifecycle
	return nil
}

func (r *DatabaseResource) Type() string {
	return "database"
}

// DB returns the sql.DB instance
func (r *DatabaseResource) DB() *sql.DB {
	return r.db
}

// Helper functions

func parseDatabaseConfig(config map[string]interface{}) (*DatabaseConfig, error) {
	// TODO: Implement proper parsing with validation
	// For now, return a basic config
	return &DatabaseConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}, nil
}

func buildConnectionString(config *DatabaseConfig) (string, error) {
	switch config.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.Username, config.Password, config.Database), nil
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.Username, config.Password, config.Host, config.Port, config.Database), nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", config.Driver)
	}
}
