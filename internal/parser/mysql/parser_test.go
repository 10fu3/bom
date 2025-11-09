package mysql

import (
	"context"
	"testing"
)

func TestParseCreateTable(t *testing.T) {
	src := `
CREATE TABLE IF NOT EXISTS video (
  id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  title VARCHAR(255) NOT NULL COMMENT 'title',
  slug VARCHAR(255) UNIQUE,
  author_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_video_author FOREIGN KEY (author_id) REFERENCES author(id) ON DELETE CASCADE ON UPDATE RESTRICT,
  KEY idx_title (title)
) COMMENT='videos';
`

	parser := New()
	ir, err := parser.Parse(context.Background(), src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ir.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(ir.Tables))
	}
	table := ir.Tables[0]
	if table.Name != "video" {
		t.Fatalf("unexpected table %s", table.Name)
	}
	if len(table.Columns) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(table.Columns))
	}
	id := table.Columns[0]
	if id.GoType != "uint64" {
		t.Fatalf("expected uint64 go type, got %s", id.GoType)
	}
	title := table.Columns[1]
	if title.Nullable {
		t.Fatalf("title should be not nullable")
	}
	if title.Comment == nil || *title.Comment != "title" {
		t.Fatalf("expected title comment, got %v", title.Comment)
	}
	created := table.Columns[4]
	if created.Default == nil || *created.Default != "CURRENT_TIMESTAMP" {
		t.Fatalf("expected CURRENT_TIMESTAMP default, got %v", created.Default)
	}
	if len(table.PrimaryKey) != 1 || table.PrimaryKey[0] != "id" {
		t.Fatalf("expected primary key id, got %v", table.PrimaryKey)
	}
	if len(table.Uniques) != 1 || table.Uniques[0].Cols[0] != "slug" {
		t.Fatalf("expected unique on slug, got %#v", table.Uniques)
	}
	if len(table.Indexes) != 1 || table.Indexes[0].Name != "idx_title" {
		t.Fatalf("expected secondary index, got %#v", table.Indexes)
	}
	if len(table.ForeignKeys) != 1 {
		t.Fatalf("expected foreign key, got %#v", table.ForeignKeys)
	}
	fk := table.ForeignKeys[0]
	if fk.OnDelete != "CASCADE" || fk.OnUpdate != "RESTRICT" {
		t.Fatalf("unexpected fk actions: %s %s", fk.OnDelete, fk.OnUpdate)
	}
	if table.Comment == nil || *table.Comment != "videos" {
		t.Fatalf("expected table comment, got %v", table.Comment)
	}
}

func TestParseAlterTable(t *testing.T) {
	src := `
CREATE TABLE video (
  id BIGINT PRIMARY KEY,
  title VARCHAR(100)
);
ALTER TABLE video ADD COLUMN slug VARCHAR(50) COMMENT 'slug';
ALTER TABLE video MODIFY COLUMN title VARCHAR(255) NOT NULL;
ALTER TABLE video CHANGE COLUMN slug permalink VARCHAR(64);
ALTER TABLE video DROP COLUMN permalink;
ALTER TABLE video ADD CONSTRAINT uq_title UNIQUE (title);
ALTER TABLE video DROP INDEX uq_title;
ALTER TABLE video COMMENT = 'updated';
`
	parser := New()
	ir, err := parser.Parse(context.Background(), src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ir.Tables) != 1 {
		t.Fatalf("expected 1 table")
	}
	table := ir.Tables[0]
	if len(table.Columns) != 2 {
		t.Fatalf("expected 2 columns after drop, got %d", len(table.Columns))
	}
	title := table.Columns[1]
	if title.DBType != "VARCHAR(255)" {
		t.Fatalf("expected modified column type, got %s", title.DBType)
	}
	if title.Nullable {
		t.Fatalf("title should be NOT NULL after modify")
	}
	if len(table.Uniques) != 0 {
		t.Fatalf("unique constraint should have been dropped, got %#v", table.Uniques)
	}
	if table.Comment == nil || *table.Comment != "updated" {
		t.Fatalf("expected updated table comment, got %v", table.Comment)
	}
}

func TestParseCreateIndexStatements(t *testing.T) {
	src := `
CREATE TABLE post (
  id BIGINT PRIMARY KEY,
  title VARCHAR(100)
);
CREATE INDEX idx_post_title ON post(title);
CREATE UNIQUE INDEX uq_post_title ON post(title);
`
	parser := New()
	ir, err := parser.Parse(context.Background(), src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ir.Tables) != 1 {
		t.Fatalf("expected 1 table")
	}
	table := ir.Tables[0]
	if len(table.Indexes) != 1 || table.Indexes[0].Name != "idx_post_title" {
		t.Fatalf("expected non-unique index, got %#v", table.Indexes)
	}
	if len(table.Uniques) != 1 || table.Uniques[0].Name != "uq_post_title" {
		t.Fatalf("expected unique index, got %#v", table.Uniques)
	}
}
