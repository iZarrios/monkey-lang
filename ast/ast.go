package ast

import "github.com/iZarrios/monkey-lang/token"

type Node interface {
	TokenLiteral() string // Used for testing and debugging
}

/*
 NOTE:
	These interfaces only contain dummy methods called statementNode and
	expressionNode respectively. They are not strictly necessary but help us by guiding the Go
	compiler and possibly causing it to throw errors when we use a Statement where an Expression
	should’ve been used, and vice versa.
*/

type Statement interface {
	Node
	statementNode() // Dummy method to throw errors
}

type Expression interface {
	Node
	expressionNode() // Dummy method to throw errors
}

// NOTE: The program Node is going to be the root node of every AST our parser produces
// A program is just a series of statements
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

type Identifier struct {
	Token token.Token // the token.IDENT tokekn
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

type ReturnStatement struct {
	Token            token.Token // the 'return' token
	ReturnExrepssion Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
