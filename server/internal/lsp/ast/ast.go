package ast

import sitter "github.com/smacker/go-tree-sitter"

type Position struct {
	Line, Column uint
}

const (
	ResolveStatusPending = iota
	ResolveStatusDone
)

type ASTNodeBase struct {
	StartPos, EndPos Position
	Attributes       []string
}

func (n ASTNodeBase) Start() Position {
	return n.StartPos
}

func (n ASTNodeBase) End() Position {
	return n.EndPos
}

func (n *ASTNodeBase) SetPos(start sitter.Point, end sitter.Point) {
	n.StartPos = Position{Line: uint(start.Row), Column: uint(start.Column)}
	n.EndPos = Position{Line: uint(end.Row), Column: uint(end.Column)}
}

type ASTNode interface {
	Start() Position
	End() Position
}

type File struct {
	ASTNodeBase
	Modules []Module
}

type Module struct {
	ASTNodeBase
	Name              string
	GenericParameters []string
	Declarations      []Declaration
	Imports           []string
}

type Declaration interface {
	ASTNode
}

type VariableDecl struct {
	ASTNodeBase
	Names []Identifier
	Type  TypeInfo
	//Initializer Initializer
}

type EnumDecl struct {
	ASTNodeBase
	Name       string
	BaseType   TypeInfo
	Properties []EnumProperty
	Members    []EnumMember
}

type EnumProperty struct {
	ASTNodeBase
	Type TypeInfo
	Name Identifier
}

type EnumMember struct {
	ASTNodeBase
	Name       Identifier
	Value      string
	Properties []PropertyValue
}

type PropertyValue struct {
	ASTNodeBase
	Name  string
	Value Expression
}

type FunctionDecl struct {
	ASTNodeBase
	Name       *Identifier
	Parameters []*Identifier
	ReturnType *Identifier
	Body       Block
}

type Block struct {
	ASTNodeBase
	Statements []ASTNode
}

type FunctionCall struct {
	ASTNodeBase
}

type TypeInfo struct {
	ASTNodeBase
	ResolveStatus int
	Name          string
	Pointer       uint
	Optional      bool
	BuiltIn       bool
	Generics      []TypeInfo
}

type Identifier struct {
	ASTNodeBase
	Name string
}

type Expression interface {
	ASTNode
}

type Literal struct {
	ASTNodeBase
	Value string
}

type CompositeLiteral struct {
	ASTNodeBase
	Values []Expression
}

// BinaryExpr representa una expresión binaria (como suma, resta, etc.)
type BinaryExpr struct {
	ASTNodeBase
	Left     ASTNode
	Operator string
	Right    ASTNode
}