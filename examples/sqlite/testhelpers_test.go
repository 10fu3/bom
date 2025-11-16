package sqlite_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"bom/examples/sqlite/generated"
	"bom/pkg/bom"
	"bom/pkg/opt"
)

type stmtData struct {
	sql  string
	args []any
}

var sqliteSchemaSQL = strings.Join([]string{
	`CREATE TABLE author (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`,
	`CREATE TABLE author_profile (
		id INTEGER PRIMARY KEY,
		author_id INTEGER NOT NULL UNIQUE,
		bio TEXT,
		avatar_url TEXT,
		created_at TEXT NOT NULL
	);`,
	`CREATE TABLE video (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		slug TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		description TEXT,
		created_at TEXT NOT NULL
	);`,
	`CREATE TABLE comment (
		id INTEGER PRIMARY KEY,
		video_id INTEGER NOT NULL,
		author_id INTEGER NOT NULL,
		body TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`,
	`CREATE TABLE tag (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE
	);`,
	`CREATE TABLE video_tag (
		video_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (video_id, tag_id)
	);`,
}, "\n")

var testFixtures = []stmtData{
	{`INSERT INTO author (id, name, email, created_at) VALUES (?, ?, ?, ?)`, []any{1, "Alice", "alice@example.com", "2024-01-01"}},
	{`INSERT INTO author (id, name, email, created_at) VALUES (?, ?, ?, ?)`, []any{2, "Bob", "bob@example.com", "2024-01-01"}},
	{`INSERT INTO author_profile (id, author_id, bio, avatar_url, created_at) VALUES (?, ?, ?, ?, ?)`, []any{1, 1, "Gopher", "https://example.com/alice.png", "2024-01-01"}},
	{`INSERT INTO video (id, title, slug, author_id, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`, []any{1, "Intro", "intro", 1, "welcome", "2024-01-01"}},
	{`INSERT INTO video (id, title, slug, author_id, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`, []any{2, "Spammy", "spammy", 1, "beware", "2024-01-02"}},
	{`INSERT INTO video (id, title, slug, author_id, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`, []any{3, "Clean", "clean", 2, "quality", "2024-01-03"}},
	{`INSERT INTO comment (id, video_id, author_id, body, created_at) VALUES (?, ?, ?, ?, ?)`, []any{1, 1, 1, "great", "2024-01-04"}},
	{`INSERT INTO comment (id, video_id, author_id, body, created_at) VALUES (?, ?, ?, ?, ?)`, []any{2, 2, 2, "spam", "2024-01-05"}},
	{`INSERT INTO comment (id, video_id, author_id, body, created_at) VALUES (?, ?, ?, ?, ?)`, []any{3, 3, 2, "fine", "2024-01-06"}},
}

func execScript(t *testing.T, db *sql.DB, script string) {
	t.Helper()
	stmts := strings.Split(script, ";")
	for _, raw := range stmts {
		sql := strings.TrimSpace(raw)
		if sql == "" {
			continue
		}
		if _, err := db.Exec(sql); err != nil {
			t.Fatalf("exec %q failed: %v", sql, err)
		}
	}
}

type sqlQuerier struct {
	db *sql.DB
}

func (s *sqlQuerier) QueryContext(ctx context.Context, query string, args ...any) (bom.RowsPublic, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *sqlQuerier) ExecContext(ctx context.Context, query string, args ...any) (bom.Result, error) {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return sqlResult{res: res}, nil
}

type sqlResult struct {
	res sql.Result
}

func (r sqlResult) LastInsertId() (int64, error) { return r.res.LastInsertId() }

func (r sqlResult) RowsAffected() (int64, error) { return r.res.RowsAffected() }

func runRelationQueryAssertions(t *testing.T, ctx context.Context, querier bom.Querier) {
	t.Helper()

	authorsSome, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Where: &generated.AuthorWhereInput{
			Video: &generated.AuthorVideoRelation{
				Some: &generated.VideoWhereInput{
					Comment: &generated.VideoCommentRelation{
						Some: &generated.CommentWhereInput{
							Body: opt.OVal("great"),
						},
					},
				},
			},
		},
		OrderBy: []generated.AuthorOrderByInput{
			{Field: generated.AuthorFieldId, Direction: generated.OrderDirectionASC},
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
	})
	if err != nil {
		t.Fatalf("FindManyAuthor some failed: %v", err)
	}
	if len(authorsSome) != 1 || authorsSome[0].Id != 1 {
		t.Fatalf("expected only author 1 for SOME, got %#v", authorsSome)
	}

	videosNone, err := generated.FindManyVideo[generated.Video](ctx, querier, generated.VideoFindMany{
		Where: &generated.VideoWhereInput{
			Comment: &generated.VideoCommentRelation{
				None: &generated.CommentWhereInput{
					Body: opt.OVal("spam"),
				},
			},
		},
		OrderBy: []generated.VideoOrderByInput{
			{Field: generated.VideoFieldId, Direction: generated.OrderDirectionASC},
		},
		Select: generated.VideoSelect{
			generated.VideoFieldId,
		},
	})
	if err != nil {
		t.Fatalf("FindManyVideo none failed: %v", err)
	}
	if len(videosNone) != 2 || videosNone[0].Id != 1 || videosNone[1].Id != 3 {
		t.Fatalf("expected videos 1 and 3 for NONE, got %#v", videosNone)
	}

	authorsEvery, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Where: &generated.AuthorWhereInput{
			Video: &generated.AuthorVideoRelation{
				Every: &generated.VideoWhereInput{
					Comment: &generated.VideoCommentRelation{
						None: &generated.CommentWhereInput{
							Body: opt.OVal("spam"),
						},
					},
				},
			},
		},
		OrderBy: []generated.AuthorOrderByInput{
			{Field: generated.AuthorFieldId, Direction: generated.OrderDirectionASC},
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
	})
	if err != nil {
		t.Fatalf("FindManyAuthor every failed: %v", err)
	}
	if len(authorsEvery) != 1 || authorsEvery[0].Id != 2 {
		t.Fatalf("expected only author 2 for EVERY, got %#v", authorsEvery)
	}
}
