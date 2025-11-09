package bom

import (
	"context"
	"database/sql"
)

// Querier is the minimal subset of *sql.DB used by the generated query helpers.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (RowsPublic, error)
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
}

type RowsPublic interface {
	Close() error
	ColumnTypes() ([]*sql.ColumnType, error)
	Columns() ([]string, error)
	Err() error
	Next() bool
	NextResultSet() bool
	Scan(dest ...any) error
}

type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// BuildOpt mutates BuildOptions to tweak how a query is planned.
type BuildOpt func(*BuildOptions)

// BuildOptions is a generic key/value bag propagated to the planner.
// Keys are unconstrained so downstream packages can experiment without
// touching the public surface again.
type BuildOptions struct {
	data map[string]any
}

// ApplyBuildOpts materializes the aggregate BuildOptions for a build call.
func ApplyBuildOpts(opts ...BuildOpt) BuildOptions {
	var cfg BuildOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// Set stores an arbitrary value within the options bag.
func (o *BuildOptions) Set(key string, value any) {
	if o.data == nil {
		o.data = make(map[string]any)
	}
	o.data[key] = value
}

// Get retrieves a value previously stored via Set.
func (o BuildOptions) Get(key string) (any, bool) {
	if o.data == nil {
		return nil, false
	}
	v, ok := o.data[key]
	return v, ok
}
