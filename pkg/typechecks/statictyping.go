// Package typechecks provides methods for ensuring
// that a parsed program fulfills type constraints
package typechecks

import (
	"chogopy/pkg/ast"
)

type StaticTyping struct {
	localEnv    LocalEnvironment
	returnType  Type
	visitedType Type
	ast.BaseVisitor
}

// Analyze performs static type checking according to the rules defined in
// chapter 5 of the chocopy language reference:
//
// https://chocopy.org/chocopy_language_reference.pdf
func (st *StaticTyping) Analyze(program *ast.Program) {
	envBuilder := EnvironmentBuilder{}
	envBuilder.Build(program)

	st.localEnv = envBuilder.LocalEnv
	st.returnType = bottomType

	for _, definition := range program.Definitions {
		definition.Visit(st)
	}
	for _, statement := range program.Statements {
		statement.Visit(st)
	}
}

// Traverse determines whether the StaticTyping visitor should traverse further down the AST after a call to any nodes' Visit() method.
// If this returns true, each node will implicitly call the Visit() method on each of its
// child nodes whenever it is visited, which is something we want to avoid here.
// Instead, we will define when exactly a node will visit its child nodes via explicit calls
// to the child nodes' Visit() methods in each of the Visitors interface methods.
func (st *StaticTyping) Traverse() bool {
	return false
}
