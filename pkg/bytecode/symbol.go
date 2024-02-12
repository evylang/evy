package bytecode

// SymbolScope defines a type of scope that a symbol can be defined inside.
type SymbolScope string

const (
	// GlobalScope is the top level scope of an evy program.
	GlobalScope SymbolScope = "GLOBAL"
)

// Symbol is a variable inside an evy program.
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable is a mapping of string identifiers to symbols.
type SymbolTable struct {
	store map[string]Symbol
}

// NewSymbolTable returns a new SymbolTable.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

// Define adds a symbol definition to the table or returns an
// already defined symbol with the same name.
func (s *SymbolTable) Define(name string) Symbol {
	if existing, found := s.store[name]; found {
		return existing
	}
	symbol := Symbol{Name: name, Index: len(s.store), Scope: GlobalScope}
	s.store[name] = symbol
	return symbol
}

// Resolve returns the Symbol with the specified name,
// or false if there is no such Symbol.
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	return obj, ok
}
