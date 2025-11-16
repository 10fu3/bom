package sqlite

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"bom/internal/schema"
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

func (l *lexer) Lex(lval *sqliteSymType) int {
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
		switch {
		case r == 0:
			return 0
		case r == '(':
			l.pos += w
			return LPAREN
		case r == ')':
			l.pos += w
			return RPAREN
		case r == ',':
			l.pos += w
			return COMMA
		case r == ';':
			l.pos += w
			return SEMICOLON
		case r == '.':
			l.pos += w
			return DOT
		case r == '+':
			l.pos += w
			return PLUS
		case r == '-':
			l.pos += w
			return MINUS
		case r == '\'':
			val, err := l.readString()
			if err != nil {
				l.err = err
				return 0
			}
			lval.str = val
			return STRING
		case r == '"' || r == '`' || r == '[':
			ident, err := l.readQuotedIdent(r)
			if err != nil {
				l.err = err
				return 0
			}
			lval.str = ident
			return IDENT
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
				case IDENT:
					lval.str = ident
					return IDENT
				case STRING:
					lval.str = ident
					return STRING
				case NUMBER:
					lval.str = ident
					return NUMBER
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
	if l.pos >= len(l.src) {
		return false
	}
	switch {
	case strings.HasPrefix(l.src[l.pos:], "--"),
		strings.HasPrefix(l.src[l.pos:], "//"),
		l.src[l.pos] == '#':
		l.skipLine()
		return true
	case strings.HasPrefix(l.src[l.pos:], "/*"):
		if err := l.consumeBlockComment(); err != nil {
			l.err = err
		}
		return true
	default:
		return false
	}
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
	l.pos += 2 // skip /*
	depth := 1
	for l.pos < len(l.src) {
		switch {
		case strings.HasPrefix(l.src[l.pos:], "/*"):
			l.pos += 2
			depth++
		case strings.HasPrefix(l.src[l.pos:], "*/"):
			l.pos += 2
			depth--
			if depth == 0 {
				return nil
			}
		default:
			_, w := l.peek()
			l.pos += w
		}
	}
	return fmt.Errorf("unterminated block comment")
}

func (l *lexer) peek() (rune, int) {
	if l.pos >= len(l.src) {
		return 0, 0
	}
	r, w := utf8.DecodeRuneInString(l.src[l.pos:])
	return r, w
}

func (l *lexer) readIdent() string {
	start := l.pos
	for l.pos < len(l.src) {
		r, w := l.peek()
		if isIdentPart(r) {
			l.pos += w
			continue
		}
		break
	}
	return l.src[start:l.pos]
}

func (l *lexer) readNumber() string {
	start := l.pos
	sawDot := false
	for l.pos < len(l.src) {
		r, w := l.peek()
		switch {
		case unicode.IsDigit(r):
			l.pos += w
		case r == '.' && !sawDot:
			sawDot = true
			l.pos += w
		case r == 'e' || r == 'E':
			l.pos += w
			if l.pos < len(l.src) {
				if sign, width := l.peek(); sign == '+' || sign == '-' {
					l.pos += width
				}
			}
			for l.pos < len(l.src) {
				r2, w2 := l.peek()
				if unicode.IsDigit(r2) {
					l.pos += w2
				} else {
					break
				}
			}
		default:
			goto done
		}
	}
done:
	return l.src[start:l.pos]
}

func (l *lexer) readString() (string, error) {
	var sb strings.Builder
	l.pos++ // skip opening quote
	for l.pos < len(l.src) {
		r, w := l.peek()
		l.pos += w
		if r == '\'' {
			if next, width := l.peek(); next == '\'' {
				l.pos += width
				sb.WriteRune('\'')
				continue
			}
			return sb.String(), nil
		}
		sb.WriteRune(r)
	}
	return "", fmt.Errorf("unterminated string literal")
}

func (l *lexer) readQuotedIdent(quote rune) (string, error) {
	var sb strings.Builder
	l.pos++ // skip opening quote
	closing := quote
	if quote == '[' {
		closing = ']'
	}
	for l.pos < len(l.src) {
		r, w := l.peek()
		l.pos += w
		if r == closing {
			// doubled closing escapes the character
			if next, width := l.peek(); next == closing {
				l.pos += width
				sb.WriteRune(closing)
				continue
			}
			return sb.String(), nil
		}
		sb.WriteRune(r)
	}
	return "", fmt.Errorf("unterminated identifier")
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
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
