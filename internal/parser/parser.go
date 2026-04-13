package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/81120/thrift2x/internal/ast"
)

type tokenType int

const (
	tokEOF tokenType = iota
	tokIdent
	tokNumber
	tokString
	tokSymbol
)

type token struct {
	typ  tokenType
	lit  string
	line int
	col  int
}

type lexer struct {
	src  []rune
	pos  int
	line int
	col  int
}

func newLexer(input string) *lexer {
	return &lexer{src: []rune(input), line: 1, col: 1}
}

func (l *lexer) eof() bool {
	return l.pos >= len(l.src)
}

func (l *lexer) peek() rune {
	if l.eof() {
		return 0
	}
	return l.src[l.pos]
}

func (l *lexer) peekN(n int) rune {
	idx := l.pos + n
	if idx >= len(l.src) {
		return 0
	}
	return l.src[idx]
}

func (l *lexer) next() rune {
	if l.eof() {
		return 0
	}
	r := l.src[l.pos]
	l.pos++
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *lexer) skipWhitespaceAndComments() {
	for !l.eof() {
		r := l.peek()
		if unicode.IsSpace(r) {
			l.next()
			continue
		}
		if r == '/' && l.peekN(1) == '/' {
			for !l.eof() && l.next() != '\n' {
			}
			continue
		}
		if r == '#' {
			for !l.eof() && l.next() != '\n' {
			}
			continue
		}
		if r == '/' && l.peekN(1) == '*' {
			l.next()
			l.next()
			for !l.eof() {
				if l.peek() == '*' && l.peekN(1) == '/' {
					l.next()
					l.next()
					break
				}
				l.next()
			}
			continue
		}
		break
	}
}

func (l *lexer) nextToken() token {
	l.skipWhitespaceAndComments()
	if l.eof() {
		return token{typ: tokEOF, line: l.line, col: l.col}
	}
	startLine, startCol := l.line, l.col
	r := l.peek()

	if unicode.IsLetter(r) || r == '_' {
		var b strings.Builder
		for !l.eof() {
			rr := l.peek()
			if unicode.IsLetter(rr) || unicode.IsDigit(rr) || rr == '_' || rr == '.' {
				b.WriteRune(l.next())
			} else {
				break
			}
		}
		return token{typ: tokIdent, lit: b.String(), line: startLine, col: startCol}
	}

	if unicode.IsDigit(r) || (r == '-' && unicode.IsDigit(l.peekN(1))) {
		var b strings.Builder
		if r == '-' {
			b.WriteRune(l.next())
		}
		for !l.eof() {
			rr := l.peek()
			if unicode.IsDigit(rr) {
				b.WriteRune(l.next())
			} else {
				break
			}
		}
		return token{typ: tokNumber, lit: b.String(), line: startLine, col: startCol}
	}

	if r == '"' || r == '\'' {
		quote := l.next()
		var b strings.Builder
		for !l.eof() {
			rr := l.next()
			if rr == '\\' && !l.eof() {
				b.WriteRune(rr)
				b.WriteRune(l.next())
				continue
			}
			if rr == quote {
				break
			}
			b.WriteRune(rr)
		}
		return token{typ: tokString, lit: b.String(), line: startLine, col: startCol}
	}

	sym := l.next()
	return token{typ: tokSymbol, lit: string(sym), line: startLine, col: startCol}
}

type Parser struct {
	lx     *lexer
	cur    token
	peeked *token
}

func Parse(path string, input string) (*ast.File, error) {
	p := &Parser{lx: newLexer(input)}
	p.cur = p.nextToken()
	f := &ast.File{Path: path}
	for p.cur.typ != tokEOF {
		if p.isIdent("include") {
			inc, err := p.parseInclude()
			if err != nil {
				return nil, err
			}
			f.Includes = append(f.Includes, inc)
			continue
		}
		if p.isIdent("namespace") {
			ns, err := p.parseNamespace()
			if err != nil {
				return nil, err
			}
			f.Namespaces = append(f.Namespaces, ns)
			continue
		}

		d, err := p.parseDecl()
		if err != nil {
			return nil, err
		}
		if d != nil {
			f.Decls = append(f.Decls, d)
		}
	}
	return f, nil
}

func (p *Parser) nextToken() token {
	if p.peeked != nil {
		t := *p.peeked
		p.peeked = nil
		return t
	}
	return p.lx.nextToken()
}

