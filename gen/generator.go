package gen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/dave/jennifer/jen"
)

const (
	EngineVivard = "Vivard"
)

type NullsHandlingKind int
type UnknownAnnotationBehaviour int
type AutoGenerateIDFieldBehaviour bool
type ExtendableTypeDescriptorBehaviour int
type PackagePrefixOption string

const (
	//NullableNothing - do nothing special fjr null handling (fields are not pointers, nulls - empty values) - default value
	NullableNothing NullsHandlingKind = iota
	// NullablePointers - all nullable fields will be pointers
	NullablePointers
	// NullableField - create special field for nulls handling
	NullableField
	// NullableStorableField - like NullableField but when storing in DB just store it (not convert null values to nulls in DB)
	NullableStorableField
)
const (
	//UnknownAnnotationError - stop generation if unknown annotation is met
	UnknownAnnotationError UnknownAnnotationBehaviour = iota
	//UnknownAnnotationWarning - inform about unknown annotation and continue
	UnknownAnnotationWarning
	//UnknownAnnotationIgnore - ignore unknown annotation
	UnknownAnnotationIgnore
)

const (
	GenerateStringFieldForExtandableTypes ExtendableTypeDescriptorBehaviour = iota
	GenerateIntFieldForExtandableTypes
	DoNotGenerateFieldForExtandableTypes
)
const (
	AutoGenerateIDField      AutoGenerateIDFieldBehaviour = true
	DoNotAutoGenerateIDField AutoGenerateIDFieldBehaviour = false
)

type Opts struct {
	NullsHandling       NullsHandlingKind
	UnknownAnnotation   UnknownAnnotationBehaviour
	AutoGenerateIDField AutoGenerateIDFieldBehaviour
	DefaultPackage      string
	OutputDir           string
	PackagePrefix       string
	ExtendableTypeDescr ExtendableTypeDescriptorBehaviour

	Custom map[string]interface{}
}

type DefinedType struct {
	name        string
	pckg        string
	packagePath string
	external    bool
	entry       *Entity
}

type Project struct {
	packages         map[string]*Package
	extPackages      map[string]string
	extTypes         map[string]*DefinedType
	metaProcs        []MetaProcessor
	generators       []Generator
	featureProviders []FeatureProvider
	Options          *Opts
	Files            []*File
	Warnings         []string
	Errors           []error
	hooks            []GeneratorHookHolder
}

type EngineDescriptor struct {
	Fields        *jen.Statement
	Initializator *jen.Statement
	Initialized   *jen.Statement
	Start         *jen.Statement
	Functions     *jen.Statement
	file          *jen.File
}

//New creates new Project object
func New(files []*File, o *Opts) *Project {
	cg := &CodeGenerator{}
	if o.DefaultPackage == "" {
		o.DefaultPackage = "generated"
	}
	return &Project{
		Files:            files,
		packages:         map[string]*Package{},
		extPackages:      map[string]string{},
		extTypes:         map[string]*DefinedType{},
		Options:          o,
		generators:       []Generator{cg},
		featureProviders: []FeatureProvider{cg},
		metaProcs:        []MetaProcessor{cg},
	}

}

//Options creates new Opts object and initializes it with given values
// first two values, if strings, are OutputDir and DefaultPackage (may be omitted)
//  PackagePrefix can be set with using corresponding type (DefaultPackageOption)
func Options(opts ...interface{}) *Opts {
	o := &Opts{Custom: map[string]interface{}{}}
	idx := 0
	if len(opts) > 0 {
		if od, ok := opts[idx].(string); ok {
			o.OutputDir = od
			idx++
			if len(opts) > 1 {
				if dp, ok := opts[idx].(string); ok {
					o.DefaultPackage = dp
					idx++
				}
			}
		}
	}
	return o.With(opts[idx:]...)
}

