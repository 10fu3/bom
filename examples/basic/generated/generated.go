package generated

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "strings"

    "bom/internal/planner"
    "bom/pkg/bom"
    "bom/pkg/dialect"
    dialectmysql "bom/pkg/dialect/mysql"
    "bom/pkg/opt"
)

var activeDialect dialect.Dialect = dialectmysql.New()

type OrderDirection string

const (
    OrderDirectionASC  OrderDirection = "ASC"
    OrderDirectionDESC OrderDirection = "DESC"
)

type argState struct {
    d    dialect.Dialect
    next int
    args []any
}

func newArgState(d dialect.Dialect) *argState {
    return &argState{d: d}
}

func (s *argState) Add(value any) string {
    s.next++
    s.args = append(s.args, value)
    return s.d.Placeholder(s.next)
}

func (s *argState) Args() []any {
    return s.args
}

type whereBuilder struct {
    d     dialect.Dialect
    alias string
    args  *argState
}

func newWhereBuilder(d dialect.Dialect, alias string, args *argState) *whereBuilder {
    return &whereBuilder{d: d, alias: alias, args: args}
}

func (w *whereBuilder) columnRef(column string) string {
    return fmt.Sprintf("%s.%s", w.d.QuoteIdent(w.alias), w.d.QuoteIdent(column))
}

func (w *whereBuilder) addArg(value any) string {
    if w.args == nil {
        return ""
    }
    return w.args.Add(value)
}

func (w *whereBuilder) eq(column string, value any) string {
    return w.d.Eq(w.columnRef(column), w.addArg(value))
}

func (w *whereBuilder) combineAnd(preds []string) string {
    if len(preds) == 0 {
        return ""
    }
    if len(preds) == 1 {
        return preds[0]
    }
    return w.d.And(preds...)
}

func (w *whereBuilder) combineOr(preds []string) string {
    if len(preds) == 0 {
        return ""
    }
    if len(preds) == 1 {
        return preds[0]
    }
    return w.d.Or(preds...)
}

func columnEquals(d dialect.Dialect, leftAlias, leftCol, rightAlias, rightCol string) string {
    return fmt.Sprintf("%s.%s = %s.%s", d.QuoteIdent(leftAlias), d.QuoteIdent(leftCol), d.QuoteIdent(rightAlias), d.QuoteIdent(rightCol))
}

func jsonPair(key, expr string) string {
    safe := strings.ReplaceAll(key, "'", "''")
    return fmt.Sprintf("'%s', %s", safe, expr)
}

func convertOpt(v opt.Opt[int]) *int64 {
    if !v.IsSome() {
        return nil
    }
    val := int64(v.Value())
    return &val
}

func one() *int64 {
    v := int64(1)
    return &v
}

var AuthorAllColumns = []string{
    "id",
    "name",
    "email",
    "created_at",
}

type Author struct {
    Id uint64 `json:"id"`

    Name string `json:"name"`

    Email string `json:"email"`

    CreatedAt string  `json:"created_at"`
    Video     []Video `json:"video,omitempty"`
}

type AuthorModel interface {
    Author
}

type AuthorField string

const (
    AuthorFieldUnknown   AuthorField = ""
    AuthorFieldId        AuthorField = "id"
    AuthorFieldName      AuthorField = "name"
    AuthorFieldEmail     AuthorField = "email"
    AuthorFieldCreatedAt AuthorField = "created_at"
)

type AuthorWhereInput struct {
    AND       []AuthorWhereInput
    OR        []AuthorWhereInput
    NOT       []AuthorWhereInput
    Id        opt.Opt[uint64]
    Name      opt.Opt[string]
    Email     opt.Opt[string]
    CreatedAt opt.Opt[string]
    Video     *AuthorVideoRelation
}

type AuthorWhereUniqueInput struct {
    Id *uint64
}

type AuthorOrderByInput struct {
    Field     AuthorField
    Direction OrderDirection
}
type AuthorVideoRelation struct {
    Some *VideoWhereInput
    None *VideoWhereInput
}

type AuthorSelect []AuthorSelectItem

type AuthorSelectItem interface {
    isAuthorSelectItem()
}

