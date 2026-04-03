package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/10fu3/bom/examples/sqlite/generated"
	"github.com/10fu3/bom/pkg/bom"
	"github.com/10fu3/bom/pkg/opt"
)

type stmtData struct {
	sql  string
	args []any
}

var sqlDebugEnabled = func() bool {
	val := strings.TrimSpace(os.Getenv("BOM_DEBUG_SQL"))
	if val == "" {
		return false
	}
	switch strings.ToLower(val) {
	case "1", "true", "t", "yes", "on":
		return true
	default:
		return false
	}
}()

func logSQL(kind, query string, args []any) {
	if !sqlDebugEnabled {
		return
	}
	fmt.Printf("[BOM DEBUG] %s: %s\n", kind, query)
	if len(args) > 0 {
		fmt.Printf("[BOM DEBUG] args: %#v\n", args)
	}
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
	logSQL("QUERY", query, args)
	return s.db.QueryContext(ctx, query, args...)
}

func (s *sqlQuerier) ExecContext(ctx context.Context, query string, args ...any) (bom.Result, error) {
	logSQL("EXEC", query, args)
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

func runCreateAssertions(t *testing.T, ctx context.Context, querier bom.Querier) {
	t.Helper()

	createdAuthor, err := generated.CreateOneAuthor[generated.Author](ctx, querier, generated.AuthorCreate{
		Data: generated.AuthorCreateData{
			Id:        opt.OVal(uint64(99)),
			Name:      opt.OVal("Carol"),
			Email:     opt.OVal("carol@example.com"),
			CreatedAt: opt.OVal("2024-02-01"),
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
			generated.AuthorFieldName,
			generated.AuthorFieldEmail,
		},
	})
	if err != nil {
		t.Fatalf("CreateOneAuthor failed: %v", err)
	}
	if createdAuthor == nil || createdAuthor.Id != 99 {
		t.Fatalf("unexpected author create result: %#v", createdAuthor)
	}
	if createdAuthor.Email != "carol@example.com" {
		t.Fatalf("expected email preserved on create, got %#v", createdAuthor.Email)
	}

	_, err = generated.CreateOneVideo[generated.Video](ctx, querier, generated.VideoCreate{
		Data: generated.VideoCreateData{
			Id:          opt.OVal(uint64(50)),
			Title:       opt.OVal("New Post"),
			Slug:        opt.OVal("new-post"),
			AuthorId:    opt.OVal(createdAuthor.Id),
			Description: opt.OVal("launch"),
			CreatedAt:   opt.OVal("2024-02-03"),
		},
		Select: generated.VideoSelect{
			generated.VideoFieldId,
			generated.VideoFieldAuthorId,
		},
	})
	if err != nil {
		t.Fatalf("CreateOneVideo failed: %v", err)
	}

	fetched, err := generated.FindUniqueVideo[generated.Video](ctx, querier, generated.VideoFindUnique[generated.VideoUK_Id]{
		Where: generated.VideoUK_Id{Id: 50},
		Select: generated.VideoSelect{
			generated.VideoFieldId,
			generated.VideoFieldAuthorId,
			generated.VideoFieldSlug,
		},
	})
	if err != nil {
		t.Fatalf("FindUniqueVideo after create failed: %v", err)
	}
	if fetched == nil || fetched.Slug != "new-post" || fetched.AuthorId != createdAuthor.Id {
		t.Fatalf("unexpected video after create: %#v", fetched)
	}

	tagRows := []generated.TagCreateData{
		{
			Id:   opt.OVal(uint64(201)),
			Name: opt.OVal("bulk-201"),
		},
		{
			Id:   opt.OVal(uint64(202)),
			Name: opt.OVal("bulk-202"),
		},
	}
	inserted, err := generated.CreateManyTag(ctx, querier, generated.TagCreateMany{
		Data: tagRows,
	})
	if err != nil {
		t.Fatalf("CreateManyTag failed: %v", err)
	}
	if inserted != int64(len(tagRows)) {
		t.Fatalf("expected %d tags inserted, got %d", len(tagRows), inserted)
	}
	tag, err := generated.FindUniqueTag[generated.Tag](ctx, querier, generated.TagFindUnique[generated.TagUK_Id]{
		Where: generated.TagUK_Id{Id: 201},
		Select: generated.TagSelect{
			generated.TagFieldId,
			generated.TagFieldName,
		},
	})
	if err != nil {
		t.Fatalf("FindUniqueTag after CreateMany failed: %v", err)
	}
	if tag == nil || tag.Name != "bulk-201" {
		t.Fatalf("unexpected tag after CreateMany: %#v", tag)
	}

	nestedAuthor, err := generated.CreateOneAuthor[generated.Author](ctx, querier, generated.AuthorCreate{
		Data: generated.AuthorCreateData{
			Id:        opt.OVal(uint64(150)),
			Name:      opt.OVal("NestedParent"),
			Email:     opt.OVal("nested@example.com"),
			CreatedAt: opt.OVal("2024-03-01"),
			Video: []generated.VideoCreateData{
				{
					Id:        opt.OVal(uint64(160)),
					Title:     opt.OVal("Nested Child"),
					Slug:      opt.OVal("nested-child"),
					CreatedAt: opt.OVal("2024-03-02"),
					Comment: []generated.CommentCreateData{
						{
							Id:        opt.OVal(uint64(170)),
							Body:      opt.OVal("nice"),
							CreatedAt: opt.OVal("2024-03-03"),
						},
					},
				},
			},
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
	})
	if err != nil {
		t.Fatalf("CreateOneAuthor nested failed: %v", err)
	}
	if nestedAuthor == nil || nestedAuthor.Id != 150 {
		t.Fatalf("unexpected nested author: %#v", nestedAuthor)
	}
	videoRecord, err := generated.FindUniqueVideo[generated.Video](ctx, querier, generated.VideoFindUnique[generated.VideoUK_Slug]{
		Where: generated.VideoUK_Slug{Slug: "nested-child"},
		Select: generated.VideoSelect{
			generated.VideoFieldId,
			generated.VideoFieldAuthorId,
			generated.VideoSelectComment{
				Args: generated.VideoCommentSelectArgs{
					Select: generated.CommentSelect{
						generated.CommentFieldBody,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("FindUniqueVideo nested failed: %v", err)
	}
	if videoRecord == nil || len(videoRecord.Comment) != 1 || videoRecord.Comment[0].Body != "nice" {
		t.Fatalf("nested relation not inserted: %#v", videoRecord)
	}
}

func runClauseInjectionAssertions(t *testing.T, ctx context.Context, querier bom.Querier) {
	t.Helper()

	assertInjectionError := func(kind string, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s: expected error for malicious identifier", kind)
		}
		if !strings.Contains(strings.ToLower(err.Error()), "no such column") {
			t.Fatalf("%s: unexpected error %v", kind, err)
		}
	}

	maliciousIdent := `name" DESC; DROP TABLE author; --`

	_, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Select: generated.AuthorSelect{
			generated.AuthorField(maliciousIdent),
		},
	})
	assertInjectionError("select", err)

	_, err = generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
		OrderBy: []generated.AuthorOrderByInput{
			{Field: generated.AuthorField(maliciousIdent), Direction: generated.OrderDirectionASC},
		},
	})
	assertInjectionError("order", err)

	distinctRows, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
		Distinct: []generated.AuthorField{
			generated.AuthorField(maliciousIdent),
		},
	})
	if err != nil {
		t.Fatalf("distinct: unexpected error %v", err)
	}
	if len(distinctRows) < 2 {
		t.Fatalf("distinct should still return rows, got %#v", distinctRows)
	}

	limited, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
		OrderBy: []generated.AuthorOrderByInput{
			{Field: generated.AuthorFieldId, Direction: generated.OrderDirectionASC},
		},
		Take: opt.OVal(1),
		Skip: opt.OVal(1),
	})
	if err != nil {
		t.Fatalf("FindManyAuthor limit/offset failed: %v", err)
	}
	if len(limited) != 1 || limited[0].Id != 2 {
		t.Fatalf("limit/offset did not return expected record: %#v", limited)
	}

	after, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
			generated.AuthorFieldEmail,
		},
		OrderBy: []generated.AuthorOrderByInput{
			{Field: generated.AuthorFieldId, Direction: generated.OrderDirectionASC},
		},
	})
	if err != nil {
		t.Fatalf("FindManyAuthor after injection attempts failed: %v", err)
	}
	if len(after) < 2 || after[0].Email != "alice@example.com" || after[1].Email != "bob@example.com" {
		t.Fatalf("author records changed unexpectedly: %#v", after)
	}
}

