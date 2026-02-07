package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type OpCode byte

const (
	OP_NOOP OpCode = iota
	OP_PUSH_INT
	OP_PUSH_STR
	OP_PUSH_BOOL
	OP_PUSH_NULL
	OP_LOAD
	OP_STORE
	OP_PRINT
	OP_INPUT
	OP_WAIT
	OP_CHECK
	OP_JMP
	OP_JMP_IF_FALSE
	OP_CALL
	OP_RET
	OP_POP
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_EQ
	OP_NEQ
	OP_LT
	OP_GT
	OP_AND
	OP_OR
	OP_NOT
	OP_NEG
	OP_MAKE_LIST
)

type Instr struct {
	Op   OpCode
	Str  string
	Int  int
	Bool bool
}

type Function struct {
	Params []string
	Code   []Instr
}

type Program struct {
	Version   int
	Main      []Instr
	Functions map[string]Function
}

type tokenKind int

const (
	tokLine tokenKind = iota
	tokLBrace
	tokRBrace
)

type Token struct {
	Kind tokenKind
	Text string
}

type Compiler struct {
	tokens []Token
	idx    int

	functions map[string]Function
}

type loopCtx struct {
	start         int
	breakJumps    []int
	continueJumps []int
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  emeraldc build <file.emer>")
		fmt.Println("  emeraldc run <file.emer|file.emec>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	path := os.Args[2]

	switch cmd {
	case "build":
		if !strings.HasSuffix(strings.ToLower(path), ".emer") {
			fmt.Println("build expects a .emer source file")
			os.Exit(1)
		}
		prog, err := compileFile(path)
		if err != nil {
			fmt.Printf("Compile error: %v\n", err)
			os.Exit(1)
		}
		outPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".emec"
		if err := writeProgram(outPath, prog); err != nil {
			fmt.Printf("Write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %s\n", outPath)
	case "run":
		if strings.HasSuffix(strings.ToLower(path), ".emec") {
			prog, err := readProgram(path)
			if err != nil {
				fmt.Printf("Load error: %v\n", err)
				os.Exit(1)
			}
			vm := NewVM()
			if _, err := vm.Exec(prog, prog.Main); err != nil {
				fmt.Printf("Runtime error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		if !strings.HasSuffix(strings.ToLower(path), ".emer") {
			fmt.Println("run expects a .emer or .emec file")
			os.Exit(1)
		}
		prog, err := compileFile(path)
		if err != nil {
			fmt.Printf("Compile error: %v\n", err)
			os.Exit(1)
		}
		vm := NewVM()
		if _, err := vm.Exec(prog, prog.Main); err != nil {
			fmt.Printf("Runtime error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}

func compileFile(path string) (*Program, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tokens, err := tokenize(string(data))
	if err != nil {
		return nil, err
	}
	c := &Compiler{tokens: tokens, functions: map[string]Function{}}
	mainCode := []Instr{}
	if err := c.compileBlock(&mainCode, nil); err != nil {
		return nil, err
	}
	prog := &Program{Version: 2, Main: mainCode, Functions: c.functions}
	return prog, nil
}

func writeProgram(path string, prog *Program) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(prog)
}

func readProgram(path string) (*Program, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	var prog Program
	if err := dec.Decode(&prog); err != nil {
		return nil, err
	}
	return &prog, nil
}

func tokenize(src string) ([]Token, error) {
	tokens := []Token{}
	var line strings.Builder
	inString := false
	var quote rune

	flushLine := func() {
		s := strings.TrimSpace(line.String())
		if s != "" {
			tokens = append(tokens, Token{Kind: tokLine, Text: s})
		}
		line.Reset()
	}

	for _, r := range src {
		if inString {
			line.WriteRune(r)
			if r == quote {
				inString = false
			}
			continue
		}
		switch r {
		case '\'', '"':
			inString = true
			quote = r
			line.WriteRune(r)
		case '{':
			flushLine()
			tokens = append(tokens, Token{Kind: tokLBrace, Text: "{"})
		case '}':
			flushLine()
			tokens = append(tokens, Token{Kind: tokRBrace, Text: "}"})
		case '\n':
			flushLine()
		case '\r':
		default:
			line.WriteRune(r)
		}
	}
	flushLine()
	return tokens, nil
}

func (c *Compiler) next() *Token {
	if c.idx >= len(c.tokens) {
		return nil
	}
	t := &c.tokens[c.idx]
	c.idx++
	return t
}

func (c *Compiler) peek() *Token {
	if c.idx >= len(c.tokens) {
		return nil
	}
	return &c.tokens[c.idx]
}

func (c *Compiler) compileBlock(code *[]Instr, loopStack []loopCtx) error {
	for {
		t := c.next()
		if t == nil {
			return nil
		}
		if t.Kind == tokRBrace {
			return nil
		}
		if t.Kind == tokLBrace {
			return errors.New("unexpected '{'")
		}

		line := t.Text
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "--") {
			continue
		}

		if strings.HasPrefix(line, "fnc ") {
			name, params, err := parseFnDecl(line)
			if err != nil {
				return err
			}
			if p := c.next(); p == nil || p.Kind != tokLBrace {
				return errors.New("fnc must be followed by '{'")
			}
			fnCode := []Instr{}
			if err := c.compileBlock(&fnCode, nil); err != nil {
				return err
			}
			c.functions[name] = Function{Params: params, Code: fnCode}
			continue
		}

		if strings.HasPrefix(line, "var") {
			if err := c.compileVar(line, code); err != nil {
				return err
			}
			continue
		}

		if strings.HasPrefix(line, "print(") && strings.HasSuffix(line, ")") {
			inner := strings.TrimSpace(line[len("print(") : len(line)-1])
			if err := c.compileExpr(inner, code); err != nil {
				return err
			}
			*code = append(*code, Instr{Op: OP_PRINT})
			continue
		}

		if strings.HasPrefix(line, "input(plc(") && strings.HasSuffix(line, "))") {
			inner := strings.TrimSpace(line[len("input(plc(") : len(line)-2])
			s, ok := parseStringLiteral(inner)
			if !ok {
				return errors.New("input requires a string literal prompt")
			}
			*code = append(*code, Instr{Op: OP_INPUT, Str: s})
			continue
		}

		if strings.HasPrefix(line, "wait(") && strings.HasSuffix(line, ")") {
			inner := strings.TrimSpace(line[len("wait(") : len(line)-1])
			if err := c.compileExpr(inner, code); err != nil {
				return err
			}
			*code = append(*code, Instr{Op: OP_WAIT})
			continue
		}

		if strings.HasPrefix(line, "check ") {
			inner := strings.TrimSpace(strings.TrimPrefix(line, "check "))
			if err := c.compileExpr(inner, code); err != nil {
				return err
			}
			*code = append(*code, Instr{Op: OP_CHECK, Str: inner})
			continue
		}

		if line == "brk" {
			if len(loopStack) == 0 {
				return errors.New("brk used outside loop")
			}
			idx := len(*code)
			*code = append(*code, Instr{Op: OP_JMP, Int: -1})
			loopStack[len(loopStack)-1].breakJumps = append(loopStack[len(loopStack)-1].breakJumps, idx)
			continue
		}

		if line == "cont" {
			if len(loopStack) == 0 {
				return errors.New("cont used outside loop")
			}
			idx := len(*code)
			*code = append(*code, Instr{Op: OP_JMP, Int: -1})
			loopStack[len(loopStack)-1].continueJumps = append(loopStack[len(loopStack)-1].continueJumps, idx)
			continue
		}

		if line == "return" || strings.HasPrefix(line, "return ") {
			expr := strings.TrimSpace(strings.TrimPrefix(line, "return"))
			if expr != "" {
				if err := c.compileExpr(expr, code); err != nil {
					return err
				}
				*code = append(*code, Instr{Op: OP_RET})
			} else {
				*code = append(*code, Instr{Op: OP_PUSH_NULL})
				*code = append(*code, Instr{Op: OP_RET})
			}
			continue
		}

		if strings.HasPrefix(line, "for(") && strings.HasSuffix(line, ")") {
			cond := strings.TrimSpace(line[len("for(") : len(line)-1])
			if p := c.next(); p == nil || p.Kind != tokLBrace {
				return errors.New("for must be followed by '{'")
			}
			loopStart := len(*code)
			hasCond := strings.ToLower(cond) != "true"
			var jmpIfFalseIdx int
			if hasCond {
				if err := c.compileExpr(cond, code); err != nil {
					return err
				}
				jmpIfFalseIdx = len(*code)
				*code = append(*code, Instr{Op: OP_JMP_IF_FALSE, Int: -1})
			}

			ctx := loopCtx{start: loopStart}
			loopStack = append(loopStack, ctx)
			if err := c.compileBlock(code, loopStack); err != nil {
				return err
			}
			*code = append(*code, Instr{Op: OP_JMP, Int: loopStart})
			loopEnd := len(*code)
			if hasCond {
				(*code)[jmpIfFalseIdx].Int = loopEnd
			}
			cur := loopStack[len(loopStack)-1]
			for _, bi := range cur.breakJumps {
				(*code)[bi].Int = loopEnd
			}
			for _, ci := range cur.continueJumps {
				(*code)[ci].Int = loopStart
			}
			loopStack = loopStack[:len(loopStack)-1]
			continue
		}

		if strings.HasPrefix(line, "while(") && strings.HasSuffix(line, ")") {
			cond := strings.TrimSpace(line[len("while(") : len(line)-1])
			if p := c.next(); p == nil || p.Kind != tokLBrace {
				return errors.New("while must be followed by '{'")
			}
			loopStart := len(*code)
			if err := c.compileExpr(cond, code); err != nil {
				return err
			}
			jmpIfFalseIdx := len(*code)
			*code = append(*code, Instr{Op: OP_JMP_IF_FALSE, Int: -1})
			ctx := loopCtx{start: loopStart}
			loopStack = append(loopStack, ctx)
			if err := c.compileBlock(code, loopStack); err != nil {
				return err
			}
			*code = append(*code, Instr{Op: OP_JMP, Int: loopStart})
			loopEnd := len(*code)
			(*code)[jmpIfFalseIdx].Int = loopEnd
			cur := loopStack[len(loopStack)-1]
			for _, bi := range cur.breakJumps {
				(*code)[bi].Int = loopEnd
			}
			for _, ci := range cur.continueJumps {
				(*code)[ci].Int = loopStart
			}
			loopStack = loopStack[:len(loopStack)-1]
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			name := strings.TrimSpace(parts[0])
			expr := strings.TrimSpace(parts[1])
			if name == "" {
				return errors.New("invalid assignment")
			}
			if err := c.compileExpr(expr, code); err != nil {
				return err
			}
			*code = append(*code, Instr{Op: OP_STORE, Str: name})
			continue
		}

		if strings.HasSuffix(line, ")") {
			if err := c.compileExpr(line, code); err == nil {
				*code = append(*code, Instr{Op: OP_POP})
				continue
			}
		}

		return fmt.Errorf("unknown statement: %s", line)
	}
}

func parseFnDecl(line string) (string, []string, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(line, "fnc "))
	if rest == "" {
		return "", nil, errors.New("fnc requires a name")
	}
	name := rest
	params := []string{}
	if i := strings.Index(rest, "("); i != -1 {
		if !strings.HasSuffix(rest, ")") {
			return "", nil, errors.New("fnc params must end with ')' before '{'")
		}
		name = strings.TrimSpace(rest[:i])
		inner := strings.TrimSpace(rest[i+1 : len(rest)-1])
		if inner != "" {
			parts := splitTopLevel(inner, ',')
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				params = append(params, p)
			}
		}
	}
	if name == "" {
		return "", nil, errors.New("fnc requires a name")
	}
	return name, params, nil
}

func (c *Compiler) compileVar(line string, code *[]Instr) error {
	trimmed := strings.TrimSpace(line)
	if trimmed == "var" {
		if p := c.next(); p == nil || p.Kind != tokLBrace {
			return errors.New("var block must be followed by '{'")
		}
		itemsText := c.collectBlockText()
		specs := splitTopLevel(itemsText, ',')
		for _, spec := range specs {
			spec = strings.TrimSpace(spec)
			if spec == "" {
				continue
			}
			if err := c.compileVarSpec(spec, code); err != nil {
				return err
			}
		}
		return nil
	}

	if strings.HasPrefix(trimmed, "var ") {
		spec := strings.TrimSpace(strings.TrimPrefix(trimmed, "var "))
		if strings.HasPrefix(spec, "{") {
			itemsText := strings.TrimSpace(strings.TrimPrefix(spec, "{"))
			if strings.HasSuffix(itemsText, "}") {
				itemsText = strings.TrimSpace(strings.TrimSuffix(itemsText, "}"))
			} else {
				itemsText += " " + c.collectBlockText()
			}
			specs := splitTopLevel(itemsText, ',')
			for _, s := range specs {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				if err := c.compileVarSpec(s, code); err != nil {
					return err
				}
			}
			return nil
		}
		return c.compileVarSpec(spec, code)
	}

	return nil
}

func (c *Compiler) collectBlockText() string {
	var b strings.Builder
	depth := 1
	for depth > 0 {
		t := c.next()
		if t == nil {
			break
		}
		if t.Kind == tokLBrace {
			depth++
			continue
		}
		if t.Kind == tokRBrace {
			depth--
			if depth == 0 {
				break
			}
			continue
		}
		if t.Kind == tokLine {
			if b.Len() > 0 {
				b.WriteString(" ")
			}
			b.WriteString(t.Text)
		}
	}
	return b.String()
}
func (c *Compiler) compileVarSpec(spec string, code *[]Instr) error {
	name, typ, expr, hasExpr, err := parseVarSpec(spec)
	if err != nil {
		return err
	}
	if !hasExpr {
		*code = append(*code, Instr{Op: OP_PUSH_NULL})
		*code = append(*code, Instr{Op: OP_STORE, Str: name})
		return nil
	}
	if err := c.compileExpr(expr, code); err != nil {
		return err
	}
	_ = typ
	*code = append(*code, Instr{Op: OP_STORE, Str: name})
	return nil
}

func parseVarSpec(spec string) (name string, typ string, expr string, hasExpr bool, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", "", "", false, errors.New("empty var spec")
	}

	nameEnd := strings.IndexAny(spec, "(=")
	if nameEnd == -1 {
		return strings.TrimSpace(spec), "", "", false, nil
	}
	name = strings.TrimSpace(spec[:nameEnd])
	rest := strings.TrimSpace(spec[nameEnd:])
	if strings.HasPrefix(rest, "(") {
		closeIdx := strings.Index(rest, ")")
		if closeIdx == -1 {
			return "", "", "", false, errors.New("missing ')' in type")
		}
		typ = strings.TrimSpace(rest[1:closeIdx])
		rest = strings.TrimSpace(rest[closeIdx+1:])
	}

	if strings.HasPrefix(rest, "=") {
		rest = strings.TrimSpace(rest[1:])
		if strings.HasPrefix(rest, "(") && strings.HasSuffix(rest, ")") {
			expr = strings.TrimSpace(rest[1 : len(rest)-1])
		} else {
			expr = strings.TrimSpace(rest)
		}
		hasExpr = expr != ""
	}

	if name == "" {
		return "", "", "", false, errors.New("invalid var name")
	}
	return name, typ, expr, hasExpr, nil
}

func splitTopLevel(s string, sep rune) []string {
	parts := []string{}
	var cur strings.Builder
	inString := false
	var quote rune
	parenDepth := 0
	bracketDepth := 0

	for _, r := range s {
		if inString {
			cur.WriteRune(r)
			if r == quote {
				inString = false
			}
			continue
		}
		switch r {
		case '\'', '"':
			inString = true
			quote = r
			cur.WriteRune(r)
		case '(':
			parenDepth++
			cur.WriteRune(r)
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
			cur.WriteRune(r)
		case '[':
			bracketDepth++
			cur.WriteRune(r)
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			cur.WriteRune(r)
		default:
			if r == sep && parenDepth == 0 && bracketDepth == 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			} else {
				cur.WriteRune(r)
			}
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// Expression parsing

type exprTokenKind int

const (
	exTokEOF exprTokenKind = iota
	exTokIdent
	exTokNumber
	exTokString
	exTokLParen
	exTokRParen
	exTokLBracket
	exTokRBracket
	exTokComma
	exTokOp
)

type exprToken struct {
	kind exprTokenKind
	text string
}

type exprParser struct {
	tokens []exprToken
	pos    int
}

type ExprKind int

const (
	exprLiteral ExprKind = iota
	exprIdent
	exprUnary
	exprBinary
	exprCall
	exprList
)

type Expr struct {
	Kind     ExprKind
	LitKind  string
	I        int
	S        string
	B        bool
	Name     string
	Op       string
	Left     *Expr
	Right    *Expr
	Args     []*Expr
	Elements []*Expr
}

func (p *exprParser) peek() exprToken {
	if p.pos >= len(p.tokens) {
		return exprToken{kind: exTokEOF}
	}
	return p.tokens[p.pos]
}

func (p *exprParser) next() exprToken {
	t := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return t
}

func (p *exprParser) expect(kind exprTokenKind, text string) error {
	t := p.next()
	if t.kind != kind || (text != "" && t.text != text) {
		return fmt.Errorf("unexpected token: %s", t.text)
	}
	return nil
}

func (p *exprParser) parseExpr() (*Expr, error) {
	return p.parseOr()
}

func (p *exprParser) parseOr() (*Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.kind == exTokOp && t.text == "or" {
			p.next()
			right, err := p.parseAnd()
			if err != nil {
				return nil, err
			}
			left = &Expr{Kind: exprBinary, Op: "or", Left: left, Right: right}
			continue
		}
		break
	}
	return left, nil
}

func (p *exprParser) parseAnd() (*Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.kind == exTokOp && t.text == "and" {
			p.next()
			right, err := p.parseEquality()
			if err != nil {
				return nil, err
			}
			left = &Expr{Kind: exprBinary, Op: "and", Left: left, Right: right}
			continue
		}
		break
	}
	return left, nil
}

func (p *exprParser) parseEquality() (*Expr, error) {
	left, err := p.parseCompare()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.kind == exTokOp && (t.text == "==" || t.text == "!=") {
			p.next()
			right, err := p.parseCompare()
			if err != nil {
				return nil, err
			}
			left = &Expr{Kind: exprBinary, Op: t.text, Left: left, Right: right}
			continue
		}
		break
	}
	return left, nil
}

func (p *exprParser) parseCompare() (*Expr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.kind == exTokOp && (t.text == "<" || t.text == ">") {
			p.next()
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			left = &Expr{Kind: exprBinary, Op: t.text, Left: left, Right: right}
			continue
		}
		break
	}
	return left, nil
}

func (p *exprParser) parseTerm() (*Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.kind == exTokOp && (t.text == "+" || t.text == "-") {
			p.next()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			left = &Expr{Kind: exprBinary, Op: t.text, Left: left, Right: right}
			continue
		}
		break
	}
	return left, nil
}