func (AuthorField) isAuthorSelectItem() {}

type AuthorSelectVideo struct {
    Args AuthorVideoSelectArgs
}

func (AuthorSelectVideo) isAuthorSelectItem() {}

var AuthorSelectAll = AuthorSelect{
	AuthorFieldId,
	AuthorFieldName,
	AuthorFieldEmail,
	AuthorFieldCreatedAt,
}

type AuthorVideoSelectArgs struct {
    Where   *VideoWhereInput
    OrderBy []VideoOrderByInput
    Take    opt.Opt[int]
    Skip    opt.Opt[int]
    Select  VideoSelect
}

type AuthorFindMany struct {
    Where    *AuthorWhereInput
    OrderBy  []AuthorOrderByInput
    Take     opt.Opt[int]
    Skip     opt.Opt[int]
    Distinct []AuthorField
    Select   AuthorSelect
}

type AuthorFindFirst struct {
    Where   *AuthorWhereInput
    OrderBy []AuthorOrderByInput
    Skip    opt.Opt[int]
    Select  AuthorSelect
}

func buildAuthorWhere(d dialect.Dialect, alias string, args *argState, where *AuthorWhereInput) string {
    if where == nil {
        return ""
    }
    w := newWhereBuilder(d, alias, args)
    return buildAuthorWherePredicates(w, where)
}

func buildAuthorWherePredicates(w *whereBuilder, where *AuthorWhereInput) string {
    var preds []string
    if where.Id.IsSome() {
        preds = append(preds, w.eq("id", where.Id.Value()))
    }
    if where.Name.IsSome() {
        preds = append(preds, w.eq("name", where.Name.Value()))
    }
    if where.Email.IsSome() {
        preds = append(preds, w.eq("email", where.Email.Value()))
    }
    if where.CreatedAt.IsSome() {
        preds = append(preds, w.eq("created_at", where.CreatedAt.Value()))
    }
    if where.Video != nil {
        if clause := buildAuthorVideoRelation(w.d, w.alias, w.args, where.Video); clause != "" {
            preds = append(preds, clause)
        }
    }
    for i := range where.AND {
        if clause := buildAuthorWherePredicates(w, &where.AND[i]); clause != "" {
            preds = append(preds, clause)
        }
    }
    var orPreds []string
    for i := range where.OR {
        if clause := buildAuthorWherePredicates(w, &where.OR[i]); clause != "" {
            orPreds = append(orPreds, clause)
        }
    }
    if orClause := w.combineOr(orPreds); orClause != "" {
        preds = append(preds, orClause)
    }
    for i := range where.NOT {
        if clause := buildAuthorWherePredicates(w, &where.NOT[i]); clause != "" {
            preds = append(preds, w.d.Not(clause))
        }
    }
    return w.combineAnd(preds)
}
func buildAuthorVideoRelation(d dialect.Dialect, parentAlias string, args *argState, rel *AuthorVideoRelation) string {
    if rel == nil {
        return ""
    }
    var preds []string
    if rel.Some != nil {
        if clause := buildAuthorVideoRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
            preds = append(preds, clause)
        }
    }
    if rel.None != nil {
        if clause := buildAuthorVideoRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
            preds = append(preds, clause)
        }
    }
    if len(preds) == 0 {
        return ""
    }
    if len(preds) == 1 {
        return preds[0]
    }
    return d.And(preds...)
}

func buildAuthorVideoRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *VideoWhereInput, negate bool) string {
    childAlias := parentAlias + "_video"
    tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
    joinConds := make([]string, 0, 1)
    joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
    var combined string
    if len(joinConds) > 0 {
        combined = strings.Join(joinConds, " AND ")
    }
    if whereClause := buildVideoWhere(d, childAlias, args, where); whereClause != "" {
        if combined != "" {
            combined = d.And(combined, whereClause)
        } else {
            combined = whereClause
        }
    }
    if combined == "" {
        combined = "1=1"
    }
    expr := fmt.Sprintf("SELECT 1 FROM %s WHERE %s", tableRef, combined)
    if negate {
        return fmt.Sprintf("NOT EXISTS (%s)", expr)
    }
    return fmt.Sprintf("EXISTS (%s)", expr)
}

