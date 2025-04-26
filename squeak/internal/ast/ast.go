package ast

// Node represent any type of node in an AST. Other than providing a means of viewing its literal representation it does
// not define any functional behaviours. This is by design as the Node abstraction is little more than a means of
// categorizing data.
type Node interface {
	Literal() string
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