func (p *exprParser) parseFactor() (*Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.kind == exTokOp && (t.text == "*" || t.text == "/") {
			p.next()
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &Expr{Kind: exprBinary, Op: t.text, Left: left, Right: right}
			continue
		}
		break
	}
	return left, nil
}

func (p *exprParser) parseUnary() (*Expr, error) {
	t := p.peek()
	if t.kind == exTokOp && (t.text == "not" || t.text == "-") {
		p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &Expr{Kind: exprUnary, Op: t.text, Left: expr}, nil
	}
	return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (*Expr, error) {
	t := p.next()
	switch t.kind {
	case exTokNumber:
		i, _ := strconv.Atoi(t.text)
		return &Expr{Kind: exprLiteral, LitKind: "int", I: i}, nil
	case exTokString:
		return &Expr{Kind: exprLiteral, LitKind: "string", S: t.text}, nil
	case exTokIdent:
		switch t.text {
		case "true":
			return &Expr{Kind: exprLiteral, LitKind: "bool", B: true}, nil
		case "false":
			return &Expr{Kind: exprLiteral, LitKind: "bool", B: false}, nil
		case "null":
			return &Expr{Kind: exprLiteral, LitKind: "null"}, nil
		}
		if p.peek().kind == exTokLParen {
			p.next()
			args := []*Expr{}
			if p.peek().kind != exTokRParen {
				for {
					arg, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
					if p.peek().kind == exTokComma {
						p.next()
						continue
					}
					break
				}
			}
			if err := p.expect(exTokRParen, ""); err != nil {
				return nil, err
			}
			return &Expr{Kind: exprCall, Name: t.text, Args: args}, nil
		}
		return &Expr{Kind: exprIdent, Name: t.text}, nil
	case exTokLParen:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(exTokRParen, ""); err != nil {
			return nil, err
		}
		return expr, nil
	case exTokLBracket:
		elems := []*Expr{}
		if p.peek().kind != exTokRBracket {
			for {
				e, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				elems = append(elems, e)
				if p.peek().kind == exTokComma {
					p.next()
					continue
				}
				break
			}
		}
		if err := p.expect(exTokRBracket, ""); err != nil {
			return nil, err
		}
		return &Expr{Kind: exprList, Elements: elems}, nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", t.text)
	}
}

func lexExpr(s string) ([]exprToken, error) {
	toks := []exprToken{}
	for i := 0; i < len(s); {
		r := s[i]
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			i++
			continue
		}
		if isLetter(r) || r == '_' {
			start := i
			i++
			for i < len(s) && (isLetter(s[i]) || isDigit(s[i]) || s[i] == '_') {
				i++
			}
			text := s[start:i]
			if text == "and" || text == "or" || text == "not" {
				toks = append(toks, exprToken{kind: exTokOp, text: text})
			} else {
				toks = append(toks, exprToken{kind: exTokIdent, text: text})
			}
			continue
		}
		if isDigit(r) {
			start := i
			i++
			for i < len(s) && isDigit(s[i]) {
				i++
			}
			toks = append(toks, exprToken{kind: exTokNumber, text: s[start:i]})
			continue
		}
		if r == '\'' || r == '"' {
			quote := r
			start := i + 1
			i++
			for i < len(s) && s[i] != quote {
				i++
			}
			if i >= len(s) {
				return nil, errors.New("unterminated string")
			}
			text := s[start:i]
			i++
			toks = append(toks, exprToken{kind: exTokString, text: text})
			continue
		}

		switch r {
		case '(':
			toks = append(toks, exprToken{kind: exTokLParen, text: "("})
			i++
		case ')':
			toks = append(toks, exprToken{kind: exTokRParen, text: ")"})
			i++
		case '[':
			toks = append(toks, exprToken{kind: exTokLBracket, text: "["})
			i++
		case ']':
			toks = append(toks, exprToken{kind: exTokRBracket, text: "]"})
			i++
		case ',':
			toks = append(toks, exprToken{kind: exTokComma, text: ","})
			i++
		case '+', '-', '*', '/', '<', '>':
			toks = append(toks, exprToken{kind: exTokOp, text: string(r)})
			i++
		case '=':
			if i+1 < len(s) && s[i+1] == '=' {
				toks = append(toks, exprToken{kind: exTokOp, text: "=="})
				i += 2
			} else {
				return nil, errors.New("unexpected '='")
			}
		case '!':
			if i+1 < len(s) && s[i+1] == '=' {
				toks = append(toks, exprToken{kind: exTokOp, text: "!="})
				i += 2
			} else {
				return nil, errors.New("unexpected '!'")
			}
		default:
			return nil, fmt.Errorf("unexpected character: %c", r)
		}
	}
	toks = append(toks, exprToken{kind: exTokEOF})
	return toks, nil
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func (c *Compiler) compileExpr(expr string, code *[]Instr) error {
	toks, err := lexExpr(expr)
	if err != nil {
		return err
	}
	p := &exprParser{tokens: toks}
	root, err := p.parseExpr()
	if err != nil {
		return err
	}
	if p.peek().kind != exTokEOF {
		return errors.New("unexpected trailing tokens")
	}
	return c.compileExprNode(root, code)
}

func (c *Compiler) compileExprNode(node *Expr, code *[]Instr) error {
	switch node.Kind {
	case exprLiteral:
		switch node.LitKind {
		case "int":
			*code = append(*code, Instr{Op: OP_PUSH_INT, Int: node.I})
		case "string":
			*code = append(*code, Instr{Op: OP_PUSH_STR, Str: node.S})
		case "bool":
			*code = append(*code, Instr{Op: OP_PUSH_BOOL, Bool: node.B})
		case "null":
			*code = append(*code, Instr{Op: OP_PUSH_NULL})
		default:
			return errors.New("unknown literal")
		}
		return nil
	case exprIdent:
		*code = append(*code, Instr{Op: OP_LOAD, Str: node.Name})
		return nil
	case exprUnary:
		if err := c.compileExprNode(node.Left, code); err != nil {
			return err
		}
		switch node.Op {
		case "not":
			*code = append(*code, Instr{Op: OP_NOT})
		case "-":
			*code = append(*code, Instr{Op: OP_NEG})
		default:
			return fmt.Errorf("unsupported unary op: %s", node.Op)
		}
		return nil
	case exprBinary:
		if err := c.compileExprNode(node.Left, code); err != nil {
			return err
		}
		if err := c.compileExprNode(node.Right, code); err != nil {
			return err
		}
		switch node.Op {
		case "+":
			*code = append(*code, Instr{Op: OP_ADD})
		case "-":
			*code = append(*code, Instr{Op: OP_SUB})
		case "*":
			*code = append(*code, Instr{Op: OP_MUL})
		case "/":
			*code = append(*code, Instr{Op: OP_DIV})
		case "==":
			*code = append(*code, Instr{Op: OP_EQ})
		case "!=":
			*code = append(*code, Instr{Op: OP_NEQ})
		case "<":
			*code = append(*code, Instr{Op: OP_LT})
		case ">":
			*code = append(*code, Instr{Op: OP_GT})
		case "and":
			*code = append(*code, Instr{Op: OP_AND})
		case "or":
			*code = append(*code, Instr{Op: OP_OR})
		default:
			return fmt.Errorf("unsupported op: %s", node.Op)
		}
		return nil
	case exprCall:
		for _, arg := range node.Args {
			if err := c.compileExprNode(arg, code); err != nil {
				return err
			}
		}
		*code = append(*code, Instr{Op: OP_CALL, Str: node.Name, Int: len(node.Args)})
		return nil
	case exprList:
		for _, e := range node.Elements {
			if err := c.compileExprNode(e, code); err != nil {
				return err
			}
		}
		*code = append(*code, Instr{Op: OP_MAKE_LIST, Int: len(node.Elements)})
		return nil
	default:
		return errors.New("unknown expr")
	}
}

func parseStringLiteral(s string) (string, bool) {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1], true
	}
	return "", false
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !(r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
				return false
			}
			continue
		}
		if !(r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}