//With add options to object
func (o *Opts) With(opts ...interface{}) *Opts {
	for _, op := range opts {
		switch opt := op.(type) {
		case NullsHandlingKind:
			o.NullsHandling = opt
		case UnknownAnnotationBehaviour:
			o.UnknownAnnotation = opt
		case AutoGenerateIDFieldBehaviour:
			o.AutoGenerateIDField = opt
		case PackagePrefixOption:
			o.PackagePrefix = string(opt)
		default:
			panic(fmt.Sprintf("undefined option: %#v (%T)", op, op))
		}
	}
	return o
}
func (o *Opts) SetAutoGenerateIDField(set AutoGenerateIDFieldBehaviour) *Opts {
	o.AutoGenerateIDField = set
	return o
}

func (o *Opts) WithCustom(name string, op interface{}) *Opts {
	o.Custom[name] = op
	return o
}

//With registers Generator gen
func (p *Project) With(gen Generator) *Project {
	p.generators = append(p.generators, gen)
	if fp, ok := gen.(FeatureProvider); ok {
		p.featureProviders = append(p.featureProviders, fp)
	}
	if mp, ok := gen.(MetaProcessor); ok {
		p.WithMetaProcessor(mp)
	}
	if hh, ok := gen.(GeneratorHookHolder); ok {
		p.WithHookHolder(hh)
	}
	return p
}

//WithMetaProcessor registers meta processor (registered Generator will be added automatically if it implements MetaProcessor interface)
func (p *Project) WithMetaProcessor(mp MetaProcessor) *Project {
	p.metaProcs = append(p.metaProcs, mp)
	return p
}

//WithHookHolder registers HookHolder (registered Generator will be added automatically if it implements HookHolder interface)
func (p *Project) WithHookHolder(hh GeneratorHookHolder) *Project {
	p.hooks = append(p.hooks, hh)
	return p
}

func (p *Project) AddWarning(warn string) {
	p.Warnings = append(p.Warnings, warn)
}
func (p *Project) AddError(err error) {
	p.Errors = append(p.Errors, err)
}
func (p *Project) HasErrors() bool {
	return len(p.Errors) > 0
}

//GetFeature looks for feature in obj (*Entity, *Field or *Method); returns nil if feature not found
func (p *Project) GetFeature(obj interface{}, kind FeatureKind, name string) interface{} {
	var f Features
	switch v := obj.(type) {
	case *Entity:
		f = v.Features
	case *Field:
		f = v.Features
	case *Method:
		f = v.Features
	default:
		panic(fmt.Sprintf("GetFeature was called for unknown type: %T", obj))
	}
	if feat, ok := f.Get(kind, name); ok {
		return feat
	}
	for _, fp := range p.featureProviders {
		if feat, ok := fp.ProvideFeature(kind, name, obj); ok != FeatureNotProvided {
			if ok == FeatureProvided {
				f[f.Name(kind, name)] = feat
			}
			return feat
		}
	}
	return nil
}

//GetFeatureMust looks for feature in obj (*Entity, *Field or *Method); panics if feature not found
func (p *Project) GetFeatureMust(obj interface{}, kind FeatureKind, name string) interface{} {
	if f := p.GetFeature(obj, kind, name); f != nil {
		return f
	}
	panic(fmt.Sprintf("no feature provider found for feature %s:%s (%T)", kind, name, obj))
}

//CallFeatureFunc looks for feature with given params, tries to assert it to CodeHelperFunc and call; panics if feature not found
func (p *Project) CallFeatureFunc(obj interface{}, kind FeatureKind, name string, args ...interface{}) jen.Code {
	if f, ok := p.GetFeatureMust(obj, kind, name).(CodeHelperFunc); ok {
		return f(args...)
	}
	panic(fmt.Sprintf("feature %s:%s is not a feature function", kind, name))
}

//CallFeatureHookFunc looks for feature with given params, tries to assert it to HookFeatureFunc and call
func (p *Project) CallFeatureHookFunc(obj interface{}, kind FeatureKind, name string, args HookArgsDescriptor) jen.Code {
	if f, ok := p.GetFeature(obj, kind, name).(HookFeatureFunc); ok {
		return f(args)
	}
	panic(fmt.Sprintf("feature %s:%s is not a hook function", kind, name))
}

