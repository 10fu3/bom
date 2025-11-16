package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bom/internal/schema"
)

func TestGenerateCreatesTypes(t *testing.T) {
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
			{Name: "title", GoType: "string"},
			{Name: "author_id", GoType: "int64"},
		},
		PrimaryKey: []string{"id"},
		Relations: []schema.Relation{
			{
				Name: "author",
				To:   "author",
			},
		},
	})

	tmp := t.TempDir()
	gen := New()
	if err := gen.Generate(ir, tmp, "mysql"); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tmp, "generated.go"))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	out := string(got)
	if !strings.Contains(out, "type VideoWhereInput struct") {
		t.Fatalf("expected where input, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoFindMany struct") {
		t.Fatalf("expected find many, got:\n%s", out)
	}
	if !strings.Contains(out, "type Video struct {") {
		t.Fatalf("expected model struct, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoModel interface") {
		t.Fatalf("expected model constraint, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoUnique interface") {
		t.Fatalf("expected unique interface, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoFindUnique[U VideoUnique] struct") {
		t.Fatalf("expected find unique struct, got:\n%s", out)
	}
	if !strings.Contains(out, "func FindManyVideo[T VideoModel]") {
		t.Fatalf("expected generic FindMany function, got:\n%s", out)
	}
	if !strings.Contains(out, "func FindUniqueVideo[T VideoModel, U VideoUnique]") {
		t.Fatalf("expected generic FindUnique function, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoCreateData struct") {
		t.Fatalf("expected create data struct, got:\n%s", out)
	}
	if !strings.Contains(out, "func CreateOneVideo[T VideoModel]") {
		t.Fatalf("expected create helper, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoCreateMany struct") {
		t.Fatalf("expected create many struct, got:\n%s", out)
	}
	if !strings.Contains(out, "func CreateManyVideo") {
		t.Fatalf("expected CreateMany helper, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoSelectAuthor struct") {
		t.Fatalf("expected relation select union type, got:\n%s", out)
	}
	if !strings.Contains(out, "type VideoAuthorSelectArgs struct") {
		t.Fatalf("expected relation select args type, got:\n%s", out)
	}
	if !strings.Contains(out, "Select: AuthorSelectAll") {
		t.Fatalf("expected relation select all reference, got:\n%s", out)
	}
}