type Value struct {
	Kind string
	I    int
	S    string
	B    bool
	L    []Value
	M    map[string]Value
}

type VM struct {
	globals map[string]Value
	reader  *bufio.Reader
}

func NewVM() *VM {
	return &VM{globals: map[string]Value{}, reader: bufio.NewReader(os.Stdin)}
}

func (vm *VM) Exec(prog *Program, code []Instr) (Value, error) {
	return vm.execWithLocals(prog, code, nil)
}

func (vm *VM) execWithLocals(prog *Program, code []Instr, locals map[string]Value) (Value, error) {
	stack := []Value{}
	pc := 0
	for pc < len(code) {
		ins := code[pc]
		switch ins.Op {
		case OP_PUSH_INT:
			stack = append(stack, Value{Kind: "int", I: ins.Int})
		case OP_PUSH_STR:
			stack = append(stack, Value{Kind: "string", S: ins.Str})
		case OP_PUSH_BOOL:
			stack = append(stack, Value{Kind: "bool", B: ins.Bool})
		case OP_PUSH_NULL:
			stack = append(stack, Value{Kind: "null"})
		case OP_LOAD:
			if locals != nil {
				if v, ok := locals[ins.Str]; ok {
					stack = append(stack, v)
					break
				}
			}
			v, ok := vm.globals[ins.Str]
			if !ok {
				v = Value{Kind: "null"}
			}
			stack = append(stack, v)
		case OP_STORE:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if locals != nil {
				locals[ins.Str] = v
			} else {
				vm.globals[ins.Str] = v
			}
		case OP_PRINT:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			fmt.Println(v.String())
		case OP_INPUT:
			fmt.Print(ins.Str)
			text, _ := vm.reader.ReadString('\n')
			text = strings.TrimRight(text, "\r\n")
			vm.globals["LAST_INPUT"] = Value{Kind: "string", S: text}
		case OP_WAIT:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			ms := v.AsInt()
			time.Sleep(time.Duration(ms) * time.Millisecond)
		case OP_CHECK:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !v.Truthy() {
				return Value{}, fmt.Errorf("check failed: %s", ins.Str)
			}
		case OP_JMP:
			pc = ins.Int
			continue
		case OP_JMP_IF_FALSE:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !v.Truthy() {
				pc = ins.Int
				continue
			}
		case OP_CALL:
			args, err := popArgs(&stack, ins.Int)
			if err != nil {
				return Value{}, err
			}
			if builtin, ok := vm.callBuiltin(ins.Str, args); ok {
				stack = append(stack, builtin)
				break
			}
			fn, ok := prog.Functions[ins.Str]
			if !ok {
				return Value{}, fmt.Errorf("unknown function: %s", ins.Str)
			}
			localsFrame := map[string]Value{}
			for i, name := range fn.Params {
				if i < len(args) {
					localsFrame[name] = args[i]
				} else {
					localsFrame[name] = Value{Kind: "null"}
				}
			}
			ret, err := vm.execWithLocals(prog, fn.Code, localsFrame)
			if err != nil {
				return Value{}, err
			}
			stack = append(stack, ret)
		case OP_RET:
			if len(stack) == 0 {
				return Value{Kind: "null"}, nil
			}
			v := stack[len(stack)-1]
			return v, nil
		case OP_POP:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			stack = stack[:len(stack)-1]
		case OP_ADD:
			if err := binOp(&stack, opAdd); err != nil {
				return Value{}, err
			}
		case OP_SUB:
			if err := binOp(&stack, opSub); err != nil {
				return Value{}, err
			}
		case OP_MUL:
			if err := binOp(&stack, opMul); err != nil {
				return Value{}, err
			}
		case OP_DIV:
			if err := binOp(&stack, opDiv); err != nil {
				return Value{}, err
			}
		case OP_EQ:
			if err := binOp(&stack, opEq); err != nil {
				return Value{}, err
			}
		case OP_NEQ:
			if err := binOp(&stack, opNeq); err != nil {
				return Value{}, err
			}
		case OP_LT:
			if err := binOp(&stack, opLt); err != nil {
				return Value{}, err
			}
		case OP_GT:
			if err := binOp(&stack, opGt); err != nil {
				return Value{}, err
			}
		case OP_AND:
			if err := binOp(&stack, opAnd); err != nil {
				return Value{}, err
			}
		case OP_OR:
			if err := binOp(&stack, opOr); err != nil {
				return Value{}, err
			}
		case OP_NOT:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			stack = append(stack, Value{Kind: "bool", B: !v.Truthy()})
		case OP_NEG:
			if len(stack) == 0 {
				return Value{}, errors.New("stack underflow")
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if v.Kind != "int" {
				return Value{}, errors.New("unary '-' expects int")
			}
			stack = append(stack, Value{Kind: "int", I: -v.I})
		case OP_MAKE_LIST:
			if len(stack) < ins.Int {
				return Value{}, errors.New("stack underflow")
			}
			items := make([]Value, ins.Int)
			for i := ins.Int - 1; i >= 0; i-- {
				items[i] = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, Value{Kind: "list", L: items})
		case OP_NOOP:
			// no-op
		default:
			return Value{}, errors.New("unknown opcode")
		}
		pc++
	}
	return Value{Kind: "null"}, nil
}

func (vm *VM) callBuiltin(name string, args []Value) (Value, bool) {
	switch name {
	case "table", "dict":
		if len(args)%2 != 0 {
			return Value{}, true
		}
		m := map[string]Value{}
		for i := 0; i < len(args); i += 2 {
			key := args[i]
			k := key.String()
			m[k] = args[i+1]
		}
		return Value{Kind: "dict", M: m}, true
	default:
		return Value{}, false
	}
}

func popArgs(stack *[]Value, n int) ([]Value, error) {
	if len(*stack) < n {
		return nil, errors.New("stack underflow")
	}
	args := make([]Value, n)
	for i := n - 1; i >= 0; i-- {
		args[i] = (*stack)[len(*stack)-1]
		*stack = (*stack)[:len(*stack)-1]
	}
	return args, nil
}

func binOp(stack *[]Value, fn func(Value, Value) (Value, error)) error {
	if len(*stack) < 2 {
		return errors.New("stack underflow")
	}
	r := (*stack)[len(*stack)-1]
	l := (*stack)[len(*stack)-2]
	*stack = (*stack)[:len(*stack)-2]
	v, err := fn(l, r)
	if err != nil {
		return err
	}
	*stack = append(*stack, v)
	return nil
}

func opAdd(a, b Value) (Value, error) {
	if a.Kind == "int" && b.Kind == "int" {
		return Value{Kind: "int", I: a.I + b.I}, nil
	}
	if a.Kind == "string" || b.Kind == "string" {
		return Value{Kind: "string", S: a.String() + b.String()}, nil
	}
	if a.Kind == "list" && b.Kind == "list" {
		out := append([]Value{}, a.L...)
		out = append(out, b.L...)
		return Value{Kind: "list", L: out}, nil
	}
	return Value{}, errors.New("'+' expects int, string, or list")
}

func opSub(a, b Value) (Value, error) {
	if a.Kind == "int" && b.Kind == "int" {
		return Value{Kind: "int", I: a.I - b.I}, nil
	}
	return Value{}, errors.New("'-' expects int")
}

func opMul(a, b Value) (Value, error) {
	if a.Kind == "int" && b.Kind == "int" {
		return Value{Kind: "int", I: a.I * b.I}, nil
	}
	return Value{}, errors.New("'*' expects int")
}

func opDiv(a, b Value) (Value, error) {
	if a.Kind == "int" && b.Kind == "int" {
		if b.I == 0 {
			return Value{}, errors.New("division by zero")
		}
		return Value{Kind: "int", I: a.I / b.I}, nil
	}
	return Value{}, errors.New("'/' expects int")
}

func opEq(a, b Value) (Value, error) {
	return Value{Kind: "bool", B: a.Equal(b)}, nil
}

func opNeq(a, b Value) (Value, error) {
	return Value{Kind: "bool", B: !a.Equal(b)}, nil
}

func opLt(a, b Value) (Value, error) {
	if a.Kind == "int" && b.Kind == "int" {
		return Value{Kind: "bool", B: a.I < b.I}, nil
	}
	if a.Kind == "string" && b.Kind == "string" {
		return Value{Kind: "bool", B: a.S < b.S}, nil
	}
	return Value{}, errors.New("'<' expects int or string")
}

func opGt(a, b Value) (Value, error) {
	if a.Kind == "int" && b.Kind == "int" {
		return Value{Kind: "bool", B: a.I > b.I}, nil
	}
	if a.Kind == "string" && b.Kind == "string" {
		return Value{Kind: "bool", B: a.S > b.S}, nil
	}
	return Value{}, errors.New("'>' expects int or string")
}

func opAnd(a, b Value) (Value, error) {
	return Value{Kind: "bool", B: a.Truthy() && b.Truthy()}, nil
}

func opOr(a, b Value) (Value, error) {
	return Value{Kind: "bool", B: a.Truthy() || b.Truthy()}, nil
}

func (v Value) Truthy() bool {
	switch v.Kind {
	case "null":
		return false
	case "bool":
		return v.B
	case "int":
		return v.I != 0
	case "string":
		return v.S != ""
	case "list":
		return len(v.L) > 0
	case "dict":
		return len(v.M) > 0
	default:
		return false
	}
}

func (v Value) AsInt() int {
	switch v.Kind {
	case "int":
		return v.I
	case "bool":
		if v.B {
			return 1
		}
		return 0
	case "string":
		if i, err := strconv.Atoi(v.S); err == nil {
			return i
		}
	}
	return 0
}

func (v Value) Equal(o Value) bool {
	if v.Kind != o.Kind {
		return false
	}
	switch v.Kind {
	case "null":
		return true
	case "bool":
		return v.B == o.B
	case "int":
		return v.I == o.I
	case "string":
		return v.S == o.S
	case "list":
		if len(v.L) != len(o.L) {
			return false
		}
		for i := range v.L {
			if !v.L[i].Equal(o.L[i]) {
				return false
			}
		}
		return true
	case "dict":
		if len(v.M) != len(o.M) {
			return false
		}
		for k, vv := range v.M {
			ov, ok := o.M[k]
			if !ok || !vv.Equal(ov) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (v Value) String() string {
	switch v.Kind {
	case "null":
		return "null"
	case "bool":
		if v.B {
			return "true"
		}
		return "false"
	case "int":
		return strconv.Itoa(v.I)
	case "string":
		return v.S
	case "list":
		parts := make([]string, 0, len(v.L))
		for _, it := range v.L {
			parts = append(parts, it.String())
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case "dict":
		parts := []string{}
		for k, val := range v.M {
			parts = append(parts, k+": "+val.String())
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return ""
	}
}