func buildAuthorWhereUnique(d dialect.Dialect, alias string, args *argState, where *AuthorWhereUniqueInput) string {
    if where == nil {
        return ""
    }
    w := newWhereBuilder(d, alias, args)
    var preds []string
    if where.Id != nil {
        preds = append(preds, w.eq("id", *where.Id))
    }
    return w.combineAnd(preds)
}

func buildAuthorOrderBy(d dialect.Dialect, alias string, order []AuthorOrderByInput) []string {
    var cols []string
    for _, o := range order {
        if o.Field == "" {
            continue
        }
        dir := "ASC"
        if o.Direction == OrderDirectionDESC {
            dir = "DESC"
        }
        cols = append(cols, fmt.Sprintf("%s.%s %s", d.QuoteIdent(alias), d.QuoteIdent(string(o.Field)), dir))
    }
    return cols
}

func buildAuthorDistinct(d dialect.Dialect, alias string, distinct []AuthorField) []string {
    var out []string
    for _, field := range distinct {
        if field == "" {
            continue
        }
        out = append(out, fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(field))))
    }
    return out
}

func buildAuthorJSONExpr(d dialect.Dialect, alias string, args *argState, sel AuthorSelect) (string, error) {
    items := sel
    includeAll := len(items) == 0
    var pairs []string
    if includeAll {
        pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
        pairs = append(pairs, jsonPair("name", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("name"))))
        pairs = append(pairs, jsonPair("email", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("email"))))
        pairs = append(pairs, jsonPair("created_at", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("created_at"))))
        expr, err := buildAuthorVideoJSON(d, alias, args, &AuthorVideoSelectArgs{})
        if err != nil {
            return "", err
        }
        pairs = append(pairs, jsonPair("video", expr))
    } else {
        seen := make(map[AuthorField]struct{})
        relationSeen := make(map[string]struct{})
        for _, item := range items {
            switch v := item.(type) {
            case AuthorField:
                if v == AuthorFieldUnknown {
                    continue
                }
                if _, ok := seen[v]; ok {
                    continue
                }
                seen[v] = struct{}{}
                pairs = append(pairs, jsonPair(string(v), fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(v)))))
            case AuthorSelectVideo:
                if _, ok := relationSeen["Video"]; ok {
                    continue
                }
                relationSeen["Video"] = struct{}{}
                expr, err := buildAuthorVideoJSON(d, alias, args, &v.Args)
                if err != nil {
                    return "", err
                }
                pairs = append(pairs, jsonPair("video", expr))
            }
        }
    }
    if len(pairs) == 0 {
        pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
    }
    return d.JSONBuildObject(pairs...), nil
}
func buildAuthorVideoJSON(d dialect.Dialect, parentAlias string, args *argState, sel *AuthorVideoSelectArgs) (string, error) {
    if sel == nil {
        sel = &AuthorVideoSelectArgs{}
    }
    childAlias := parentAlias + "_video"
    childJSON, err := buildVideoJSONExpr(d, childAlias, args, sel.Select)
    if err != nil {
        return "", err
    }
    joinConds := make([]string, 0, 1)
    joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
    joinClause := strings.Join(joinConds, " AND ")
    var whereClause string
    whereClause = buildVideoWhere(d, childAlias, args, sel.Where)
    if joinClause != "" {
        if whereClause != "" {
            whereClause = d.And(joinClause, whereClause)
        } else {
            whereClause = joinClause
        }
    }
    var whereSQL string
    if whereClause != "" {
        whereSQL = " WHERE " + whereClause
    }
    var orderSQL string
    var lo string
    orderClause := buildVideoOrderBy(d, childAlias, sel.OrderBy)
    if len(orderClause) > 0 {
        orderSQL = " ORDER BY " + strings.Join(orderClause, ", ")
    }
    limit := convertOpt(sel.Take)
    offset := convertOpt(sel.Skip)
    lo = d.LimitOffset(limit, offset)
    tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
    arrayExpr := d.JSONArrayAgg(childJSON)
    expr := fmt.Sprintf("(SELECT %s FROM %s%s%s", d.CoalesceJSONAgg(arrayExpr, d.JSONArrayEmpty()), tableRef, whereSQL, orderSQL)
    if lo != "" {
        expr += " " + lo
    }
    expr += ")"
    return expr, nil
}

