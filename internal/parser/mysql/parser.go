package mysql

import (
	"context"
	"strings"

	mysqlparser "github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
	_ "github.com/pingcap/tidb/parser/test_driver"

	"bom/internal/schema"
)

type Parser struct{}

func New() *Parser { return &Parser{} }

func (*Parser) Dialect() string { return "mysql" }

func (p *Parser) Parse(_ context.Context, ddl string) (schema.IR, error) {
	out := schema.New()
	pr := mysqlparser.New()
	stmts, _, err := pr.Parse(ddl, "", "")
	if err != nil {
		return out, err
	}

	table := func(schemaName, tableName string) *schema.Table {
		if t := out.Table(schemaName, tableName); t != nil {
			return t
		}
		out.AddTable(schema.Table{Schema: schemaName, Name: tableName})
		return out.Table(schemaName, tableName)
	}

	for _, stmt := range stmts {
		switch n := stmt.(type) {
		case *ast.CreateTableStmt:
			t := table(n.Table.Schema.O, n.Table.Name.O)
			applyCreateTable(t, n)
		case *ast.AlterTableStmt:
			t := table(n.Table.Schema.O, n.Table.Name.O)
			applyAlterTable(&out, t, n)
		case *ast.CreateIndexStmt:
			t := table(n.Table.Schema.O, n.Table.Name.O)
			applyCreateIndex(t, n)
		}
	}

	return out, nil
}

func applyCreateTable(t *schema.Table, ct *ast.CreateTableStmt) {
	t.Schema = ct.Table.Schema.O
	t.Name = ct.Table.Name.O
	t.Columns = nil
	t.PrimaryKey = nil
	t.Uniques = nil
	t.Indexes = nil
	t.ForeignKeys = nil
	t.Comment = tableComment(ct.Options)

	for _, colDef := range ct.Cols {
		upsertColumn(t, buildColumn(colDef))
		applyInlineColumnOptions(t, colDef)
	}
	for _, cons := range ct.Constraints {
		handleConstraint(t, cons)
	}
}

func applyAlterTable(ir *schema.IR, t *schema.Table, at *ast.AlterTableStmt) {
	if t == nil {
		return
	}
	for _, spec := range at.Specs {
		switch spec.Tp {
		case ast.AlterTableAddColumns:
			for _, col := range spec.NewColumns {
				upsertColumn(t, buildColumn(col))
				applyInlineColumnOptions(t, col)
			}
		case ast.AlterTableAddConstraint:
			handleConstraint(t, spec.Constraint)
		case ast.AlterTableDropColumn:
			if spec.OldColumnName != nil {
				dropColumn(t, spec.OldColumnName.Name.O)
			}
		case ast.AlterTableDropPrimaryKey:
			t.PrimaryKey = nil
		case ast.AlterTableDropIndex:
			dropIndexOrUnique(t, spec.Name)
		case ast.AlterTableDropForeignKey:
			dropForeignKey(t, spec.Name)
		case ast.AlterTableModifyColumn:
			for _, col := range spec.NewColumns {
				updateColumn(t, col.Name.Name.O, col)
				applyInlineColumnOptions(t, col)
			}
		case ast.AlterTableChangeColumn:
			if len(spec.NewColumns) == 0 {
				continue
			}
			oldName := ""
			if spec.OldColumnName != nil {
				oldName = spec.OldColumnName.Name.O
			}
			updateColumnWithRename(t, oldName, spec.NewColumns[0])
			applyInlineColumnOptions(t, spec.NewColumns[0])
		case ast.AlterTableRenameColumn:
			if spec.OldColumnName != nil && spec.NewColumnName != nil {
				renameColumn(t, spec.OldColumnName.Name.O, spec.NewColumnName.Name.O)
			}
		case ast.AlterTableRenameTable:
			if spec.NewTable != nil {
				renameTable(ir, t, spec.NewTable.Schema.O, spec.NewTable.Name.O)
			}
		case ast.AlterTableOption:
			applyTableOptions(t, spec.Options)
		}
	}
}

func applyCreateIndex(t *schema.Table, stmt *ast.CreateIndexStmt) {
	if t == nil {
		return
	}
	cols := colsFromIndexParts(stmt.IndexPartSpecifications)
	if len(cols) == 0 {
		return
	}
	if stmt.KeyType == ast.IndexKeyTypeUnique {
		t.Uniques = append(t.Uniques, schema.Unique{
			Name: stmt.IndexName,
			Cols: cols,
		})
		return
	}
	t.Indexes = append(t.Indexes, schema.Index{
		Name:    stmt.IndexName,
		Cols:    cols,
		Unique:  false,
		Comment: indexComment(stmt.IndexOption),
	})
}

