package postgres

//go:generate go run golang.org/x/tools/cmd/goyacc@v0.38.0 -o ddl_gen.go -p postgres ddl.y

import (
	"context"
	"fmt"
	"strings"

	"bom/internal/schema"
)

// Parser implements schema parsing for PostgreSQL DDL using a goyacc-generated grammar.
type Parser struct{}

// New returns a PostgreSQL DDL parser.
func New() *Parser { return &Parser{} }

// Dialect identifies the parser dialect.
func (*Parser) Dialect() string { return "postgres" }

// Parse consumes PostgreSQL-compatible DDL and produces the intermediate representation.
func (p *Parser) Parse(_ context.Context, ddl string) (schema.IR, error) {
	ir := schema.New()
	lex := newLexer(ddl, &ir)
	if code := postgresParse(lex); code != 0 && lex.err == nil {
		lex.err = fmt.Errorf("parse failed with code %d", code)
	}
	if lex.err != nil {
		return schema.IR{}, lex.err
	}
	return ir, nil
}

type tableBuilder struct {
	schema      string
	name        string
	columns     []*columnDef
	primaryKey  []string
	uniques     []schema.Unique
	foreignKeys []schema.FK
	comment     *string
}

type columnDef struct {
	name        string
	dbType      string
	notNull     bool
	primaryKey  bool
	unique      bool
	defaultExpr *string
	fk          *fkTarget
	comment     *string
}

type fkTarget struct {
	table   qualifiedName
	columns []string
	actions fkActions
}

type fkActions struct {
	onDelete string
	onUpdate string
}

type constraintKind int

const (
	constraintPrimaryKey constraintKind = iota
	constraintUnique
	constraintForeignKey
)

type constraintDef struct {
	kind    constraintKind
	name    string
	columns []string
	fk      fkTarget
}

type columnConstraintKind int

const (
	columnConstraintNotNull columnConstraintKind = iota
	columnConstraintNull
	columnConstraintPrimaryKey
	columnConstraintUnique
	columnConstraintDefault
	columnConstraintForeignKey
	columnConstraintNoop
)

type columnConstraint struct {
	kind  columnConstraintKind
	value string
	fk    *fkTarget
}

type tableElement struct {
	column     *columnDef
	constraint *constraintDef
}

type qualifiedName struct {
	schema string
	name   string
}

func (q qualifiedName) String() string {
	if q.schema != "" {
		return q.schema + "." + q.name
	}
	return q.name
}

func newTableBuilder(schemaName, tableName string) *tableBuilder {
	return &tableBuilder{schema: schemaName, name: tableName}
}

func (tb *tableBuilder) addColumn(col *columnDef) {
	if col == nil {
		return
	}
	tb.columns = append(tb.columns, col)
}

func (tb *tableBuilder) applyConstraint(cons *constraintDef) {
	if cons == nil {
		return
	}
	switch cons.kind {
	case constraintPrimaryKey:
		tb.primaryKey = append([]string{}, cons.columns...)
	case constraintUnique:
		tb.uniques = appendUnique(tb.uniques, cons.name, cons.columns)
	case constraintForeignKey:
		tb.foreignKeys = append(tb.foreignKeys, schema.FK{
			Name:       cons.name,
			Local:      append([]string{}, cons.columns...),
			RefTable:   cons.fk.table.String(),
			RefColumns: append([]string{}, cons.fk.columns...),
			OnDelete:   cons.fk.actions.onDelete,
			OnUpdate:   cons.fk.actions.onUpdate,
		})
	}
}

func (tb *tableBuilder) finalize() schema.Table {
	var result schema.Table
	result.Schema = tb.schema
	result.Name = tb.name
	result.Comment = tb.comment
	for _, col := range tb.columns {
		if col == nil {
			continue
		}
		result.Columns = append(result.Columns, schema.Column{
			Name:     col.name,
			DBType:   normalizeType(col.dbType),
			GoType:   mapType(col.dbType),
			Nullable: !col.notNull,
			Default:  cloneString(col.defaultExpr),
			Comment:  cloneString(col.comment),
		})
		if col.primaryKey {
			result.PrimaryKey = append(result.PrimaryKey, col.name)
		}
		if col.unique {
			result.Uniques = appendUnique(result.Uniques, "", []string{col.name})
		}
		if col.fk != nil {
			result.ForeignKeys = append(result.ForeignKeys, schema.FK{
				Local:      []string{col.name},
				RefTable:   col.fk.table.String(),
				RefColumns: append([]string{}, col.fk.columns...),
				OnDelete:   col.fk.actions.onDelete,
				OnUpdate:   col.fk.actions.onUpdate,
			})
		}
	}
	if len(tb.primaryKey) > 0 {
		result.PrimaryKey = append([]string{}, tb.primaryKey...)
	}
	if len(tb.uniques) > 0 {
		result.Uniques = append(result.Uniques, tb.uniques...)
	}
	if len(tb.foreignKeys) > 0 {
		result.ForeignKeys = append(result.ForeignKeys, tb.foreignKeys...)
	}
	return result
}

func appendUnique(existing []schema.Unique, name string, cols []string) []schema.Unique {
	if len(cols) == 0 {
		return existing
	}
	for i := range existing {
		if sameColumns(existing[i].Cols, cols) {
			return existing
		}
	}
	name = strings.TrimSpace(name)
	existing = append(existing, schema.Unique{
		Name: name,
		Cols: append([]string{}, cols...),
	})
	return existing
}

func sameColumns(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !strings.EqualFold(a[i], b[i]) {
			return false
		}
	}
	return true
}

func cloneString(in *string) *string {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func normalizeType(dbType string) string {
	return strings.ToUpper(strings.TrimSpace(dbType))
}

func mapType(dbType string) string {
	upper := strings.ToUpper(dbType)
	switch {
	case strings.Contains(upper, "SERIAL"), strings.Contains(upper, "BIGINT"), strings.Contains(upper, "SMALLINT"), strings.Contains(upper, "INTEGER"), strings.Contains(upper, "INT"):
		return "int64"
	case strings.Contains(upper, "UUID"):
		return "string"
	case strings.Contains(upper, "BOOL"):
		return "bool"
	case strings.Contains(upper, "JSON"):
		return "string"
	case strings.Contains(upper, "CHAR"), strings.Contains(upper, "TEXT"), strings.Contains(upper, "CITEXT"), strings.Contains(upper, "XML"):
		return "string"
	case strings.Contains(upper, "BYTEA"):
		return "[]byte"
	case strings.Contains(upper, "DOUBLE"), strings.Contains(upper, "REAL"), strings.Contains(upper, "NUMERIC"), strings.Contains(upper, "DEC"):
		return "float64"
	case strings.Contains(upper, "DATE"), strings.Contains(upper, "TIME"):
		return "string"
	default:
		return "string"
	}
}

func applyColumnConstraints(col *columnDef, specs []columnConstraint) {
	for _, spec := range specs {
		switch spec.kind {
		case columnConstraintNotNull:
			col.notNull = true
		case columnConstraintNull:
			col.notNull = false
		case columnConstraintPrimaryKey:
			col.primaryKey = true
			col.notNull = true
		case columnConstraintUnique:
			col.unique = true
		case columnConstraintDefault:
			val := spec.value
			col.defaultExpr = &val
		case columnConstraintForeignKey:
			if spec.fk != nil {
				col.fk = spec.fk
			}
		case columnConstraintNoop:
			continue
		}
	}
}

type indexDef struct {
	schema  string
	table   string
	name    string
	columns []string
	unique  bool
}