func (p *Project) Generate() (err error) {
	// desc.generators = []Generator{&CodeGenerator{}, &GQLGenerator{}}
	p.start()

	for _, file := range p.Files {
		pname := file.Package
		if pname == "" {
			pname = p.Options.DefaultPackage
		}
		desc := p.GetPackage(pname)
		desc.Files = append(desc.Files, file)
	}
	for _, pckg := range p.packages {
		err = pckg.postParsed()
		if err != nil {
			return
		}
	}
	for _, pckg := range p.packages {
		err = pckg.prepare()
		if err != nil {
			return
		}
	}
	for _, pckg := range p.packages {
		for _, file := range pckg.Files {
			bldr := &Builder{
				File:       file,
				JenFile:    jen.NewFile(pckg.Name),
				Descriptor: pckg,
				Types:      &jen.Statement{},
				vars:       map[string][]*jen.Statement{},
				consts:     map[string][]*jen.Statement{},
				Functions:  &jen.Statement{},
				Project:    p}

			bldr.Generator = &jen.Statement{}
			bldr.JenFile.HeaderComment(fmt.Sprintf("Code generated from file %s by vivgen. DO NOT EDIT.", file.FileName))
			err = pckg.doGenerate(bldr)
			if err != nil {
				return
			}
			if len(p.Errors) > 0 {
				return p.Errors[0]
			}

			fname := fmt.Sprintf("%sFile", file.Name)
			fname = strings.ToUpper(fname[:1]) + fname[1:]

			gen := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Block(
				bldr.Generator,
			).Line()
			bldr.JenFile.Add(bldr.Types)
			for _, stmts := range bldr.consts {
				if len(stmts) == 1 {
					bldr.JenFile.Add(jen.Const().Add(stmts[0]))
				} else {
					multiLineConst := jen.Options{
						Close: ")",
						Multi: true,
						Open:  "(",
					}
					bldr.JenFile.Add(jen.Const().CustomFunc(
						multiLineConst,
						func(g *jen.Group) {
							for _, stmt := range stmts {
								g.Add(stmt)
							}
						},
					))
				}
			}

			for _, stmts := range bldr.vars {
				if len(stmts) == 1 {
					bldr.JenFile.Add(jen.Var().Add(stmts[0]))
				} else {
					multiLineConst := jen.Options{
						Close: ")",
						Multi: true,
						Open:  "(",
					}
					bldr.JenFile.Add(jen.Var().CustomFunc(
						multiLineConst,
						func(g *jen.Group) {
							for _, stmt := range stmts {
								g.Add(stmt)
							}
						},
					))
				}
			}

			bldr.JenFile.Add(gen)
			bldr.JenFile.Add(bldr.Functions)
			pckg.builders = append(pckg.builders, bldr)
			pckg.Engine.Initializator.Add(jen.Id(EngineVar).Dot(fname).Params()).Line()
		}
	}

	return
}

func (p *Project) Print() {
	for _, desc := range p.packages {
		for _, bldr := range desc.builders {
			fmt.Printf("%s/%s.go\n%#v", bldr.File.Package, bldr.File.Name, bldr.JenFile)
		}
		fmt.Printf("\nengine.go\n%#v", desc.Engine.file)
	}
}

