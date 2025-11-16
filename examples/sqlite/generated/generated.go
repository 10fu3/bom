package generated

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"bom/internal/planner"
	"bom/pkg/bom"
	"bom/pkg/dialect"
	dialectsqlite "bom/pkg/dialect/sqlite"
	"bom/pkg/opt"
)

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

func normalizeJSONRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		var decoded string
		if err := json.Unmarshal(raw, &decoded); err == nil {
			return json.RawMessage(decoded)
		}
	}
	return raw
}

func wrapJSONValue(d dialect.Dialect, expr string) string {
	if expr == "" {
		return expr
	}
	return d.JSONValue(expr)
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

var identityGenerator bom.IdentityGenerator = &bom.DefaultIdentityGenerator{}

func buildInsertSQL(d dialect.Dialect, table string, columns, placeholders []string) string {
	tableIdent := d.QuoteIdent(table)
	if len(columns) == 0 {
		switch strings.ToLower(d.Name()) {
		case "postgres":
			return fmt.Sprintf("INSERT INTO %s DEFAULT VALUES", tableIdent)
		default:
			return fmt.Sprintf("INSERT INTO %s () VALUES ()", tableIdent)
		}
	}
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableIdent,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
}

func maxParameters(d dialect.Dialect) int {
	if cap := d.Cap().MaxParameters; cap > 0 {
		return cap
	}
	return math.MaxInt
}

func ensureParamLimit(d dialect.Dialect, count int) error {
	if count == 0 {
		return nil
	}
	limit := maxParameters(d)
	if count > limit {
		return fmt.Errorf("%s statements support up to %d parameters (got %d)", d.Name(), limit, count)
	}
	return nil
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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

	CreatedAt     string         `json:"created_at"`
	AuthorProfile *AuthorProfile `json:"authorProfile,omitempty"`
	Comment       []Comment      `json:"comment,omitempty"`
	Video         []Video        `json:"video,omitempty"`
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
	AND           []AuthorWhereInput
	OR            []AuthorWhereInput
	NOT           []AuthorWhereInput
	Id            opt.Opt[uint64]
	Name          opt.Opt[string]
	Email         opt.Opt[string]
	CreatedAt     opt.Opt[string]
	AuthorProfile *AuthorAuthorProfileRelation
	Comment       *AuthorCommentRelation
	Video         *AuthorVideoRelation
}

type AuthorWhereUniqueInput struct {
	Id *uint64
}

type AuthorOrderByInput struct {
	Field     AuthorField
	Direction OrderDirection
}
type AuthorAuthorProfileRelation struct {
	Some  *AuthorProfileWhereInput
	None  *AuthorProfileWhereInput
	Every *AuthorProfileWhereInput
}
type AuthorCommentRelation struct {
	Some  *CommentWhereInput
	None  *CommentWhereInput
	Every *CommentWhereInput
}
type AuthorVideoRelation struct {
	Some  *VideoWhereInput
	None  *VideoWhereInput
	Every *VideoWhereInput
}

type AuthorSelect []AuthorSelectItem

type AuthorSelectItem interface {
	isAuthorSelectItem()
}

func (AuthorField) isAuthorSelectItem() {}

type AuthorSelectAuthorProfile struct {
	Args AuthorAuthorProfileSelectArgs
}

func (AuthorSelectAuthorProfile) isAuthorSelectItem() {}

type AuthorSelectComment struct {
	Args AuthorCommentSelectArgs
}

func (AuthorSelectComment) isAuthorSelectItem() {}

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

type AuthorAuthorProfileSelectArgs struct {
	Select AuthorProfileSelect
}

type AuthorCommentSelectArgs struct {
	Where   *CommentWhereInput
	OrderBy []CommentOrderByInput
	Take    opt.Opt[int]
	Skip    opt.Opt[int]
	Select  CommentSelect
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
type AuthorCreateData struct {
	Id            opt.Opt[uint64]
	Name          opt.Opt[string]
	Email         opt.Opt[string]
	CreatedAt     opt.Opt[string]
	AuthorProfile *AuthorProfileCreateData
	Comment       []CommentCreateData
	Video         []VideoCreateData
}

type AuthorCreate struct {
	Data   AuthorCreateData
	Select AuthorSelect
}

type AuthorCreateMany struct {
	Data []AuthorCreateData
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
	if where.AuthorProfile != nil {
		if clause := buildAuthorAuthorProfileRelation(w.d, w.alias, w.args, where.AuthorProfile); clause != "" {
			preds = append(preds, clause)
		}
	}
	if where.Comment != nil {
		if clause := buildAuthorCommentRelation(w.d, w.alias, w.args, where.Comment); clause != "" {
			preds = append(preds, clause)
		}
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
func buildAuthorAuthorProfileRelation(d dialect.Dialect, parentAlias string, args *argState, rel *AuthorAuthorProfileRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildAuthorAuthorProfileRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildAuthorAuthorProfileRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildAuthorAuthorProfileRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildAuthorAuthorProfileRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *AuthorProfileWhereInput, negate bool) string {
	childAlias := parentAlias + "_authorProfile"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author_profile"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
	var combined string
	if len(joinConds) > 0 {
		combined = strings.Join(joinConds, " AND ")
	}
	if whereClause := buildAuthorProfileWhere(d, childAlias, args, where); whereClause != "" {
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

func buildAuthorAuthorProfileRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *AuthorProfileWhereInput) string {
	childAlias := parentAlias + "_authorProfile" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author_profile"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
	var negated string
	if whereClause := buildAuthorProfileWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}
func buildAuthorCommentRelation(d dialect.Dialect, parentAlias string, args *argState, rel *AuthorCommentRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildAuthorCommentRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildAuthorCommentRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildAuthorCommentRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildAuthorCommentRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *CommentWhereInput, negate bool) string {
	childAlias := parentAlias + "_comment"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("comment"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
	var combined string
	if len(joinConds) > 0 {
		combined = strings.Join(joinConds, " AND ")
	}
	if whereClause := buildCommentWhere(d, childAlias, args, where); whereClause != "" {
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

func buildAuthorCommentRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *CommentWhereInput) string {
	childAlias := parentAlias + "_comment" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("comment"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
	var negated string
	if whereClause := buildCommentWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
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
	if rel.Every != nil {
		if clause := buildAuthorVideoRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildAuthorVideoRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *VideoWhereInput) string {
	childAlias := parentAlias + "_video" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
	var negated string
	if whereClause := buildVideoWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
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
		{
			expr, err := buildAuthorAuthorProfileJSON(d, alias, args, &AuthorAuthorProfileSelectArgs{
				Select: AuthorProfileSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("authorProfile", wrapJSONValue(d, expr)))
		}
		{
			expr, err := buildAuthorCommentJSON(d, alias, args, &AuthorCommentSelectArgs{
				Select: CommentSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("comment", wrapJSONValue(d, expr)))
		}
		{
			expr, err := buildAuthorVideoJSON(d, alias, args, &AuthorVideoSelectArgs{
				Select: VideoSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("video", wrapJSONValue(d, expr)))
		}
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
			case AuthorSelectAuthorProfile:
				if _, ok := relationSeen["AuthorProfile"]; ok {
					continue
				}
				relationSeen["AuthorProfile"] = struct{}{}
				{
					expr, err := buildAuthorAuthorProfileJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("authorProfile", wrapJSONValue(d, expr)))
				}
			case AuthorSelectComment:
				if _, ok := relationSeen["Comment"]; ok {
					continue
				}
				relationSeen["Comment"] = struct{}{}
				{
					expr, err := buildAuthorCommentJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("comment", wrapJSONValue(d, expr)))
				}
			case AuthorSelectVideo:
				if _, ok := relationSeen["Video"]; ok {
					continue
				}
				relationSeen["Video"] = struct{}{}
				{
					expr, err := buildAuthorVideoJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("video", wrapJSONValue(d, expr)))
				}
			}
		}
	}
	if len(pairs) == 0 {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
	}
	return d.JSONBuildObject(pairs...), nil
}
func buildAuthorAuthorProfileJSON(d dialect.Dialect, parentAlias string, args *argState, sel *AuthorAuthorProfileSelectArgs) (string, error) {
	if sel == nil {
		sel = &AuthorAuthorProfileSelectArgs{}
	}
	childAlias := parentAlias + "_authorProfile"
	childJSON, err := buildAuthorProfileJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = joinClause
	var whereSQL string
	if whereClause != "" {
		whereSQL = " WHERE " + whereClause
	}
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author_profile"), d.QuoteIdent(childAlias))
	expr := fmt.Sprintf("(SELECT %s FROM %s%s", childJSON, tableRef, whereSQL)
	expr += ")"
	return expr, nil
}
func buildAuthorCommentJSON(d dialect.Dialect, parentAlias string, args *argState, sel *AuthorCommentSelectArgs) (string, error) {
	if sel == nil {
		sel = &AuthorCommentSelectArgs{}
	}
	childAlias := parentAlias + "_comment"
	childJSON, err := buildCommentJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "author_id", parentAlias, "id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = buildCommentWhere(d, childAlias, args, sel.Where)
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
	orderClause := buildCommentOrderBy(d, childAlias, sel.OrderBy)
	if len(orderClause) > 0 {
		orderSQL = " ORDER BY " + strings.Join(orderClause, ", ")
	}
	limit := convertOpt(sel.Take)
	offset := convertOpt(sel.Skip)
	lo = d.LimitOffset(limit, offset)
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("comment"), d.QuoteIdent(childAlias))
	arrayExpr := d.JSONArrayAgg(childJSON)
	expr := fmt.Sprintf("(SELECT %s FROM %s%s%s", d.CoalesceJSONAgg(arrayExpr, d.JSONArrayEmpty()), tableRef, whereSQL, orderSQL)
	if lo != "" {
		expr += " " + lo
	}
	expr += ")"
	return expr, nil
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
	childJSON = wrapJSONValue(d, childJSON)
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
	d := dialectsqlite.New()
	state := newArgState(d)
	rootAlias := "t0"
	jsonExpr, err := buildAuthorJSONExpr(d, rootAlias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	whereClause := buildAuthorWhere(d, rootAlias, state, q.Where)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
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
func CreateOneAuthor[T AuthorModel](ctx context.Context, db bom.Querier, q AuthorCreate) (*T, error) {
	d := dialectsqlite.New()
	data := q.Data
	if err := createAuthorRecord(ctx, db, d, &data); err != nil {
		return nil, err
	}
	if err := createAuthorRelations(ctx, db, d, &data); err != nil {
		return nil, err
	}
	var fetch func(context.Context, AuthorSelect) (*T, error)
	if fetch == nil {
		if data.Id.IsSome() {
			lookup := AuthorUK_Id{
				Id: data.Id.Value(),
			}
			fetch = func(ctx context.Context, sel AuthorSelect) (*T, error) {
				return FindUniqueAuthor[T](ctx, db, AuthorFindUnique[AuthorUK_Id]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		if data.Email.IsSome() {
			lookup := AuthorUK_Email{
				Email: data.Email.Value(),
			}
			fetch = func(ctx context.Context, sel AuthorSelect) (*T, error) {
				return FindUniqueAuthor[T](ctx, db, AuthorFindUnique[AuthorUK_Email]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		return nil, fmt.Errorf("AuthorCreate requires values for a unique constraint")
	}
	return fetch(ctx, q.Select)
}

func CreateManyAuthor(ctx context.Context, db bom.Querier, q AuthorCreateMany) (int64, error) {
	d := dialectsqlite.New()
	if len(q.Data) == 0 {
		return 0, nil
	}
	var total int64
	for i := range q.Data {
		data := q.Data[i]
		if err := createAuthorRecord(ctx, db, d, &data); err != nil {
			return total, err
		}
		if err := createAuthorRelations(ctx, db, d, &data); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func createAuthorRecord(ctx context.Context, db bom.Querier, d dialect.Dialect, data *AuthorCreateData) error {
	state := newArgState(d)
	var columns []string
	var placeholders []string
	if identityGenerator == nil {
		identityGenerator = &bom.DefaultIdentityGenerator{}
	}
	var wantsAutoId bool
	var wantsAutoName bool
	var wantsAutoEmail bool
	var wantsAutoCreatedAt bool
	if data.Id.IsSome() {
		val := data.Id.Value()
		columns = append(columns, d.QuoteIdent("id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
		wantsAutoId = true
	}
	if data.Name.IsSome() {
		val := data.Name.Value()
		columns = append(columns, d.QuoteIdent("name"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.Email.IsSome() {
		val := data.Email.Value()
		columns = append(columns, d.QuoteIdent("email"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.CreatedAt.IsSome() {
		val := data.CreatedAt.Value()
		columns = append(columns, d.QuoteIdent("created_at"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return err
	}
	sqlStr := buildInsertSQL(d, "author", columns, placeholders)
	autoCount := 0
	if wantsAutoId {
		autoCount++
	}
	if wantsAutoName {
		autoCount++
	}
	if wantsAutoEmail {
		autoCount++
	}
	if wantsAutoCreatedAt {
		autoCount++
	}
	var returning bool
	if d.Cap().InsertReturning && autoCount > 0 {
		returning = true
		var returningCols []string
		if wantsAutoId {
			returningCols = append(returningCols, d.QuoteIdent("id"))
		}
		if wantsAutoName {
			returningCols = append(returningCols, d.QuoteIdent("name"))
		}
		if wantsAutoEmail {
			returningCols = append(returningCols, d.QuoteIdent("email"))
		}
		if wantsAutoCreatedAt {
			returningCols = append(returningCols, d.QuoteIdent("created_at"))
		}
		if len(returningCols) > 0 {
			sqlStr = sqlStr + " RETURNING " + strings.Join(returningCols, ", ")
		}
	}
	if returning {
		rows, err := db.QueryContext(ctx, sqlStr, state.Args()...)
		if err != nil {
			return err
		}
		defer rows.Close()
		var scanTargets []any
		var retId uint64
		if wantsAutoId {
			scanTargets = append(scanTargets, &retId)
		}
		var retName string
		if wantsAutoName {
			scanTargets = append(scanTargets, &retName)
		}
		var retEmail string
		if wantsAutoEmail {
			scanTargets = append(scanTargets, &retEmail)
		}
		var retCreatedAt string
		if wantsAutoCreatedAt {
			scanTargets = append(scanTargets, &retCreatedAt)
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return fmt.Errorf("Author insert returned no rows")
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(retId)
		}
		if wantsAutoName {
			data.Name = opt.OVal(retName)
		}
		if wantsAutoEmail {
			data.Email = opt.OVal(retEmail)
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(retCreatedAt)
		}
		return nil
	}
	res, err := db.ExecContext(ctx, sqlStr, state.Args()...)
	if err != nil {
		return err
	}
	if autoCount == 1 {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(uint64(id))
		}
		if wantsAutoName {
			data.Name = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoEmail {
			data.Email = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(strconv.FormatInt(id, 10))
		}
	} else if autoCount > 1 {
		return fmt.Errorf("%s cannot populate multiple auto columns for author without RETURNING", d.Name())
	}
	return nil
}

func createAuthorRelations(ctx context.Context, db bom.Querier, d dialect.Dialect, data *AuthorCreateData) error {
	if data.AuthorProfile != nil {
		child := data.AuthorProfile
		if !child.AuthorId.IsSome() {
			if !data.Id.IsSome() {
				return fmt.Errorf("Author: missing id for relation AuthorProfile")
			}
			child.AuthorId = data.Id
		}
		if err := createAuthorProfileRecord(ctx, db, d, child); err != nil {
			return err
		}
		if err := createAuthorProfileRelations(ctx, db, d, child); err != nil {
			return err
		}
	}
	if len(data.Comment) > 0 {
		for i := range data.Comment {
			child := &data.Comment[i]
			if !child.AuthorId.IsSome() {
				if !data.Id.IsSome() {
					return fmt.Errorf("Author: missing id for relation Comment")
				}
				child.AuthorId = data.Id
			}
			if err := createCommentRecord(ctx, db, d, child); err != nil {
				return err
			}
			if err := createCommentRelations(ctx, db, d, child); err != nil {
				return err
			}
		}
	}
	if len(data.Video) > 0 {
		for i := range data.Video {
			child := &data.Video[i]
			if !child.AuthorId.IsSome() {
				if !data.Id.IsSome() {
					return fmt.Errorf("Author: missing id for relation Video")
				}
				child.AuthorId = data.Id
			}
			if err := createVideoRecord(ctx, db, d, child); err != nil {
				return err
			}
			if err := createVideoRelations(ctx, db, d, child); err != nil {
				return err
			}
		}
	}
	return nil
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
				item = normalizeJSONRaw(item)
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
		raw = normalizeJSONRaw(raw)
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

var AuthorProfileAllColumns = []string{
	"id",
	"author_id",
	"bio",
	"avatar_url",
	"created_at",
}

type AuthorProfile struct {
	Id uint64 `json:"id"`

	AuthorId uint64 `json:"author_id"`

	Bio opt.Opt[string] `json:"bio,omitempty"`

	AvatarUrl opt.Opt[string] `json:"avatar_url,omitempty"`

	CreatedAt string  `json:"created_at"`
	Author    *Author `json:"author,omitempty"`
}

type AuthorProfileModel interface {
	AuthorProfile
}

type AuthorProfileField string

const (
	AuthorProfileFieldUnknown   AuthorProfileField = ""
	AuthorProfileFieldId        AuthorProfileField = "id"
	AuthorProfileFieldAuthorId  AuthorProfileField = "author_id"
	AuthorProfileFieldBio       AuthorProfileField = "bio"
	AuthorProfileFieldAvatarUrl AuthorProfileField = "avatar_url"
	AuthorProfileFieldCreatedAt AuthorProfileField = "created_at"
)

type AuthorProfileWhereInput struct {
	AND       []AuthorProfileWhereInput
	OR        []AuthorProfileWhereInput
	NOT       []AuthorProfileWhereInput
	Id        opt.Opt[uint64]
	AuthorId  opt.Opt[uint64]
	Bio       opt.Opt[string]
	AvatarUrl opt.Opt[string]
	CreatedAt opt.Opt[string]
	Author    *AuthorProfileAuthorRelation
}

type AuthorProfileWhereUniqueInput struct {
	Id *uint64
}

type AuthorProfileOrderByInput struct {
	Field     AuthorProfileField
	Direction OrderDirection
}
type AuthorProfileAuthorRelation struct {
	Some  *AuthorWhereInput
	None  *AuthorWhereInput
	Every *AuthorWhereInput
}

type AuthorProfileSelect []AuthorProfileSelectItem

type AuthorProfileSelectItem interface {
	isAuthorProfileSelectItem()
}

func (AuthorProfileField) isAuthorProfileSelectItem() {}

type AuthorProfileSelectAuthor struct {
	Args AuthorProfileAuthorSelectArgs
}

func (AuthorProfileSelectAuthor) isAuthorProfileSelectItem() {}

var AuthorProfileSelectAll = AuthorProfileSelect{
	AuthorProfileFieldId,
	AuthorProfileFieldAuthorId,
	AuthorProfileFieldBio,
	AuthorProfileFieldAvatarUrl,
	AuthorProfileFieldCreatedAt,
}

type AuthorProfileAuthorSelectArgs struct {
	Select AuthorSelect
}

type AuthorProfileFindMany struct {
	Where    *AuthorProfileWhereInput
	OrderBy  []AuthorProfileOrderByInput
	Take     opt.Opt[int]
	Skip     opt.Opt[int]
	Distinct []AuthorProfileField
	Select   AuthorProfileSelect
}

type AuthorProfileFindFirst struct {
	Where   *AuthorProfileWhereInput
	OrderBy []AuthorProfileOrderByInput
	Skip    opt.Opt[int]
	Select  AuthorProfileSelect
}
type AuthorProfileCreateData struct {
	Id        opt.Opt[uint64]
	AuthorId  opt.Opt[uint64]
	Bio       opt.Opt[string]
	AvatarUrl opt.Opt[string]
	CreatedAt opt.Opt[string]
}

type AuthorProfileCreate struct {
	Data   AuthorProfileCreateData
	Select AuthorProfileSelect
}

type AuthorProfileCreateMany struct {
	Data []AuthorProfileCreateData
}

func buildAuthorProfileWhere(d dialect.Dialect, alias string, args *argState, where *AuthorProfileWhereInput) string {
	if where == nil {
		return ""
	}
	w := newWhereBuilder(d, alias, args)
	return buildAuthorProfileWherePredicates(w, where)
}

func buildAuthorProfileWherePredicates(w *whereBuilder, where *AuthorProfileWhereInput) string {
	var preds []string
	if where.Id.IsSome() {
		preds = append(preds, w.eq("id", where.Id.Value()))
	}
	if where.AuthorId.IsSome() {
		preds = append(preds, w.eq("author_id", where.AuthorId.Value()))
	}
	if where.Bio.IsSome() {
		preds = append(preds, w.eq("bio", where.Bio.Value()))
	}
	if where.AvatarUrl.IsSome() {
		preds = append(preds, w.eq("avatar_url", where.AvatarUrl.Value()))
	}
	if where.CreatedAt.IsSome() {
		preds = append(preds, w.eq("created_at", where.CreatedAt.Value()))
	}
	if where.Author != nil {
		if clause := buildAuthorProfileAuthorRelation(w.d, w.alias, w.args, where.Author); clause != "" {
			preds = append(preds, clause)
		}
	}
	for i := range where.AND {
		if clause := buildAuthorProfileWherePredicates(w, &where.AND[i]); clause != "" {
			preds = append(preds, clause)
		}
	}
	var orPreds []string
	for i := range where.OR {
		if clause := buildAuthorProfileWherePredicates(w, &where.OR[i]); clause != "" {
			orPreds = append(orPreds, clause)
		}
	}
	if orClause := w.combineOr(orPreds); orClause != "" {
		preds = append(preds, orClause)
	}
	for i := range where.NOT {
		if clause := buildAuthorProfileWherePredicates(w, &where.NOT[i]); clause != "" {
			preds = append(preds, w.d.Not(clause))
		}
	}
	return w.combineAnd(preds)
}
func buildAuthorProfileAuthorRelation(d dialect.Dialect, parentAlias string, args *argState, rel *AuthorProfileAuthorRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildAuthorProfileAuthorRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildAuthorProfileAuthorRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildAuthorProfileAuthorRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildAuthorProfileAuthorRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *AuthorWhereInput, negate bool) string {
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

func buildAuthorProfileAuthorRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *AuthorWhereInput) string {
	childAlias := parentAlias + "_author" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "author_id"))
	var negated string
	if whereClause := buildAuthorWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}

func buildAuthorProfileWhereUnique(d dialect.Dialect, alias string, args *argState, where *AuthorProfileWhereUniqueInput) string {
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

func buildAuthorProfileOrderBy(d dialect.Dialect, alias string, order []AuthorProfileOrderByInput) []string {
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

func buildAuthorProfileDistinct(d dialect.Dialect, alias string, distinct []AuthorProfileField) []string {
	var out []string
	for _, field := range distinct {
		if field == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(field))))
	}
	return out
}

func buildAuthorProfileJSONExpr(d dialect.Dialect, alias string, args *argState, sel AuthorProfileSelect) (string, error) {
	items := sel
	includeAll := len(items) == 0
	var pairs []string
	if includeAll {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
		pairs = append(pairs, jsonPair("author_id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("author_id"))))
		pairs = append(pairs, jsonPair("bio", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("bio"))))
		pairs = append(pairs, jsonPair("avatar_url", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("avatar_url"))))
		pairs = append(pairs, jsonPair("created_at", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("created_at"))))
		{
			expr, err := buildAuthorProfileAuthorJSON(d, alias, args, &AuthorProfileAuthorSelectArgs{
				Select: AuthorSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("author", wrapJSONValue(d, expr)))
		}
	} else {
		seen := make(map[AuthorProfileField]struct{})
		relationSeen := make(map[string]struct{})
		for _, item := range items {
			switch v := item.(type) {
			case AuthorProfileField:
				if v == AuthorProfileFieldUnknown {
					continue
				}
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				pairs = append(pairs, jsonPair(string(v), fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(v)))))
			case AuthorProfileSelectAuthor:
				if _, ok := relationSeen["Author"]; ok {
					continue
				}
				relationSeen["Author"] = struct{}{}
				{
					expr, err := buildAuthorProfileAuthorJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("author", wrapJSONValue(d, expr)))
				}
			}
		}
	}
	if len(pairs) == 0 {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
	}
	return d.JSONBuildObject(pairs...), nil
}
func buildAuthorProfileAuthorJSON(d dialect.Dialect, parentAlias string, args *argState, sel *AuthorProfileAuthorSelectArgs) (string, error) {
	if sel == nil {
		sel = &AuthorProfileAuthorSelectArgs{}
	}
	childAlias := parentAlias + "_author"
	childJSON, err := buildAuthorJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
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

func FindManyAuthorProfile[T AuthorProfileModel](ctx context.Context, db bom.Querier, q AuthorProfileFindMany) ([]T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	rootAlias := "t0"
	jsonExpr, err := buildAuthorProfileJSONExpr(d, rootAlias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	whereClause := buildAuthorProfileWhere(d, rootAlias, state, q.Where)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "author_profile",
		Alias: rootAlias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where:     whereClause,
		Args:      state.Args(),
		OrderBy:   buildAuthorProfileOrderBy(d, rootAlias, q.OrderBy),
		Distinct:  buildAuthorProfileDistinct(d, rootAlias, q.Distinct),
		Limit:     convertOpt(q.Take),
		Offset:    convertOpt(q.Skip),
		JSONArray: true,
	}
	return queryAuthorProfileRows[T](ctx, db, d, input, true)
}

func FindFirstAuthorProfile[T AuthorProfileModel](ctx context.Context, db bom.Querier, q AuthorProfileFindFirst) (*T, error) {
	fm := AuthorProfileFindMany{
		Where:   q.Where,
		OrderBy: q.OrderBy,
		Skip:    q.Skip,
		Select:  q.Select,
	}
	fm.Take = opt.OVal(1)
	rows, err := FindManyAuthorProfile[T](ctx, db, fm)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	out := rows[0]
	return &out, nil
}
func CreateOneAuthorProfile[T AuthorProfileModel](ctx context.Context, db bom.Querier, q AuthorProfileCreate) (*T, error) {
	d := dialectsqlite.New()
	data := q.Data
	if err := createAuthorProfileRecord(ctx, db, d, &data); err != nil {
		return nil, err
	}
	if err := createAuthorProfileRelations(ctx, db, d, &data); err != nil {
		return nil, err
	}
	var fetch func(context.Context, AuthorProfileSelect) (*T, error)
	if fetch == nil {
		if data.Id.IsSome() {
			lookup := AuthorProfileUK_Id{
				Id: data.Id.Value(),
			}
			fetch = func(ctx context.Context, sel AuthorProfileSelect) (*T, error) {
				return FindUniqueAuthorProfile[T](ctx, db, AuthorProfileFindUnique[AuthorProfileUK_Id]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		if data.AuthorId.IsSome() {
			lookup := AuthorProfileUK_AuthorId{
				AuthorId: data.AuthorId.Value(),
			}
			fetch = func(ctx context.Context, sel AuthorProfileSelect) (*T, error) {
				return FindUniqueAuthorProfile[T](ctx, db, AuthorProfileFindUnique[AuthorProfileUK_AuthorId]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		return nil, fmt.Errorf("AuthorProfileCreate requires values for a unique constraint")
	}
	return fetch(ctx, q.Select)
}

func CreateManyAuthorProfile(ctx context.Context, db bom.Querier, q AuthorProfileCreateMany) (int64, error) {
	d := dialectsqlite.New()
	if len(q.Data) == 0 {
		return 0, nil
	}
	var total int64
	for i := range q.Data {
		data := q.Data[i]
		if err := createAuthorProfileRecord(ctx, db, d, &data); err != nil {
			return total, err
		}
		if err := createAuthorProfileRelations(ctx, db, d, &data); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func createAuthorProfileRecord(ctx context.Context, db bom.Querier, d dialect.Dialect, data *AuthorProfileCreateData) error {
	state := newArgState(d)
	var columns []string
	var placeholders []string
	if identityGenerator == nil {
		identityGenerator = &bom.DefaultIdentityGenerator{}
	}
	var wantsAutoId bool
	var wantsAutoAuthorId bool
	var wantsAutoBio bool
	var wantsAutoAvatarUrl bool
	var wantsAutoCreatedAt bool
	if data.Id.IsSome() {
		val := data.Id.Value()
		columns = append(columns, d.QuoteIdent("id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
		wantsAutoId = true
	}
	if data.AuthorId.IsSome() {
		val := data.AuthorId.Value()
		columns = append(columns, d.QuoteIdent("author_id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.Bio.IsSome() {
		val := data.Bio.Value()
		columns = append(columns, d.QuoteIdent("bio"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.AvatarUrl.IsSome() {
		val := data.AvatarUrl.Value()
		columns = append(columns, d.QuoteIdent("avatar_url"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.CreatedAt.IsSome() {
		val := data.CreatedAt.Value()
		columns = append(columns, d.QuoteIdent("created_at"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return err
	}
	sqlStr := buildInsertSQL(d, "author_profile", columns, placeholders)
	autoCount := 0
	if wantsAutoId {
		autoCount++
	}
	if wantsAutoAuthorId {
		autoCount++
	}
	if wantsAutoBio {
		autoCount++
	}
	if wantsAutoAvatarUrl {
		autoCount++
	}
	if wantsAutoCreatedAt {
		autoCount++
	}
	var returning bool
	if d.Cap().InsertReturning && autoCount > 0 {
		returning = true
		var returningCols []string
		if wantsAutoId {
			returningCols = append(returningCols, d.QuoteIdent("id"))
		}
		if wantsAutoAuthorId {
			returningCols = append(returningCols, d.QuoteIdent("author_id"))
		}
		if wantsAutoBio {
			returningCols = append(returningCols, d.QuoteIdent("bio"))
		}
		if wantsAutoAvatarUrl {
			returningCols = append(returningCols, d.QuoteIdent("avatar_url"))
		}
		if wantsAutoCreatedAt {
			returningCols = append(returningCols, d.QuoteIdent("created_at"))
		}
		if len(returningCols) > 0 {
			sqlStr = sqlStr + " RETURNING " + strings.Join(returningCols, ", ")
		}
	}
	if returning {
		rows, err := db.QueryContext(ctx, sqlStr, state.Args()...)
		if err != nil {
			return err
		}
		defer rows.Close()
		var scanTargets []any
		var retId uint64
		if wantsAutoId {
			scanTargets = append(scanTargets, &retId)
		}
		var retAuthorId uint64
		if wantsAutoAuthorId {
			scanTargets = append(scanTargets, &retAuthorId)
		}
		var retBio string
		if wantsAutoBio {
			scanTargets = append(scanTargets, &retBio)
		}
		var retAvatarUrl string
		if wantsAutoAvatarUrl {
			scanTargets = append(scanTargets, &retAvatarUrl)
		}
		var retCreatedAt string
		if wantsAutoCreatedAt {
			scanTargets = append(scanTargets, &retCreatedAt)
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return fmt.Errorf("AuthorProfile insert returned no rows")
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(retId)
		}
		if wantsAutoAuthorId {
			data.AuthorId = opt.OVal(retAuthorId)
		}
		if wantsAutoBio {
			data.Bio = opt.OVal(retBio)
		}
		if wantsAutoAvatarUrl {
			data.AvatarUrl = opt.OVal(retAvatarUrl)
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(retCreatedAt)
		}
		return nil
	}
	res, err := db.ExecContext(ctx, sqlStr, state.Args()...)
	if err != nil {
		return err
	}
	if autoCount == 1 {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(uint64(id))
		}
		if wantsAutoAuthorId {
			data.AuthorId = opt.OVal(uint64(id))
		}
		if wantsAutoBio {
			data.Bio = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoAvatarUrl {
			data.AvatarUrl = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(strconv.FormatInt(id, 10))
		}
	} else if autoCount > 1 {
		return fmt.Errorf("%s cannot populate multiple auto columns for author_profile without RETURNING", d.Name())
	}
	return nil
}

func createAuthorProfileRelations(ctx context.Context, db bom.Querier, d dialect.Dialect, data *AuthorProfileCreateData) error {
	return nil
}

func queryAuthorProfileRows[T AuthorProfileModel](ctx context.Context, db bom.Querier, d dialect.Dialect, input planner.FindManyInput, expectArray bool) ([]T, error) {
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
				item = normalizeJSONRaw(item)
				var rec AuthorProfile
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
		raw = normalizeJSONRaw(raw)
		var rec AuthorProfile
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

	CreatedAt string     `json:"created_at"`
	Comment   []Comment  `json:"comment,omitempty"`
	VideoTag  []VideoTag `json:"videoTag,omitempty"`
	Author    *Author    `json:"author,omitempty"`
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
	Comment     *VideoCommentRelation
	VideoTag    *VideoVideoTagRelation
	Author      *VideoAuthorRelation
}

type VideoWhereUniqueInput struct {
	Id *uint64
}

type VideoOrderByInput struct {
	Field     VideoField
	Direction OrderDirection
}
type VideoCommentRelation struct {
	Some  *CommentWhereInput
	None  *CommentWhereInput
	Every *CommentWhereInput
}
type VideoVideoTagRelation struct {
	Some  *VideoTagWhereInput
	None  *VideoTagWhereInput
	Every *VideoTagWhereInput
}
type VideoAuthorRelation struct {
	Some  *AuthorWhereInput
	None  *AuthorWhereInput
	Every *AuthorWhereInput
}

type VideoSelect []VideoSelectItem

type VideoSelectItem interface {
	isVideoSelectItem()
}

func (VideoField) isVideoSelectItem() {}

type VideoSelectComment struct {
	Args VideoCommentSelectArgs
}

func (VideoSelectComment) isVideoSelectItem() {}

type VideoSelectVideoTag struct {
	Args VideoVideoTagSelectArgs
}

func (VideoSelectVideoTag) isVideoSelectItem() {}

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

type VideoCommentSelectArgs struct {
	Where   *CommentWhereInput
	OrderBy []CommentOrderByInput
	Take    opt.Opt[int]
	Skip    opt.Opt[int]
	Select  CommentSelect
}

type VideoVideoTagSelectArgs struct {
	Where   *VideoTagWhereInput
	OrderBy []VideoTagOrderByInput
	Take    opt.Opt[int]
	Skip    opt.Opt[int]
	Select  VideoTagSelect
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
type VideoCreateData struct {
	Id          opt.Opt[uint64]
	Title       opt.Opt[string]
	Slug        opt.Opt[string]
	AuthorId    opt.Opt[uint64]
	Description opt.Opt[string]
	CreatedAt   opt.Opt[string]
	Comment     []CommentCreateData
	VideoTag    []VideoTagCreateData
}

type VideoCreate struct {
	Data   VideoCreateData
	Select VideoSelect
}

type VideoCreateMany struct {
	Data []VideoCreateData
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
	if where.Comment != nil {
		if clause := buildVideoCommentRelation(w.d, w.alias, w.args, where.Comment); clause != "" {
			preds = append(preds, clause)
		}
	}
	if where.VideoTag != nil {
		if clause := buildVideoVideoTagRelation(w.d, w.alias, w.args, where.VideoTag); clause != "" {
			preds = append(preds, clause)
		}
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
func buildVideoCommentRelation(d dialect.Dialect, parentAlias string, args *argState, rel *VideoCommentRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildVideoCommentRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildVideoCommentRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildVideoCommentRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildVideoCommentRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *CommentWhereInput, negate bool) string {
	childAlias := parentAlias + "_comment"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("comment"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "video_id", parentAlias, "id"))
	var combined string
	if len(joinConds) > 0 {
		combined = strings.Join(joinConds, " AND ")
	}
	if whereClause := buildCommentWhere(d, childAlias, args, where); whereClause != "" {
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

func buildVideoCommentRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *CommentWhereInput) string {
	childAlias := parentAlias + "_comment" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("comment"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "video_id", parentAlias, "id"))
	var negated string
	if whereClause := buildCommentWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}
func buildVideoVideoTagRelation(d dialect.Dialect, parentAlias string, args *argState, rel *VideoVideoTagRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildVideoVideoTagRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildVideoVideoTagRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildVideoVideoTagRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildVideoVideoTagRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *VideoTagWhereInput, negate bool) string {
	childAlias := parentAlias + "_videoTag"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video_tag"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "video_id", parentAlias, "id"))
	var combined string
	if len(joinConds) > 0 {
		combined = strings.Join(joinConds, " AND ")
	}
	if whereClause := buildVideoTagWhere(d, childAlias, args, where); whereClause != "" {
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

func buildVideoVideoTagRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *VideoTagWhereInput) string {
	childAlias := parentAlias + "_videoTag" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video_tag"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "video_id", parentAlias, "id"))
	var negated string
	if whereClause := buildVideoTagWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
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
	if rel.Every != nil {
		if clause := buildVideoAuthorRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildVideoAuthorRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *AuthorWhereInput) string {
	childAlias := parentAlias + "_author" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "author_id"))
	var negated string
	if whereClause := buildAuthorWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
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
		{
			expr, err := buildVideoCommentJSON(d, alias, args, &VideoCommentSelectArgs{
				Select: CommentSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("comment", wrapJSONValue(d, expr)))
		}
		{
			expr, err := buildVideoVideoTagJSON(d, alias, args, &VideoVideoTagSelectArgs{
				Select: VideoTagSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("videoTag", wrapJSONValue(d, expr)))
		}
		{
			expr, err := buildVideoAuthorJSON(d, alias, args, &VideoAuthorSelectArgs{
				Select: AuthorSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("author", wrapJSONValue(d, expr)))
		}
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
			case VideoSelectComment:
				if _, ok := relationSeen["Comment"]; ok {
					continue
				}
				relationSeen["Comment"] = struct{}{}
				{
					expr, err := buildVideoCommentJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("comment", wrapJSONValue(d, expr)))
				}
			case VideoSelectVideoTag:
				if _, ok := relationSeen["VideoTag"]; ok {
					continue
				}
				relationSeen["VideoTag"] = struct{}{}
				{
					expr, err := buildVideoVideoTagJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("videoTag", wrapJSONValue(d, expr)))
				}
			case VideoSelectAuthor:
				if _, ok := relationSeen["Author"]; ok {
					continue
				}
				relationSeen["Author"] = struct{}{}
				{
					expr, err := buildVideoAuthorJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("author", wrapJSONValue(d, expr)))
				}
			}
		}
	}
	if len(pairs) == 0 {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
	}
	return d.JSONBuildObject(pairs...), nil
}
func buildVideoCommentJSON(d dialect.Dialect, parentAlias string, args *argState, sel *VideoCommentSelectArgs) (string, error) {
	if sel == nil {
		sel = &VideoCommentSelectArgs{}
	}
	childAlias := parentAlias + "_comment"
	childJSON, err := buildCommentJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "video_id", parentAlias, "id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = buildCommentWhere(d, childAlias, args, sel.Where)
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
	orderClause := buildCommentOrderBy(d, childAlias, sel.OrderBy)
	if len(orderClause) > 0 {
		orderSQL = " ORDER BY " + strings.Join(orderClause, ", ")
	}
	limit := convertOpt(sel.Take)
	offset := convertOpt(sel.Skip)
	lo = d.LimitOffset(limit, offset)
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("comment"), d.QuoteIdent(childAlias))
	arrayExpr := d.JSONArrayAgg(childJSON)
	expr := fmt.Sprintf("(SELECT %s FROM %s%s%s", d.CoalesceJSONAgg(arrayExpr, d.JSONArrayEmpty()), tableRef, whereSQL, orderSQL)
	if lo != "" {
		expr += " " + lo
	}
	expr += ")"
	return expr, nil
}
func buildVideoVideoTagJSON(d dialect.Dialect, parentAlias string, args *argState, sel *VideoVideoTagSelectArgs) (string, error) {
	if sel == nil {
		sel = &VideoVideoTagSelectArgs{}
	}
	childAlias := parentAlias + "_videoTag"
	childJSON, err := buildVideoTagJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "video_id", parentAlias, "id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = buildVideoTagWhere(d, childAlias, args, sel.Where)
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
	orderClause := buildVideoTagOrderBy(d, childAlias, sel.OrderBy)
	if len(orderClause) > 0 {
		orderSQL = " ORDER BY " + strings.Join(orderClause, ", ")
	}
	limit := convertOpt(sel.Take)
	offset := convertOpt(sel.Skip)
	lo = d.LimitOffset(limit, offset)
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video_tag"), d.QuoteIdent(childAlias))
	arrayExpr := d.JSONArrayAgg(childJSON)
	expr := fmt.Sprintf("(SELECT %s FROM %s%s%s", d.CoalesceJSONAgg(arrayExpr, d.JSONArrayEmpty()), tableRef, whereSQL, orderSQL)
	if lo != "" {
		expr += " " + lo
	}
	expr += ")"
	return expr, nil
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
	childJSON = wrapJSONValue(d, childJSON)
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
	d := dialectsqlite.New()
	state := newArgState(d)
	rootAlias := "t0"
	jsonExpr, err := buildVideoJSONExpr(d, rootAlias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	whereClause := buildVideoWhere(d, rootAlias, state, q.Where)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
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
func CreateOneVideo[T VideoModel](ctx context.Context, db bom.Querier, q VideoCreate) (*T, error) {
	d := dialectsqlite.New()
	data := q.Data
	if err := createVideoRecord(ctx, db, d, &data); err != nil {
		return nil, err
	}
	if err := createVideoRelations(ctx, db, d, &data); err != nil {
		return nil, err
	}
	var fetch func(context.Context, VideoSelect) (*T, error)
	if fetch == nil {
		if data.Id.IsSome() {
			lookup := VideoUK_Id{
				Id: data.Id.Value(),
			}
			fetch = func(ctx context.Context, sel VideoSelect) (*T, error) {
				return FindUniqueVideo[T](ctx, db, VideoFindUnique[VideoUK_Id]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		if data.Slug.IsSome() {
			lookup := VideoUK_Slug{
				Slug: data.Slug.Value(),
			}
			fetch = func(ctx context.Context, sel VideoSelect) (*T, error) {
				return FindUniqueVideo[T](ctx, db, VideoFindUnique[VideoUK_Slug]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		return nil, fmt.Errorf("VideoCreate requires values for a unique constraint")
	}
	return fetch(ctx, q.Select)
}

func CreateManyVideo(ctx context.Context, db bom.Querier, q VideoCreateMany) (int64, error) {
	d := dialectsqlite.New()
	if len(q.Data) == 0 {
		return 0, nil
	}
	var total int64
	for i := range q.Data {
		data := q.Data[i]
		if err := createVideoRecord(ctx, db, d, &data); err != nil {
			return total, err
		}
		if err := createVideoRelations(ctx, db, d, &data); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func createVideoRecord(ctx context.Context, db bom.Querier, d dialect.Dialect, data *VideoCreateData) error {
	state := newArgState(d)
	var columns []string
	var placeholders []string
	if identityGenerator == nil {
		identityGenerator = &bom.DefaultIdentityGenerator{}
	}
	var wantsAutoId bool
	var wantsAutoTitle bool
	var wantsAutoSlug bool
	var wantsAutoAuthorId bool
	var wantsAutoDescription bool
	var wantsAutoCreatedAt bool
	if data.Id.IsSome() {
		val := data.Id.Value()
		columns = append(columns, d.QuoteIdent("id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
		wantsAutoId = true
	}
	if data.Title.IsSome() {
		val := data.Title.Value()
		columns = append(columns, d.QuoteIdent("title"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.Slug.IsSome() {
		val := data.Slug.Value()
		columns = append(columns, d.QuoteIdent("slug"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.AuthorId.IsSome() {
		val := data.AuthorId.Value()
		columns = append(columns, d.QuoteIdent("author_id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.Description.IsSome() {
		val := data.Description.Value()
		columns = append(columns, d.QuoteIdent("description"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.CreatedAt.IsSome() {
		val := data.CreatedAt.Value()
		columns = append(columns, d.QuoteIdent("created_at"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return err
	}
	sqlStr := buildInsertSQL(d, "video", columns, placeholders)
	autoCount := 0
	if wantsAutoId {
		autoCount++
	}
	if wantsAutoTitle {
		autoCount++
	}
	if wantsAutoSlug {
		autoCount++
	}
	if wantsAutoAuthorId {
		autoCount++
	}
	if wantsAutoDescription {
		autoCount++
	}
	if wantsAutoCreatedAt {
		autoCount++
	}
	var returning bool
	if d.Cap().InsertReturning && autoCount > 0 {
		returning = true
		var returningCols []string
		if wantsAutoId {
			returningCols = append(returningCols, d.QuoteIdent("id"))
		}
		if wantsAutoTitle {
			returningCols = append(returningCols, d.QuoteIdent("title"))
		}
		if wantsAutoSlug {
			returningCols = append(returningCols, d.QuoteIdent("slug"))
		}
		if wantsAutoAuthorId {
			returningCols = append(returningCols, d.QuoteIdent("author_id"))
		}
		if wantsAutoDescription {
			returningCols = append(returningCols, d.QuoteIdent("description"))
		}
		if wantsAutoCreatedAt {
			returningCols = append(returningCols, d.QuoteIdent("created_at"))
		}
		if len(returningCols) > 0 {
			sqlStr = sqlStr + " RETURNING " + strings.Join(returningCols, ", ")
		}
	}
	if returning {
		rows, err := db.QueryContext(ctx, sqlStr, state.Args()...)
		if err != nil {
			return err
		}
		defer rows.Close()
		var scanTargets []any
		var retId uint64
		if wantsAutoId {
			scanTargets = append(scanTargets, &retId)
		}
		var retTitle string
		if wantsAutoTitle {
			scanTargets = append(scanTargets, &retTitle)
		}
		var retSlug string
		if wantsAutoSlug {
			scanTargets = append(scanTargets, &retSlug)
		}
		var retAuthorId uint64
		if wantsAutoAuthorId {
			scanTargets = append(scanTargets, &retAuthorId)
		}
		var retDescription string
		if wantsAutoDescription {
			scanTargets = append(scanTargets, &retDescription)
		}
		var retCreatedAt string
		if wantsAutoCreatedAt {
			scanTargets = append(scanTargets, &retCreatedAt)
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return fmt.Errorf("Video insert returned no rows")
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(retId)
		}
		if wantsAutoTitle {
			data.Title = opt.OVal(retTitle)
		}
		if wantsAutoSlug {
			data.Slug = opt.OVal(retSlug)
		}
		if wantsAutoAuthorId {
			data.AuthorId = opt.OVal(retAuthorId)
		}
		if wantsAutoDescription {
			data.Description = opt.OVal(retDescription)
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(retCreatedAt)
		}
		return nil
	}
	res, err := db.ExecContext(ctx, sqlStr, state.Args()...)
	if err != nil {
		return err
	}
	if autoCount == 1 {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(uint64(id))
		}
		if wantsAutoTitle {
			data.Title = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoSlug {
			data.Slug = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoAuthorId {
			data.AuthorId = opt.OVal(uint64(id))
		}
		if wantsAutoDescription {
			data.Description = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(strconv.FormatInt(id, 10))
		}
	} else if autoCount > 1 {
		return fmt.Errorf("%s cannot populate multiple auto columns for video without RETURNING", d.Name())
	}
	return nil
}

func createVideoRelations(ctx context.Context, db bom.Querier, d dialect.Dialect, data *VideoCreateData) error {
	if len(data.Comment) > 0 {
		for i := range data.Comment {
			child := &data.Comment[i]
			if !child.VideoId.IsSome() {
				if !data.Id.IsSome() {
					return fmt.Errorf("Video: missing id for relation Comment")
				}
				child.VideoId = data.Id
			}
			if !child.AuthorId.IsSome() && data.AuthorId.IsSome() {
				child.AuthorId = data.AuthorId
			}
			if err := createCommentRecord(ctx, db, d, child); err != nil {
				return err
			}
			if err := createCommentRelations(ctx, db, d, child); err != nil {
				return err
			}
		}
	}
	if len(data.VideoTag) > 0 {
		for i := range data.VideoTag {
			child := &data.VideoTag[i]
			if !child.VideoId.IsSome() {
				if !data.Id.IsSome() {
					return fmt.Errorf("Video: missing id for relation VideoTag")
				}
				child.VideoId = data.Id
			}
			if err := createVideoTagRecord(ctx, db, d, child); err != nil {
				return err
			}
			if err := createVideoTagRelations(ctx, db, d, child); err != nil {
				return err
			}
		}
	}
	return nil
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
				item = normalizeJSONRaw(item)
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
		raw = normalizeJSONRaw(raw)
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

var CommentAllColumns = []string{
	"id",
	"video_id",
	"author_id",
	"body",
	"created_at",
}

type Comment struct {
	Id uint64 `json:"id"`

	VideoId uint64 `json:"video_id"`

	AuthorId uint64 `json:"author_id"`

	Body string `json:"body"`

	CreatedAt string  `json:"created_at"`
	Video     *Video  `json:"video,omitempty"`
	Author    *Author `json:"author,omitempty"`
}

type CommentModel interface {
	Comment
}

type CommentField string

const (
	CommentFieldUnknown   CommentField = ""
	CommentFieldId        CommentField = "id"
	CommentFieldVideoId   CommentField = "video_id"
	CommentFieldAuthorId  CommentField = "author_id"
	CommentFieldBody      CommentField = "body"
	CommentFieldCreatedAt CommentField = "created_at"
)

type CommentWhereInput struct {
	AND       []CommentWhereInput
	OR        []CommentWhereInput
	NOT       []CommentWhereInput
	Id        opt.Opt[uint64]
	VideoId   opt.Opt[uint64]
	AuthorId  opt.Opt[uint64]
	Body      opt.Opt[string]
	CreatedAt opt.Opt[string]
	Video     *CommentVideoRelation
	Author    *CommentAuthorRelation
}

type CommentWhereUniqueInput struct {
	Id *uint64
}

type CommentOrderByInput struct {
	Field     CommentField
	Direction OrderDirection
}
type CommentVideoRelation struct {
	Some  *VideoWhereInput
	None  *VideoWhereInput
	Every *VideoWhereInput
}
type CommentAuthorRelation struct {
	Some  *AuthorWhereInput
	None  *AuthorWhereInput
	Every *AuthorWhereInput
}

type CommentSelect []CommentSelectItem

type CommentSelectItem interface {
	isCommentSelectItem()
}

func (CommentField) isCommentSelectItem() {}

type CommentSelectVideo struct {
	Args CommentVideoSelectArgs
}

func (CommentSelectVideo) isCommentSelectItem() {}

type CommentSelectAuthor struct {
	Args CommentAuthorSelectArgs
}

func (CommentSelectAuthor) isCommentSelectItem() {}

var CommentSelectAll = CommentSelect{
	CommentFieldId,
	CommentFieldVideoId,
	CommentFieldAuthorId,
	CommentFieldBody,
	CommentFieldCreatedAt,
}

type CommentVideoSelectArgs struct {
	Select VideoSelect
}

type CommentAuthorSelectArgs struct {
	Select AuthorSelect
}

type CommentFindMany struct {
	Where    *CommentWhereInput
	OrderBy  []CommentOrderByInput
	Take     opt.Opt[int]
	Skip     opt.Opt[int]
	Distinct []CommentField
	Select   CommentSelect
}

type CommentFindFirst struct {
	Where   *CommentWhereInput
	OrderBy []CommentOrderByInput
	Skip    opt.Opt[int]
	Select  CommentSelect
}
type CommentCreateData struct {
	Id        opt.Opt[uint64]
	VideoId   opt.Opt[uint64]
	AuthorId  opt.Opt[uint64]
	Body      opt.Opt[string]
	CreatedAt opt.Opt[string]
}

type CommentCreate struct {
	Data   CommentCreateData
	Select CommentSelect
}

type CommentCreateMany struct {
	Data []CommentCreateData
}

func buildCommentWhere(d dialect.Dialect, alias string, args *argState, where *CommentWhereInput) string {
	if where == nil {
		return ""
	}
	w := newWhereBuilder(d, alias, args)
	return buildCommentWherePredicates(w, where)
}

func buildCommentWherePredicates(w *whereBuilder, where *CommentWhereInput) string {
	var preds []string
	if where.Id.IsSome() {
		preds = append(preds, w.eq("id", where.Id.Value()))
	}
	if where.VideoId.IsSome() {
		preds = append(preds, w.eq("video_id", where.VideoId.Value()))
	}
	if where.AuthorId.IsSome() {
		preds = append(preds, w.eq("author_id", where.AuthorId.Value()))
	}
	if where.Body.IsSome() {
		preds = append(preds, w.eq("body", where.Body.Value()))
	}
	if where.CreatedAt.IsSome() {
		preds = append(preds, w.eq("created_at", where.CreatedAt.Value()))
	}
	if where.Video != nil {
		if clause := buildCommentVideoRelation(w.d, w.alias, w.args, where.Video); clause != "" {
			preds = append(preds, clause)
		}
	}
	if where.Author != nil {
		if clause := buildCommentAuthorRelation(w.d, w.alias, w.args, where.Author); clause != "" {
			preds = append(preds, clause)
		}
	}
	for i := range where.AND {
		if clause := buildCommentWherePredicates(w, &where.AND[i]); clause != "" {
			preds = append(preds, clause)
		}
	}
	var orPreds []string
	for i := range where.OR {
		if clause := buildCommentWherePredicates(w, &where.OR[i]); clause != "" {
			orPreds = append(orPreds, clause)
		}
	}
	if orClause := w.combineOr(orPreds); orClause != "" {
		preds = append(preds, orClause)
	}
	for i := range where.NOT {
		if clause := buildCommentWherePredicates(w, &where.NOT[i]); clause != "" {
			preds = append(preds, w.d.Not(clause))
		}
	}
	return w.combineAnd(preds)
}
func buildCommentVideoRelation(d dialect.Dialect, parentAlias string, args *argState, rel *CommentVideoRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildCommentVideoRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildCommentVideoRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildCommentVideoRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildCommentVideoRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *VideoWhereInput, negate bool) string {
	childAlias := parentAlias + "_video"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "video_id"))
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

func buildCommentVideoRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *VideoWhereInput) string {
	childAlias := parentAlias + "_video" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "video_id"))
	var negated string
	if whereClause := buildVideoWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}
func buildCommentAuthorRelation(d dialect.Dialect, parentAlias string, args *argState, rel *CommentAuthorRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildCommentAuthorRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildCommentAuthorRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildCommentAuthorRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildCommentAuthorRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *AuthorWhereInput, negate bool) string {
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

func buildCommentAuthorRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *AuthorWhereInput) string {
	childAlias := parentAlias + "_author" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("author"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "author_id"))
	var negated string
	if whereClause := buildAuthorWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}

func buildCommentWhereUnique(d dialect.Dialect, alias string, args *argState, where *CommentWhereUniqueInput) string {
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

func buildCommentOrderBy(d dialect.Dialect, alias string, order []CommentOrderByInput) []string {
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

func buildCommentDistinct(d dialect.Dialect, alias string, distinct []CommentField) []string {
	var out []string
	for _, field := range distinct {
		if field == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(field))))
	}
	return out
}

func buildCommentJSONExpr(d dialect.Dialect, alias string, args *argState, sel CommentSelect) (string, error) {
	items := sel
	includeAll := len(items) == 0
	var pairs []string
	if includeAll {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
		pairs = append(pairs, jsonPair("video_id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("video_id"))))
		pairs = append(pairs, jsonPair("author_id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("author_id"))))
		pairs = append(pairs, jsonPair("body", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("body"))))
		pairs = append(pairs, jsonPair("created_at", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("created_at"))))
		{
			expr, err := buildCommentVideoJSON(d, alias, args, &CommentVideoSelectArgs{
				Select: VideoSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("video", wrapJSONValue(d, expr)))
		}
		{
			expr, err := buildCommentAuthorJSON(d, alias, args, &CommentAuthorSelectArgs{
				Select: AuthorSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("author", wrapJSONValue(d, expr)))
		}
	} else {
		seen := make(map[CommentField]struct{})
		relationSeen := make(map[string]struct{})
		for _, item := range items {
			switch v := item.(type) {
			case CommentField:
				if v == CommentFieldUnknown {
					continue
				}
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				pairs = append(pairs, jsonPair(string(v), fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(v)))))
			case CommentSelectVideo:
				if _, ok := relationSeen["Video"]; ok {
					continue
				}
				relationSeen["Video"] = struct{}{}
				{
					expr, err := buildCommentVideoJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("video", wrapJSONValue(d, expr)))
				}
			case CommentSelectAuthor:
				if _, ok := relationSeen["Author"]; ok {
					continue
				}
				relationSeen["Author"] = struct{}{}
				{
					expr, err := buildCommentAuthorJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("author", wrapJSONValue(d, expr)))
				}
			}
		}
	}
	if len(pairs) == 0 {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
	}
	return d.JSONBuildObject(pairs...), nil
}
func buildCommentVideoJSON(d dialect.Dialect, parentAlias string, args *argState, sel *CommentVideoSelectArgs) (string, error) {
	if sel == nil {
		sel = &CommentVideoSelectArgs{}
	}
	childAlias := parentAlias + "_video"
	childJSON, err := buildVideoJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "video_id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = joinClause
	var whereSQL string
	if whereClause != "" {
		whereSQL = " WHERE " + whereClause
	}
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
	expr := fmt.Sprintf("(SELECT %s FROM %s%s", childJSON, tableRef, whereSQL)
	expr += ")"
	return expr, nil
}
func buildCommentAuthorJSON(d dialect.Dialect, parentAlias string, args *argState, sel *CommentAuthorSelectArgs) (string, error) {
	if sel == nil {
		sel = &CommentAuthorSelectArgs{}
	}
	childAlias := parentAlias + "_author"
	childJSON, err := buildAuthorJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
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

func FindManyComment[T CommentModel](ctx context.Context, db bom.Querier, q CommentFindMany) ([]T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	rootAlias := "t0"
	jsonExpr, err := buildCommentJSONExpr(d, rootAlias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	whereClause := buildCommentWhere(d, rootAlias, state, q.Where)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "comment",
		Alias: rootAlias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where:     whereClause,
		Args:      state.Args(),
		OrderBy:   buildCommentOrderBy(d, rootAlias, q.OrderBy),
		Distinct:  buildCommentDistinct(d, rootAlias, q.Distinct),
		Limit:     convertOpt(q.Take),
		Offset:    convertOpt(q.Skip),
		JSONArray: true,
	}
	return queryCommentRows[T](ctx, db, d, input, true)
}

func FindFirstComment[T CommentModel](ctx context.Context, db bom.Querier, q CommentFindFirst) (*T, error) {
	fm := CommentFindMany{
		Where:   q.Where,
		OrderBy: q.OrderBy,
		Skip:    q.Skip,
		Select:  q.Select,
	}
	fm.Take = opt.OVal(1)
	rows, err := FindManyComment[T](ctx, db, fm)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	out := rows[0]
	return &out, nil
}
func CreateOneComment[T CommentModel](ctx context.Context, db bom.Querier, q CommentCreate) (*T, error) {
	d := dialectsqlite.New()
	data := q.Data
	if err := createCommentRecord(ctx, db, d, &data); err != nil {
		return nil, err
	}
	if err := createCommentRelations(ctx, db, d, &data); err != nil {
		return nil, err
	}
	var fetch func(context.Context, CommentSelect) (*T, error)
	if fetch == nil {
		if data.Id.IsSome() {
			lookup := CommentUK_Id{
				Id: data.Id.Value(),
			}
			fetch = func(ctx context.Context, sel CommentSelect) (*T, error) {
				return FindUniqueComment[T](ctx, db, CommentFindUnique[CommentUK_Id]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		return nil, fmt.Errorf("CommentCreate requires values for a unique constraint")
	}
	return fetch(ctx, q.Select)
}

func CreateManyComment(ctx context.Context, db bom.Querier, q CommentCreateMany) (int64, error) {
	d := dialectsqlite.New()
	if len(q.Data) == 0 {
		return 0, nil
	}
	var total int64
	for i := range q.Data {
		data := q.Data[i]
		if err := createCommentRecord(ctx, db, d, &data); err != nil {
			return total, err
		}
		if err := createCommentRelations(ctx, db, d, &data); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func createCommentRecord(ctx context.Context, db bom.Querier, d dialect.Dialect, data *CommentCreateData) error {
	state := newArgState(d)
	var columns []string
	var placeholders []string
	if identityGenerator == nil {
		identityGenerator = &bom.DefaultIdentityGenerator{}
	}
	var wantsAutoId bool
	var wantsAutoVideoId bool
	var wantsAutoAuthorId bool
	var wantsAutoBody bool
	var wantsAutoCreatedAt bool
	if data.Id.IsSome() {
		val := data.Id.Value()
		columns = append(columns, d.QuoteIdent("id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
		wantsAutoId = true
	}
	if data.VideoId.IsSome() {
		val := data.VideoId.Value()
		columns = append(columns, d.QuoteIdent("video_id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.AuthorId.IsSome() {
		val := data.AuthorId.Value()
		columns = append(columns, d.QuoteIdent("author_id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.Body.IsSome() {
		val := data.Body.Value()
		columns = append(columns, d.QuoteIdent("body"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.CreatedAt.IsSome() {
		val := data.CreatedAt.Value()
		columns = append(columns, d.QuoteIdent("created_at"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return err
	}
	sqlStr := buildInsertSQL(d, "comment", columns, placeholders)
	autoCount := 0
	if wantsAutoId {
		autoCount++
	}
	if wantsAutoVideoId {
		autoCount++
	}
	if wantsAutoAuthorId {
		autoCount++
	}
	if wantsAutoBody {
		autoCount++
	}
	if wantsAutoCreatedAt {
		autoCount++
	}
	var returning bool
	if d.Cap().InsertReturning && autoCount > 0 {
		returning = true
		var returningCols []string
		if wantsAutoId {
			returningCols = append(returningCols, d.QuoteIdent("id"))
		}
		if wantsAutoVideoId {
			returningCols = append(returningCols, d.QuoteIdent("video_id"))
		}
		if wantsAutoAuthorId {
			returningCols = append(returningCols, d.QuoteIdent("author_id"))
		}
		if wantsAutoBody {
			returningCols = append(returningCols, d.QuoteIdent("body"))
		}
		if wantsAutoCreatedAt {
			returningCols = append(returningCols, d.QuoteIdent("created_at"))
		}
		if len(returningCols) > 0 {
			sqlStr = sqlStr + " RETURNING " + strings.Join(returningCols, ", ")
		}
	}
	if returning {
		rows, err := db.QueryContext(ctx, sqlStr, state.Args()...)
		if err != nil {
			return err
		}
		defer rows.Close()
		var scanTargets []any
		var retId uint64
		if wantsAutoId {
			scanTargets = append(scanTargets, &retId)
		}
		var retVideoId uint64
		if wantsAutoVideoId {
			scanTargets = append(scanTargets, &retVideoId)
		}
		var retAuthorId uint64
		if wantsAutoAuthorId {
			scanTargets = append(scanTargets, &retAuthorId)
		}
		var retBody string
		if wantsAutoBody {
			scanTargets = append(scanTargets, &retBody)
		}
		var retCreatedAt string
		if wantsAutoCreatedAt {
			scanTargets = append(scanTargets, &retCreatedAt)
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return fmt.Errorf("Comment insert returned no rows")
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(retId)
		}
		if wantsAutoVideoId {
			data.VideoId = opt.OVal(retVideoId)
		}
		if wantsAutoAuthorId {
			data.AuthorId = opt.OVal(retAuthorId)
		}
		if wantsAutoBody {
			data.Body = opt.OVal(retBody)
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(retCreatedAt)
		}
		return nil
	}
	res, err := db.ExecContext(ctx, sqlStr, state.Args()...)
	if err != nil {
		return err
	}
	if autoCount == 1 {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(uint64(id))
		}
		if wantsAutoVideoId {
			data.VideoId = opt.OVal(uint64(id))
		}
		if wantsAutoAuthorId {
			data.AuthorId = opt.OVal(uint64(id))
		}
		if wantsAutoBody {
			data.Body = opt.OVal(strconv.FormatInt(id, 10))
		}
		if wantsAutoCreatedAt {
			data.CreatedAt = opt.OVal(strconv.FormatInt(id, 10))
		}
	} else if autoCount > 1 {
		return fmt.Errorf("%s cannot populate multiple auto columns for comment without RETURNING", d.Name())
	}
	return nil
}

func createCommentRelations(ctx context.Context, db bom.Querier, d dialect.Dialect, data *CommentCreateData) error {
	return nil
}

func queryCommentRows[T CommentModel](ctx context.Context, db bom.Querier, d dialect.Dialect, input planner.FindManyInput, expectArray bool) ([]T, error) {
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
				item = normalizeJSONRaw(item)
				var rec Comment
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
		raw = normalizeJSONRaw(raw)
		var rec Comment
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

var TagAllColumns = []string{
	"id",
	"name",
}

type Tag struct {
	Id uint64 `json:"id"`

	Name     string     `json:"name"`
	VideoTag []VideoTag `json:"videoTag,omitempty"`
}

type TagModel interface {
	Tag
}

type TagField string

const (
	TagFieldUnknown TagField = ""
	TagFieldId      TagField = "id"
	TagFieldName    TagField = "name"
)

type TagWhereInput struct {
	AND      []TagWhereInput
	OR       []TagWhereInput
	NOT      []TagWhereInput
	Id       opt.Opt[uint64]
	Name     opt.Opt[string]
	VideoTag *TagVideoTagRelation
}

type TagWhereUniqueInput struct {
	Id *uint64
}

type TagOrderByInput struct {
	Field     TagField
	Direction OrderDirection
}
type TagVideoTagRelation struct {
	Some  *VideoTagWhereInput
	None  *VideoTagWhereInput
	Every *VideoTagWhereInput
}

type TagSelect []TagSelectItem

type TagSelectItem interface {
	isTagSelectItem()
}

func (TagField) isTagSelectItem() {}

type TagSelectVideoTag struct {
	Args TagVideoTagSelectArgs
}

func (TagSelectVideoTag) isTagSelectItem() {}

var TagSelectAll = TagSelect{
	TagFieldId,
	TagFieldName,
}

type TagVideoTagSelectArgs struct {
	Where   *VideoTagWhereInput
	OrderBy []VideoTagOrderByInput
	Take    opt.Opt[int]
	Skip    opt.Opt[int]
	Select  VideoTagSelect
}

type TagFindMany struct {
	Where    *TagWhereInput
	OrderBy  []TagOrderByInput
	Take     opt.Opt[int]
	Skip     opt.Opt[int]
	Distinct []TagField
	Select   TagSelect
}

type TagFindFirst struct {
	Where   *TagWhereInput
	OrderBy []TagOrderByInput
	Skip    opt.Opt[int]
	Select  TagSelect
}
type TagCreateData struct {
	Id       opt.Opt[uint64]
	Name     opt.Opt[string]
	VideoTag []VideoTagCreateData
}

type TagCreate struct {
	Data   TagCreateData
	Select TagSelect
}

type TagCreateMany struct {
	Data []TagCreateData
}

func buildTagWhere(d dialect.Dialect, alias string, args *argState, where *TagWhereInput) string {
	if where == nil {
		return ""
	}
	w := newWhereBuilder(d, alias, args)
	return buildTagWherePredicates(w, where)
}

func buildTagWherePredicates(w *whereBuilder, where *TagWhereInput) string {
	var preds []string
	if where.Id.IsSome() {
		preds = append(preds, w.eq("id", where.Id.Value()))
	}
	if where.Name.IsSome() {
		preds = append(preds, w.eq("name", where.Name.Value()))
	}
	if where.VideoTag != nil {
		if clause := buildTagVideoTagRelation(w.d, w.alias, w.args, where.VideoTag); clause != "" {
			preds = append(preds, clause)
		}
	}
	for i := range where.AND {
		if clause := buildTagWherePredicates(w, &where.AND[i]); clause != "" {
			preds = append(preds, clause)
		}
	}
	var orPreds []string
	for i := range where.OR {
		if clause := buildTagWherePredicates(w, &where.OR[i]); clause != "" {
			orPreds = append(orPreds, clause)
		}
	}
	if orClause := w.combineOr(orPreds); orClause != "" {
		preds = append(preds, orClause)
	}
	for i := range where.NOT {
		if clause := buildTagWherePredicates(w, &where.NOT[i]); clause != "" {
			preds = append(preds, w.d.Not(clause))
		}
	}
	return w.combineAnd(preds)
}
func buildTagVideoTagRelation(d dialect.Dialect, parentAlias string, args *argState, rel *TagVideoTagRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildTagVideoTagRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildTagVideoTagRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildTagVideoTagRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildTagVideoTagRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *VideoTagWhereInput, negate bool) string {
	childAlias := parentAlias + "_videoTag"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video_tag"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "tag_id", parentAlias, "id"))
	var combined string
	if len(joinConds) > 0 {
		combined = strings.Join(joinConds, " AND ")
	}
	if whereClause := buildVideoTagWhere(d, childAlias, args, where); whereClause != "" {
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

func buildTagVideoTagRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *VideoTagWhereInput) string {
	childAlias := parentAlias + "_videoTag" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video_tag"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "tag_id", parentAlias, "id"))
	var negated string
	if whereClause := buildVideoTagWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}

func buildTagWhereUnique(d dialect.Dialect, alias string, args *argState, where *TagWhereUniqueInput) string {
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

func buildTagOrderBy(d dialect.Dialect, alias string, order []TagOrderByInput) []string {
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

func buildTagDistinct(d dialect.Dialect, alias string, distinct []TagField) []string {
	var out []string
	for _, field := range distinct {
		if field == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(field))))
	}
	return out
}

func buildTagJSONExpr(d dialect.Dialect, alias string, args *argState, sel TagSelect) (string, error) {
	items := sel
	includeAll := len(items) == 0
	var pairs []string
	if includeAll {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
		pairs = append(pairs, jsonPair("name", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("name"))))
		{
			expr, err := buildTagVideoTagJSON(d, alias, args, &TagVideoTagSelectArgs{
				Select: VideoTagSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("videoTag", wrapJSONValue(d, expr)))
		}
	} else {
		seen := make(map[TagField]struct{})
		relationSeen := make(map[string]struct{})
		for _, item := range items {
			switch v := item.(type) {
			case TagField:
				if v == TagFieldUnknown {
					continue
				}
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				pairs = append(pairs, jsonPair(string(v), fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(v)))))
			case TagSelectVideoTag:
				if _, ok := relationSeen["VideoTag"]; ok {
					continue
				}
				relationSeen["VideoTag"] = struct{}{}
				{
					expr, err := buildTagVideoTagJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("videoTag", wrapJSONValue(d, expr)))
				}
			}
		}
	}
	if len(pairs) == 0 {
		pairs = append(pairs, jsonPair("id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("id"))))
	}
	return d.JSONBuildObject(pairs...), nil
}
func buildTagVideoTagJSON(d dialect.Dialect, parentAlias string, args *argState, sel *TagVideoTagSelectArgs) (string, error) {
	if sel == nil {
		sel = &TagVideoTagSelectArgs{}
	}
	childAlias := parentAlias + "_videoTag"
	childJSON, err := buildVideoTagJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "tag_id", parentAlias, "id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = buildVideoTagWhere(d, childAlias, args, sel.Where)
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
	orderClause := buildVideoTagOrderBy(d, childAlias, sel.OrderBy)
	if len(orderClause) > 0 {
		orderSQL = " ORDER BY " + strings.Join(orderClause, ", ")
	}
	limit := convertOpt(sel.Take)
	offset := convertOpt(sel.Skip)
	lo = d.LimitOffset(limit, offset)
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video_tag"), d.QuoteIdent(childAlias))
	arrayExpr := d.JSONArrayAgg(childJSON)
	expr := fmt.Sprintf("(SELECT %s FROM %s%s%s", d.CoalesceJSONAgg(arrayExpr, d.JSONArrayEmpty()), tableRef, whereSQL, orderSQL)
	if lo != "" {
		expr += " " + lo
	}
	expr += ")"
	return expr, nil
}

func FindManyTag[T TagModel](ctx context.Context, db bom.Querier, q TagFindMany) ([]T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	rootAlias := "t0"
	jsonExpr, err := buildTagJSONExpr(d, rootAlias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	whereClause := buildTagWhere(d, rootAlias, state, q.Where)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "tag",
		Alias: rootAlias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where:     whereClause,
		Args:      state.Args(),
		OrderBy:   buildTagOrderBy(d, rootAlias, q.OrderBy),
		Distinct:  buildTagDistinct(d, rootAlias, q.Distinct),
		Limit:     convertOpt(q.Take),
		Offset:    convertOpt(q.Skip),
		JSONArray: true,
	}
	return queryTagRows[T](ctx, db, d, input, true)
}

func FindFirstTag[T TagModel](ctx context.Context, db bom.Querier, q TagFindFirst) (*T, error) {
	fm := TagFindMany{
		Where:   q.Where,
		OrderBy: q.OrderBy,
		Skip:    q.Skip,
		Select:  q.Select,
	}
	fm.Take = opt.OVal(1)
	rows, err := FindManyTag[T](ctx, db, fm)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	out := rows[0]
	return &out, nil
}
func CreateOneTag[T TagModel](ctx context.Context, db bom.Querier, q TagCreate) (*T, error) {
	d := dialectsqlite.New()
	data := q.Data
	if err := createTagRecord(ctx, db, d, &data); err != nil {
		return nil, err
	}
	if err := createTagRelations(ctx, db, d, &data); err != nil {
		return nil, err
	}
	var fetch func(context.Context, TagSelect) (*T, error)
	if fetch == nil {
		if data.Id.IsSome() {
			lookup := TagUK_Id{
				Id: data.Id.Value(),
			}
			fetch = func(ctx context.Context, sel TagSelect) (*T, error) {
				return FindUniqueTag[T](ctx, db, TagFindUnique[TagUK_Id]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		if data.Name.IsSome() {
			lookup := TagUK_Name{
				Name: data.Name.Value(),
			}
			fetch = func(ctx context.Context, sel TagSelect) (*T, error) {
				return FindUniqueTag[T](ctx, db, TagFindUnique[TagUK_Name]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		return nil, fmt.Errorf("TagCreate requires values for a unique constraint")
	}
	return fetch(ctx, q.Select)
}

func CreateManyTag(ctx context.Context, db bom.Querier, q TagCreateMany) (int64, error) {
	d := dialectsqlite.New()
	if len(q.Data) == 0 {
		return 0, nil
	}
	var total int64
	for i := range q.Data {
		data := q.Data[i]
		if err := createTagRecord(ctx, db, d, &data); err != nil {
			return total, err
		}
		if err := createTagRelations(ctx, db, d, &data); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func createTagRecord(ctx context.Context, db bom.Querier, d dialect.Dialect, data *TagCreateData) error {
	state := newArgState(d)
	var columns []string
	var placeholders []string
	if identityGenerator == nil {
		identityGenerator = &bom.DefaultIdentityGenerator{}
	}
	var wantsAutoId bool
	var wantsAutoName bool
	if data.Id.IsSome() {
		val := data.Id.Value()
		columns = append(columns, d.QuoteIdent("id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
		wantsAutoId = true
	}
	if data.Name.IsSome() {
		val := data.Name.Value()
		columns = append(columns, d.QuoteIdent("name"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return err
	}
	sqlStr := buildInsertSQL(d, "tag", columns, placeholders)
	autoCount := 0
	if wantsAutoId {
		autoCount++
	}
	if wantsAutoName {
		autoCount++
	}
	var returning bool
	if d.Cap().InsertReturning && autoCount > 0 {
		returning = true
		var returningCols []string
		if wantsAutoId {
			returningCols = append(returningCols, d.QuoteIdent("id"))
		}
		if wantsAutoName {
			returningCols = append(returningCols, d.QuoteIdent("name"))
		}
		if len(returningCols) > 0 {
			sqlStr = sqlStr + " RETURNING " + strings.Join(returningCols, ", ")
		}
	}
	if returning {
		rows, err := db.QueryContext(ctx, sqlStr, state.Args()...)
		if err != nil {
			return err
		}
		defer rows.Close()
		var scanTargets []any
		var retId uint64
		if wantsAutoId {
			scanTargets = append(scanTargets, &retId)
		}
		var retName string
		if wantsAutoName {
			scanTargets = append(scanTargets, &retName)
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return fmt.Errorf("Tag insert returned no rows")
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(retId)
		}
		if wantsAutoName {
			data.Name = opt.OVal(retName)
		}
		return nil
	}
	res, err := db.ExecContext(ctx, sqlStr, state.Args()...)
	if err != nil {
		return err
	}
	if autoCount == 1 {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if wantsAutoId {
			data.Id = opt.OVal(uint64(id))
		}
		if wantsAutoName {
			data.Name = opt.OVal(strconv.FormatInt(id, 10))
		}
	} else if autoCount > 1 {
		return fmt.Errorf("%s cannot populate multiple auto columns for tag without RETURNING", d.Name())
	}
	return nil
}

func createTagRelations(ctx context.Context, db bom.Querier, d dialect.Dialect, data *TagCreateData) error {
	if len(data.VideoTag) > 0 {
		for i := range data.VideoTag {
			child := &data.VideoTag[i]
			if !child.TagId.IsSome() {
				if !data.Id.IsSome() {
					return fmt.Errorf("Tag: missing id for relation VideoTag")
				}
				child.TagId = data.Id
			}
			if err := createVideoTagRecord(ctx, db, d, child); err != nil {
				return err
			}
			if err := createVideoTagRelations(ctx, db, d, child); err != nil {
				return err
			}
		}
	}
	return nil
}

func queryTagRows[T TagModel](ctx context.Context, db bom.Querier, d dialect.Dialect, input planner.FindManyInput, expectArray bool) ([]T, error) {
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
				item = normalizeJSONRaw(item)
				var rec Tag
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
		raw = normalizeJSONRaw(raw)
		var rec Tag
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

var VideoTagAllColumns = []string{
	"video_id",
	"tag_id",
}

type VideoTag struct {
	VideoId uint64 `json:"video_id"`

	TagId uint64 `json:"tag_id"`
	Video *Video `json:"video,omitempty"`
	Tag   *Tag   `json:"tag,omitempty"`
}

type VideoTagModel interface {
	VideoTag
}

type VideoTagField string

const (
	VideoTagFieldUnknown VideoTagField = ""
	VideoTagFieldVideoId VideoTagField = "video_id"
	VideoTagFieldTagId   VideoTagField = "tag_id"
)

type VideoTagWhereInput struct {
	AND     []VideoTagWhereInput
	OR      []VideoTagWhereInput
	NOT     []VideoTagWhereInput
	VideoId opt.Opt[uint64]
	TagId   opt.Opt[uint64]
	Video   *VideoTagVideoRelation
	Tag     *VideoTagTagRelation
}

type VideoTagWhereUniqueInput struct {
	VideoId *uint64
	TagId   *uint64
}

type VideoTagOrderByInput struct {
	Field     VideoTagField
	Direction OrderDirection
}
type VideoTagVideoRelation struct {
	Some  *VideoWhereInput
	None  *VideoWhereInput
	Every *VideoWhereInput
}
type VideoTagTagRelation struct {
	Some  *TagWhereInput
	None  *TagWhereInput
	Every *TagWhereInput
}

type VideoTagSelect []VideoTagSelectItem

type VideoTagSelectItem interface {
	isVideoTagSelectItem()
}

func (VideoTagField) isVideoTagSelectItem() {}

type VideoTagSelectVideo struct {
	Args VideoTagVideoSelectArgs
}

func (VideoTagSelectVideo) isVideoTagSelectItem() {}

type VideoTagSelectTag struct {
	Args VideoTagTagSelectArgs
}

func (VideoTagSelectTag) isVideoTagSelectItem() {}

var VideoTagSelectAll = VideoTagSelect{
	VideoTagFieldVideoId,
	VideoTagFieldTagId,
}

type VideoTagVideoSelectArgs struct {
	Select VideoSelect
}

type VideoTagTagSelectArgs struct {
	Select TagSelect
}

type VideoTagFindMany struct {
	Where    *VideoTagWhereInput
	OrderBy  []VideoTagOrderByInput
	Take     opt.Opt[int]
	Skip     opt.Opt[int]
	Distinct []VideoTagField
	Select   VideoTagSelect
}

type VideoTagFindFirst struct {
	Where   *VideoTagWhereInput
	OrderBy []VideoTagOrderByInput
	Skip    opt.Opt[int]
	Select  VideoTagSelect
}
type VideoTagCreateData struct {
	VideoId opt.Opt[uint64]
	TagId   opt.Opt[uint64]
}

type VideoTagCreate struct {
	Data   VideoTagCreateData
	Select VideoTagSelect
}

type VideoTagCreateMany struct {
	Data []VideoTagCreateData
}

func buildVideoTagWhere(d dialect.Dialect, alias string, args *argState, where *VideoTagWhereInput) string {
	if where == nil {
		return ""
	}
	w := newWhereBuilder(d, alias, args)
	return buildVideoTagWherePredicates(w, where)
}

func buildVideoTagWherePredicates(w *whereBuilder, where *VideoTagWhereInput) string {
	var preds []string
	if where.VideoId.IsSome() {
		preds = append(preds, w.eq("video_id", where.VideoId.Value()))
	}
	if where.TagId.IsSome() {
		preds = append(preds, w.eq("tag_id", where.TagId.Value()))
	}
	if where.Video != nil {
		if clause := buildVideoTagVideoRelation(w.d, w.alias, w.args, where.Video); clause != "" {
			preds = append(preds, clause)
		}
	}
	if where.Tag != nil {
		if clause := buildVideoTagTagRelation(w.d, w.alias, w.args, where.Tag); clause != "" {
			preds = append(preds, clause)
		}
	}
	for i := range where.AND {
		if clause := buildVideoTagWherePredicates(w, &where.AND[i]); clause != "" {
			preds = append(preds, clause)
		}
	}
	var orPreds []string
	for i := range where.OR {
		if clause := buildVideoTagWherePredicates(w, &where.OR[i]); clause != "" {
			orPreds = append(orPreds, clause)
		}
	}
	if orClause := w.combineOr(orPreds); orClause != "" {
		preds = append(preds, orClause)
	}
	for i := range where.NOT {
		if clause := buildVideoTagWherePredicates(w, &where.NOT[i]); clause != "" {
			preds = append(preds, w.d.Not(clause))
		}
	}
	return w.combineAnd(preds)
}
func buildVideoTagVideoRelation(d dialect.Dialect, parentAlias string, args *argState, rel *VideoTagVideoRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildVideoTagVideoRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildVideoTagVideoRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildVideoTagVideoRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildVideoTagVideoRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *VideoWhereInput, negate bool) string {
	childAlias := parentAlias + "_video"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "video_id"))
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

func buildVideoTagVideoRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *VideoWhereInput) string {
	childAlias := parentAlias + "_video" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "video_id"))
	var negated string
	if whereClause := buildVideoWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}
func buildVideoTagTagRelation(d dialect.Dialect, parentAlias string, args *argState, rel *VideoTagTagRelation) string {
	if rel == nil {
		return ""
	}
	var preds []string
	if rel.Some != nil {
		if clause := buildVideoTagTagRelationExists(d, parentAlias, args, rel.Some, false); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.None != nil {
		if clause := buildVideoTagTagRelationExists(d, parentAlias, args, rel.None, true); clause != "" {
			preds = append(preds, clause)
		}
	}
	if rel.Every != nil {
		if clause := buildVideoTagTagRelationEvery(d, parentAlias, args, rel.Every); clause != "" {
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

func buildVideoTagTagRelationExists(d dialect.Dialect, parentAlias string, args *argState, where *TagWhereInput, negate bool) string {
	childAlias := parentAlias + "_tag"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("tag"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "tag_id"))
	var combined string
	if len(joinConds) > 0 {
		combined = strings.Join(joinConds, " AND ")
	}
	if whereClause := buildTagWhere(d, childAlias, args, where); whereClause != "" {
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

func buildVideoTagTagRelationEvery(d dialect.Dialect, parentAlias string, args *argState, where *TagWhereInput) string {
	childAlias := parentAlias + "_tag" + "_every"
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("tag"), d.QuoteIdent(childAlias))
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "tag_id"))
	var negated string
	if whereClause := buildTagWhere(d, childAlias, args, where); whereClause != "" {
		negated = d.Not(whereClause)
	} else {
		negated = "0=1"
	}
	var combined []string
	if len(joinConds) > 0 {
		combined = append(combined, strings.Join(joinConds, " AND "))
	}
	combined = append(combined, negated)
	return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s WHERE %s)", tableRef, strings.Join(combined, " AND "))
}

func buildVideoTagWhereUnique(d dialect.Dialect, alias string, args *argState, where *VideoTagWhereUniqueInput) string {
	if where == nil {
		return ""
	}
	w := newWhereBuilder(d, alias, args)
	var preds []string
	if where.VideoId != nil {
		preds = append(preds, w.eq("video_id", *where.VideoId))
	}
	if where.TagId != nil {
		preds = append(preds, w.eq("tag_id", *where.TagId))
	}
	return w.combineAnd(preds)
}

func buildVideoTagOrderBy(d dialect.Dialect, alias string, order []VideoTagOrderByInput) []string {
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

func buildVideoTagDistinct(d dialect.Dialect, alias string, distinct []VideoTagField) []string {
	var out []string
	for _, field := range distinct {
		if field == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(field))))
	}
	return out
}

func buildVideoTagJSONExpr(d dialect.Dialect, alias string, args *argState, sel VideoTagSelect) (string, error) {
	items := sel
	includeAll := len(items) == 0
	var pairs []string
	if includeAll {
		pairs = append(pairs, jsonPair("video_id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("video_id"))))
		pairs = append(pairs, jsonPair("tag_id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("tag_id"))))
		{
			expr, err := buildVideoTagVideoJSON(d, alias, args, &VideoTagVideoSelectArgs{
				Select: VideoSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("video", wrapJSONValue(d, expr)))
		}
		{
			expr, err := buildVideoTagTagJSON(d, alias, args, &VideoTagTagSelectArgs{
				Select: TagSelectAll,
			})
			if err != nil {
				return "", err
			}
			pairs = append(pairs, jsonPair("tag", wrapJSONValue(d, expr)))
		}
	} else {
		seen := make(map[VideoTagField]struct{})
		relationSeen := make(map[string]struct{})
		for _, item := range items {
			switch v := item.(type) {
			case VideoTagField:
				if v == VideoTagFieldUnknown {
					continue
				}
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				pairs = append(pairs, jsonPair(string(v), fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent(string(v)))))
			case VideoTagSelectVideo:
				if _, ok := relationSeen["Video"]; ok {
					continue
				}
				relationSeen["Video"] = struct{}{}
				{
					expr, err := buildVideoTagVideoJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("video", wrapJSONValue(d, expr)))
				}
			case VideoTagSelectTag:
				if _, ok := relationSeen["Tag"]; ok {
					continue
				}
				relationSeen["Tag"] = struct{}{}
				{
					expr, err := buildVideoTagTagJSON(d, alias, args, &v.Args)
					if err != nil {
						return "", err
					}
					pairs = append(pairs, jsonPair("tag", wrapJSONValue(d, expr)))
				}
			}
		}
	}
	if len(pairs) == 0 {
		pairs = append(pairs, jsonPair("video_id", fmt.Sprintf("%s.%s", d.QuoteIdent(alias), d.QuoteIdent("video_id"))))
	}
	return d.JSONBuildObject(pairs...), nil
}
func buildVideoTagVideoJSON(d dialect.Dialect, parentAlias string, args *argState, sel *VideoTagVideoSelectArgs) (string, error) {
	if sel == nil {
		sel = &VideoTagVideoSelectArgs{}
	}
	childAlias := parentAlias + "_video"
	childJSON, err := buildVideoJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "video_id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = joinClause
	var whereSQL string
	if whereClause != "" {
		whereSQL = " WHERE " + whereClause
	}
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("video"), d.QuoteIdent(childAlias))
	expr := fmt.Sprintf("(SELECT %s FROM %s%s", childJSON, tableRef, whereSQL)
	expr += ")"
	return expr, nil
}
func buildVideoTagTagJSON(d dialect.Dialect, parentAlias string, args *argState, sel *VideoTagTagSelectArgs) (string, error) {
	if sel == nil {
		sel = &VideoTagTagSelectArgs{}
	}
	childAlias := parentAlias + "_tag"
	childJSON, err := buildTagJSONExpr(d, childAlias, args, sel.Select)
	if err != nil {
		return "", err
	}
	childJSON = wrapJSONValue(d, childJSON)
	joinConds := make([]string, 0, 1)
	joinConds = append(joinConds, columnEquals(d, childAlias, "id", parentAlias, "tag_id"))
	joinClause := strings.Join(joinConds, " AND ")
	var whereClause string
	whereClause = joinClause
	var whereSQL string
	if whereClause != "" {
		whereSQL = " WHERE " + whereClause
	}
	tableRef := fmt.Sprintf("%s %s", d.QuoteIdent("tag"), d.QuoteIdent(childAlias))
	expr := fmt.Sprintf("(SELECT %s FROM %s%s", childJSON, tableRef, whereSQL)
	expr += ")"
	return expr, nil
}

func FindManyVideoTag[T VideoTagModel](ctx context.Context, db bom.Querier, q VideoTagFindMany) ([]T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	rootAlias := "t0"
	jsonExpr, err := buildVideoTagJSONExpr(d, rootAlias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	whereClause := buildVideoTagWhere(d, rootAlias, state, q.Where)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "video_tag",
		Alias: rootAlias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where:     whereClause,
		Args:      state.Args(),
		OrderBy:   buildVideoTagOrderBy(d, rootAlias, q.OrderBy),
		Distinct:  buildVideoTagDistinct(d, rootAlias, q.Distinct),
		Limit:     convertOpt(q.Take),
		Offset:    convertOpt(q.Skip),
		JSONArray: true,
	}
	return queryVideoTagRows[T](ctx, db, d, input, true)
}

func FindFirstVideoTag[T VideoTagModel](ctx context.Context, db bom.Querier, q VideoTagFindFirst) (*T, error) {
	fm := VideoTagFindMany{
		Where:   q.Where,
		OrderBy: q.OrderBy,
		Skip:    q.Skip,
		Select:  q.Select,
	}
	fm.Take = opt.OVal(1)
	rows, err := FindManyVideoTag[T](ctx, db, fm)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	out := rows[0]
	return &out, nil
}
func CreateOneVideoTag[T VideoTagModel](ctx context.Context, db bom.Querier, q VideoTagCreate) (*T, error) {
	d := dialectsqlite.New()
	data := q.Data
	if err := createVideoTagRecord(ctx, db, d, &data); err != nil {
		return nil, err
	}
	if err := createVideoTagRelations(ctx, db, d, &data); err != nil {
		return nil, err
	}
	var fetch func(context.Context, VideoTagSelect) (*T, error)
	if fetch == nil {
		if data.VideoId.IsSome() && data.TagId.IsSome() {
			lookup := VideoTagUK_VideoIdTagId{
				VideoId: data.VideoId.Value(),
				TagId:   data.TagId.Value(),
			}
			fetch = func(ctx context.Context, sel VideoTagSelect) (*T, error) {
				return FindUniqueVideoTag[T](ctx, db, VideoTagFindUnique[VideoTagUK_VideoIdTagId]{
					Where:  lookup,
					Select: sel,
				})
			}
		}
	}
	if fetch == nil {
		return nil, fmt.Errorf("VideoTagCreate requires values for a unique constraint")
	}
	return fetch(ctx, q.Select)
}

func CreateManyVideoTag(ctx context.Context, db bom.Querier, q VideoTagCreateMany) (int64, error) {
	d := dialectsqlite.New()
	if len(q.Data) == 0 {
		return 0, nil
	}
	var total int64
	for i := range q.Data {
		data := q.Data[i]
		if err := createVideoTagRecord(ctx, db, d, &data); err != nil {
			return total, err
		}
		if err := createVideoTagRelations(ctx, db, d, &data); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func createVideoTagRecord(ctx context.Context, db bom.Querier, d dialect.Dialect, data *VideoTagCreateData) error {
	state := newArgState(d)
	var columns []string
	var placeholders []string
	if identityGenerator == nil {
		identityGenerator = &bom.DefaultIdentityGenerator{}
	}
	var wantsAutoVideoId bool
	var wantsAutoTagId bool
	if data.VideoId.IsSome() {
		val := data.VideoId.Value()
		columns = append(columns, d.QuoteIdent("video_id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if data.TagId.IsSome() {
		val := data.TagId.Value()
		columns = append(columns, d.QuoteIdent("tag_id"))
		placeholders = append(placeholders, state.Add(val))
	} else {
	}
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return err
	}
	sqlStr := buildInsertSQL(d, "video_tag", columns, placeholders)
	autoCount := 0
	if wantsAutoVideoId {
		autoCount++
	}
	if wantsAutoTagId {
		autoCount++
	}
	var returning bool
	if d.Cap().InsertReturning && autoCount > 0 {
		returning = true
		var returningCols []string
		if wantsAutoVideoId {
			returningCols = append(returningCols, d.QuoteIdent("video_id"))
		}
		if wantsAutoTagId {
			returningCols = append(returningCols, d.QuoteIdent("tag_id"))
		}
		if len(returningCols) > 0 {
			sqlStr = sqlStr + " RETURNING " + strings.Join(returningCols, ", ")
		}
	}
	if returning {
		rows, err := db.QueryContext(ctx, sqlStr, state.Args()...)
		if err != nil {
			return err
		}
		defer rows.Close()
		var scanTargets []any
		var retVideoId uint64
		if wantsAutoVideoId {
			scanTargets = append(scanTargets, &retVideoId)
		}
		var retTagId uint64
		if wantsAutoTagId {
			scanTargets = append(scanTargets, &retTagId)
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return fmt.Errorf("VideoTag insert returned no rows")
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}
		if wantsAutoVideoId {
			data.VideoId = opt.OVal(retVideoId)
		}
		if wantsAutoTagId {
			data.TagId = opt.OVal(retTagId)
		}
		return nil
	}
	res, err := db.ExecContext(ctx, sqlStr, state.Args()...)
	if err != nil {
		return err
	}
	if autoCount == 1 {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if wantsAutoVideoId {
			data.VideoId = opt.OVal(uint64(id))
		}
		if wantsAutoTagId {
			data.TagId = opt.OVal(uint64(id))
		}
	} else if autoCount > 1 {
		return fmt.Errorf("%s cannot populate multiple auto columns for video_tag without RETURNING", d.Name())
	}
	return nil
}

func createVideoTagRelations(ctx context.Context, db bom.Querier, d dialect.Dialect, data *VideoTagCreateData) error {
	return nil
}

func queryVideoTagRows[T VideoTagModel](ctx context.Context, db bom.Querier, d dialect.Dialect, input planner.FindManyInput, expectArray bool) ([]T, error) {
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
				item = normalizeJSONRaw(item)
				var rec VideoTag
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
		raw = normalizeJSONRaw(raw)
		var rec VideoTag
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
	d := dialectsqlite.New()
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
	jsonExpr = wrapJSONValue(d, jsonExpr)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
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

type AuthorProfileUnique interface {
	AuthorProfileUK_Id | AuthorProfileUK_AuthorId
}

type AuthorProfileUK_Id struct {
	Id uint64
}

type AuthorProfileUK_AuthorId struct {
	AuthorId uint64
}

type AuthorProfileFindUnique[U AuthorProfileUnique] struct {
	Where  U
	Select AuthorProfileSelect
}

func buildAuthorProfileUniquePredicate[U AuthorProfileUnique](d dialect.Dialect, alias string, args *argState, where U) (string, error) {
	w := newWhereBuilder(d, alias, args)
	switch v := any(where).(type) {
	case AuthorProfileUK_Id:
		return buildAuthorProfileUK_IdPredicate(w, v), nil
	case AuthorProfileUK_AuthorId:
		return buildAuthorProfileUK_AuthorIdPredicate(w, v), nil
	default:
		return "", fmt.Errorf("unsupported unique type %T", where)
	}
}

func buildAuthorProfileUK_IdPredicate(w *whereBuilder, where AuthorProfileUK_Id) string {
	var preds []string
	preds = append(preds, w.eq("id", where.Id))
	return w.combineAnd(preds)
}

func buildAuthorProfileUK_AuthorIdPredicate(w *whereBuilder, where AuthorProfileUK_AuthorId) string {
	var preds []string
	preds = append(preds, w.eq("author_id", where.AuthorId))
	return w.combineAnd(preds)
}

func FindUniqueAuthorProfile[T AuthorProfileModel, U AuthorProfileUnique](ctx context.Context, db bom.Querier, q AuthorProfileFindUnique[U]) (*T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	alias := "t0"
	whereClause, err := buildAuthorProfileUniquePredicate(d, alias, state, q.Where)
	if err != nil {
		return nil, err
	}
	jsonExpr, err := buildAuthorProfileJSONExpr(d, alias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "author_profile",
		Alias: alias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where: whereClause,
		Args:  state.Args(),
		Limit: one(),
	}
	rows, err := queryAuthorProfileRows[T](ctx, db, d, input, false)
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
	d := dialectsqlite.New()
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
	jsonExpr = wrapJSONValue(d, jsonExpr)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
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

type CommentUnique interface {
	CommentUK_Id
}

type CommentUK_Id struct {
	Id uint64
}

type CommentFindUnique[U CommentUnique] struct {
	Where  U
	Select CommentSelect
}

func buildCommentUniquePredicate[U CommentUnique](d dialect.Dialect, alias string, args *argState, where U) (string, error) {
	w := newWhereBuilder(d, alias, args)
	switch v := any(where).(type) {
	case CommentUK_Id:
		return buildCommentUK_IdPredicate(w, v), nil
	default:
		return "", fmt.Errorf("unsupported unique type %T", where)
	}
}

func buildCommentUK_IdPredicate(w *whereBuilder, where CommentUK_Id) string {
	var preds []string
	preds = append(preds, w.eq("id", where.Id))
	return w.combineAnd(preds)
}

func FindUniqueComment[T CommentModel, U CommentUnique](ctx context.Context, db bom.Querier, q CommentFindUnique[U]) (*T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	alias := "t0"
	whereClause, err := buildCommentUniquePredicate(d, alias, state, q.Where)
	if err != nil {
		return nil, err
	}
	jsonExpr, err := buildCommentJSONExpr(d, alias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "comment",
		Alias: alias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where: whereClause,
		Args:  state.Args(),
		Limit: one(),
	}
	rows, err := queryCommentRows[T](ctx, db, d, input, false)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	out := rows[0]
	return &out, nil
}

type TagUnique interface {
	TagUK_Id | TagUK_Name
}

type TagUK_Id struct {
	Id uint64
}

type TagUK_Name struct {
	Name string
}

type TagFindUnique[U TagUnique] struct {
	Where  U
	Select TagSelect
}

func buildTagUniquePredicate[U TagUnique](d dialect.Dialect, alias string, args *argState, where U) (string, error) {
	w := newWhereBuilder(d, alias, args)
	switch v := any(where).(type) {
	case TagUK_Id:
		return buildTagUK_IdPredicate(w, v), nil
	case TagUK_Name:
		return buildTagUK_NamePredicate(w, v), nil
	default:
		return "", fmt.Errorf("unsupported unique type %T", where)
	}
}

func buildTagUK_IdPredicate(w *whereBuilder, where TagUK_Id) string {
	var preds []string
	preds = append(preds, w.eq("id", where.Id))
	return w.combineAnd(preds)
}

func buildTagUK_NamePredicate(w *whereBuilder, where TagUK_Name) string {
	var preds []string
	preds = append(preds, w.eq("name", where.Name))
	return w.combineAnd(preds)
}

func FindUniqueTag[T TagModel, U TagUnique](ctx context.Context, db bom.Querier, q TagFindUnique[U]) (*T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	alias := "t0"
	whereClause, err := buildTagUniquePredicate(d, alias, state, q.Where)
	if err != nil {
		return nil, err
	}
	jsonExpr, err := buildTagJSONExpr(d, alias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "tag",
		Alias: alias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where: whereClause,
		Args:  state.Args(),
		Limit: one(),
	}
	rows, err := queryTagRows[T](ctx, db, d, input, false)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	out := rows[0]
	return &out, nil
}

type VideoTagUnique interface {
	VideoTagUK_VideoIdTagId
}

type VideoTagUK_VideoIdTagId struct {
	VideoId uint64
	TagId   uint64
}

type VideoTagFindUnique[U VideoTagUnique] struct {
	Where  U
	Select VideoTagSelect
}

func buildVideoTagUniquePredicate[U VideoTagUnique](d dialect.Dialect, alias string, args *argState, where U) (string, error) {
	w := newWhereBuilder(d, alias, args)
	switch v := any(where).(type) {
	case VideoTagUK_VideoIdTagId:
		return buildVideoTagUK_VideoIdTagIdPredicate(w, v), nil
	default:
		return "", fmt.Errorf("unsupported unique type %T", where)
	}
}

func buildVideoTagUK_VideoIdTagIdPredicate(w *whereBuilder, where VideoTagUK_VideoIdTagId) string {
	var preds []string
	preds = append(preds, w.eq("video_id", where.VideoId))
	preds = append(preds, w.eq("tag_id", where.TagId))
	return w.combineAnd(preds)
}

func FindUniqueVideoTag[T VideoTagModel, U VideoTagUnique](ctx context.Context, db bom.Querier, q VideoTagFindUnique[U]) (*T, error) {
	d := dialectsqlite.New()
	state := newArgState(d)
	alias := "t0"
	whereClause, err := buildVideoTagUniquePredicate(d, alias, state, q.Where)
	if err != nil {
		return nil, err
	}
	jsonExpr, err := buildVideoTagJSONExpr(d, alias, state, q.Select)
	if err != nil {
		return nil, err
	}
	jsonExpr = wrapJSONValue(d, jsonExpr)
	if err := ensureParamLimit(d, len(state.Args())); err != nil {
		return nil, err
	}
	input := planner.FindManyInput{
		Table: "video_tag",
		Alias: alias,
		Projections: []planner.Projection{
			{Expr: jsonExpr, Alias: "__bom_json"},
		},
		Where: whereClause,
		Args:  state.Args(),
		Limit: one(),
	}
	rows, err := queryVideoTagRows[T](ctx, db, d, input, false)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	out := rows[0]
	return &out, nil
}
