package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/akojo/monkey/token"
)

type Node interface {
	TokenLiteral() string
	String() string
	PrettyPrint(level int) string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) == 0 {
		return ""
	}
	return p.Statements[0].TokenLiteral()
}

func (p *Program) String() string {
	return p.PrettyPrint(0)
}

func (p *Program) PrettyPrint(level int) string {
	var out bytes.Buffer

	for _, stmt := range p.Statements {
		out.WriteString(stmt.PrettyPrint(level) + ";\n")
	}

	return out.String()
}

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

func (ls *LetStatement) String() string {
	return ls.PrettyPrint(0)
}

func (ls *LetStatement) PrettyPrint(level int) string {
	return fmt.Sprintf("%slet %s = %s", indent(level), ls.Name.String(), ls.Value.PrettyPrint(0))
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

func (rs *ReturnStatement) String() string {
	return rs.PrettyPrint(0)
}

func (rs *ReturnStatement) PrettyPrint(level int) string {
	return fmt.Sprintf("%sreturn %s", indent(level), rs.ReturnValue.String())
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string       { return es.PrettyPrint(0) }
func (es *ExpressionStatement) PrettyPrint(level int) string {
	return es.Expression.PrettyPrint(level)
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }
func (i *Identifier) PrettyPrint(level int) string {
	return fmt.Sprintf("%s%s", indent(level), i.Value)
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
func (il *IntegerLiteral) PrettyPrint(level int) string {
	return fmt.Sprintf("%s%s", indent(level), il.Token.Literal)
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }
func (b *Boolean) PrettyPrint(level int) string {
	return fmt.Sprintf("%s%s", indent(level), b.Token.Literal)
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string       { return pe.PrettyPrint(0) }
func (pe *PrefixExpression) PrettyPrint(level int) string {
	return fmt.Sprintf("%s(%s%s)", indent(level), pe.Operator, pe.Right)
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string       { return ie.PrettyPrint(0) }
func (ie *InfixExpression) PrettyPrint(level int) string {
	return fmt.Sprintf("%s(%s %s %s)", indent(level), ie.Left, ie.Operator, ie.Right)
}

type IFExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IFExpression) expressionNode()      {}
func (ie *IFExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IFExpression) String() string       { return ie.PrettyPrint(0) }
func (ie *IFExpression) PrettyPrint(level int) string {
	in := indent(level)
	if ie.Alternative == nil {
		return fmt.Sprintf("%sif %s %s", in, ie.Condition.String(), ie.Consequence.PrettyPrint(level))
	}
	return fmt.Sprintf("%sif %s %s else %s", in, ie.Condition.String(), ie.Consequence.PrettyPrint(level), ie.Alternative.PrettyPrint(level))

}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string       { return bs.PrettyPrint(0) }
func (bs *BlockStatement) PrettyPrint(level int) string {
	var out bytes.Buffer

	out.WriteString("{\n")
	for _, s := range bs.Statements {
		fmt.Fprintf(&out, "%s;\n", s.PrettyPrint(level+1))
	}
	out.WriteString(indent(level) + "}")
	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string       { return fl.PrettyPrint(0) }
func (fl *FunctionLiteral) PrettyPrint(level int) string {
	params := make([]string, 0)
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("%s%s (%s) %s", indent(level), fl.TokenLiteral(), strings.Join(params, ", "), fl.Body.PrettyPrint(level))
}

type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string       { return ce.PrettyPrint(0) }
func (ce *CallExpression) PrettyPrint(level int) string {
	args := make([]string, 0)
	for _, arg := range ce.Arguments {
		args = append(args, arg.String())
	}

	return fmt.Sprintf("%s(%s)", ce.Function.PrettyPrint(level), strings.Join(args, ", "))
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }
func (sl *StringLiteral) PrettyPrint(level int) string {
	return fmt.Sprintf("%s\"%s\"", indent(level), sl.Token.Literal)
}

func indent(level int) string {
	var out bytes.Buffer
	for range level {
		out.WriteString("  ")
	}
	return out.String()
}
