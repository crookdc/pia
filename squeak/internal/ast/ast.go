package ast

import "github.com/crookdc/pia/squeak/internal/token"

// Node represent any type of node in an AST. It does not define any functional behaviours. This is by design as the
// Node abstraction is little more than a means of categorizing data. It is not considered incorrect to implement the
// Node.Node method as nothing but a panic since it should never be called.
type Node interface {
	Node()
}

// ExpressionNode is specialization of [ast.Node] that does not provide any additional functional behaviors directly.
// The extra method ExpressionNode.ExpressionNode() is never meant to be called and the correct behavior of any concrete
// implementation is to panic when this method is invoked. Values that implement the ExpressionNode interface agrees to
// be processed as expressions in an AST.
type ExpressionNode interface {
	Node
	ExpressionNode()
}

// Expression is the default concrete implementation of [ast.ExpressionNode] and should be the first tool to reach for
// when defining a new expression. Its intended usage is to be embedded in structs that themselves provide the necessary
// expressive data.
type Expression struct{}

// ExpressionNode does nothing but panic. An explanation as to why is given in the documentation for
// [ast.ExpressionNode].
func (e Expression) ExpressionNode() {
	panic("Expression.ExpressionNode's behavior is classed as undefined and should never be invoked")
}

// Node does nothing but panic. An explanation as to why is given in the documentation for [ast.Node].
func (e Expression) Node() {
	panic("Expression.Node's behavior is classed as undefined and should never be invoked")
}

// StatementNode is specialization of [ast.Node] that does not provide any additional functional behaviors directly. The
// extra method StatementNode.StatementNode() is never meant to be called and the correct behavior of any concrete
// implementation is to panic when this method is invoked. Values that implement the StatementNode interface agrees to
// be processed as statements in an AST.
type StatementNode interface {
	Node
	StatementNode()
}

// Statement is the default concrete implementation of [ast.StatementNode] and should be the first tool to reach for
// when defining a new statement. Its intended usage is to be embedded in structs that themselves provide the necessary
// data to represent the statement.
type Statement struct{}

// StatementNode does nothing but panic. An explanation as to why is given in the documentation for [ast.StatementNode].
func (s Statement) StatementNode() {
	panic("Statement.StatementNode's behavior is classed as undefined and should never be invoked")
}

// Node does nothing but panic. An explanation as to why is given in the documentation for [ast.Node].
func (s Statement) Node() {
	panic("Statement.Node's behavior is classed as undefined and should never be invoked")
}

// LetStatement represents a statement which creates and initializes a new variable, for example: "let a = 5;".
type LetStatement struct {
	Statement
	Assignment ExpressionNode
}

// IfStatement represents a statement that allows the user to choose one of several branched statements to execute
// exclusively.
type IfStatement struct {
	Statement
	Condition   ExpressionNode
	Consequence StatementNode
	// Alternative represents any number of if-else pairs chained together or a single else statement which is
	// represented as an alternative statement with a Condition that always resolves to true.
	Alternative *IfStatement
}

// ExpressionStatement represents an expression that exists in isolation within a Squeak script, meaning that it is not
// defined as part of a statement and will thus be considered a statement by itself.
type ExpressionStatement struct {
	Statement
	Expression ExpressionNode
}

// IdentifierExpression represents an expression in the format of just an identifier.
type IdentifierExpression struct {
	Expression
	Identifier string
}

// IntegerExpression represents an expression which holds a primitive integer literal.
type IntegerExpression struct {
	Expression
	Integer int
}

// StringExpression represents an expression which holds a string literal.
type StringExpression struct {
	Expression
	String string
}

// BooleanExpression represents an expression which holds a boolean literal.
type BooleanExpression struct {
	Expression
	Boolean bool
}

type PrefixExpression struct {
	Expression
	Operator token.Token
	RHS      ExpressionNode
}

type InfixExpression struct {
	Expression
	Operator token.Token
	LHS      ExpressionNode
	RHS      ExpressionNode
}
