// Package analyzer computes architectural metrics from a parsed Program:
// afferent/efferent coupling, instability, abstraction, main-sequence
// distance, cyclomatic stats, connascence (CoN, CoM), orphan blocks and
// god blocks.
package analyzer

import (
	"sort"
	"strings"

	"github.com/mpraes/archi/internal/model"
)

// Analyze computes metrics on p.Modules in place and returns a Summary.
func Analyze(p *model.Program) model.Summary {
	// 1. Resolve efferent coupling: normalize import paths to internal module names.
	modByName := map[string]*model.Module{}
	for i := range p.Modules {
		modByName[p.Modules[i].Name] = &p.Modules[i]
	}

	// Normalize imports: project internal imports are full import paths; module
	// names use forward slashes and are prefixed by the module path. Drop
	// imports that don't resolve to a known module (external / stdlib).
	for i := range p.Modules {
		m := &p.Modules[i]
		keep := map[string]struct{}{}
		for im := range m.Imports {
			if target, ok := modByName[im]; ok && target != m {
				keep[im] = struct{}{}
			} else {
				// try suffix match: module name == im
				for _, other := range p.Modules {
					if other.Name == m.Name {
						continue
					}
					if other.Name == im {
						keep[other.Name] = struct{}{}
						continue
					}
				}
			}
		}
		m.Imports = keep
	}

	// 2. Aff/eff coupling.
	for i := range p.Modules {
		p.Modules[i].Efferent = len(p.Modules[i].Imports)
	}
	ca := make(map[string]int)
	for i := range p.Modules {
		for dep := range p.Modules[i].Imports {
			ca[dep]++
		}
	}
	for i := range p.Modules {
		p.Modules[i].Afferent = ca[p.Modules[i].Name]
	}

	// 3. I, A, D.
	for i := range p.Modules {
		m := &p.Modules[i]
		total := m.Afferent + m.Efferent
		if total > 0 {
			m.Instability = float64(m.Efferent) / float64(total)
		}
		comp := m.Abstracts + m.Concretes
		if comp > 0 {
			m.Abstraction = float64(m.Abstracts) / float64(comp)
		}
		d := m.Abstraction + m.Instability - 1
		if d < 0 {
			d = -d
		}
		m.Distance = d
	}

	// 4. Orphan blocks: functions/methods never referenced by any other block.
	calledNames := map[string]int{}
	// collect resolved calls per module.
	for i := range p.Files {
		for _, b := range p.Files[i].Blocks {
			for _, c := range b.Calls {
				if strings.HasPrefix(c.Target, "?") {
					continue
				}
				// last segment as the function name.
				name := c.Target
				if dot := strings.LastIndex(name, "."); dot >= 0 {
					name = name[dot+1:]
				}
				calledNames[name]++
			}
		}
	}
	for i := range p.Modules {
		m := &p.Modules[i]
		var orphans []string
		for _, fi := range p.Files {
			if fi.Module != m.Name {
				continue
			}
			for _, b := range fi.Blocks {
				if b.Kind == model.BlockType {
					continue
				}
				if calledNames[b.Name] == 0 {
					orphans = append(orphans, b.Name)
				}
			}
		}
		m.OrphanBlocks = orphans
	}

	// 5. God blocks: complexity above heuristic threshold (top decile or >= 15).
	const godThreshold = 15
	for i := range p.Modules {
		m := &p.Modules[i]
		var gods []string
		for _, fi := range p.Files {
			if fi.Module != m.Name {
				continue
			}
			for _, b := range fi.Blocks {
				if b.Complexity >= godThreshold {
					gods = append(gods, b.Name)
				}
			}
		}
		m.GodBlocks = gods
	}

	// 6. Connascence.
	var conns []model.Connascence
	conns = append(conns, connascenceName(p, modByName)...)
	conns = append(conns, connascenceMeaning(p, modByName)...)

	// Attach per-module connascence and global list.
	for i := range p.Modules {
		var local []model.Connascence
		for _, c := range conns {
			if c.From == p.Modules[i].Name || c.To == p.Modules[i].Name {
				local = append(local, c)
			}
		}
		p.Modules[i].Connascence = local
	}

	// 7. Hotspots: modules in the pain zone (I high & A low) or with god blocks.
	var hotspots []string
	for _, m := range p.Modules {
		if isBoilerplateModule(m) {
			continue
		}
		pain := m.Instability > 0.5 && m.Abstraction < 0.3 && m.Distance > 0.5
		if pain || len(m.GodBlocks) > 0 {
			hotspots = append(hotspots, m.Name)
		}
	}

	return buildSummary(p, hotspots, conns)
}

