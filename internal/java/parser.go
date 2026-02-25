// Copyright 2025 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate antlr4 -Dlanguage=Go -package java20 -o java20 Java20Lexer.g4 Java20Parser.g4
package java

import (
	"sync"

	"github.com/antlr4-go/antlr/v4"

	"github.com/ossf/scorecard/v5/internal/java/java20"
)

// parserMutex protects concurrent access to the Java parser.
// ANTLR-generated parsers have shared static state that isn't thread-safe.
var parserMutex sync.Mutex

// File represents a parsed Java source file.
type File struct {
	TypeNames []*TypeNameSpec
}

// TypeNameSpec represents a type name occurrence.
type TypeNameSpec struct {
	pos  antlr.Token
	Name string
}

// Pos returns the position of the type name.
func (t *TypeNameSpec) Pos() antlr.Token {
	return t.pos
}

// typeNameListener walks the parse tree to collect type name occurrences.
type typeNameListener struct {
	*java20.BaseJava20ParserListener
	tokens    *antlr.CommonTokenStream
	typeNames []*TypeNameSpec
}

func (l *typeNameListener) EnterTypeName(ctx *java20.TypeNameContext) {
	// Collect TypeName occurrences (used in import statements)
	typeName := ctx.GetText()
	startToken := ctx.GetStart()

	l.typeNames = append(l.typeNames, &TypeNameSpec{
		Name: typeName,
		pos:  startToken,
	})
}

func (l *typeNameListener) EnterUnannType(ctx *java20.UnannTypeContext) {
	// Collect type references (in variable declarations, method parameters, field declarations, casts, etc.)
	typeName := ctx.GetText()
	startToken := ctx.GetStart()

	l.typeNames = append(l.typeNames, &TypeNameSpec{
		Name: typeName,
		pos:  startToken,
	})
}

func (l *typeNameListener) EnterReferenceType(ctx *java20.ReferenceTypeContext) {
	// Collect reference type references (in casts, instanceof, etc.)
	typeName := ctx.GetText()
	startToken := ctx.GetStart()

	l.typeNames = append(l.typeNames, &TypeNameSpec{
		Name: typeName,
		pos:  startToken,
	})
}

// ParseFile parses Java source code and returns a File with type name information.
func ParseFile(content []byte) (*File, error) {
	// Lock to protect shared static state in ANTLR-generated parser
	parserMutex.Lock()
	defer parserMutex.Unlock()

	is := antlr.NewInputStream(string(content))
	lexer := java20.NewJava20Lexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := java20.NewJava20Parser(stream)

	// Disable error output for now
	parser.RemoveErrorListeners()

	tree := parser.CompilationUnit()

	listener := &typeNameListener{
		tokens: stream,
	}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)

	return &File{
		TypeNames: listener.typeNames,
	}, nil
}
