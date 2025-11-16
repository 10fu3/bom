package sqlite

import (
	"context"
	"testing"
)

func TestParseCreateTableBasic(t *testing.T) {
	src := `
CREATE TABLE author (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT UNIQUE,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
`
	parser := New()
	ir, err := parser.Parse(context.Background(), src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ir.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(ir.Tables))
	}
	author := ir.Tables[0]
	if author.Name != "author" {
		t.Fatalf("unexpected table name %s", author.Name)
	}
	if len(author.Columns) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(author.Columns))
	}
	if author.PrimaryKey == nil || len(author.PrimaryKey) != 1 || author.PrimaryKey[0] != "id" {
		t.Fatalf("primary key not detected: %#v", author.PrimaryKey)
	}
	if len(author.Uniques) != 1 {
		t.Fatalf("expected unique constraint for email, got %#v", author.Uniques)
	}
	email := author.Columns[2]
	if !email.Nullable {
		t.Fatalf("email should be nullable when UNIQUE only")
	}
	created := author.Columns[3]
	if created.Default == nil || *created.Default != "CURRENT_TIMESTAMP" {
		t.Fatalf("expected CURRENT_TIMESTAMP default, got %v", created.Default)
	}
}

func TestParseForeignKeysAndIndexes(t *testing.T) {
	src := `
CREATE TABLE video (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  author_id INTEGER NOT NULL,
  CONSTRAINT fk_video_author FOREIGN KEY(author_id) REFERENCES author(id) ON DELETE CASCADE ON UPDATE RESTRICT
);
CREATE INDEX idx_video_title ON video(title);
CREATE UNIQUE INDEX uq_video_title ON video(title);
`
	parser := New()
	ir, err := parser.Parse(context.Background(), src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ir.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(ir.Tables))
	}
	video := ir.Tables[0]
	if len(video.ForeignKeys) != 1 {
		t.Fatalf("expected foreign key, got %#v", video.ForeignKeys)
	}
	fk := video.ForeignKeys[0]
	if fk.Name != "fk_video_author" {
		t.Fatalf("unexpected fk name %s", fk.Name)
	}
	if fk.OnDelete != "CASCADE" || fk.OnUpdate != "RESTRICT" {
		t.Fatalf("unexpected fk actions: %s %s", fk.OnDelete, fk.OnUpdate)
	}
	if len(video.Indexes) != 1 || video.Indexes[0].Name != "idx_video_title" {
		t.Fatalf("expected non-unique index, got %#v", video.Indexes)
	}
	if len(video.Uniques) != 1 || video.Uniques[0].Name != "uq_video_title" {
		t.Fatalf("expected unique index recorded as constraint, got %#v", video.Uniques)
	}
}