func FindManyAuthor[T AuthorModel](ctx context.Context, db bom.Querier, q AuthorFindMany) ([]T, error) {
    d := activeDialect
    state := newArgState(d)
    rootAlias := "t0"
    jsonExpr, err := buildAuthorJSONExpr(d, rootAlias, state, q.Select)
    if err != nil {
        return nil, err
    }
    whereClause := buildAuthorWhere(d, rootAlias, state, q.Where)
    input := planner.FindManyInput{
        Table: "author",
        Alias: rootAlias,
        Projections: []planner.Projection{
            {Expr: jsonExpr, Alias: "__bom_json"},
        },
        Where:     whereClause,
        Args:      state.Args(),
        OrderBy:   buildAuthorOrderBy(d, rootAlias, q.OrderBy),
        Distinct:  buildAuthorDistinct(d, rootAlias, q.Distinct),
        Limit:     convertOpt(q.Take),
        Offset:    convertOpt(q.Skip),
        JSONArray: true,
    }
    return queryAuthorRows[T](ctx, db, d, input, true)
}

func FindFirstAuthor[T AuthorModel](ctx context.Context, db bom.Querier, q AuthorFindFirst) (*T, error) {
    fm := AuthorFindMany{
        Where:   q.Where,
        OrderBy: q.OrderBy,
        Skip:    q.Skip,
        Select:  q.Select,
    }
    fm.Take = opt.OVal(1)
    rows, err := FindManyAuthor[T](ctx, db, fm)
    if err != nil {
        return nil, err
    }
    if len(rows) == 0 {
        return nil, sql.ErrNoRows
    }
    out := rows[0]
    return &out, nil
}

