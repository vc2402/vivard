package gen

import "github.com/dave/jennifer/jen"

type Builder struct {
	*Project
	Descriptor *Package
	*File
	JenFile   *jen.File
	Types     *jen.Statement
	vars      map[string][]*jen.Statement
	consts    map[string][]*jen.Statement
	Functions *jen.Statement
	Generator *jen.Statement
}

//AddConst adds statement to const section (const word will be added automatically);
// it may be few named const sections, each will be organized as const() block
func (bldr *Builder) AddConst(section string, stmt *jen.Statement) {
	bldr.consts[section] = append(bldr.consts[section], stmt)
}

//AddGlobal adds statement to var section (var word will be added autimatically);
// it may be few named var sections, each will be organized as var() block
func (bldr *Builder) AddGlobal(section string, stmt *jen.Statement) {
	bldr.vars[section] = append(bldr.vars[section], stmt)
}

func (f *File) AddEntity(e *Entity) {
	f.Entries = append(f.Entries, e)
	f.Pckg.RegisterType(e)
}
