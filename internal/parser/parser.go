package parser

import (
	"context"

	"bom/internal/schema"
)

// DDLParser parses dialect-specific DDL into the intermediate representation.
type DDLParser interface {
	Parse(ctx context.Context, ddl string) (schema.IR, error)
	Dialect() string
}
