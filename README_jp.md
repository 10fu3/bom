# Bom ORM (日本語)

Bom は Prisma 互換の入力モデルを Go 構造体で表現する軽量 ORM です。DDL と設定ファイルからコードを自動生成し、方言差分や SQL インジェクション対策をすべてライブラリ側で吸収します。

## 主な特徴
- **Prisma と同じ概念**：`Where` / `OrderBy` / `Select` / `Distinct` / `Take` / `Skip` / ネストしたリレーション / `CreateOne` / `CreateMany` の入力を構造体で記述。
- **読み書きサポート**：`FindMany` / `FindFirst` / `FindUnique` に加え、ネストした `CreateOne` / `CreateMany` (AUTO_INCREMENT / UUIDv4/v7 / ULID / CUID 自動採番) と `UpdateOne` / `UpdateMany` / `DeleteOne` / `DeleteMany` を安全に実行。すべての書き込みは Dialect 対応のミューテーションプランナー経由で生成し、識別子引用や `RETURNING` 対応を自動調整します。
- **型安全なユニーク検索**：ユニーク制約ごとに専用構造体 + インターフェースを生成するため、誤ったキーで `FindUnique` を呼べません。
- **Dialect 抽象化**：識別子の引用・プレースホルダ・JSON 集計・ILIKE/LcLike・`DISTINCT ON` などの差分を `pkg/dialect/*` が担当。
- **スキーマ駆動コード生成**：DDL→AST→IR→アソシエーション解決→Go テンプレートというパイプラインで `pkg/generated` を構築。
- **ランタイム反射ゼロ**：SQL 文字列・JSON 変換は生成コードに焼き込み、実行時は単なる関数呼び出しだけです。

## 必要環境
- Go 1.22 以上
- 対応方言（`mysql` / `postgres` / `sqlite`）のいずれか
- TiDB 互換の MySQL DDL もしくは goyacc ベースの SQLite パーサが解釈できる `schema.sql`（他方言は順次対応予定）

## bomgen の導入
```bash
go install ./cmd/bomgen
```

生成済みコードの実行に必要なパッケージを追加します。
```bash
go get github.com/10fu3/bom/pkg/bom github.com/10fu3/bom/pkg/opt github.com/10fu3/bom/pkg/dialect/...
```

## クイックスタート
1. **スキーマと設定を用意**
   - `schema.sql` に DDL を置き、必要なら `bom.yml` で出力先やテーブルフィルタ、方言、手動アソシエーションを定義します。
   - 例:
     ```yaml
     dialect: mysql
     include_tables: ["author", "video"]
     output:
       root: ./pkg/generated
     ```
2. **コードを生成**
   ```bash
   bomgen --ddl ./schema.sql --config ./bom.yml --out ./pkg/generated
   ```
   実行後は `pkg/generated` 以下にモデル・入力型・共通ヘルパが並びます。