func queryAuthorRows[T AuthorModel](ctx context.Context, db bom.Querier, d dialect.Dialect, input planner.FindManyInput, expectArray bool) ([]T, error) {
    sqlStr, args, err := planner.BuildFindMany(d, input)
    if err != nil {
        return nil, err
    }
    rows, err := db.QueryContext(ctx, sqlStr, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []T
    if expectArray {
        if rows.Next() {
            var raw []byte
            if err := rows.Scan(&raw); err != nil {
                return nil, err
            }
            if len(raw) == 0 {
                raw = []byte("[]")
            }
            var payload []json.RawMessage
            if err := json.Unmarshal(raw, &payload); err != nil {
                return nil, err
            }
            for _, item := range payload {
                var rec Author
                if err := json.Unmarshal(item, &rec); err != nil {
                    return nil, err
                }
                result = append(result, T(rec))
            }
        }
        if err := rows.Err(); err != nil {
            return nil, err
        }
        return result, nil
    }
    for rows.Next() {
        var raw []byte
        if err := rows.Scan(&raw); err != nil {
            return nil, err
        }
        if len(raw) == 0 {
            raw = []byte("{}")
        }
        var rec Author
        if err := json.Unmarshal(raw, &rec); err != nil {
            return nil, err
        }
        result = append(result, T(rec))
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return result, nil
}

var VideoAllColumns = []string{
    "id",
    "title",
    "slug",
    "author_id",
    "description",
    "created_at",
}

type Video struct {
    Id uint64 `json:"id"`

    Title string `json:"title"`

    Slug string `json:"slug"`

    AuthorId uint64 `json:"author_id"`

    Description opt.Opt[string] `json:"description,omitempty"`

    CreatedAt string  `json:"created_at"`
    Author    *Author `json:"author,omitempty"`
}

type VideoModel interface {
    Video
}

type VideoField string

const (
    VideoFieldUnknown     VideoField = ""
    VideoFieldId          VideoField = "id"
    VideoFieldTitle       VideoField = "title"
    VideoFieldSlug        VideoField = "slug"
    VideoFieldAuthorId    VideoField = "author_id"
    VideoFieldDescription VideoField = "description"
    VideoFieldCreatedAt   VideoField = "created_at"
)

type VideoWhereInput struct {
    AND         []VideoWhereInput
    OR          []VideoWhereInput
    NOT         []VideoWhereInput
    Id          opt.Opt[uint64]
    Title       opt.Opt[string]
    Slug        opt.Opt[string]
    AuthorId    opt.Opt[uint64]
    Description opt.Opt[string]
    CreatedAt   opt.Opt[string]
    Author      *VideoAuthorRelation
}

type VideoWhereUniqueInput struct {
    Id *uint64
}

type VideoOrderByInput struct {
    Field     VideoField
    Direction OrderDirection
}
type VideoAuthorRelation struct {
    Some *AuthorWhereInput
    None *AuthorWhereInput
}

type VideoSelect []VideoSelectItem

type VideoSelectItem interface {
    isVideoSelectItem()
}

func (VideoField) isVideoSelectItem() {}

type VideoSelectAuthor struct {
    Args VideoAuthorSelectArgs
}

func (VideoSelectAuthor) isVideoSelectItem() {}

var VideoSelectAll = VideoSelect{
	VideoFieldId,
	VideoFieldTitle,
	VideoFieldSlug,
	VideoFieldAuthorId,
	VideoFieldDescription,
	VideoFieldCreatedAt,
}

type VideoAuthorSelectArgs struct {
    Select AuthorSelect
}

type VideoFindMany struct {
    Where    *VideoWhereInput
    OrderBy  []VideoOrderByInput
    Take     opt.Opt[int]
    Skip     opt.Opt[int]
    Distinct []VideoField
    Select   VideoSelect
}

type VideoFindFirst struct {
    Where   *VideoWhereInput
    OrderBy []VideoOrderByInput
    Skip    opt.Opt[int]
    Select  VideoSelect
}

func buildVideoWhere(d dialect.Dialect, alias string, args *argState, where *VideoWhereInput) string {
    if where == nil {
        return ""
    }
    w := newWhereBuilder(d, alias, args)
    return buildVideoWherePredicates(w, where)
}

func buildVideoWherePredicates(w *whereBuilder, where *VideoWhereInput) string {
    var preds []string
    if where.Id.IsSome() {
        preds = append(preds, w.eq("id", where.Id.Value()))
    }
    if where.Title.IsSome() {
        preds = append(preds, w.eq("title", where.Title.Value()))
    }
    if where.Slug.IsSome() {
        preds = append(preds, w.eq("slug", where.Slug.Value()))
    }
    if where.AuthorId.IsSome() {
        preds = append(preds, w.eq("author_id", where.AuthorId.Value()))
    }
    if where.Description.IsSome() {
        preds = append(preds, w.eq("description", where.Description.Value()))
    }
    if where.CreatedAt.IsSome() {
        preds = append(preds, w.eq("created_at", where.CreatedAt.Value()))
    }
    if where.Author != nil {
        if clause := buildVideoAuthorRelation(w.d, w.alias, w.args, where.Author); clause != "" {
            preds = append(preds, clause)
        }
    }
    for i := range where.AND {
        if clause := buildVideoWherePredicates(w, &where.AND[i]); clause != "" {
            preds = append(preds, clause)
        }
    }
    var orPreds []string
    for i := range where.OR {
        if clause := buildVideoWherePredicates(w, &where.OR[i]); clause != "" {
            orPreds = append(orPreds, clause)
        }
    }
    if orClause := w.combineOr(orPreds); orClause != "" {
        preds = append(preds, orClause)
    }
    for i := range where.NOT {
        if clause := buildVideoWherePredicates(w, &where.NOT[i]); clause != "" {
            preds = append(preds, w.d.Not(clause))
        }
    }
    return w.combineAnd(preds)
}
func buildVideoAuthorRelation(d dialect.Dialect, parentAlias string, args *argState, rel *VideoAuthorRelation) string {
    if rel == nil {
        return ""
    }
    var preds []string
    if rel.Some != nil {
        if clause := buildVideoAuthorRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
            preds = append(preds, clause)
        }
    }
    if rel.None != nil {
        if clause := buildVideoAuthorRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
            preds = append(preds, clause)
        }
    }
    if len(preds) == 0 {
        return ""
    }
    if len(preds) == 1 {
        return preds[0]
    }
    return d.And(preds...)
}

func buildVideoAuthorRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *AuthorWhereInput, negate bool) string {
    childAlias := parentAlias + "_author"
    tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author"), d.QuoteIdent(childAlias))
    joinConds := make([]string, 0, 1)
    joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "author_id"))
    var combined string
    if len(joinConds) > 0 {
        combined = strings.Join(joinConds, " AND ")
    }
    if whereClause := buildAuthorWhere(d, childAlias, args, where); whereClause != "" {
        if combined != "" {
            combined = d.And(combined, whereClause)
        } else {
            combined = whereClause
        }
    }
    if combined == "" {
        combined = "1=1"
    }
    expr := fmt.Sprintf("SELECT 1 FROM %s WHERE %s", tableRef, combined)
    if negate {
        return fmt.Sprintf("NOT EXISTS (%s)", expr)
    }
    return fmt.Sprintf("EXISTS (%s)", expr)
}

func buildVideoWhereUnique(d dialect.Dialect, alias string, args *argState, where *VideoWhereUniqueInput) string {
    if where == nil {
        return ""
    }
    w := newWhereBuilder(d, alias, args)
    var preds []string
    if where.Id != nil {
        preds = append(preds, w.eq("id", *where.Id))
    }
    return w.combineAnd(preds)
}

func buildVideoOrderBy(d dialect.Dialect, alias string, order []VideoOrderByInput) []string {
    var cols []string
    for _, o := range order {
        if o.Field == "" {
            continue
        }
        dir := "ASC"
        if o.Direction == OrderDirectionDESC {
            dir = "DESC"
        }
        cols = append(cols, fmt.Sprintf("%s.%s %s", d.QuoteIdent(alias), d.QuoteIdent(string(o.Field)), dir))
    }
    return cols
}

func buildVideoDistinct(d dialect.Dialect, alias string, distinct []VideoField) []string {
    var out []string
    for _, field := range distinct {
        if field == "" {
            continue
        }
        out = append(out, fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(field))))
    }
    return out
}

func buildVideoJSONExpr(d dialect.Dialect, alias string, args *argState, sel VideoSelect) (string, error) {
    items := sel
    includeAll := len(items) == 0
    var pairs []string
    if includeAll {
        pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
        pairs = append(pairs, jsonPair("title", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("title"))))
        pairs = append(pairs, jsonPair("slug", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("slug"))))
        pairs = append(pairs, jsonPair("author_id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("author_id"))))
        pairs = append(pairs, jsonPair("description", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("description"))))
        pairs = append(pairs, jsonPair("created_at", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("created_at"))))
        expr, err := buildVideoAuthorJSON(d, alias, args, &VideoAuthorSelectArgs{})
        if err != nil {
            return "", err
        }
        pairs = append(pairs, jsonPair("author", expr))
    } else {
        seen := make(map[VideoField]struct{})
        relationSeen := make(map[string]struct{})
        for _, item := range items {
            switch v := item.(type) {
            case VideoField:
                if v == VideoFieldUnknown {
                    continue
                }
                if _, ok := seen[v]; ok {
                    continue
                }
                seen[v] = struct{}{}
                pairs = append(pairs, jsonPair(string(v), fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(v)))))
            case VideoSelectAuthor:
                if _, ok := relationSeen["Author"]; ok {
                    continue
                }
                relationSeen["Author"] = struct{}{}
                expr, err := buildVideoAuthorJSON(d, alias, args, &v.Args)
                if err != nil {
                    return "", err
                }
                pairs = append(pairs, jsonPair("author", expr))
            }
        }
    }
    if len(pairs) == 0 {
        pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
    }
    return d.JSONBuildObject(pairs...), nil
}
func buildVideoAuthorJSON(d dialect.Dialect, parentAlias string, args *argState, sel *VideoAuthorSelectArgs) (string, error) {
    if sel == nil {
        sel = &VideoAuthorSelectArgs{}
    }
    childAlias := parentAlias + "_author"
    childJSON, err := buildAuthorJSONExpr(d, childAlias, args, sel.Select)
    if err != nil {
        return "", err
    }
    joinConds := make([]string, 0, 1)
    joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "author_id"))
    joinClause := strings.Join(joinConds, " AND ")
    var whereClause string
    whereClause = joinClause
    var whereSQL string
    if whereClause != "" {
        whereSQL = " WHERE " + whereClause
    }
    tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author"), d.QuoteIdent(childAlias))
    expr := fmt.Sprintf("(SELECT %s FROM %s%s", childJSON, tableRef, whereSQL)
    expr += ")"
    return expr, nil
}

