package assoc

import (
	"fmt"
	"strings"

	"bom/internal/config"
	"bom/internal/schema"
)

// Resolve merges inferred + user-supplied associations into the schema IR.
func Resolve(ir schema.IR, cfg config.Config) (schema.IR, error) {
	out := ir.Clone()

	inferRelations(&out)

	for tableName, entries := range cfg.Associations {
		table := lookupTable(&out, tableName)
		if table == nil {
			return out, fmt.Errorf("association: unknown table %q", tableName)
		}
		for _, entry := range entries {
			if entry.Type == "" || entry.To == "" {
				continue
			}
			manual := schema.Relation{
				Name:        entry.To,
				Kind:        entry.Type,
				To:          entry.To,
				LocalKeys:   append([]string(nil), entry.LocalKeys...),
				ForeignKeys: append([]string(nil), entry.ForeignKeys...),
				IsGenerated: true,
			}
			upsertRelation(table, manual, false)
		}
	}
	return out, nil
}

func inferRelations(ir *schema.IR) {
	for i := range ir.Tables {
		table := &ir.Tables[i]
		for _, fk := range table.ForeignKeys {
			target := lookupTable(ir, fk.RefTable)
			if target == nil {
				continue
			}
			if len(fk.Local) == 0 {
				continue
			}
			targetCols := fk.RefColumns
			if len(targetCols) == 0 {
				targetCols = target.PrimaryKey
			}
			if len(targetCols) == 0 || len(targetCols) != len(fk.Local) {
				continue
			}
			childRelation := schema.Relation{
				Name:        relationName(table, target.Name, fk.Local),
				Kind:        "BelongsTo",
				To:          target.Name,
				LocalKeys:   copyStrings(fk.Local),
				ForeignKeys: copyStrings(targetCols),
				IsGenerated: true,
			}
			upsertRelation(table, childRelation, true)

			kind := "HasMany"
			if isUniqueSet(table, fk.Local) {
				kind = "HasOne"
			}
			parentRelation := schema.Relation{
				Name:        relationName(target, table.Name, targetCols),
				Kind:        kind,
				To:          table.Name,
				LocalKeys:   copyStrings(targetCols),
				ForeignKeys: copyStrings(fk.Local),
				IsGenerated: true,
			}
			upsertRelation(target, parentRelation, true)
		}
	}
}

func lookupTable(ir *schema.IR, name string) *schema.Table {
	if name == "" {
		return nil
	}
	schemaName, tableName := splitSchemaTable(name)
	if t := ir.Table(schemaName, tableName); t != nil {
		return t
	}
	for i := range ir.Tables {
		if strings.EqualFold(ir.Tables[i].Name, tableName) {
			return &ir.Tables[i]
		}
	}
	return nil
}

func splitSchemaTable(name string) (string, string) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", strings.TrimSpace(name)
}

func relationName(table *schema.Table, preferred string, cols []string) string {
	candidates := []string{}
	if preferred != "" {
		candidates = append(candidates, preferred)
		if len(cols) > 0 {
			candidates = append(candidates, preferred+"_"+strings.Join(cols, "_"))
		}
	}
	if len(cols) > 0 {
		candidates = append(candidates, strings.Join(cols, "_"))
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if !relationNameExists(table, c) {
			return c
		}
	}
	base := preferred
	if base == "" {
		if len(cols) > 0 {
			base = cols[0]
		} else {
			base = fmt.Sprintf("rel_%d", len(table.Relations)+1)
		}
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if !relationNameExists(table, candidate) {
			return candidate
		}
	}
}

func relationNameExists(table *schema.Table, name string) bool {
	lower := strings.ToLower(name)
	for _, rel := range table.Relations {
		if strings.ToLower(rel.Name) == lower {
			return true
		}
	}
	return false
}

func upsertRelation(table *schema.Table, rel schema.Relation, generated bool) {
	for i := range table.Relations {
		if sameRelation(table.Relations[i], rel) {
			table.Relations[i] = rel
			table.Relations[i].IsGenerated = generated
			return
		}
	}
	rel.IsGenerated = generated
	table.Relations = append(table.Relations, rel)
}

func sameRelation(a, b schema.Relation) bool {
	return strings.EqualFold(a.To, b.To) &&
		equalStrings(a.LocalKeys, b.LocalKeys) &&
		equalStrings(a.ForeignKeys, b.ForeignKeys)
}

func equalStrings(a, b []string) bool {
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

func copyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func isUniqueSet(table *schema.Table, cols []string) bool {
	if matchesColumns(table.PrimaryKey, cols) {
		return true
	}
	for _, uniq := range table.Uniques {
		if matchesColumns(uniq.Cols, cols) {
			return true
		}
	}
	return false
}

func matchesColumns(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
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
