# Bom ORM

Bom is a Go-first ORM that mirrors Prisma’s object style while preserving idiomatic Go APIs. It turns your database schema into strongly typed query builders, keeping SQL injection-proof statements and dialect-specific quirks behind a single interface.

## Highlights
- **Prisma-compatible inputs** – `Where`, `OrderBy`, `Select`, `Distinct`, `Take/Skip`, nested relations, and now `CreateOne` data all expressed as plain Go structs (no fluent chaining).
- **Typed reads, creates, updates, and deletes** – generated helpers cover `FindMany`, `FindFirst`, `FindUnique`, Prisma-style `CreateOne`/`CreateMany`, typed `UpdateOne`/`UpdateMany`, and `DeleteOne`/`DeleteMany` helpers with reselectable unique keys and filters.
- **Type-safe uniques** – each unique constraint gets its own struct plus a sum-type-style interface so `FindUnique` can only be called with supported keys.
- **Dialect abstraction** – common interface for quoting, placeholders, JSON aggregation, case-insensitive LIKE, and `DISTINCT ON`. The generator wires in MySQL, PostgreSQL, or SQLite behavior without touching your application code.
- **Schema-driven codegen** – `schema.sql` + `bom.yml` flow through a DDL parser, association resolver, planner, and Go template set to emit everything under `pkg/generated`.
- **Zero reflection at runtime** – queries compile to static functions, JSON marshalling/unmarshalling is explicit, and all arguments are fed through prepared statements.

## Requirements
- Go 1.22 or newer
- A supported SQL dialect (`mysql`, `postgres`, or `sqlite`) for query rendering
- DDL file (`schema.sql`) that TiDB’s MySQL parser or the bundled goyacc-powered SQLite parser can understand (other dialect parsers are pluggable as well)

## Install the generator
```bash
go install ./cmd/bomgen
```

Add the runtime packages to your module (they are already part of this repo if vendored):
```bash
go get bom/pkg/bom bom/pkg/opt bom/pkg/dialect/...
```

## Quick start
1. **Describe your schema**
   - Place your DDL in `./schema.sql`.
   - Optionally create `bom.yml` to filter tables, set the default dialect, alias strategy, or override associations.
   - Minimal example:
     ```yaml
     dialect: mysql
     include_tables: ["author", "video"]
     output:
       root: ./pkg/generated
     ```
2. **Generate code**
   ```bash
   bomgen --ddl ./schema.sql --config ./bom.yml --out ./pkg/generated
   ```
   The generator parses the schema, resolves relations, and emits:
   ```
   pkg/generated/
     generated.go   # models, inputs, relation helpers, planner bridge
   ```
3. **Use the strongly typed API**
   ```go
   import (
     "context"

     "bom/examples/sqlite/generated"
     "bom/pkg/bom"
     "bom/pkg/opt"
   )

   func sampleHasOne(ctx context.Context, db bom.Querier) {
     generated.FindManyAuthor[generated.Author](ctx, db, generated.AuthorFindMany{
       Select: generated.AuthorSelect{
         generated.AuthorFieldName,
         generated.AuthorSelectAuthorProfile{ // HasOne 取得
           Args: generated.AuthorAuthorProfileSelectArgs{
             Select: generated.AuthorProfileSelect{
               generated.AuthorProfileFieldAvatarUrl,
             },
           },
         },
       },
     })
   }

   func sampleBelongsTo(ctx context.Context, db bom.Querier) {
     generated.FindManyVideo[generated.Video](ctx, db, generated.VideoFindMany{
       Select: generated.VideoSelect{
         generated.VideoFieldTitle,
         generated.VideoSelectAuthor{ // BelongsTo 取得
           Args: generated.VideoAuthorSelectArgs{
             Select: generated.AuthorSelect{
               generated.AuthorFieldName,
             },
           },
         },
       },
     })
   }

   func sampleRelationFilters(ctx context.Context, db bom.Querier) {
     generated.FindManyAuthor[generated.Author](ctx, db, generated.AuthorFindMany{
       Where: &generated.AuthorWhereInput{
         Video: &generated.AuthorVideoRelation{
           Some: &generated.VideoWhereInput{ /* HasMany some */ },
           None: &generated.VideoWhereInput{ /* HasMany none */ },
           Every: &generated.VideoWhereInput{
             Comment: &generated.VideoCommentRelation{
               None: &generated.CommentWhereInput{
                 Body: opt.OVal("spam"),
               },
             },
           },
         },
       },
       Select: generated.AuthorSelect{
         generated.AuthorFieldName,
       },
     })
   }

   func sampleCreate(ctx context.Context, db bom.Querier) (*generated.Author, error) {
     return generated.CreateOneAuthor[generated.Author](ctx, db, generated.AuthorCreate{
       Data: generated.AuthorCreateData{
         Name:      opt.OVal("Carol"),
         Email:     opt.OVal("carol@example.com"),
         CreatedAt: opt.OVal("2024-02-03"),
         Video: []generated.VideoCreateData{
           {
             Title:     opt.OVal("Nested"),
             Slug:      opt.OVal("nested"),
             CreatedAt: opt.OVal("2024-02-03"),
           },
         },
       },
       Select: generated.AuthorSelect{
         generated.AuthorFieldId,
         generated.AuthorFieldEmail,
         generated.AuthorSelectVideo{
           Args: generated.AuthorVideoSelectArgs{
             Select: generated.VideoSelect{
               generated.VideoFieldId,
               generated.VideoFieldSlug,
             },
           },
         },
       },
     })
   }
   ```