func runMutationInjectionAssertions(t *testing.T, ctx context.Context, querier bom.Querier) {
	t.Helper()

	insertPayload := `Mallory'); DROP TABLE author; --`
	newAuthorID := uint64(250)
	created, err := generated.CreateOneAuthor[generated.Author](ctx, querier, generated.AuthorCreate{
		Data: generated.AuthorCreateData{
			Id:        opt.OVal(newAuthorID),
			Name:      opt.OVal(insertPayload),
			Email:     opt.OVal("mallory@example.com"),
			CreatedAt: opt.OVal("2024-05-01"),
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
			generated.AuthorFieldName,
		},
	})
	if err != nil {
		t.Fatalf("CreateOneAuthor malicious insert failed: %v", err)
	}
	if created == nil || created.Name != insertPayload {
		t.Fatalf("malicious payload not stored as value: %#v", created)
	}

	updatePayload := `Alice'); UPDATE author SET name='pwned'; --`
	updated, err := generated.UpdateOneAuthor[generated.Author](ctx, querier, generated.AuthorUpdate[generated.AuthorUK_Id]{
		Where: generated.AuthorUK_Id{Id: 1},
		Data: generated.AuthorUpdateData{
			Name: opt.OVal(updatePayload),
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
			generated.AuthorFieldName,
		},
	})
	if err != nil {
		t.Fatalf("UpdateOneAuthor malicious payload failed: %v", err)
	}
	if updated == nil || updated.Name != updatePayload {
		t.Fatalf("update payload not round-tripped: %#v", updated)
	}

	deletePayload := "alice@example.com' OR 1=1 --"
	affected, err := generated.DeleteManyAuthor(ctx, querier, generated.AuthorDeleteMany{
		Where: &generated.AuthorWhereInput{
			Email: opt.OVal(deletePayload),
		},
	})
	if err != nil {
		t.Fatalf("DeleteManyAuthor malicious payload failed: %v", err)
	}
	if affected != 0 {
		t.Fatalf("malicious delete should not remove rows, deleted=%d", affected)
	}

	cleanupDeleted, err := generated.DeleteManyAuthor(ctx, querier, generated.AuthorDeleteMany{
		Where: &generated.AuthorWhereInput{
			Id: opt.OVal(newAuthorID),
		},
	})
	if err != nil {
		t.Fatalf("cleanup delete failed: %v", err)
	}
	if cleanupDeleted != 1 {
		t.Fatalf("expected to remove inserted row, deleted=%d", cleanupDeleted)
	}

	_, err = generated.UpdateOneAuthor[generated.Author](ctx, querier, generated.AuthorUpdate[generated.AuthorUK_Id]{
		Where: generated.AuthorUK_Id{Id: 1},
		Data: generated.AuthorUpdateData{
			Name: opt.OVal("Alice"),
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
	})
	if err != nil {
		t.Fatalf("resetting Alice name failed: %v", err)
	}

	finalAuthor, err := generated.FindUniqueAuthor[generated.Author](ctx, querier, generated.AuthorFindUnique[generated.AuthorUK_Id]{
		Where: generated.AuthorUK_Id{Id: 1},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
			generated.AuthorFieldName,
		},
	})
	if err != nil {
		t.Fatalf("FindUniqueAuthor final check failed: %v", err)
	}
	if finalAuthor == nil || finalAuthor.Name != "Alice" {
		t.Fatalf("author record corrupted after mutations: %#v", finalAuthor)
	}
}
