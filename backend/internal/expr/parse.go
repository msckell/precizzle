package expr

import (
	"errors"
	"strings"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// The size limits of D45 and D50.
const (
	maxOperations = 100
	maxPowers     = 10
	maxDepth      = 100
)

const sqrtToken = "sqrt("

// node is a parsed expression: an *operand, a *unary, a *binary or a *sum.
type node any

// operand is a number or a value in brackets.
type operand struct {
	value calc.Number
}

// unary is a negation ('-'), a percentage ('%') or a square root ('s'), at
// the position of its operator.
type unary struct {
	op  byte
	pos int
	x   node
}

// binary is a product ('*'), a quotient ('/') or a power ('^'), at the
// position of its operator.
type binary struct {
	op   byte
	pos  int
	x, y node
}

// sum is a sequence of terms joined by + and -.
type sum struct {
	terms []sumTerm
}

// sumTerm is one term of a sum, with the operator in front of it (0 for the
// first term) and the percentages that follow it (D43).
type sumTerm struct {
	op       byte
	pos      int
	x        node
	percents []percent
}

// percent is a term that is only a percentage, B% with or without minus signs
// in front. It follows + or - and applies to the term before it (D43).
type percent struct {
	op   byte
	negs []int
	base node
	pos  int
}

// operator is the position of an operation in the text, for the limits of D45.
type operator struct {
	pos   int
	power bool
}

// parser reads an expression from left to right with one function per rule
// of the grammar in docs/SPEC.md. It reads each token only when the grammar
// asks for it, so the first problem in the text is the one reported.
type parser struct {
	s     string
	i     int
	depth int
	ops   []operator
	// outOfRange is the position of the first operand outside the limits, or
	// -1. It is reported only after the whole syntax is checked.
	outOfRange int
}

// parse checks the syntax, the operand limits and the number of operations
// of s, in that order, and returns its tree.
func parse(s string) (node, error) {
	p := &parser{s: s, outOfRange: -1}
	n, err := p.sum()
	if err != nil {
		return nil, err
	}
	if p.peek(); p.i < len(s) {
		return nil, &Error{ErrInvalidExpression, p.i}
	}
	if p.outOfRange >= 0 {
		return nil, &Error{calc.ErrOperandOutOfRange, p.outOfRange}
	}
	powers := 0
	for k, o := range p.ops {
		if o.power {
			powers++
		}
		if k >= maxOperations || powers > maxPowers {
			return nil, &Error{ErrExpressionTooComplex, o.pos}
		}
	}
	return n, nil
}

// peek skips spaces and returns the next character, or 0 at the end. A 0
// before the end is a NUL character, so callers test the end with p.i.
func (p *parser) peek() byte {
	for p.i < len(p.s) && p.s[p.i] == ' ' {
		p.i++
	}
	if p.i == len(p.s) {
		return 0
	}
	return p.s[p.i]
}

// consume records the operator at the current position and moves past it.
func (p *parser) consume(power bool) int {
	pos := p.i
	p.ops = append(p.ops, operator{pos, power})
	p.i++
	return pos
}

// sum = term { ("+" | "-") term }
func (p *parser) sum() (node, error) {
	x, _, err := p.term()
	if err != nil {
		return nil, err
	}
	terms := []sumTerm{{x: x}}
	for c := p.peek(); c == '+' || c == '-'; c = p.peek() {
		pos := p.consume(false)
		y, isPercent, err := p.term()
		if err != nil {
			return nil, err
		}
		if isPercent {
			last := &terms[len(terms)-1]
			last.percents = append(last.percents, asPercent(c, y))
		} else {
			terms = append(terms, sumTerm{op: c, pos: pos, x: y})
		}
	}
	if len(terms) == 1 && len(terms[0].percents) == 0 {
		return x, nil
	}
	return &sum{terms}, nil
}

// asPercent unwraps a term that is only a percentage, -…-B%, as the parser built it.
func asPercent(op byte, x node) percent {
	p := percent{op: op}
	for {
		u := x.(*unary)
		if u.op == '%' {
			p.base, p.pos = u.x, u.pos
			return p
		}
		p.negs = append(p.negs, u.pos)
		x = u.x
	}
}

// term = signed { ("*" | "/") signed }. It also reports whether the term is
// only a percentage, with or without minus signs in front.
func (p *parser) term() (node, bool, error) {
	x, isPercent, err := p.signed()
	for c := p.peek(); err == nil && (c == '*' || c == '/'); c = p.peek() {
		pos := p.consume(false)
		var y node
		y, _, err = p.signed()
		x, isPercent = &binary{c, pos, x, y}, false
	}
	return x, isPercent, err
}

// signed = "-" signed | power. The minus signs are read in a loop, so a long
// run of them does not deepen the recursion.
func (p *parser) signed() (node, bool, error) {
	var negs []int
	for p.peek() == '-' {
		negs = append(negs, p.consume(false))
	}
	x, isPercent, err := p.power()
	for k := len(negs) - 1; k >= 0; k-- {
		x = &unary{'-', negs[k], x}
	}
	return x, isPercent, err
}

// power = postfix [ "^" signed ]
func (p *parser) power() (node, bool, error) {
	x, isPercent, err := p.postfix()
	if err != nil || p.peek() != '^' {
		return x, isPercent, err
	}
	pos := p.consume(true)
	y, _, err := p.signed()
	return &binary{'^', pos, x, y}, false, err
}

// postfix = primary [ "%" ]
func (p *parser) postfix() (node, bool, error) {
	x, err := p.primary()
	if err != nil || p.peek() != '%' {
		return x, false, err
	}
	pos := p.consume(false)
	return &unary{'%', pos, x}, true, nil
}

// primary = number | value | "(" sum ")" | "sqrt(" sum ")"
func (p *parser) primary() (node, error) {
	c := p.peek()
	start := p.i
	switch {
	case p.i == len(p.s):
		// An operand is missing, so the error points to the end (its length).
		return nil, &Error{ErrInvalidExpression, p.i}
	case isDigit(c) || c == '.':
		p.i = numberEnd(p.s, p.i)
		return p.operand(p.s[start:p.i], start)
	case c == '[':
		end := strings.IndexByte(p.s[start:], ']')
		if end < 0 {
			return nil, &Error{ErrInvalidExpression, start}
		}
		p.i = start + end + 1
		return p.operand(p.s[start+1:start+end], start)
	case c == '(':
		p.i++
		return p.group(start)
	case c == 's':
		// sqrt( is one token, matched character by character, so the error
		// points to the first character that differs (D54).
		for k := range len(sqrtToken) {
			if p.i == len(p.s) || p.s[p.i] != sqrtToken[k] {
				return nil, &Error{ErrInvalidExpression, p.i}
			}
			p.i++
		}
		p.ops = append(p.ops, operator{pos: start})
		x, err := p.group(start)
		if err != nil {
			return nil, err
		}
		return &unary{'s', start, x}, nil
	}
	return nil, &Error{ErrInvalidExpression, p.i}
}

// group reads a sum and its closing parenthesis. open is the position of the
// opening ( or sqrt(.
func (p *parser) group(open int) (node, error) {
	if p.depth == maxDepth {
		return nil, &Error{ErrExpressionTooComplex, open}
	}
	p.depth++
	x, err := p.sum()
	if err != nil {
		return nil, err
	}
	if p.peek() != ')' {
		if p.i == len(p.s) {
			return nil, &Error{ErrInvalidExpression, open}
		}
		return nil, &Error{ErrInvalidExpression, p.i}
	}
	p.i++
	p.depth--
	return x, nil
}

// operand reads a number or the text of a value with the operand syntax. A
// value outside the limits is recorded, because the limits are checked after
// the syntax.
func (p *parser) operand(s string, pos int) (node, error) {
	n, err := calc.Parse(s)
	switch {
	case errors.Is(err, calc.ErrOperandOutOfRange):
		if p.outOfRange < 0 {
			p.outOfRange = pos
		}
	case err != nil:
		return nil, &Error{calc.ErrInvalidNumber, pos}
	}
	return &operand{n}, nil
}

// numberEnd returns the end of the number that starts at i: a run of digits,
// '.', 'e' and 'E', and a '+' or '-' right after 'e' or 'E'.
func numberEnd(s string, i int) int {
	for ; i < len(s); i++ {
		c := s[i]
		sign := (c == '+' || c == '-') && (s[i-1] == 'e' || s[i-1] == 'E')
		if !isDigit(c) && c != '.' && c != 'e' && c != 'E' && !sign {
			break
		}
	}
	return i
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