3. **生成 API を呼び出す**
   ```go
   import (
     "context"
     "database/sql"

     "github.com/10fu3/bom/pkg/generated"
     "github.com/10fu3/bom/pkg/opt"
   )

   func listVideos(ctx context.Context, db *sql.DB) ([]generated.Video, error) {
     q := generated.VideoFindMany{
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
     return generated.FindManyVideo[generated.Video](ctx, db, q)
   }

   func createAuthor(ctx context.Context, db *sql.DB) (*generated.Author, error) {
     return generated.CreateOneAuthor[generated.Author](ctx, db, generated.AuthorCreate{
       Data: generated.AuthorCreateData{
         Name:      opt.OVal("Carol"),
         Email:     opt.OVal("carol@example.com"),
         CreatedAt: opt.OVal("2024-02-03"),
       },
       Select: generated.AuthorSelect{
         generated.AuthorFieldId,
         generated.AuthorFieldEmail,
       },
     })
   }

   func createAuthor(ctx context.Context, db *sql.DB) (*generated.Author, error) {
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

## `bom.yml` の主なキー
| キー                  | 説明                                                                 |
|----------------------|----------------------------------------------------------------------|
| `dialect`            | `mysql` / `postgres` / `sqlite`。方言ごとの SQL/JSON 戦略を切替。      |
| `include_tables`     | 生成対象テーブルのホワイトリスト。                                   |
| `exclude_tables`     | 生成対象から外すテーブルのリスト。                                   |
| `output.root`        | 生成ファイルの出力ディレクトリ。                                     |
| `alias.strategy`     | 長い識別子を短縮する戦略名。                                         |
| `associations`       | FK では判別できないリレーションを手動で宣言。                         |
| `allow_null_unique`  | ユニーク制約に NULL 許容列を含める場合に `true`。                    |
| `identity`           | `table: { column: strategy }` 形式で UUIDv4/UUIDv7/ULID/CUID を自動採番。 |

## クエリ記述上のヒント
- `SelectAll` はスカラー列のみを含みます。リレーションを取りたい場合は `ModelSelectRelation{ Args: ... }` を手動で追加してください。
- `Where` は `AND` / `OR` / `NOT` / リレーションフィルタ (`Some`/`None`/`Every`) / `opt.Opt[T]` を組み合わせ可能。
- `FindUnique` は `ModelFindUnique[ModelUK_*]` のジェネリクスでユニークキーを限定します。
- `pkg/dialect/*` の実装は `activeDialect` から差し替え可能なため、アプリ側で `postgres.New()` などを設定して使えます。

## 安全性とパフォーマンス
- 値はすべて `argState` でプレースホルダ化され、SQL 文字列へ直接連結されません。
|- カラム・テーブル名は列挙型 + Dialect の `QuoteIdent` で必ずエスケープされ、ORDER BY / DISTINCT でも安全です。
- INSERT/UPDATE/DELETE 文は共有プランナーで生成し、識別子の引用と `RETURNING` サポートの有無を Dialect ごとに切り替えます。
- JSON キーは生成時にサニタイズし、各 Dialect の `JSON_OBJECT` 系 API を使用します。
- `FindMany` ではサブクエリ＋ `JSON_ARRAYAGG` で結果を 1 レコードにまとめる方式を採用しており、アプリ側で配列展開するだけで済みます。

## 実 DB での統合テスト
- SQLite (modernc.org/sqlite)
  ```bash
  go get modernc.org/sqlite@latest
  go test -tags moderncsqlite ./examples/sqlite
  ```
- MySQL (Docker)
  ```bash
  scripts/run_mysql_integration.sh
  ```
  （`examples/mysql/docker/mysql/Dockerfile` でイメージをビルドし、`--tmpfs /var/lib/mysql` 付きでコンテナを起動して `TEST_MYSQL_DSN` を設定し、`go test -tags mysqlserver ./examples/mysql` を実行します。ホスト側のポートを変える場合は `HOST_PORT=3307 scripts/run_mysql_integration.sh` のように指定してください。）
- PostgreSQL (Docker)
  ```bash
  scripts/run_postgres_integration.sh
  ```
  （`examples/postgres/docker/postgres/Dockerfile` でイメージをビルドし、`--tmpfs /var/lib/postgresql/data` 付きでコンテナを起動して `TEST_POSTGRES_DSN` を設定し、`go test -tags postgresserver ./examples/postgres` を実行します。ホスト側のポートを変える場合は `PG_HOST_PORT=5433 scripts/run_postgres_integration.sh` のように指定してください。）

## 制限と今後の予定
- `Upsert` は未サポートです。
- `bomgen` が使用する DDL パーサは MySQL(TiDB) と goyacc 製 SQLite を同梱。PostgreSQL 連携は作業中です。
- `SelectAll` にリレーションは含まれません。循環参照を避けるための仕様です。

## ライセンス
Copyright (c) 2025 10fu3
Released under the MIT license
https://opensource.org/licenses/mit-license.php

追加のライセンス表記:
- PostgreSQL について、`internal/parser/postgres/ddl.y` と `internal/parser/postgres/ddl_gen.go` は PostgreSQL のライセンスに従います。
- SQLite について、`internal/parser/sqlite/ddl.y`、`internal/parser/sqlite/ddl_gen.go`、`internal/parser/sqlite/lexer.l`、`internal/parser/sqlite/lexer.go` は SQLite のライセンスに従います。
