package language

import "C"
import (
	"errors"
	"github.com/pherrymason/c3-lsp/lsp/document"
	"github.com/pherrymason/c3-lsp/lsp/indexables"
	"github.com/pherrymason/c3-lsp/lsp/parser"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Language will be the center of knowledge of everything parsed.
type Language struct {
	index                  IndexStore
	symbolsByModule        map[protocol.DocumentUri]indexables.IndexableCollection
	functionTreeByDocument map[protocol.DocumentUri]indexables.Function
}

func NewLanguage() Language {
	return Language{
		index:                  NewIndexStore(),
		symbolsByModule:        make(map[protocol.DocumentUri]indexables.IndexableCollection),
		functionTreeByDocument: make(map[protocol.DocumentUri]indexables.Function),
	}
}

func (l *Language) RefreshDocumentIdentifiers(doc *document.Document, parser *parser.Parser) {
	l.functionTreeByDocument[doc.URI] = parser.ExtractSymbols(doc)
}

func (l *Language) BuildCompletionList(text string, line protocol.UInteger, character protocol.UInteger) []protocol.CompletionItem {
	var items []protocol.CompletionItem
	for _, value := range l.symbolsByModule {
		for _, storedIdentifier := range value {
			tempKind := storedIdentifier.GetKind()

			items = append(items, protocol.CompletionItem{
				Label: storedIdentifier.GetName(),
				Kind:  &tempKind,
			})
		}
	}

	return items
}

func (l *Language) registerIndexable(doc *document.Document, indexable indexables.Indexable) {
	l.symbolsByModule[doc.URI] = append(l.symbolsByModule[doc.URI], indexable)
}

const (
	AnyPosition FindMode = iota
	InPosition
)

type FindMode int

func (l *Language) FindSymbolDeclarationInWorkspace(docId protocol.DocumentUri, identifier string, position protocol.Position) (indexables.Indexable, error) {

	symbol := l.findClosestSymbolDeclaration(identifier, docId, position)

	return symbol, nil
}

func (l *Language) FindHoverInformation(doc *document.Document, params *protocol.HoverParams) (protocol.Hover, error) {
	word, err := doc.WordInPosition(params.Position)
	if err != nil {
		return protocol.Hover{}, err
	}

	symbol := l.findClosestSymbolDeclaration(word, params.TextDocument.URI, params.Position)
	if symbol == nil {
		return protocol.Hover{}, nil
	}

	// expected behaviour:
	// hovering on variables: display variable type + any description
	// hovering on functions: display function signature
	// hovering on members: same as variable
	hover := protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: symbol.GetHoverInfo(),
		},
	}

	return hover, nil
}

// Finds the closest symbol based on current scope.
// If not present in current Scope:
// - Search in imported files (TODO)
// - Search in global symbols in workspace
func (l *Language) findClosestSymbolDeclaration(word string, docId protocol.DocumentUri, position protocol.Position) indexables.Indexable {
	identifier, _ := l.findSymbolDeclarationInDocPositionScope(word, docId, position)
	if identifier != nil {
		return identifier
	}

	// TODO search in imported files in docId
	// -----------

	// Not found yet, let's try search the symbol defined as global in other files
	for _, scope := range l.functionTreeByDocument {
		found, foundDepth := findDeepFirst(word, position, &scope, 0, AnyPosition)

		if found != nil && (foundDepth <= 1) {
			return found
		}
	}

	// Not found...
	return nil
}

// Search for symbol in docId
func (l *Language) findSymbolDeclarationInDocPositionScope(identifier string, docId protocol.DocumentUri, position protocol.Position) (indexables.Indexable, error) {
	scopedTree, ok := l.functionTreeByDocument[docId]
	if !ok {
		return nil, errors.New("Document is not indexed")
	}

	// Go through every element defined in scopedTree
	symbol, _ := findDeepFirst(identifier, position, &scopedTree, 0, InPosition)
	return symbol, nil
}

func findDeepFirst(identifier string, position protocol.Position, function *indexables.Function, depth uint, mode FindMode) (indexables.Indexable, uint) {
	if mode == InPosition &&
		!function.GetDeclarationRange().HasPosition(position) {
		return nil, depth
	}

	if identifier == function.GetName() {
		return function, depth
	}

	for _, child := range function.ChildrenFunctions {
		if result, resultDepth := findDeepFirst(identifier, position, &child, depth+1, mode); result != nil {
			return result, resultDepth
		}
	}

	variable, foundVariableInThisScope := function.Variables[identifier]
	if foundVariableInThisScope {
		return variable, depth
	}

	enum, foundEnumInThisScope := function.Enums[identifier]
	if foundEnumInThisScope {
		return enum, depth
	}

	var enumerator indexables.Enumerator
	foundEnumeratorInThisScope := false
	for _, scopedEnums := range function.Enums {
		if scopedEnums.HasEnumerator(identifier) {
			enumerator = scopedEnums.GetEnumerator(identifier)
			foundEnumeratorInThisScope = true
		}
	}
	if foundEnumeratorInThisScope {
		return enumerator, depth
	}

	_struct, foundStructInThisScope := function.Structs[identifier]
	if foundStructInThisScope {
		return _struct, depth
	}

	def, foundDefInScope := function.Defs[identifier]
	if foundDefInScope {
		return def, depth
	}

	return nil, depth
}