func FindManyVideo[T VideoModel](ctx context.Context, db bom.Querier, q VideoFindMany) ([]T, error) {
    d := activeDialect
    state := newArgState(d)
    rootAlias := "t0"
    jsonExpr, err := buildVideoJSONExpr(d, rootAlias, state, q.Select)
    if err != nil {
        return nil, err
    }
    whereClause := buildVideoWhere(d, rootAlias, state, q.Where)
    input := planner.FindManyInput{
        Table: "video",
        Alias: rootAlias,
        Projections: []planner.Projection{
            {Expr: jsonExpr, Alias: "__bom_json"},
        },
        Where:     whereClause,
        Args:      state.Args(),
        OrderBy:   buildVideoOrderBy(d, rootAlias, q.OrderBy),
        Distinct:  buildVideoDistinct(d, rootAlias, q.Distinct),
        Limit:     convertOpt(q.Take),
        Offset:    convertOpt(q.Skip),
        JSONArray: true,
    }
    return queryVideoRows[T](ctx, db, d, input, true)
}

func FindFirstVideo[T VideoModel](ctx context.Context, db bom.Querier, q VideoFindFirst) (*T, error) {
    fm := VideoFindMany{
        Where:   q.Where,
        OrderBy: q.OrderBy,
        Skip:    q.Skip,
        Select:  q.Select,
    }
    fm.Take = opt.OVal(1)
    rows, err := FindManyVideo[T](ctx, db, fm)
    if err != nil {
        return nil, err
    }
    if len(rows) == 0 {
        return nil, sql.ErrNoRows
    }
    out := rows[0]
    return &out, nil
}

