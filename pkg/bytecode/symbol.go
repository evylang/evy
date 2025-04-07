package bytecode

// SymbolScope defines a type of scope that a symbol can be defined inside.
type SymbolScope string

const (
	// GlobalScope is the top level scope of an evy program.
	GlobalScope SymbolScope = "GLOBAL"
	// LocalScope is any local scope in an evy program, it is distinct
	// from the GlobalScope.
	LocalScope SymbolScope = "LOCAL"
)

// Symbol is a variable inside an evy program.
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable translates the variables of an Evy program to memory locations.
// It maps variable names (or symbols) encountered during compilation to their
// corresponding index-addressed memory locations, which are used at run time.
// This enables efficient access and manipulation of variables by the Evy
// virtual machine.
//
// Global variables, which are declared at the top level of a program and
// outside of any blocks, are stored in a dedicated global memory space. All
// other variables are considered local variables and are stored on the program
// stack.
//
// Index mapping for symbols involves tracking variable counts within each
// scope of the Evy program. As new variables are defined within a scope, the
// corresponding index is incremented (see [SymbolTable.Define] method). Upon
// entering a new block scope, a dedicated SymbolTable is created and pushed
// onto a stack ([SymbolTable.Push] method) to manage variables within that
// block. When exiting the block scope, the SymbolTable is popped off the stack
// ([SymbolTable.Pop] method), and the maximum index value is propagated back
// to the parent scope. This allows the maximum number of symbols in scope at
// the same time to be known so we do not over-allocate space for them at run
// time.
type SymbolTable struct {
	store map[string]Symbol
	// index keeps track of the number of locals inside this scope,
	// an enclosed SymbolTable will inherit a starting index from
	// its outer table.
	index int
	// nestedMaxIndex is the maximum index of any inner symbol
	// tables. It propagates up the SymbolTable chain as SymbolTables
	// are Popped so know how much space to allocate for variables.
	nestedMaxIndex int
	// outer is a SymbolTable that encloses this one, Resolve will
	// travel up the stack of tables looking for a symbol. If outer
	// is nil, this SymbolTable is the global symbol table.
	outer *SymbolTable
}

// NewSymbolTable returns a new SymbolTable.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

// Push nests a new symbol table under s. The index for the nested table starts
// at the index of the outer symbol table if the outer symbol table is not the
// global symbol table. Otherwise the index starts at zero.
func (s *SymbolTable) Push() *SymbolTable {
	index := s.index
	// If our outer symbol table is the global symbol table, we do
	// not count its index as globals are stored in a separate space.
	if s.outer == nil {
		index = 0
	}
	return &SymbolTable{
		store: make(map[string]Symbol),
		outer: s,
		index: index,
	}
}

// Pop returns the outer symbol table of s and updates the outer symbol table's
// nested max index. This accumulates the maximum index of all nested scopes
// into the parent scope so we know how much storage is needed for all the
// variables of a scope and its children.
func (s *SymbolTable) Pop() *SymbolTable {
	if s.outer == nil {
		return s
	}
	s.outer.nestedMaxIndex = max(s.outer.nestedMaxIndex, s.nestedMaxIndex+s.index)
	return s.outer
}

// Define adds a symbol definition to the table or returns an
// already defined symbol with the same name.
func (s *SymbolTable) Define(name string) Symbol {
	if existing, found := s.store[name]; found {
		return existing
	}
	symbol := Symbol{Name: name, Index: s.index}
	s.index++
	if s.outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	return symbol
}

// Resolve returns the Symbol with the specified name and true, or if there is
// no matching symbol, it recurses to the outer symbol table. If there is no
// outer symbol table, false is returned.
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if ok || s.outer == nil {
		// Found symbol or we cannot recurse any more.
		return obj, ok
	}
	return s.outer.Resolve(name)
}