func (p *Parser) peekToken() token {
	if p.peeked == nil {
		t := p.lx.nextToken()
		p.peeked = &t
	}
	return *p.peeked
}

func (p *Parser) advance() {
	p.cur = p.nextToken()
}

func (p *Parser) isIdent(s string) bool {
	return p.cur.typ == tokIdent && p.cur.lit == s
}

func (p *Parser) isSymbol(s string) bool {
	return p.cur.typ == tokSymbol && p.cur.lit == s
}

func (p *Parser) expectIdent() (string, error) {
	if p.cur.typ != tokIdent {
		return "", p.errf("expected identifier, got %q", p.cur.lit)
	}
	v := p.cur.lit
	p.advance()
	return v, nil
}

func (p *Parser) expectSymbol(s string) error {
	if !p.isSymbol(s) {
		return p.errf("expected symbol %q, got %q", s, p.cur.lit)
	}
	p.advance()
	return nil
}

func (p *Parser) consumeOptionalDelim() {
	if p.isSymbol(",") || p.isSymbol(";") {
		p.advance()
	}
}

func (p *Parser) parseInclude() (ast.Include, error) {
	p.advance()
	if p.cur.typ != tokString {
		return ast.Include{}, p.errf("include requires string path")
	}
	inc := ast.Include{Path: p.cur.lit}
	p.advance()
	p.consumeOptionalDelim()
	return inc, nil
}

func (p *Parser) parseNamespace() (ast.Namespace, error) {
	p.advance()
	scope, err := p.expectIdent()
	if err != nil {
		return ast.Namespace{}, err
	}
	name, err := p.expectIdent()
	if err != nil {
		return ast.Namespace{}, err
	}
	p.consumeOptionalDelim()
	return ast.Namespace{Scope: scope, Name: name}, nil
}

func (p *Parser) parseDecl() (ast.Decl, error) {
	switch {
	case p.isIdent("typedef"):
		return p.parseTypedef()
	case p.isIdent("const"):
		return p.parseConst()
	case p.isIdent("enum"):
		return p.parseEnum()
	case p.isIdent("struct"):
		return p.parseStructLike(ast.StructKind)
	case p.isIdent("union"):
		return p.parseStructLike(ast.UnionKind)
	case p.isIdent("exception"):
		return p.parseStructLike(ast.ExceptionKind)
	case p.isIdent("service"):
		return p.parseService()
	default:
		p.skipUnknownTopLevel()
		return nil, nil
	}
}

func (p *Parser) parseTypedef() (ast.Decl, error) {
	p.advance()
	t, err := p.parseTypeRef()
	if err != nil {
		return nil, err
	}
	alias, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	p.skipAnnotations()
	p.consumeOptionalDelim()
	return ast.Typedef{Alias: alias, Type: t}, nil
}

func (p *Parser) parseConst() (ast.Decl, error) {
	p.advance()
	t, err := p.parseTypeRef()
	if err != nil {
		return nil, err
	}
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if err := p.expectSymbol("="); err != nil {
		return nil, err
	}
	value := p.readValueUntilDelim()
	p.consumeOptionalDelim()
	return ast.Const{Name: name, Type: t, Value: value}, nil
}

func (p *Parser) parseEnum() (ast.Decl, error) {
	p.advance()
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	var vals []ast.EnumValue
	for !p.isSymbol("}") && p.cur.typ != tokEOF {
		if p.cur.typ != tokIdent {
			p.advance()
			continue
		}
		item := ast.EnumValue{Name: p.cur.lit}
		p.advance()
		if p.isSymbol("=") {
			p.advance()
			if p.cur.typ == tokNumber {
				v, _ := strconv.ParseInt(p.cur.lit, 10, 64)
				item.Value = &v
				p.advance()
			}
		}
		p.skipAnnotations()
		p.consumeOptionalDelim()
		vals = append(vals, item)
	}
	if err := p.expectSymbol("}"); err != nil {
		return nil, err
	}
	p.consumeOptionalDelim()
	return ast.Enum{Name: name, Values: vals}, nil
}

func (p *Parser) parseStructLike(kind ast.StructLikeKind) (ast.Decl, error) {
	p.advance()
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	var fields []ast.Field
	for !p.isSymbol("}") && p.cur.typ != tokEOF {
		if p.cur.typ == tokIdent && p.cur.lit == "}" {
			break
		}
		f, ok, err := p.parseField()
		if err != nil {
			return nil, err
		}
		if ok {
			fields = append(fields, f)
		} else {
			p.advance()
		}
	}
	if err := p.expectSymbol("}"); err != nil {
		return nil, err
	}
	p.skipAnnotations()
	p.consumeOptionalDelim()
	return ast.StructLike{Kind: kind, Name: name, Fields: fields}, nil
}

