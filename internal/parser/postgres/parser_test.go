package postgres

import (
	"context"
	"testing"
)

func TestParserCreateTable(t *testing.T) {
	src := `
CREATE TABLE public.author (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE public.video (
    id BIGINT PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES public.author(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_video_slug UNIQUE (slug),
    CONSTRAINT fk_video_author FOREIGN KEY (author_id) REFERENCES public.author(id)
);

CREATE UNIQUE INDEX uq_video_slug_idx ON public.video (slug);
`
	parser := New()
	ir, err := parser.Parse(context.Background(), src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	author := ir.Table("public", "author")
	if author == nil {
		t.Fatalf("author table missing")
	}
	if len(author.Columns) != 4 {
		t.Fatalf("expected 4 author columns, got %d", len(author.Columns))
	}
	if author.Columns[0].Name != "id" || author.Columns[0].DBType != "BIGSERIAL" {
		t.Fatalf("unexpected id column: %#v", author.Columns[0])
	}
	if author.Columns[3].Default == nil || *author.Columns[3].Default != "now()" {
		t.Fatalf("expected now() default, got %v", author.Columns[3].Default)
	}
	video := ir.Table("public", "video")
	if video == nil {
		t.Fatalf("video table missing")
	}
	if len(video.PrimaryKey) != 1 || video.PrimaryKey[0] != "id" {
		t.Fatalf("unexpected primary key: %#v", video.PrimaryKey)
	}
	if video.Columns[4].Default == nil || *video.Columns[4].Default != "CURRENT_TIMESTAMP" {
		t.Fatalf("expected CURRENT_TIMESTAMP default, got %v", video.Columns[4].Default)
	}
	if len(video.Uniques) == 0 {
		t.Fatalf("expected unique constraints")
	}
	foundSlug := false
	for _, u := range video.Uniques {
		if len(u.Cols) == 1 && u.Cols[0] == "slug" {
			foundSlug = true
			break
		}
	}
	if !foundSlug {
		t.Fatalf("slug unique constraint missing")
	}
	if len(video.ForeignKeys) == 0 {
		t.Fatalf("expected at least 1 foreign key")
	}
	fkTarget := false
	for _, fk := range video.ForeignKeys {
		if fk.RefTable == "public.author" {
			fkTarget = true
			break
		}
	}
	if !fkTarget {
		t.Fatalf("author fk missing: %#v", video.ForeignKeys)
	}
	if len(video.Indexes) != 0 {
		t.Fatalf("expected no non-unique indexes, got %d", len(video.Indexes))
	}
}
