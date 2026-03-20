package planner

import (
	"fmt"
	"strings"
)

// InsertInput describes an INSERT statement.
type InsertInput struct {
	Table     string
	Columns   []string
	Values    []string
	Returning []string
}

// BuildInsert renders an INSERT statement for the given input.
func BuildInsert(d Dialect, in InsertInput) (string, error) {
	if in.Table == "" {
		return "", fmt.Errorf("insert: table required")
	}
	if len(in.Columns) != len(in.Values) {
		return "", fmt.Errorf("insert: columns/values mismatch")
	}
	if len(in.Columns) == 0 && len(in.Values) > 0 {
		return "", fmt.Errorf("insert: values without columns")
	}
	quote := func(cols []string) []string {
		out := make([]string, len(cols))
		for i, col := range cols {
			out[i] = d.QuoteIdent(col)
		}
		return out
	}
	builder := strings.Builder{}
	builder.WriteString("INSERT INTO ")
	builder.WriteString(d.QuoteIdent(in.Table))
	if len(in.Columns) == 0 {
		switch strings.ToLower(d.Name()) {
		case "postgres":
			builder.WriteString(" DEFAULT VALUES")
		default:
			builder.WriteString(" () VALUES ()")
		}
	} else {
		builder.WriteString(" (")
		builder.WriteString(strings.Join(quote(in.Columns), ", "))
		builder.WriteString(") VALUES (")
		builder.WriteString(strings.Join(in.Values, ", "))
		builder.WriteString(")")
	}
	if len(in.Returning) > 0 {
		if !d.Cap().InsertReturning {
			return "", fmt.Errorf("%s: RETURNING not supported", d.Name())
		}
		builder.WriteString(" RETURNING ")
		builder.WriteString(strings.Join(quote(in.Returning), ", "))
	}
	return builder.String(), nil
}

// UpdateInput describes an UPDATE statement.
type UpdateInput struct {
	Table      string
	SetClauses []string
	Where      string
}

// BuildUpdate renders an UPDATE statement for the given input.
func BuildUpdate(d Dialect, in UpdateInput) (string, error) {
	if in.Table == "" {
		return "", fmt.Errorf("update: table required")
	}
	if len(in.SetClauses) == 0 {
		return "", fmt.Errorf("update: at least one SET clause required")
	}
	builder := strings.Builder{}
	builder.WriteString("UPDATE ")
	builder.WriteString(d.QuoteIdent(in.Table))
	builder.WriteString(" SET ")
	builder.WriteString(strings.Join(in.SetClauses, ", "))
	if in.Where != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(in.Where)
	}
	return builder.String(), nil
}

// DeleteInput describes a DELETE statement.
type DeleteInput struct {
	Table string
	Where string
}

// BuildDelete renders a DELETE statement for the given input.
func BuildDelete(d Dialect, in DeleteInput) (string, error) {
	if in.Table == "" {
		return "", fmt.Errorf("delete: table required")
	}
	builder := strings.Builder{}
	builder.WriteString("DELETE FROM ")
	builder.WriteString(d.QuoteIdent(in.Table))
	if in.Where != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(in.Where)
	}
	return builder.String(), nil
}
