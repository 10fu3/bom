package assoc

import (
	"fmt"

	"bom/internal/config"
	"bom/internal/schema"
)

// Resolve merges user-supplied associations into the schema IR.
func Resolve(ir schema.IR, cfg config.Config) (schema.IR, error) {
	out := ir.Clone()

	for tableName, entries := range cfg.Associations {
		table := out.TableByName[tableName]
		if table == nil {
			return out, fmt.Errorf("association: unknown table %q", tableName)
		}
		if len(entries) == 0 {
			continue
		}
		for _, entry := range entries {
			if entry.Type == "" || entry.To == "" {
				continue
			}
			table.Relations = append(table.Relations, schema.Relation{
				Name:        entry.To,
				Kind:        entry.Type,
				To:          entry.To,
				LocalKeys:   entry.LocalKeys,
				ForeignKeys: entry.ForeignKeys,
				IsGenerated: true,
			})
		}
	}
	return out, nil
}
