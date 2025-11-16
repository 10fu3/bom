%{
package sqlite

import (
	"bufio"
	"errors"
	"fmt"
	"go/token"
	"strings"
	"unicode/utf8"

	"bom/internal/schema"

	"modernc.org/golex/lex"
)

type lexer struct {
	*lex.Lexer
	ir  *schema.IR
	err error
}

func newLexer(src string, ir *schema.IR) *lexer {
	reader := bufio.NewReader(strings.NewReader(src))
	file := token.NewFileSet().AddFile("", -1, len(src))
	lx, err := lex.New(file, reader)
	if err != nil {
		panic(err)
	}
	return &lexer{Lexer: lx, ir: ir}
}

func (l *lexer) Error(msg string) {
	if l.err == nil {
		l.err = errors.New(msg)
	}
}

func (l *lexer) emitIdentifier(lval *sqliteSymType) int {
	raw := string(l.TokenBytes(nil))
	upper := strings.ToUpper(raw)
	if tok, ok := keywords[upper]; ok {
		switch tok {
		case IDENT, STRING, NUMBER:
			lval.str = raw
		}
		return tok
	}
	lval.str = raw
	return IDENT
}

func (l *lexer) emitQuotedIdent(lval *sqliteSymType, closing rune) int {
	text := l.TokenBytes(nil)
	val, err := unescapeDuplicated(text, closing)
	if err != nil {
		l.err = err
		return 0
	}
	lval.str = val
	return IDENT
}

func (l *lexer) emitString(lval *sqliteSymType) int {
	text := l.TokenBytes(nil)
	val, err := unescapeDuplicated(text, '\'')
	if err != nil {
		l.err = err
		return 0
	}
	lval.str = val
	return STRING
}

func unescapeDuplicated(raw []byte, closing rune) (string, error) {
	if len(raw) < 2 {
		return "", fmt.Errorf("invalid literal %q", string(raw))
	}
	var sb strings.Builder
	i := 1
	limit := len(raw) - 1
	for i < limit {
		r, size := utf8.DecodeRune(raw[i:])
		if r == closing {
			next, nextSize := utf8.DecodeRune(raw[i+size:])
			if next != closing {
				return "", fmt.Errorf("unterminated literal %q", string(raw))
			}
			sb.WriteRune(closing)
			i += size + nextSize
			continue
		}
		sb.WriteRune(r)
		i += size
	}
	return sb.String(), nil
}

func (l *lexer) addTable(tb *tableBuilder) {
	if l.err != nil || tb == nil || tb.name == "" || l.ir == nil {
		return
	}
	table := tb.finalize()
	l.ir.AddTable(table)
}

func (l *lexer) addIndex(idx indexDef) {
	if l.err != nil || idx.table == "" || l.ir == nil {
		return
	}
	target := l.ir.Table(idx.schema, idx.table)
	if target == nil {
		l.err = fmt.Errorf("index %s references unknown table %s", idx.name, idx.table)
		return
	}
	if idx.unique {
		target.Uniques = appendUnique(target.Uniques, idx.name, idx.columns)
		return
	}
	target.Indexes = append(target.Indexes, schema.Index{
		Name:   idx.name,
		Cols:   append([]string{}, idx.columns...),
		Unique: false,
	})
}

var keywords = map[string]int{
	"CREATE":            CREATE_KW,
	"TABLE":             TABLE_KW,
	"TEMP":              TEMP_KW,
	"TEMPORARY":         TEMP_KW,
	"IF":                IF_KW,
	"NOT":               NOT_KW,
	"EXISTS":            EXISTS_KW,
	"PRIMARY":           PRIMARY_KW,
	"KEY":               KEY_KW,
	"UNIQUE":            UNIQUE_KW,
	"CONSTRAINT":        CONSTRAINT_KW,
	"FOREIGN":           FOREIGN_KW,
	"REFERENCES":        REFERENCES_KW,
	"DEFAULT":           DEFAULT_KW,
	"AUTOINCREMENT":     AUTOINCREMENT_KW,
	"ON":                ON_KW,
	"DELETE":            DELETE_KW,
	"UPDATE":            UPDATE_KW,
	"SET":               SET_KW,
	"CASCADE":           CASCADE_KW,
	"RESTRICT":          RESTRICT_KW,
	"NO":                NO_KW,
	"ACTION":            ACTION_KW,
	"WITHOUT":           WITHOUT_KW,
	"ROWID":             ROWID_KW,
	"INDEX":             INDEX_KW,
	"TRUE":              TRUE_KW,
	"FALSE":             FALSE_KW,
	"NULL":              NULL_KW,
	"CURRENT_TIMESTAMP": CURRENT_TIMESTAMP_KW,
	"CURRENT_DATE":      CURRENT_DATE_KW,
	"CURRENT_TIME":      CURRENT_TIME_KW,
	"COLLATE":           COLLATE_KW,
	"ASC":               ASC_KW,
	"DESC":              DESC_KW,
}
%}

%yyc c
%yyn c = l.Next()
%yym l.Mark()

ws              [ \t\r\n\f]+
commentdash     --[^\r\n]*
commenthash     #[^\r\n]*
commentdouble   //[^\r\n]*
commentblock    /\*([^*]|\*+[^*/])*\*+/
digit           [0-9]
number          {digit}+(\.{digit}+)?([eE][+-]?{digit}+)?
identifier      [_A-Za-z][_A-Za-z0-9$]*
doublequoted    "([^"\n]|"")*"
backtickquoted  `([^`\n]|``)*`
bracketquoted   \[([^\]]|\]\])*\]
stringliteral   '([^'\n]|'')*'

%%
	c = l.Rule0()

{ws}
{commentdash}
{commenthash}
{commentdouble}
{commentblock}
\(        return LPAREN
\)        return RPAREN
,          return COMMA
;          return SEMICOLON
\.         return DOT
\+         return PLUS
-          return MINUS
{stringliteral} { return l.emitString(lval) }
{number}        { lval.str = string(l.TokenBytes(nil)); return NUMBER }
{doublequoted}  { return l.emitQuotedIdent(lval, '"') }
{backtickquoted}{ return l.emitQuotedIdent(lval, '`') }
{bracketquoted} { return l.emitQuotedIdent(lval, ']') }
{identifier}    { return l.emitIdentifier(lval) }
.               { l.err = fmt.Errorf("unexpected input %q", string(l.TokenBytes(nil))); return 0 }

%%
	if c, ok := l.Abort(); ok {
		return c
	}
	goto yyAction
}
