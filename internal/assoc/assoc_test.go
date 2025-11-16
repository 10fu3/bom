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

func TestResolveInfersHasManyAndBelongsTo(t *testing.T) {
	ir := schema.New()
	ir.AddTable(schema.Table{
		Name: "author",
		Columns: []schema.Column{
			{Name: "id", GoType: "int64"},
			{Name: "name", GoType: "string"},
		},
		PrimaryKey: []string{"id"},
	})
	ir.AddTable(schema.Table{
		Name: "video",
		Columns: []schema.Column{
			{Name: "id", GoType: "int64"},
			{Name: "author_id", GoType: "int64"},
		},
		PrimaryKey: []string{"id"},
		ForeignKeys: []schema.FK{
			{
				Name:       "fk_video_author",
				Local:      []string{"author_id"},
				RefTable:   "author",
				RefColumns: []string{"id"},
			},
		},
	})

	out, err := Resolve(ir, config.Config{})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if len(out.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(out.Tables))
	}
	var author, video *schema.Table
	for i := range out.Tables {
		switch out.Tables[i].Name {
		case "author":
			author = &out.Tables[i]
		case "video":
			video = &out.Tables[i]
		}
	}
	if author == nil || video == nil {
		t.Fatalf("missing tables in output: %#v", out.Tables)
	}
	if len(author.Relations) != 1 {
		t.Fatalf("author relations mismatch: %#v", author.Relations)
	}
	if author.Relations[0].Kind != "HasMany" {
		t.Fatalf("expected HasMany on author, got %s", author.Relations[0].Kind)
	}
	if len(author.Relations[0].LocalKeys) == 0 || len(author.Relations[0].ForeignKeys) == 0 {
		t.Fatalf("expected join keys on author relation")
	}
	if len(video.Relations) != 1 {
		t.Fatalf("video relations mismatch: %#v", video.Relations)
	}
	if video.Relations[0].Kind != "BelongsTo" {
		t.Fatalf("expected BelongsTo on video, got %s", video.Relations[0].Kind)
	}
	if video.Relations[0].LocalKeys[0] != "author_id" {
		t.Fatalf("unexpected local key: %#v", video.Relations[0].LocalKeys)
	}
	if video.Relations[0].ForeignKeys[0] != "id" {
		t.Fatalf("unexpected foreign key: %#v", video.Relations[0].ForeignKeys)
	}
}

func TestResolveInfersHasOne(t *testing.T) {
	ir := schema.New()
	ir.AddTable(schema.Table{
		Name: "user",
		Columns: []schema.Column{
			{Name: "id", GoType: "int64"},
		},
		PrimaryKey: []string{"id"},
	})
	ir.AddTable(schema.Table{
		Name: "profile",
		Columns: []schema.Column{
			{Name: "id", GoType: "int64"},
			{Name: "user_id", GoType: "int64"},
		},
		PrimaryKey: []string{"id"},
		Uniques: []schema.Unique{
			{Cols: []string{"user_id"}},
		},
		ForeignKeys: []schema.FK{
			{
				Local:      []string{"user_id"},
				RefTable:   "user",
				RefColumns: []string{"id"},
			},
		},
	})

	out, err := Resolve(ir, config.Config{})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	var userTable *schema.Table
	for i := range out.Tables {
		if out.Tables[i].Name == "user" {
			userTable = &out.Tables[i]
			break
		}
	}
	if userTable == nil {
		t.Fatalf("user table missing")
	}
	found := false
	for _, rel := range userTable.Relations {
		if rel.To == "profile" {
			found = true
			if rel.Kind != "HasOne" {
				t.Fatalf("expected HasOne, got %s", rel.Kind)
			}
		}
	}
	if !found {
		t.Fatalf("expected user->profile relation: %#v", userTable.Relations)
	}
}
