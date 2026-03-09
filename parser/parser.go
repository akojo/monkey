package parser

import (
	"fmt"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/token"
)

type Parser struct {
	l *lexer.Lexer

	errors    []error
	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: make([]error, 0)}

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
		return nil, fmt.Errorf("invalid token %q at start of statement", p.curToken.Literal)
	}
}

func (p *Parser) parseExpression() ast.Expression {
	for p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.EOF {
		p.nextToken()
	}
	return nil
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, error) {
	stmt := &ast.LetStatement{Token: p.curToken}

	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, err
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if err := p.expectPeek(token.ASSIGN); err != nil {
		return nil, err
	}

	stmt.Value = p.parseExpression()

	if err := p.expectPeek(token.SEMICOLON); err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression()

	if err := p.expectPeek(token.SEMICOLON); err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) expectPeek(t token.TokenType) error {
	if p.peekToken.Type != t {
		return fmt.Errorf("expected %q, got %q", t, p.peekToken.Literal)
	}
	p.nextToken()
	return nil
}
