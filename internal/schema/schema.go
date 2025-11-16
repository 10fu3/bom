package schema

import "strings"

// IR represents the fully normalized schema extracted from DDL + config.
type IR struct {
	Tables      []Table
	TableByName map[string]*Table
}

// Table describes a SQL table and upfront metadata needed by planner/codegen.
type Table struct {
	Name        string
	Schema      string
	Columns     []Column
	PrimaryKey  []string
	Uniques     []Unique
	Indexes     []Index
	ForeignKeys []FK
	Relations   []Relation
	IsJoinTable bool
	Comment     *string
}

// Column carries column-level metadata.
type Column struct {
	Name     string
	DBType   string
	GoType   string
	Nullable bool
	Default  *string
	Comment  *string
	AutoIncrement bool
	Identity      string
}

// Unique represents a unique or primary key constraint.
type Unique struct {
	Name string
	Cols []string
}

// Index describes a non-unique or unique index.
type Index struct {
	Name    string
	Cols    []string
	Unique  bool
	Comment *string
}

// FK describes a foreign key constraint.
type FK struct {
	Name       string
	Local      []string
	RefTable   string
	RefColumns []string
	OnDelete   string
	OnUpdate   string
}

// Relation links tables after association resolution.
type Relation struct {
	Name        string
	Kind        string
	To          string
	LocalKeys   []string
	ForeignKeys []string
	IsGenerated bool
	Comment     *string
}

// New returns an empty IR ready for population.
func New() IR {
	return IR{Tables: nil, TableByName: make(map[string]*Table)}
}

// AddTable safely appends a table and indexes it by name.
func (ir *IR) AddTable(t Table) {
	if ir.TableByName == nil {
		ir.TableByName = make(map[string]*Table)
	}
	key := tableKey(t.Schema, t.Name)
	ir.Tables = append(ir.Tables, t)
	ir.TableByName[key] = &ir.Tables[len(ir.Tables)-1]
}

// Clone returns a deep copy suitable for incremental transformations.
func (ir IR) Clone() IR {
	out := IR{TableByName: make(map[string]*Table)}
	if len(ir.Tables) == 0 {
		return out
	}
	out.Tables = make([]Table, len(ir.Tables))
	copy(out.Tables, ir.Tables)
	for i := range out.Tables {
		t := &out.Tables[i]
		out.TableByName[tableKey(t.Schema, t.Name)] = t
	}
	return out
}

func (ir IR) Table(schemaName, tableName string) *Table {
	return ir.TableByName[tableKey(schemaName, tableName)]
}

func tableKey(schemaName, tableName string) string {
	if schemaName != "" {
		return strings.ToLower(schemaName) + "." + strings.ToLower(tableName)
	}
	return strings.ToLower(tableName)
}