func queryVideoRows[T VideoModel](ctx context.Context, db bom.Querier, d dialect.Dialect, input planner.FindManyInput, expectArray bool) ([]T, error) {
    sqlStr, args, err := planner.BuildFindMany(d, input)
    if err != nil {
        return nil, err
    }
    rows, err := db.QueryContext(ctx, sqlStr, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []T
    if expectArray {
        if rows.Next() {
            var raw []byte
            if err := rows.Scan(&raw); err != nil {
                return nil, err
            }
            if len(raw) == 0 {
                raw = []byte("[]")
            }
            var payload []json.RawMessage
            if err := json.Unmarshal(raw, &payload); err != nil {
                return nil, err
            }
            for _, item := range payload {
                var rec Video
                if err := json.Unmarshal(item, &rec); err != nil {
                    return nil, err
                }
                result = append(result, T(rec))
            }
        }
        if err := rows.Err(); err != nil {
            return nil, err
        }
        return result, nil
    }
    for rows.Next() {
        var raw []byte
        if err := rows.Scan(&raw); err != nil {
            return nil, err
        }
        if len(raw) == 0 {
            raw = []byte("{}")
        }
        var rec Video
        if err := json.Unmarshal(raw, &rec); err != nil {
            return nil, err
        }
        result = append(result, T(rec))
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return result, nil
}

type AuthorUnique interface {
    AuthorUK_Id | AuthorUK_Email
}

type AuthorUK_Id struct {
    Id uint64
}

type AuthorUK_Email struct {
    Email string
}

type AuthorFindUnique[U AuthorUnique] struct {
    Where  U
    Select AuthorSelect
}

func buildAuthorUniquePredicate[U AuthorUnique](d dialect.Dialect, alias string, args *argState, where U) (string, error) {
    w := newWhereBuilder(d, alias, args)
    switch v := any(where).(type) {
    case AuthorUK_Id:
        return buildAuthorUK_IdPredicate(w, v), nil
    case AuthorUK_Email:
        return buildAuthorUK_EmailPredicate(w, v), nil
    default:
        return "", fmt.Errorf("unsupported unique type %T", where)
    }
}

func buildAuthorUK_IdPredicate(w *whereBuilder, where AuthorUK_Id) string {
    var preds []string
    preds = append(preds, w.eq("id", where.Id))
    return w.combineAnd(preds)
}

func buildAuthorUK_EmailPredicate(w *whereBuilder, where AuthorUK_Email) string {
    var preds []string
    preds = append(preds, w.eq("email", where.Email))
    return w.combineAnd(preds)
}

func FindUniqueAuthor[T AuthorModel, U AuthorUnique](ctx context.Context, db bom.Querier, q AuthorFindUnique[U]) (*T, error) {
    d := activeDialect
    state := newArgState(d)
    alias := "t0"
    whereClause, err := buildAuthorUniquePredicate(d, alias, state, q.Where)
    if err != nil {
        return nil, err
    }
    jsonExpr, err := buildAuthorJSONExpr(d, alias, state, q.Select)
    if err != nil {
        return nil, err
    }
    input := planner.FindManyInput{
        Table: "author",
        Alias: alias,
        Projections: []planner.Projection{
            {Expr: jsonExpr, Alias: "__bom_json"},
        },
        Where: whereClause,
        Args:  state.Args(),
        Limit: one(),
    }
    rows, err := queryAuthorRows[T](ctx, db, d, input, false)
    if err != nil {
        return nil, err
    }
    if len(rows) == 0 {
        return nil, sql.ErrNoRows
    }
    out := rows[0]
    return &out, nil
}

type VideoUnique interface {
    VideoUK_Id | VideoUK_Slug
}

type VideoUK_Id struct {
    Id uint64
}

type VideoUK_Slug struct {
    Slug string
}

type VideoFindUnique[U VideoUnique] struct {
    Where  U
    Select VideoSelect
}

func buildVideoUniquePredicate[U VideoUnique](d dialect.Dialect, alias string, args *argState, where U) (string, error) {
    w := newWhereBuilder(d, alias, args)
    switch v := any(where).(type) {
    case VideoUK_Id:
        return buildVideoUK_IdPredicate(w, v), nil
    case VideoUK_Slug:
        return buildVideoUK_SlugPredicate(w, v), nil
    default:
        return "", fmt.Errorf("unsupported unique type %T", where)
    }
}

func buildVideoUK_IdPredicate(w *whereBuilder, where VideoUK_Id) string {
    var preds []string
    preds = append(preds, w.eq("id", where.Id))
    return w.combineAnd(preds)
}

func buildVideoUK_SlugPredicate(w *whereBuilder, where VideoUK_Slug) string {
    var preds []string
    preds = append(preds, w.eq("slug", where.Slug))
    return w.combineAnd(preds)
}

func FindUniqueVideo[T VideoModel, U VideoUnique](ctx context.Context, db bom.Querier, q VideoFindUnique[U]) (*T, error) {
    d := activeDialect
    state := newArgState(d)
    alias := "t0"
    whereClause, err := buildVideoUniquePredicate(d, alias, state, q.Where)
    if err != nil {
        return nil, err
    }
    jsonExpr, err := buildVideoJSONExpr(d, alias, state, q.Select)
    if err != nil {
        return nil, err
    }
    input := planner.FindManyInput{
        Table: "video",
        Alias: alias,
        Projections: []planner.Projection{
            {Expr: jsonExpr, Alias: "__bom_json"},
        },
        Where: whereClause,
        Args:  state.Args(),
        Limit: one(),
    }
    rows, err := queryVideoRows[T](ctx, db, d, input, false)
    if err != nil {
        return nil, err
    }
    if len(rows) == 0 {
        return nil, sql.ErrNoRows
    }
    out := rows[0]
    return &out, nil
}