func (p *Parser) parseService() (ast.Decl, error) {
	p.advance()
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	svc := ast.Service{Name: name}
	if p.isIdent("extends") {
		p.advance()
		ext, err := p.expectIdent()
		if err != nil {
			return nil, err
		}
		svc.Extends = ext
	}
	if err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	for !p.isSymbol("}") && p.cur.typ != tokEOF {
		m, ok, err := p.parseMethod()
		if err != nil {
			return nil, err
		}
		if ok {
			svc.Methods = append(svc.Methods, m)
		} else {
			p.advance()
		}
	}
	if err := p.expectSymbol("}"); err != nil {
		return nil, err
	}
	p.skipAnnotations()
	p.consumeOptionalDelim()
	return svc, nil
}

func (p *Parser) parseMethod() (ast.Method, bool, error) {
	if p.cur.typ == tokEOF || p.isSymbol("}") {
		return ast.Method{}, false, nil
	}
	m := ast.Method{}
	if p.isIdent("oneway") {
		m.Oneway = true
		p.advance()
	}
	ret, err := p.parseTypeRef()
	if err != nil {
		return ast.Method{}, false, nil
	}
	m.ReturnType = ret
	name, err := p.expectIdent()
	if err != nil {
		return ast.Method{}, false, nil
	}
	m.Name = name
	if err := p.expectSymbol("("); err != nil {
		return ast.Method{}, false, nil
	}
	params, err := p.parseFieldListUntil(")")
	if err != nil {
		return ast.Method{}, false, nil
	}
	m.Params = params
	if p.isIdent("throws") {
		p.advance()
		if err := p.expectSymbol("("); err != nil {
			return ast.Method{}, false, nil
		}
		th, err := p.parseFieldListUntil(")")
		if err != nil {
			return ast.Method{}, false, nil
		}
		m.Throws = th
	}
	p.skipAnnotations()
	p.consumeOptionalDelim()
	return m, true, nil
}

func (p *Parser) parseFieldListUntil(end string) ([]ast.Field, error) {
	var out []ast.Field
	for !p.isSymbol(end) && p.cur.typ != tokEOF {
		f, ok, err := p.parseField()
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, f)
		} else {
			p.advance()
		}
	}
	if err := p.expectSymbol(end); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *Parser) parseField() (ast.Field, bool, error) {
	f := ast.Field{Requiredness: ast.RequirednessDefault}

	if p.cur.typ == tokNumber {
		if pv := p.peekToken(); pv.typ == tokSymbol && pv.lit == ":" {
			id64, _ := strconv.ParseInt(p.cur.lit, 10, 64)
			id := int(id64)
			f.ID = &id
			p.advance()
			if err := p.expectSymbol(":"); err != nil {
				return ast.Field{}, false, err
			}
		}
	}

	if p.isIdent("required") {
		f.Requiredness = ast.RequirednessRequired
		p.advance()
	} else if p.isIdent("optional") {
		f.Requiredness = ast.RequirednessOptional
		p.advance()
	}

	t, err := p.parseTypeRef()
	if err != nil {
		return ast.Field{}, false, nil
	}
	f.Type = t

	name, err := p.expectIdent()
	if err != nil {
		return ast.Field{}, false, nil
	}
	f.Name = name

	if p.isSymbol("=") {
		p.advance()
		f.Default = p.readValueUntilDelim()
	}
	p.skipAnnotations()
	p.consumeOptionalDelim()
	return f, true, nil
}

