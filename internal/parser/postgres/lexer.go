package postgres

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/10fu3/bom/internal/schema"
)

type lexer struct {
	src string
	pos int
	ir  *schema.IR
	err error
}

func newLexer(src string, ir *schema.IR) *lexer {
	return &lexer{src: src, ir: ir}
}

func (l *lexer) Lex(lval *postgresSymType) int {
	for {
		if l.err != nil {
			return 0
		}
		l.skipWhitespace()
		if l.pos >= len(l.src) {
			return 0
		}
		if l.skipComment() {
			continue
		}
		r, w := l.peek()
		switch r {
		case 0:
			return 0
		case '(':
			l.pos += w
			return LPAREN
		case ')':
			l.pos += w
			return RPAREN
		case ',':
			l.pos += w
			return COMMA
		case ';':
			l.pos += w
			return SEMICOLON
		case '.':
			l.pos += w
			return DOT
		case '+':
			l.pos += w
			return PLUS
		case '-':
			if l.match("--") {
				l.skipLine()
				continue
			}
			l.pos += w
			return MINUS
		case '*':
			l.pos += w
			return STAR
		case '/':
			if l.match("/*") {
				if err := l.consumeBlockComment(); err != nil {
					l.err = err
					return 0
				}
				continue
			}
			l.pos += w
			return SLASH
		case '[':
			l.pos += w
			return LBRACKET
		case ']':
			l.pos += w
			return RBRACKET
		case '=':
			l.pos += w
			return EQUALS
		case ':':
			if l.match("::") {
				l.pos += 2
				return TYPECAST
			}
		case '"':
			ident, err := l.readQuotedIdent()
			if err != nil {
				l.err = err
				return 0
			}
			lval.str = ident
			return IDENT
		case '$':
			str, err := l.readDollarString()
			if err != nil {
				l.err = err
				return 0
			}
			lval.str = str
			return STRING
		case '\'':
			str, err := l.readString(false)
			if err != nil {
				l.err = err
				return 0
			}
			lval.str = str
			return STRING
		case 'E', 'e':
			if l.pos+w < len(l.src) && l.src[l.pos+w] == '\'' {
				l.pos += w
				str, err := l.readString(true)
				if err != nil {
					l.err = err
					return 0
				}
				lval.str = str
				return STRING
			}
		}
		if unicode.IsDigit(r) {
			num := l.readNumber()
			lval.str = num
			return NUMBER
		}
		if isIdentStart(r) {
			ident := l.readIdent()
			upper := strings.ToUpper(ident)
			if tok, ok := keywords[upper]; ok {
				switch tok {
				case IDENT, STRING, NUMBER:
					lval.str = ident
				default:
					return tok
				}
			}
			lval.str = ident
			return IDENT
		}
		l.err = fmt.Errorf("unexpected character %q", r)
		return 0
	}
}

func (l *lexer) Error(msg string) {
	if l.err == nil {
		l.err = errors.New(msg)
	}
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.src) {
		r, w := l.peek()
		if unicode.IsSpace(r) || r == '\ufeff' {
			l.pos += w
			continue
		}
		break
	}
}

func (l *lexer) skipComment() bool {
	if l.match("--") {
		l.skipLine()
		return true
	}
	if l.match("/*") {
		if err := l.consumeBlockComment(); err != nil {
			l.err = err
		}
		return true
	}
	return false
}

func (l *lexer) skipLine() {
	for l.pos < len(l.src) {
		r, w := l.peek()
		l.pos += w
		if r == '\n' || r == '\r' {
			break
		}
	}
}

func (l *lexer) consumeBlockComment() error {
	depth := 1
	l.pos += 2
	for l.pos < len(l.src) {
		r := l.src[l.pos]
		if r == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '*' {
			depth++
			l.pos += 2
			continue
		}
		if r == '*' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
			depth--
			l.pos += 2
			if depth == 0 {
				return nil
			}
			continue
		}
		l.pos++
	}
	return fmt.Errorf("unterminated comment")
}

func (l *lexer) readQuotedIdent() (string, error) {
	start := l.pos + 1
	for l.pos++; l.pos < len(l.src); {
		r, w := l.peek()
		if r == '"' {
			if l.pos+w < len(l.src) && l.src[l.pos+w] == '"' {
				l.pos += w * 2
				continue
			}
			ident := unescapeDuplicated(l.src[start:l.pos], '"')
			l.pos += w
			return ident, nil
		}
		l.pos += w
	}
	return "", fmt.Errorf("unterminated quoted identifier")
}

func (l *lexer) readString(_ bool) (string, error) {
	var sb strings.Builder
	l.pos++ // skip opening quote
	for l.pos < len(l.src) {
		r, w := l.peek()
		if r == '\'' {
			if l.pos+w < len(l.src) && l.src[l.pos+w] == '\'' {
				sb.WriteRune('\'')
				l.pos += 2 * w
				continue
			}
			l.pos += w
			return sb.String(), nil
		}
		sb.WriteRune(r)
		l.pos += w
	}
	return "", fmt.Errorf("unterminated string literal")
}

