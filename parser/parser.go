package parser

import (
	"fmt"
	"strconv"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/token"
)

type (
	prefixParseFn func() (ast.Expression, error)
	infixParseFn  func(ast.Expression) (ast.Expression, error)
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

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NE:       EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: make([]error, 0)}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NE, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)

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

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, error) {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = make([]ast.Statement, 0)

	p.nextToken()

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		block.Statements = append(block.Statements, stmt)
		p.nextToken()
	}

	return block, nil
}

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil, fmt.Errorf("unrecognized token type: %q", p.curToken.Type)
	}
	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}

	for p.peekToken.Type != token.SEMICOLON && precedence < p.peekPrecedence() {
		infix, found := p.infixParseFns[p.peekToken.Type]
		if !found {
			return leftExp, nil
		}

		p.nextToken()

		leftExp, err = infix(leftExp)
		if err != nil {
			return nil, err
		}
	}

	return leftExp, err
}

func (p *Parser) parseGroupedExpression() (ast.Expression, error) {
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if err := p.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}

	return exp, nil
}

func (p *Parser) parseIdentifier() (ast.Expression, error) {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}, nil
}

func (p *Parser) parseIntegerLiteral() (ast.Expression, error) {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer: %q", p.curToken.Literal)
	}

	lit.Value = value
	return lit, nil
}

func (p *Parser) parseBoolean() (ast.Expression, error) {
	return &ast.Boolean{Token: p.curToken, Value: p.curToken.Type == token.TRUE}, nil
}

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	var err error

	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right, err = p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) parseIfExpression() (ast.Expression, error) {
	var err error
	expression := &ast.IFExpression{Token: p.curToken}

	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, err
	}

	p.nextToken()
	expression.Condition, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if err := p.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, err
	}

	expression.Consequence, err = p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	if p.peekToken.Type != token.ELSE {
		return expression, nil
	}

	p.nextToken()

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, err
	}

	expression.Alternative, err = p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) parseFunctionLiteral() (ast.Expression, error) {
	var err error
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, fmt.Errorf("function: %w", err)
	}

	lit.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		return nil, fmt.Errorf("function: %w", err)
	}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, fmt.Errorf("function: %w", err)
	}

	lit.Body, err = p.parseBlockStatement()
	if err != nil {
		return nil, fmt.Errorf("function body: %w", err)
	}

	return lit, nil
}

func (p *Parser) parseFunctionParameters() ([]*ast.Identifier, error) {
	identifiers := make([]*ast.Identifier, 0)
	makeIdentifier := func() *ast.Identifier { return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal} }

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return identifiers, nil
	}

	p.nextToken()

	if p.curToken.Type != token.IDENT {
		return nil, fmt.Errorf("parameters: expected %s, got %s", token.IDENT, p.curToken)
	}
	identifiers = append(identifiers, makeIdentifier())

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()

		if p.curToken.Type != token.IDENT {
			return nil, fmt.Errorf("parameters: expected %s, got %s", token.IDENT, p.curToken)
		}
		identifiers = append(identifiers, makeIdentifier())
	}

	if err := p.expectPeek(token.RPAREN); err != nil {
		return nil, fmt.Errorf("parameters: %w", err)
	}

	return identifiers, nil
}

func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	var err error

	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	expression.Right, err = p.parseExpression(precedence)

	if err != nil {
		return nil, err
	}
	return expression, nil
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

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