func handleConstraint(t *schema.Table, cons *ast.Constraint) {
	if t == nil || cons == nil {
		return
	}
	switch cons.Tp {
	case ast.ConstraintPrimaryKey:
		if cols := colsFromIndexParts(cons.Keys); len(cols) > 0 {
			t.PrimaryKey = cols
		}
	case ast.ConstraintUniq, ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
		if cols := colsFromIndexParts(cons.Keys); len(cols) > 0 {
			t.Uniques = append(t.Uniques, schema.Unique{
				Name: cons.Name,
				Cols: cols,
			})
		}
	case ast.ConstraintForeignKey:
		local := colsFromIndexParts(cons.Keys)
		var refCols []string
		refTable := ""
		if cons.Refer != nil {
			refCols = colsFromIndexParts(cons.Refer.IndexPartSpecifications)
			if cons.Refer.Table != nil {
				refTable = cons.Refer.Table.Name.O
				if schemaName := cons.Refer.Table.Schema.O; schemaName != "" {
					refTable = schemaName + "." + refTable
				}
			}
		}
		t.ForeignKeys = append(t.ForeignKeys, schema.FK{
			Name:       cons.Name,
			Local:      local,
			RefTable:   refTable,
			RefColumns: refCols,
			OnDelete:   referAction(cons.Refer, true),
			OnUpdate:   referAction(cons.Refer, false),
		})
	case ast.ConstraintIndex, ast.ConstraintKey:
		if cols := colsFromIndexParts(cons.Keys); len(cols) > 0 {
			t.Indexes = append(t.Indexes, schema.Index{
				Name:    cons.Name,
				Cols:    cols,
				Unique:  false,
				Comment: indexComment(cons.Option),
			})
		}
	}
}

func referAction(ref *ast.ReferenceDef, onDelete bool) string {
	if ref == nil {
		return ""
	}
	var optStr string
	if onDelete {
		if ref.OnDelete == nil {
			return ""
		}
		optStr = ref.OnDelete.ReferOpt.String()
	} else {
		if ref.OnUpdate == nil {
			return ""
		}
		optStr = ref.OnUpdate.ReferOpt.String()
	}
	if optStr == "" {
		return ""
	}
	return strings.ToUpper(optStr)
}

func colsFromIndexParts(parts []*ast.IndexPartSpecification) []string {
	var cols []string
	for _, part := range parts {
		if part == nil || part.Column == nil {
			continue
		}
		cols = append(cols, part.Column.Name.O)
	}
	return cols
}

func buildColumn(colDef *ast.ColumnDef) schema.Column {
	col := schema.Column{
		Name:     colDef.Name.Name.O,
		DBType:   strings.ToUpper(colDef.Tp.InfoSchemaStr()),
		GoType:   mapType(colDef),
		Nullable: !hasNotNull(colDef),
		Default:  renderDefault(colDef),
		Comment:  columnComment(colDef),
	}
	return col
}

func upsertColumn(t *schema.Table, col schema.Column) {
	if t == nil {
		return
	}
	if idx := columnIndex(t.Columns, col.Name); idx >= 0 {
		t.Columns[idx] = col
		return
	}
	t.Columns = append(t.Columns, col)
}

func updateColumn(t *schema.Table, name string, colDef *ast.ColumnDef) {
	if colDef == nil {
		return
	}
	target := name
	if target == "" {
		target = colDef.Name.Name.O
	}
	col := buildColumn(colDef)
	if idx := columnIndex(t.Columns, target); idx >= 0 {
		t.Columns[idx] = col
		return
	}
	upsertColumn(t, col)
}

func updateColumnWithRename(t *schema.Table, oldName string, colDef *ast.ColumnDef) {
	if colDef == nil {
		return
	}
	col := buildColumn(colDef)
	if idx := columnIndex(t.Columns, oldName); idx >= 0 {
		t.Columns[idx] = col
		return
	}
	upsertColumn(t, col)
}

func renameColumn(t *schema.Table, oldName, newName string) {
	if oldName == "" || newName == "" {
		return
	}
	if idx := columnIndex(t.Columns, oldName); idx >= 0 {
		t.Columns[idx].Name = newName
	}
}

func dropColumn(t *schema.Table, name string) {
	if idx := columnIndex(t.Columns, name); idx >= 0 {
		t.Columns = append(t.Columns[:idx], t.Columns[idx+1:]...)
	}
}

func dropForeignKey(t *schema.Table, name string) {
	lower := strings.ToLower(name)
	for i := range t.ForeignKeys {
		if strings.ToLower(t.ForeignKeys[i].Name) == lower {
			t.ForeignKeys = append(t.ForeignKeys[:i], t.ForeignKeys[i+1:]...)
			return
		}
	}
}

