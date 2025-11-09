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
    // Close は Rows を閉じて、関連するリソースを解放します。
    Close() error

    // ColumnTypes は結果セットの各カラムの型情報を返します。
    ColumnTypes() ([]*sql.ColumnType, error)

    // Columns は結果セットのカラム名を返します。
    Columns() ([]string, error)

    // Err はイテレーション中に発生したエラーを返します。
    Err() error

    // Next は次の行を読み込める場合に true を返します。
    Next() bool

    // NextResultSet は複数の結果セットがある場合に次のセットへ進みます（Go 1.8+）。
    NextResultSet() bool

    // Scan は現在の行を指定された変数に読み込みます。
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
