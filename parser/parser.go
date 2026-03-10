package parser

import (
	"fmt"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	errors    []error
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

const (
	_ int = iota
	LOWEST
	EQUALS      // ==, !=
	LESSGREATER // <, >
	SUM         // +, -
	PRODUCT     // *, /
	PREFIX      // -x, !x
	CALL        // function(x)
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: make([]error, 0)}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []error {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	for p.curToken.Type != token.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			pos := p.l.Position
			p.errors = append(p.errors, fmt.Errorf("%s:%d:%d: %w", pos.Filename, pos.Line, pos.Column, err))
		} else {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	var err error
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, error) {
	var err error
	stmt := &ast.LetStatement{Token: p.curToken}

	if err = p.expectPeek(token.IDENT); err != nil {
		return nil, err
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if err = p.expectPeek(token.ASSIGN); err != nil {
		return nil, err
	}

	for p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.EOF {
		p.nextToken()
	}

	if err = p.expectPeek(token.SEMICOLON); err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	for p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.EOF {
		p.nextToken()
	}

	if err := p.expectPeek(token.SEMICOLON); err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil, nil
	}
	leftExp := prefix()

	return leftExp, nil
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) expectPeek(t token.TokenType) error {
	if p.peekToken.Type != t {
		return fmt.Errorf("expected %q, got %q", t, p.peekToken.Literal)
	}
	p.nextToken()
	return nil
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
