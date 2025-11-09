package assoc

import (
	"testing"

	"bom/internal/config"
	"bom/internal/schema"
)

func TestResolve(t *testing.T) {
	ir := schema.New()
	ir.AddTable(schema.Table{Name: "video"})
	cfg := config.Config{
		Associations: map[string][]config.Association{
			"video": {{
				Type:        "HasMany",
				To:          "video_actor",
				LocalKeys:   []string{"id"},
				ForeignKeys: []string{"video_id"},
			}},
		},
	}

	out, err := Resolve(ir, cfg)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if len(out.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(out.Tables))
	}
	table := out.Tables[0]
	if len(table.Relations) != 1 {
		t.Fatalf("expected relation, got %d", len(table.Relations))
	}
	if table.Relations[0].Kind != "HasMany" {
		t.Fatalf("expected HasMany, got %s", table.Relations[0].Kind)
	}
}
