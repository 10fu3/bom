package dialect

// Strategy encodes how the dialect handles case-insensitive comparisons.
type Strategy int

const (
	// ILike uses the ILIKE keyword (PostgreSQL).
	ILike Strategy = iota
	// LowerLike lower-cases both sides of a comparison (MySQL/SQLite default).
	LowerLike
	// CollateCI uses an explicit _ci collation (MySQL).
	CollateCI
)

// Capabilities describe the behavior toggles a dialect exposes.
type Capabilities struct {
	DistinctOn      bool
	CaseInsensitive Strategy
	Placeholder     string
}

// Dialect builds SQL snippets in a dialect-aware manner.
type Dialect interface {
	Name() string
	Cap() Capabilities

	QuoteIdent(id string) string
	Placeholder(n int) string

	Eq(l, r string) string
	And(preds ...string) string
	Or(preds ...string) string
	Not(pred string) string
	InsensitiveLike(lhs, placeholder string) string

	JSONBuildObject(pairs ...string) string
	JSONArrayAgg(expr string) string
	JSONArrayEmpty() string
	CoalesceJSONAgg(expr, empty string) string
	JSONValue(expr string) string

	LimitOffset(limit, offset *int64) string
	DistinctProjection(cols []string) (prefix string, needsWrap bool)
}