## Configuration overview (`bom.yml`)
| Key               | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| `dialect`         | `mysql`, `postgres`, or `sqlite`. Controls placeholders, quoting, JSON ops. |
| `include_tables`  | Optional allowlist (array).                                                 |
| `exclude_tables`  | Optional blocklist (array).                                                 |
| `output.root`     | Target directory for generated files.                                       |
| `alias.strategy`  | Identifier shortening strategy (`base62`, etc.).                            |
| `associations`    | Manual relation declarations when FK naming conventions are not enough.     |
| `allow_null_unique` | Permit nullable columns inside unique constraints.                        |
| `identity`        | Map of `table: { column: strategy }` for UUIDv4/UUIDv7/ULID/CUID auto-population. |

## Query model
- `Select` values default to scalar columns only. To fetch relations, add `ModelSelectRelation{ Args: ... }`.
- `Where` inputs support `AND`, `OR`, `NOT`, nested relation filters (`Some`/`None`/`Every`), and optional values via `opt.Opt[T]`.
- `FindUnique` takes `ModelFindUnique[ModelUK_X]` where `ModelUK_X` is a generated struct for each unique key.
- Dialect implementations live under `pkg/dialect/*` and plug into generated code with an `activeDialect` variable you can override if needed.

## Safety and performance
- All user inputs are converted into argument placeholders (`?`, `$1`, etc.) via `argState`, so SQL strings never contain interpolated values.
- Identifier names (columns, tables, aliases) are generated enums and always quoted, removing injection vectors in ORDER BY/DISTINCT clauses.
- JSON shape building uses dialect-specific functions and escapes JSON keys up front.
- Queries can run as raw SELECTs or can wrap results into a top-level JSON array (for `FindMany` pagination helpers) without reflection.

## Integration tests
SQLite/MySQL 向けの実データベース検証はビルドタグで切り替えています。

1. **SQLite (modernc.org/sqlite)**
   ```bash
   go get modernc.org/sqlite@latest
   go test -tags moderncsqlite ./examples/sqlite
   ```
2. **MySQL (Docker)**
   ```bash
   scripts/run_mysql_integration.sh
   ```
   （`examples/mysql/docker/mysql/Dockerfile` でイメージをビルドし、`--tmpfs /var/lib/mysql` 付きでコンテナを起動したうえで `TEST_MYSQL_DSN` を設定し、`go test -tags mysqlserver ./examples/mysql` を実行します。ポートを変える場合は `HOST_PORT=3307 scripts/run_mysql_integration.sh` のように指定してください。）
3. **PostgreSQL (Docker)**
   ```bash
   scripts/run_postgres_integration.sh
   ```
   （`examples/postgres/docker/postgres/Dockerfile` でイメージをビルドし、`--tmpfs /var/lib/postgresql/data` 付きでコンテナを起動したうえで `TEST_POSTGRES_DSN` を設定し、`go test -tags postgresserver ./examples/postgres` を実行します。ポートを変える場合は `PG_HOST_PORT=5433 scripts/run_postgres_integration.sh` のように指定してください。）

## Limitations & roadmap
- Upsert flows remain intentionally out of scope; use the generated helpers above for creates/updates/deletes.
- The CLI currently bootstraps using the TiDB MySQL parser or the goyacc-based SQLite loader selected via `dialect`; a PostgreSQL DDL loader exists internally but is not wired into `bomgen` yet.
- Generated `SelectAll` constants include scalar columns only—relations must be opted in to avoid infinite recursion across circular relations.

## License
MIT