func dropIndexOrUnique(t *schema.Table, name string) {
	lower := strings.ToLower(name)
	for i := range t.Indexes {
		if strings.ToLower(t.Indexes[i].Name) == lower {
			t.Indexes = append(t.Indexes[:i], t.Indexes[i+1:]...)
			return
		}
	}
	for i := range t.Uniques {
		if strings.ToLower(t.Uniques[i].Name) == lower {
			t.Uniques = append(t.Uniques[:i], t.Uniques[i+1:]...)
			return
		}
	}
}

func columnIndex(cols []schema.Column, name string) int {
	lower := strings.ToLower(name)
	for i := range cols {
		if strings.ToLower(cols[i].Name) == lower {
			return i
		}
	}
	return -1
}

func applyInlineColumnOptions(t *schema.Table, colDef *ast.ColumnDef) {
	if t == nil || colDef == nil {
		return
	}
	colName := colDef.Name.Name.O
	for _, opt := range colDef.Options {
		switch opt.Tp {
		case ast.ColumnOptionPrimaryKey:
			ensurePrimaryKeyColumn(t, colName)
		case ast.ColumnOptionUniqKey:
			ensureSingleColumnUnique(t, colName)
		}
	}
}

func ensurePrimaryKeyColumn(t *schema.Table, col string) {
	for _, existing := range t.PrimaryKey {
		if strings.EqualFold(existing, col) {
			return
		}
	}
	t.PrimaryKey = append(t.PrimaryKey, col)
}

func ensureSingleColumnUnique(t *schema.Table, col string) {
	for _, uniq := range t.Uniques {
		if len(uniq.Cols) == 1 && strings.EqualFold(uniq.Cols[0], col) {
			return
		}
	}
	t.Uniques = append(t.Uniques, schema.Unique{
		Name: col,
		Cols: []string{col},
	})
}

func hasNotNull(col *ast.ColumnDef) bool {
	for _, opt := range col.Options {
		if opt.Tp == ast.ColumnOptionNotNull {
			return true
		}
	}
	return false
}

func renderDefault(col *ast.ColumnDef) *string {
	for _, opt := range col.Options {
		if opt.Tp == ast.ColumnOptionDefaultValue && opt.Expr != nil {
			s := exprString(opt.Expr)
			if suffix := "()"; strings.HasSuffix(s, suffix) {
				s = strings.TrimSuffix(s, suffix)
			}
			return &s
		}
	}
	return nil
}

func columnComment(col *ast.ColumnDef) *string {
	for _, opt := range col.Options {
		if opt.Tp == ast.ColumnOptionComment && opt.Expr != nil {
			if ve, ok := opt.Expr.(ast.ValueExpr); ok {
				c := ve.GetString()
				return &c
			}
			c := exprString(opt.Expr)
			return &c
		}
	}
	return nil
}

func tableComment(opts []*ast.TableOption) *string {
	for _, opt := range opts {
		if opt.Tp == ast.TableOptionComment && opt.StrValue != "" {
			c := opt.StrValue
			return &c
		}
	}
	return nil
}

func applyTableOptions(t *schema.Table, opts []*ast.TableOption) {
	if comment := tableComment(opts); comment != nil {
		t.Comment = comment
	}
}

func exprString(node ast.ExprNode) string {
	var buf strings.Builder
	ctx := format.NewRestoreCtx(format.DefaultRestoreFlags, &buf)
	_ = node.Restore(ctx)
	return buf.String()
}

func indexComment(opt *ast.IndexOption) *string {
	if opt == nil || opt.Comment == "" {
		return nil
	}
	c := opt.Comment
	return &c
}

func mapType(def *ast.ColumnDef) string {
	dbType := strings.ToUpper(def.Tp.InfoSchemaStr())
	switch {
	case strings.Contains(dbType, "BIGINT") && strings.Contains(dbType, "UNSIGNED"):
		return "uint64"
	case strings.Contains(dbType, "BIGINT"):
		return "int64"
	case strings.Contains(dbType, "INT"):
		return "int32"
	case strings.Contains(dbType, "TINYINT(1)"):
		return "bool"
	case strings.Contains(dbType, "CHAR"), strings.Contains(dbType, "TEXT"):
		return "string"
	case strings.Contains(dbType, "JSON"):
		return "[]byte"
	default:
		return "string"
	}
}

func renameTable(ir *schema.IR, t *schema.Table, schemaName, tableName string) {
	if ir == nil || t == nil {
		return
	}
	oldKey := makeTableKey(t.Schema, t.Name)
	delete(ir.TableByName, oldKey)
	if schemaName != "" {
		t.Schema = schemaName
	}
	if tableName != "" {
		t.Name = tableName
	}
	ir.TableByName[makeTableKey(t.Schema, t.Name)] = t
}

func makeTableKey(schemaName, tableName string) string {
	if schemaName != "" {
		return strings.ToLower(schemaName) + "." + strings.ToLower(tableName)
	}
	return strings.ToLower(tableName)
}