// connascenceName: cross-module method calls where the receiver type is
// defined in another module (rigid textual signature dependency).
func connascenceName(p *model.Program, byName map[string]*model.Module) []model.Connascence {
	// Build a map of method names to the module that declares them via receiver.
	methodOwner := map[string]string{}
	for _, f := range p.Files {
		for _, b := range f.Blocks {
			if b.Kind == model.BlockMethod && b.Receiver != "" {
				methodOwner[b.Receiver+"."+b.Name] = f.Module
				methodOwner[b.Name] = f.Module
			}
		}
	}
	// Also map type names to their module.
	typeOwner := map[string]string{}
	for _, f := range p.Files {
		for _, b := range f.Blocks {
			if b.Kind == model.BlockType && b.Receiver != "" {
				typeOwner[b.Name] = f.Module
			}
		}
	}

	seen := map[string]struct{}{}
	var out []model.Connascence
	for _, f := range p.Files {
		for _, b := range f.Blocks {
			for _, c := range b.Calls {
				t := c.Target
				// "recv.Method" form.
				if dot := strings.Index(t, "."); dot > 0 {
					recv := t[:dot]
					method := t[dot+1:]
					owner, ok := methodOwner[recv+"."+method]
					if !ok {
						owner, ok = typeOwner[recv]
					}
					if ok && owner != "" && owner != f.Module {
						key := f.Module + "|" + owner + "|" + t
						if _, dup := seen[key]; dup {
							continue
						}
						seen[key] = struct{}{}
						out = append(out, model.Connascence{
							Kind:   "name",
							From:   f.Module,
							To:     owner,
							Detail: "calls " + t,
						})
					}
				}
			}
		}
	}
	return out
}

// connascenceMeaning: identical literal values shared between files of
// different modules (ocult meaning dependency on a magic value).
func connascenceMeaning(p *model.Program, byName map[string]*model.Module) []model.Connascence {
	type occ struct {
		mod  string
		file string
		line int
	}
	// Group literals by value; only consider non-trivial strings/numbers.
	// Exclude short strings (<=1 char) and zero.
	byValue := map[string][]occ{}
	for _, f := range p.Files {
		for _, l := range f.Literals {
			if l.Value == "" || l.Value == "0" {
				continue
			}
			if l.Kind == model.LitString && len(l.Value) < 4 {
				continue
			}
			byValue[l.Value] = append(byValue[l.Value], occ{mod: f.Module, file: f.Path, line: l.Line})
		}
	}
	seen := map[string]struct{}{}
	var out []model.Connascence
	for val, occs := range byValue {
		if len(occs) < 2 {
			continue
		}
		// dedupe across module pairs.
		pairs := map[string]struct{}{}
		for _, a := range occs {
			for _, b := range occs {
				if a.mod != b.mod && a.mod != "" && b.mod != "" {
					key := a.mod + "|" + b.mod
					if _, dup := pairs[key]; dup {
						continue
					}
					pairs[key] = struct{}{}
					dedupKey := key + "|" + val
					if _, dup := seen[dedupKey]; dup {
						continue
					}
					seen[dedupKey] = struct{}{}
					out = append(out, model.Connascence{
						Kind:   "meaning",
						From:   a.mod,
						To:     b.mod,
						Detail: "shared literal " + quote(val),
					})
				}
			}
		}
	}
	return out
}

func quote(s string) string {
	if len(s) > 40 {
		s = s[:37] + "..."
	}
	return "\"" + s + "\""
}

func buildSummary(p *model.Program, hotspots []string, conns []model.Connascence) model.Summary {
	mods := make([]model.ModuleMetrics, 0, len(p.Modules))
	for _, m := range p.Modules {
		orphan := m.OrphanBlocks
		gods := m.GodBlocks
		if orphan == nil {
			orphan = []string{}
		}
		if gods == nil {
			gods = []string{}
		}
		mods = append(mods, model.ModuleMetrics{
			Module:          m.Name,
			Path:            m.Path,
			Language:        m.Language,
			Files:           m.Files,
			Afferent:        m.Afferent,
			Efferent:        m.Efferent,
			Instability:     round(m.Instability),
			Abstraction:     round(m.Abstraction),
			Distance:        round(m.Distance),
			MaxComplexity:   m.MaxCyclo,
			TotalComplexity: m.SumCyclo,
			Abstracts:       m.Abstracts,
			Concretes:       m.Concretes,
			OrphanBlocks:    orphan,
			GodBlocks:       gods,
		})
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].Module < mods[j].Module })
	if hotspots == nil {
		hotspots = []string{}
	}
	if conns == nil {
		conns = []model.Connascence{}
	}
	return model.Summary{
		ProjectName: p.ProjectName,
		ModuleCount: len(mods),
		Modules:     mods,
		Connascence: conns,
		Hotspots:    hotspots,
	}
}

func round(f float64) float64 {
	v := float64(int(f*1000)) / 1000
	return v
}

func isBoilerplateModule(m model.Module) bool {
	needle := strings.ToLower(m.Name + " " + m.Path)
	return strings.Contains(needle, "template") || strings.Contains(needle, "boilerplate")
}