func (p *Project) WriteToFiles() (err error) {
	for _, desc := range p.packages {
		for _, bldr := range desc.builders {
			err = os.MkdirAll(path.Join(p.Options.OutputDir, desc.Name), os.ModeDir|os.ModePerm)
			if err != nil {
				return
			}
			fname := path.Join(p.Options.OutputDir, desc.Name, bldr.File.Name+".go")
			err = bldr.JenFile.Save(fname)
			if err != nil {
				return
			}
		}
		err := desc.Engine.file.Save(path.Join(p.Options.OutputDir, desc.Name, "engine.go"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (desc *Package) Options() *Opts {
	return desc.Project.Options
}

func (p *Project) GetPackage(name string) *Package {
	ret, ok := p.packages[name]
	if !ok {
		fp := name
		if p.Options.PackagePrefix != "" {
			fp = fmt.Sprintf("%s/%s", p.Options.PackagePrefix, name)
		}
		ret = &Package{
			Name:        name,
			types:       map[string]*DefinedType{},
			Project:     p,
			Features:    Features{},
			extEngines:  map[string]string{},
			fullPackage: fp,
		}
		p.packages[name] = ret
	}
	return ret
}

func (p *Project) FindType(name string) (t *DefinedType, ok bool) {
	t, ok = p.extTypes[name]
	if !ok {
		parts := strings.SplitN(name, ".", 2)
		if len(parts) == 1 {
			return
		}
		for _, pckg := range p.packages {
			if pckg.Name == parts[0] {
				return pckg.FindType(parts[1])
			}
		}
	}
	return
}

func (p *Project) GetTypePackage(t *DefinedType) string {
	return t.packagePath
}

func (p *Project) GetFullPackage(alias string) string {
	if pc, ok := p.packages[alias]; ok {
		return pc.fullPackage
	}
	if ep, ok := p.extPackages[alias]; ok {
		return ep
	}
	p.AddWarning(fmt.Sprintf("package alias '%s' not found", alias))
	return alias
}

func (p *Project) addExternal(e *Entity) error {
	ann := e.Annotations.Find(AnnotationGo, AnnGoPackage)
	if ann == nil || len(ann.Values) == 0 {
		return fmt.Errorf("at %v: no package for %s type %s", e.Pos, TypeModifierExternal, e.Name)
	}
	pckg := ""
	pckgAlias := ""
	name := e.Name
	if idx := strings.LastIndex(name, "."); idx != -1 {
		pckgAlias = name[:idx]
		e.Name = name[idx+1:]
	}
	if ann.Values[0].Value != nil && ann.Values[0].Value.String != nil {
		pckgAlias = ann.Values[0].Key
		pckg = *ann.Values[0].Value.String
	} else {
		pckg = ann.Values[0].Key
		if idx := strings.LastIndex(pckg, "/"); idx != -1 {
			pckgAlias = pckg[idx+1:]
		}
	}
	if pckgAlias != "" {
		p.extPackages[pckgAlias] = pckg
	}
	p.extTypes[fmt.Sprintf("%s.%s", pckgAlias, e.Name)] = &DefinedType{name: e.Name, entry: e, external: true, packagePath: pckg, pckg: pckgAlias}
	return nil
}

func (o *Opts) CustomToStruct(name string, to interface{}) (found bool, err error) {
	tov := reflect.ValueOf(to)
	if tov.Kind() != reflect.Ptr {
		return false, errors.New("only pointer can be used with CustomToStruct")
	}

	if o, ok := o.Custom[name]; ok {
		switch v := o.(type) {
		case map[string]interface{}:
			b, e := json.Marshal(v)
			if e != nil {
				return true, e
			}
			e = json.Unmarshal(b, to)
			if e != nil {
				return true, nil
			}
		default:
			vv := reflect.Indirect(reflect.ValueOf(o))
			tov = reflect.Indirect(tov)
			if vv.Type() == tov.Type() {
				tov.Set(vv)
				return true, nil
			} else {
				return true, fmt.Errorf("can not set %T from %T", to, o)
			}
		}
	}
	return false, nil
}

func (desc *Package) FindTypes(cb func(*Entity) bool) (ret []*Entity) {
	for _, t := range desc.types {
		if cb(t.entry) {
			ret = append(ret, t.entry)
		}
	}
	return
}

func (desc *Package) processMetas() error {
	for _, f := range desc.Files {
		for _, m := range f.Meta {
			err := desc.processMeta(m)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (desc *Package) processMeta(m *Meta) error {
	if m.TypeName != "" {
		if tr, ok := desc.FindType(m.TypeName); ok {
			m.TypeRef = tr.Entity()
		} else {
			desc.AddWarning(fmt.Sprintf("at %v: type %s not find", m.Pos, m.TypeName))
		}
	}
META_LOOP:
	for m.Next() {
		ok, err := desc.ParseAnnotationMeta(m)
		if ok {
			continue
		}
		for _, mp := range desc.Project.metaProcs {
			ok, e := mp.ProcessMeta(m)
			if ok {
				continue META_LOOP
			}
			if e != nil {
				if err != nil {
					desc.AddWarning(err.Error())
				}
				err = e
			}
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("at %s: %d: meta slice not understandable", m.Pos.Filename, m.Pos.Line+m.start)
	}
	return nil
}

type annotationMetaParser struct {
	meta       *Meta
	line       int
	pos        int
	annotation string
	field      string
	specifier  string
	tags       []*AnnotationTag
	pckg       *Package
}

type ampTags struct {
	Tags []*AnnotationTag ` (@@)* `
}

func (amp annotationMetaParser) current() string {
	return strings.Trim(amp.meta.Current()[amp.line], " \t")
}
func (amp *annotationMetaParser) next() bool {
	amp.line++
	return amp.line < len(amp.meta.Current())
}
func (amp annotationMetaParser) ready() bool { return amp.annotation != "" }
func (amp *annotationMetaParser) flush() error {
	if amp.field != "" {
		fld := amp.meta.TypeRef.GetField(amp.field)
		if fld.parent != amp.meta.TypeRef {
			amp.pckg.AddWarning(fmt.Sprintf("at %v: changing annotation '%s' of BaseClass", amp.meta.Pos, amp.annotation))
		}
		ann := &Annotation{Pos: amp.meta.Position(), Name: amp.annotation}
		ann.Values = amp.tags
		if amp.specifier != "" {
			ann.Name += ":" + amp.specifier
		}
		fld.Annotations.Append(Annotations{ann.Name: ann})
	}
	amp.tags = nil
	return nil
}
func (amp *annotationMetaParser) parseNext() (ok bool, err error) {
	l := amp.current()
	r := regexp.MustCompile(`^(\$|\.|:)([a-zA-Z_][a-zA-Z0-9_-]+)[ \t]*(:([a-zA-Z_][a-zA-Z0-9_-]+))?[ \t]*$`)
	op := r.FindStringSubmatch(l)
	if op == nil {
		if !amp.ready() {
			if amp.line == 0 {
				return false, nil
			} else {
				return false, fmt.Errorf("unexpected '%s' while parsing annotations", l)
			}
		}
		tags := &ampTags{}
		err = metaAnnParser.ParseString(amp.current(), tags)
		if err != nil {
			return
		}
		amp.tags = append(amp.tags, tags.Tags...)
		return true, nil
	}
	switch op[1] {
	case "$":
		// amp.tags = []*AnnotationTag{}
		amp.annotation = op[2]
		if op[4] != "" {
			amp.specifier = op[4]
		}
		amp.flush()
		return true, nil
	case ":":
		amp.specifier = op[2]
		return true, nil
	case ".":
		amp.field = op[2]
		amp.flush()
		return true, nil
	default:

	}
	return
}

var metaAnnParser = participle.MustBuild(&ampTags{},
	// participle.Lexer(regexLex),
	participle.Lexer(lex),
	participle.Elide("Comment", "Whitespace"),
	participle.Unquote("String"),
	participle.Map(annotationMapper, "AnnotationTag", "HookTag", "MetaLine"),
	participle.UseLookahead(1),
)

//ParseAnnotationMeta tries to parse meta as annotations set
func (desc *Package) ParseAnnotationMeta(m *Meta) (ok bool, err error) {
	if m.TypeRef == nil {
		return
	}
	amp := annotationMetaParser{meta: m, line: -1}

	for amp.next() {
		if ok, err = amp.parseNext(); !ok {
			return
		}
	}
	amp.flush()
	ok = true
	return
}

func (p *Project) start() {
	for _, g := range p.generators {
		if da, ok := g.(DescriptorAware); ok {
			da.SetDescriptor(p)
		}
	}
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase - thanks to stower (https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6)
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
