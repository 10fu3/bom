# Bom ORM (日本語)

Bom は Prisma 互換の入力モデルを Go 構造体で表現する、読み取り専用の軽量 ORM です。DDL と設定ファイルからコードを自動生成し、方言差分や SQL インジェクション対策をすべてライブラリ側で吸収します。

## 主な特徴
- **Prisma と同じ概念**：`Where` / `OrderBy` / `Select` / `Distinct` / `Take` / `Skip` / ネストしたリレーションをチェーンではなく構造体で記述。
- **読み取り専用**：`FindMany` / `FindFirst` / `FindUnique` のみを提供し、全クエリはプレースホルダ付きで安全に発行。
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
go get bom/pkg/bom bom/pkg/opt bom/pkg/dialect/...
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

     "bom/pkg/generated"
     "bom/pkg/opt"
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

## クエリ記述上のヒント
- `SelectAll` はスカラー列のみを含みます。リレーションを取りたい場合は `ModelSelectRelation{ Args: ... }` を手動で追加してください。
- `Where` は `AND` / `OR` / `NOT` / リレーションフィルタ / `opt.Opt[T]` を組み合わせ可能。
- `FindUnique` は `ModelFindUnique[ModelUK_*]` のジェネリクスでユニークキーを限定します。
- `pkg/dialect/*` の実装は `activeDialect` から差し替え可能なため、アプリ側で `postgres.New()` などを設定して使えます。

## 安全性とパフォーマンス
- 値はすべて `argState` でプレースホルダ化され、SQL 文字列へ直接連結されません。
|- カラム・テーブル名は列挙型 + Dialect の `QuoteIdent` で必ずエスケープされ、ORDER BY / DISTINCT でも安全です。
- JSON キーは生成時にサニタイズし、各 Dialect の `JSON_OBJECT` 系 API を使用します。
- `FindMany` ではサブクエリ＋ `JSON_ARRAYAGG` で結果を 1 レコードにまとめる方式を採用しており、アプリ側で配列展開するだけで済みます。

## 制限と今後の予定
- 書き込み系 (INSERT/UPDATE/DELETE) は非対応です。
- `bomgen` が使用する DDL パーサは MySQL(TiDB) と goyacc 製 SQLite を同梱。PostgreSQL 連携は作業中です。
- `SelectAll` にリレーションは含まれません。循環参照を避けるための仕様です。

## ライセンス
Copyright (c) 2025 10fu3
Released under the MIT license
https://opensource.org/licenses/mit-license.php