func (p *Parser) parseTypeRef() (ast.TypeRef, error) {
	if p.isIdent("list") {
		p.advance()
		if err := p.expectSymbol("<"); err != nil {
			return nil, err
		}
		elem, err := p.parseTypeRef()
		if err != nil {
			return nil, err
		}
		if err := p.expectSymbol(">"); err != nil {
			return nil, err
		}
		return ast.List{Elem: elem}, nil
	}
	if p.isIdent("set") {
		p.advance()
		if err := p.expectSymbol("<"); err != nil {
			return nil, err
		}
		elem, err := p.parseTypeRef()
		if err != nil {
			return nil, err
		}
		if err := p.expectSymbol(">"); err != nil {
			return nil, err
		}
		return ast.Set{Elem: elem}, nil
	}
	if p.isIdent("map") {
		p.advance()
		if err := p.expectSymbol("<"); err != nil {
			return nil, err
		}
		k, err := p.parseTypeRef()
		if err != nil {
			return nil, err
		}
		if err := p.expectSymbol(","); err != nil {
			return nil, err
		}
		v, err := p.parseTypeRef()
		if err != nil {
			return nil, err
		}
		if err := p.expectSymbol(">"); err != nil {
			return nil, err
		}
		return ast.Map{Key: k, Value: v}, nil
	}

	if p.cur.typ != tokIdent {
		return nil, p.errf("expected type, got %q", p.cur.lit)
	}
	name := p.cur.lit
	p.advance()

	switch name {
	case string(ast.TypeBool):
		return ast.Builtin{Name: ast.TypeBool}, nil
	case string(ast.TypeByte):
		return ast.Builtin{Name: ast.TypeByte}, nil
	case string(ast.TypeI16):
		return ast.Builtin{Name: ast.TypeI16}, nil
	case string(ast.TypeI32):
		return ast.Builtin{Name: ast.TypeI32}, nil
	case string(ast.TypeI64):
		return ast.Builtin{Name: ast.TypeI64}, nil
	case string(ast.TypeInt):
		return ast.Builtin{Name: ast.TypeInt}, nil
	case string(ast.TypeDouble):
		return ast.Builtin{Name: ast.TypeDouble}, nil
	case string(ast.TypeString):
		return ast.Builtin{Name: ast.TypeString}, nil
	case string(ast.TypeBinary):
		return ast.Builtin{Name: ast.TypeBinary}, nil
	case string(ast.TypeVoid):
		return ast.Builtin{Name: ast.TypeVoid}, nil
	default:
		return ast.Identifier{Name: name}, nil
	}
}

func (p *Parser) readValueUntilDelim() string {
	var b strings.Builder
	depth := 0
	hasRead := false
	for p.cur.typ != tokEOF {
		if p.isSymbol("{") || p.isSymbol("[") || p.isSymbol("(") {
			depth++
		}
		if p.isSymbol("}") || p.isSymbol("]") || p.isSymbol(")") {
			if depth == 0 {
				break
			}
			depth--
		}
		if depth == 0 && (p.isSymbol(",") || p.isSymbol(";")) {
			break
		}
		if depth == 0 && hasRead {
			if p.isSymbol("}") {
				break
			}
			if p.cur.typ == tokIdent && isTopLevelDeclKeyword(p.cur.lit) {
				break
			}
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		if p.cur.typ == tokString {
			b.WriteString(strconv.Quote(p.cur.lit))
		} else {
			b.WriteString(p.cur.lit)
		}
		hasRead = true
		p.advance()
	}
	return strings.TrimSpace(b.String())
}

func isTopLevelDeclKeyword(s string) bool {
	switch s {
	case "include", "namespace", "typedef", "const", "enum", "struct", "union", "exception", "service":
		return true
	default:
		return false
	}
}

func (p *Parser) skipAnnotations() {
	if !p.isSymbol("(") {
		return
	}
	depth := 0
	for p.cur.typ != tokEOF {
		if p.isSymbol("(") {
			depth++
		}
		if p.isSymbol(")") {
			depth--
			p.advance()
			if depth == 0 {
				break
			}
			continue
		}
		p.advance()
	}
}

func (p *Parser) skipUnknownTopLevel() {
	for p.cur.typ != tokEOF {
		if p.isSymbol(";") {
			p.advance()
			return
		}
		if p.isSymbol("{") {
			depth := 1
			p.advance()
			for p.cur.typ != tokEOF && depth > 0 {
				if p.isSymbol("{") {
					depth++
				} else if p.isSymbol("}") {
					depth--
				}
				p.advance()
			}
			p.consumeOptionalDelim()
			return
		}
		if p.isIdent("include") || p.isIdent("namespace") || p.isIdent("typedef") || p.isIdent("const") || p.isIdent("enum") || p.isIdent("struct") || p.isIdent("union") || p.isIdent("exception") || p.isIdent("service") {
			return
		}
		p.advance()
	}
}

func (p *Parser) errf(format string, args ...any) error {
	return fmt.Errorf("parse error at %d:%d: %s", p.cur.line, p.cur.col, fmt.Sprintf(format, args...))
}
