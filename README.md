# Bom ORM

Bom is a Go-first, read-only ORM that mirrors Prisma’s object style while preserving idiomatic Go APIs. It turns your database schema into strongly typed query builders, keeping SQL injection-proof statements and dialect-specific quirks behind a single interface.

## Highlights
- **Prisma-compatible inputs** – `Where`, `OrderBy`, `Select`, `Distinct`, `Take/Skip`, and nested relations expressed as plain Go structs (no fluent chaining).
- **Read-only by design** – generated helpers cover `FindMany`, `FindFirst`, and `FindUnique`. Every query is parameterized and safe.
- **Type-safe uniques** – each unique constraint gets its own struct plus a sum-type-style interface so `FindUnique` can only be called with supported keys.
- **Dialect abstraction** – common interface for quoting, placeholders, JSON aggregation, case-insensitive LIKE, and `DISTINCT ON`. The generator wires in MySQL, PostgreSQL, or SQLite behavior without touching your application code.
- **Schema-driven codegen** – `schema.sql` + `bom.yml` flow through a DDL parser, association resolver, planner, and Go template set to emit everything under `pkg/generated`.
- **Zero reflection at runtime** – queries compile to static functions, JSON marshalling/unmarshalling is explicit, and all arguments are fed through prepared statements.

## Requirements
- Go 1.22 or newer
- A supported SQL dialect (`mysql`, `postgres`, or `sqlite`) for query rendering
- DDL file (`schema.sql`) that TiDB’s MySQL parser can understand (other dialect parsers are pluggable, but the default CLI currently uses the MySQL/TiDB version)

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
     author.go
     video.go
     generated.go   # shared helpers (arg state, planner integration, JSON shaping)
   ```
3. **Use the strongly typed API**
   ```go
   package main

   import (
     "context"
     "database/sql"

     "bom/pkg/generated"
     "bom/pkg/opt"
   )

   func listVideos(ctx context.Context, db *sql.DB) ([]generated.Video, error) {
     query := generated.VideoFindMany{
       Where: &generated.VideoWhereInput{
         AuthorId: opt.OVal(uint64(1)),
       },
       Select: generated.VideoSelect{
         generated.VideoFieldId,
         generated.VideoFieldTitle,
         generated.VideoSelectAuthor{
           Args: generated.VideoAuthorSelectArgs{
             Select: generated.AuthorSelect{
               generated.AuthorFieldName,
             },
           },
         },
       },
     }
     return generated.FindManyVideo[generated.Video](ctx, db, query)
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

## Query model
- `Select` values default to scalar columns only. To fetch relations, add `ModelSelectRelation{ Args: ... }`.
- `Where` inputs support `AND`, `OR`, `NOT`, nested relation filters, and optional values via `opt.Opt[T]`.
- `FindUnique` takes `ModelFindUnique[ModelUK_X]` where `ModelUK_X` is a generated struct for each unique key.
- Dialect implementations live under `pkg/dialect/*` and plug into generated code with an `activeDialect` variable you can override if needed.

## Safety and performance
- All user inputs are converted into argument placeholders (`?`, `$1`, etc.) via `argState`, so SQL strings never contain interpolated values.
- Identifier names (columns, tables, aliases) are generated enums and always quoted, removing injection vectors in ORDER BY/DISTINCT clauses.
- JSON shape building uses dialect-specific functions and escapes JSON keys up front.
- Queries can run as raw SELECTs or can wrap results into a top-level JSON array (for `FindMany` pagination helpers) without reflection.

## Limitations & roadmap
- Write operations (INSERT/UPDATE/DELETE) are intentionally out of scope.
- The CLI currently bootstraps using the TiDB MySQL parser; PostgreSQL and SQLite DDL loaders exist internally but are not wired into `bomgen` yet.
- Generated `SelectAll` constants include scalar columns only—relations must be opted in to avoid infinite recursion across circular relations.

## License
MIT
