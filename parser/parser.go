package parser

import (
	"errors"

	"github.com/iZarrios/monkey-lang/ast"
	"github.com/iZarrios/monkey-lang/lexer"
	"github.com/iZarrios/monkey-lang/token"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func NewParser(l *lexer.Lexer) (*Parser, error) {
	if l == nil {
		return nil, errors.New("can't pass lexer as nil")
	}

	p := &Parser{
		l:              l,
		errors:         make([]string, 0),
		prefixParseFns: make(map[token.TokenType]prefixParseFn),
		infixParseFns:  make(map[token.TokenType]infixParseFn),
	}

	// Read two tokens (to set both curToken and peekToken)
	p.nextToken()
	p.nextToken()

	{ // PREFIX
		p.registerPrefix(token.IDENT, p.parseIdentifier)
		p.registerPrefix(token.INT, p.parseIntegerLiteral)
		p.registerPrefix(token.BANG, p.parsePrefixExpression)
		p.registerPrefix(token.MINUS, p.parsePrefixExpression)
		p.registerPrefix(token.TRUE, p.parseBoolean)
		p.registerPrefix(token.FALSE, p.parseBoolean)
		p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
		p.registerPrefix(token.IF, p.parseIfExpression)
		p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
		p.registerPrefix(token.STRING, p.parseStringLiteral)
		p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
		p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	}

	{ // INFIX
		p.infixParseFns = make(map[token.TokenType]infixParseFn)
		p.registerInfix(token.PLUS, p.parseInfixExpression)
		p.registerInfix(token.MINUS, p.parseInfixExpression)
		p.registerInfix(token.SLASH, p.parseInfixExpression)
		p.registerInfix(token.ASTERISK, p.parseInfixExpression)
		p.registerInfix(token.EQ, p.parseInfixExpression)
		p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
		p.registerInfix(token.LT, p.parseInfixExpression)
		p.registerInfix(token.GT, p.parseInfixExpression)
		p.registerInfix(token.LPAREN, p.parseCallExpression)
		p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	}

	return p, nil
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()

		program.Statements = append(program.Statements, stmt)

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExrepssionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.curToken.Type != token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExrepssionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))

	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != token.SEMICOLON && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil

	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == token.ELSE {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		stmt := p.parseStatement()

		// //TODO: not needed?
		// if stmt != nil {
		block.Statements = append(block.Statements, stmt)
		// }
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fnLiteral := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fnLiteral.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fnLiteral.Body = p.parseBlockStatement()

	return fnLiteral
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {

	idents := []*ast.Identifier{}

	// if there are no params
	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return idents
	}

	// we have params in the current function literal
	p.nextToken()

	// get first param
	ident := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	idents = append(idents, ident)

	for p.peekToken.Type == token.COMMA {
		// make current = comma
		p.nextToken()
		// make current = ident
		p.nextToken()
		ident := &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		idents = append(idents, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		// no closing ')'
		return nil
	}

	return idents
}

// `function` is the left operand
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	callExpr := &ast.CallExpression{
		Token:    p.curToken,
		Function: function,
	}

	callExpr.Arguments = p.parseExpressionList(token.RPAREN)

	return callExpr
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekToken.Type == end {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.curToken}

	arr.Elements = p.parseExpressionList(token.RBRACKET)

	return arr
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken}

	exp.Left = left

	p.nextToken()

	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for p.peekToken.Type != token.RBRACE {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		if p.peekToken.Type != token.RBRACE && !p.expectPeek(token.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}