func (l *lexer) readDollarString() (string, error) {
	start := l.pos
	end := start + 1
	for end < len(l.src) {
		ch := l.src[end]
		if ch == '$' {
			break
		}
		if !isDollarTagChar(ch) {
			return "", fmt.Errorf("invalid dollar quote")
		}
		end++
	}
	if end >= len(l.src) || l.src[end] != '$' {
		return "", fmt.Errorf("invalid dollar quote")
	}
	tag := l.src[start : end+1]
	l.pos = end + 1
	closing := strings.Index(l.src[l.pos:], tag)
	if closing < 0 {
		return "", fmt.Errorf("unterminated dollar quote")
	}
	content := l.src[l.pos : l.pos+closing]
	l.pos += closing + len(tag)
	return content, nil
}

func (l *lexer) readIdent() string {
	start := l.pos
	for l.pos < len(l.src) {
		r, w := l.peek()
		if !isIdentPart(r) {
			break
		}
		l.pos += w
	}
	return l.src[start:l.pos]
}

func (l *lexer) readNumber() string {
	start := l.pos
	for l.pos < len(l.src) {
		r := l.src[l.pos]
		if (r < '0' || r > '9') && r != '.' {
			break
		}
		l.pos++
	}
	if l.pos < len(l.src) && (l.src[l.pos] == 'e' || l.src[l.pos] == 'E') {
		l.pos++
		if l.pos < len(l.src) && (l.src[l.pos] == '+' || l.src[l.pos] == '-') {
			l.pos++
		}
		for l.pos < len(l.src) {
			ch := l.src[l.pos]
			if ch < '0' || ch > '9' {
				break
			}
			l.pos++
		}
	}
	return l.src[start:l.pos]
}

func (l *lexer) peek() (rune, int) {
	if l.pos >= len(l.src) {
		return 0, 0
	}
	r, w := utf8.DecodeRuneInString(l.src[l.pos:])
	return r, w
}

func (l *lexer) match(prefix string) bool {
	return strings.HasPrefix(l.src[l.pos:], prefix)
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

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '$'
}

func isIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

func isDollarTagChar(b byte) bool {
	return b == '_' || (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func unescapeDuplicated(raw string, closing rune) string {
	var sb strings.Builder
	for i := 0; i < len(raw); {
		r, size := utf8.DecodeRuneInString(raw[i:])
		if r == closing && i+size < len(raw) {
			next, nextSize := utf8.DecodeRuneInString(raw[i+size:])
			if next == closing {
				sb.WriteRune(closing)
				i += size + nextSize
				continue
			}
		}
		sb.WriteRune(r)
		i += size
	}
	return sb.String()
}

var keywords = map[string]int{
	"CREATE":            CREATE_KW,
	"TABLE":             TABLE_KW,
	"INDEX":             INDEX_KW,
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
	"ON":                ON_KW,
	"DELETE":            DELETE_KW,
	"UPDATE":            UPDATE_KW,
	"SET":               SET_KW,
	"CASCADE":           CASCADE_KW,
	"RESTRICT":          RESTRICT_KW,
	"NO":                NO_KW,
	"ACTION":            ACTION_KW,
	"NULL":              NULL_KW,
	"TRUE":              TRUE_KW,
	"FALSE":             FALSE_KW,
	"CURRENT_TIMESTAMP": CURRENT_TIMESTAMP_KW,
	"CURRENT_DATE":      CURRENT_DATE_KW,
	"CURRENT_TIME":      CURRENT_TIME_KW,
	"ALTER":             ALTER_KW,
	"ADD":               ADD_KW,
	"COLUMN":            COLUMN_KW,
	"DROP":              DROP_KW,
	"ONLY":              ONLY_KW,
	"TABLESPACE":        TABLESPACE_KW,
	"WITH":              WITH_KW,
	"CHECK":             CHECK_KW,
	"MATCH":             MATCH_KW,
	"DEFERRABLE":        DEFERRABLE_KW,
	"INITIALLY":         INITIALLY_KW,
	"DEFERRED":          DEFERRED_KW,
	"IMMEDIATE":         IMMEDIATE_KW,
	"GENERATED":         GENERATED_KW,
	"ALWAYS":            ALWAYS_KW,
	"BY":                BY_KW,
	"AS":                AS_KW,
	"IDENTITY":          IDENTITY_KW,
	"USING":             USING_KW,
	"INCLUDE":           INCLUDE_KW,
	"COMMENT":           COMMENT_KW,
	"IS":                IS_KW,
	"COLLATE":           COLLATE_KW,
}
