// Package model defines the architectural domain types shared across parser,
// analyzer, server, and AI layers.
package model

// Program is the complete parsed representation of a scanned project.
type Program struct {
	ProjectName string
	ModulePath  string // go module path for Go projects ("" otherwise)
	Files       []File
	Modules     []Module
}

// File holds parsed structural information about a single source file.
type File struct {
	Path     string
	Module   string
	Lines    int
	Blocks   []Block
	Imports  []string
	Literals []Literal
	Errors   []string
}

// Block represents an atomic anatomical unit (function, method, or type aggregate).
type Block struct {
	Name       string
	Kind       BlockKind
	Receiver   string // receiver type name for methods, "" otherwise
	StartLine  int
	EndLine    int
	Complexity int
	Calls      []Call
}

type BlockKind int

const (
	BlockFunc BlockKind = iota
	BlockMethod
	BlockType
)

func (k BlockKind) String() string {
	switch k {
	case BlockFunc:
		return "func"
	case BlockMethod:
		return "method"
	case BlockType:
		return "type"
	}
	return "?"
}

// Call is an invocation site resolved to a target module if internal.
type Call struct {
	// Textual target as it appears in source (e.g. "pkg.Foo", "x.Bar", "Foo").
	Target string
	// Module the call resolves to, if internal; "" if external/unknown.
	ResolvedModule string
}

type Literal struct {
	Kind  LiteralKind
	Value string
	Line  int
}

type LiteralKind int

const (
	LitString LiteralKind = iota
	LitNumber
)

// Module is a unit of architectural analysis (a package / package directory).
type Module struct {
	Name     string // import-path-style or directory-relative name
	Path     string // filesystem path
	Language string // go, js, ts, py
	Files    int
	// Discovered structure.
	Imports   map[string]struct{} // internal modules this depends on (efferent)
	Abstracts int                 // interfaces / signatures
	Concretes int                 // structs / concrete types
	MaxCyclo  int
	SumCyclo  int
	// Computed metrics (set by analyzer).
	Afferent     int
	Efferent     int
	Instability  float64
	Abstraction  float64
	Distance     float64
	Connascence  []Connascence
	OrphanBlocks []string
	GodBlocks    []string
}

// Connascence between two modules.
type Connascence struct {
	Kind   string `json:"kind"` // "name" (CoN) or "meaning" (CoM)
	From   string `json:"from"` // module owning the dependent block
	To     string `json:"to"`   // module owning the dependency target
	Detail string `json:"detail"`
}

// Summary is the top-level payload served to the frontend and AI. It omits
// source code entirely (RNF-009).
type Summary struct {
	ProjectName string          `json:"projectName"`
	ModuleCount int             `json:"moduleCount"`
	Modules     []ModuleMetrics `json:"modules"`
	Connascence []Connascence   `json:"connascence"`
	Hotspots    []string        `json:"hotspots"`
}

// ModuleMetrics is the per-module JSON view served to the frontend and AI.
type ModuleMetrics struct {
	Module          string   `json:"module"`
	Path            string   `json:"path"`
	Language        string   `json:"language"`
	Files           int      `json:"files"`
	Afferent        int      `json:"afferent"`
	Efferent        int      `json:"efferent"`
	Instability     float64  `json:"instability"`
	Abstraction     float64  `json:"abstraction"`
	Distance        float64  `json:"distance"`
	MaxComplexity   int      `json:"maxComplexity"`
	TotalComplexity int      `json:"totalComplexity"`
	Abstracts       int      `json:"abstracts"`
	Concretes       int      `json:"concretes"`
	OrphanBlocks    []string `json:"orphanBlocks"`
	GodBlocks       []string `json:"godBlocks"`
}
