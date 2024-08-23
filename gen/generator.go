package gen

import (
	"errors"
	"fmt"
	"github.com/alecthomas/participle/lexer"
	"github.com/vc2402/vivard/utils"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/dave/jennifer/jen"
)

const (
	EngineVivard = "Vivard"

	InternalPackageName = "vivintrnl"
)

type NullsHandlingKind int
type UnknownAnnotationBehaviour int
type AutoGenerateIDFieldBehaviour bool
type ExtendableTypeDescriptorBehaviour int
type DefaultPackageOption string
type OutputDirectoryOption string
type ClientOutputDirOption string
type PackagePrefixOption string

const (
	// NullablePointers - all nullable fields will be pointers
	NullablePointers NullsHandlingKind = iota
	// NullableField - create special field for nulls handling
	NullableField
	// NullableStorableField - like NullableField but when storing in DB just store it (not convert null values to nulls in DB)
	NullableStorableField
	// NullableNothing - do nothing special for null handling (fields are not pointers, nulls - empty values) - default value
	NullableNothing
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
	GenerateStringFieldForExtendableTypes ExtendableTypeDescriptorBehaviour = iota
	GenerateIntFieldForExtendableTypes
	DoNotGenerateFieldForExtendableTypes
)
const (
	AutoGenerateIDField      AutoGenerateIDFieldBehaviour = true
	DoNotAutoGenerateIDField AutoGenerateIDFieldBehaviour = false
)

type GenerationStage int

const (
	StageParsing GenerationStage = iota
	StageBeforePrepare
	StagePrepare
	StageMetaProcessing
	StageGenerating
)

type Opts struct {
	NullsHandling       NullsHandlingKind
	UnknownAnnotation   UnknownAnnotationBehaviour
	AutoGenerateIDField AutoGenerateIDFieldBehaviour
	DefaultPackage      string
	OutputDir           string
	ClientOutputDir     string
	PackagePrefix       string
	ExtendableTypeDescr ExtendableTypeDescriptorBehaviour

	Custom map[string]interface{}
}

type DefinedType struct {
	pos         lexer.Position
	name        string
	pckg        string
	packagePath string
	external    bool
	entry       *Entity
	enum        *Enum
}

type Project struct {
	packages          map[string]*Package
	extPackages       map[string]string
	extTypes          map[string]*DefinedType
	metaProcs         []MetaProcessor
	generators        []Generator
	featureProviders  []FeatureProvider
	fragmentProviders []CodeFragmentProvider
	Options           *Opts
	Files             []*File
	Warnings          []string
	Errors            []error
	hooks             []GeneratorHookHolder
	stage             GenerationStage
}

type EngineDescriptor struct {
	Fields         *jen.Statement
	Initializator  *jen.Statement
	SingletonInits map[string]*jen.Statement
	Initialized    *jen.Statement
	Start          *jen.Statement
	Functions      *jen.Statement
	file           *jen.File
	startAdd       *jen.Statement
	prepAdd        *jen.Statement
}

var plugins = map[string]Generator{}

func RegisterPlugin(plugin Generator) {
	if _, ok := plugins[plugin.Name()]; ok {
		panic(fmt.Sprintf("duplicate plugin name: %s", plugin.Name()))
	}
	plugins[plugin.Name()] = plugin
}

// New creates new Project object
func New(files []*File, o *Opts) *Project {
	cg := &CodeGenerator{}
	if o.DefaultPackage == "" {
		o.DefaultPackage = "generated"
	}
	return &Project{
		Files:             files,
		packages:          map[string]*Package{},
		extPackages:       map[string]string{},
		extTypes:          map[string]*DefinedType{},
		Options:           o,
		generators:        []Generator{cg},
		featureProviders:  []FeatureProvider{cg},
		metaProcs:         []MetaProcessor{cg},
		fragmentProviders: []CodeFragmentProvider{cg},
	}

}

// Options creates new Opts object and initializes it with given values
// first three values, if strings, are OutputDir, DefaultPackage and ClientOutputDir (may be omitted)
//
//	PackagePrefix can be set with using corresponding type (DefaultPackageOption)
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
					if len(opts) > 2 {
						if dp, ok := opts[idx].(string); ok {
							o.ClientOutputDir = dp
							idx++
						}
					}

				}
			}
		}
	}
	return o.With(opts[idx:]...)
}

// SetOutputDir sets output dir for generator
func (o *Opts) SetOutputDir(od string) *Opts {
	o.OutputDir = od
	return o
}

// SetDefaultPackage sets default package for generator
func (o *Opts) SetDefaultPackage(dp string) *Opts {
	o.DefaultPackage = dp
	return o
}

// SetClientOutputDir sets default package for generator
func (o *Opts) SetClientOutputDir(cd string) *Opts {
	o.ClientOutputDir = cd
	return o
}

func (o *Opts) FromAny(options any) error {
	if opts, ok := options.(map[string]interface{}); ok {
		for name, opt := range opts {
			var val any
		nameCase:
			switch name {
			case "NullableHandling", "nullable_handling", "nullable-handling":
				switch opt {
				case "none":
					val = NullableNothing
				case "pointer":
					val = NullablePointers
				case "field":
					val = NullableField
				case "storable-field":
					val = NullableStorableField
				default:
					break nameCase
				}
			case "UnknownAnnotation", "unknown_annotation", "unknown-annotation":
				switch opt {
				case "ignore":
					val = UnknownAnnotationIgnore
				case "warn", "warning":
					val = UnknownAnnotationWarning
				case "err", "error":
					val = UnknownAnnotationError
				default:
					break nameCase
				}
			//AutoGenerateIDField: not implemented
			case "DefaultPackage", "default_package", "default-package":
				val = DefaultPackageOption(opt.(string))
			case "OutputDir", "OutputDirectory", "output_dir", "output_directory", "output-dir", "output-directory":
				val = OutputDirectoryOption(opt.(string))
			case "ClientOutputDir", "ClientOutputDirectory", "client_output_dir", "client_output_directory", "ClientOut", "client-out":
				val = ClientOutputDirOption(opt.(string))
			case "PackagePrefix", "package_prefix", "package-prefix":
				val = PackagePrefixOption(opt.(string))
			case "ExtendableTypeField", "extendable_type_field", "extendable-type-field":
				switch opt {
				case "string":
					val = GenerateStringFieldForExtendableTypes
				case "int":
					val = GenerateIntFieldForExtendableTypes
				case "none", "ignore", "skip":
					val = DoNotGenerateFieldForExtendableTypes
				default:
					break nameCase
				}
			case "CodeGenerator", "code-generator", "code_generator":
				setCodeGeneratorOptions := func(opts map[string]any) {
					if o.Custom == nil {
						o.Custom = map[string]any{CodeGeneratorOptionsName: opts}
					} else {
						o.Custom[CodeGeneratorOptionsName] = opts
					}
				}
				if opts, ok := opt.(map[string]any); ok {
					setCodeGeneratorOptions(opts)
				} else if opts, ok := opt.([]map[string]any); ok {
					options := map[string]any{}
					for _, opt := range opts {
						for k, v := range opt {
							options[k] = v
						}
					}
					setCodeGeneratorOptions(options)
				} else {
					break nameCase
				}
				continue
			}
			if val != nil {
				o.With(val)
			} else {
				return fmt.Errorf("undefined value for option %s: %v", name, opt)
			}
		}
	} else if opts, ok := options.([]map[string]any); ok {
		for _, opt := range opts {
			err := o.FromAny(opt)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("invalid type for options %T", options)
	}
	return nil
}

// With add options to object
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
		case DefaultPackageOption:
			o.DefaultPackage = string(opt)
		case OutputDirectoryOption:
			o.OutputDir = string(opt)
		case ClientOutputDirOption:
			o.ClientOutputDir = string(opt)
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

// With registers Generator gen
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
	if cfp, ok := gen.(CodeFragmentProvider); ok {
		p.fragmentProviders = append(p.fragmentProviders, cfp)
	}
	return p
}

// WithPlugin adds registered plugin as generator
func (p *Project) WithPlugin(name string, options interface{}) error {
	if gen, ok := plugins[name]; ok {
		p.With(gen)
		if os, ok := gen.(OptionsSetter); ok {
			// from viper options can come as array of maps...
			if opts, ok := options.([]map[string]any); ok {
				for _, opt := range opts {
					err := os.SetOptions(opt)
					if err != nil {
						return err
					}
				}
			} else {
				return os.SetOptions(options)
			}
		}
		if options != nil {
			return fmt.Errorf("plugin %s is not accepting options", name)
		}
		return nil
	}
	return fmt.Errorf("plugin '%s' not found", name)
}

// WithPluginMust adds registered plugin as generator; panics in case of error
func (p *Project) WithPluginMust(name string, options interface{}) *Project {
	err := p.WithPlugin(name, options)
	if err != nil {
		panic(err.Error())
	}
	return p
}

// WithMetaProcessor registers meta processor (registered Generator will be added automatically if it implements MetaProcessor interface)
func (p *Project) WithMetaProcessor(mp MetaProcessor) *Project {
	p.metaProcs = append(p.metaProcs, mp)
	return p
}

// WithHookHolder registers HookHolder (registered Generator will be added automatically if it implements HookHolder interface)
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

// GetFeature looks for feature in obj (*Package (for *Package and *Builder), *Entity, *Field or *Method); returns nil if feature not found
func (p *Project) GetFeature(obj interface{}, kind FeatureKind, name string) interface{} {
	var f Features
	switch v := obj.(type) {
	case *Entity:
		f = v.Features
	case *Field:
		f = v.Features
	case *Method:
		f = v.Features
	case *Package:
		f = v.Features
	case *Builder:
		f = v.Descriptor.Features
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

// GetFeatureMust looks for feature in obj (*Entity, *Field or *Method); panics if feature not found
func (p *Project) GetFeatureMust(obj interface{}, kind FeatureKind, name string) interface{} {
	if f := p.GetFeature(obj, kind, name); f != nil {
		return f
	}
	var objName string
	var pos lexer.Position
	switch v := obj.(type) {
	case *Entity:
		objName = v.Name
		pos = v.Pos
	case *Field:
		objName = v.Name
		pos = v.Pos
	case *Method:
		objName = v.Name
		pos = v.Pos
	case *Package:
		objName = v.Name
		if len(v.Files) > 0 {
			pos = v.Files[0].Pos
		}
	case *Builder:
		objName = v.Name
		pos = v.Pos
	}
	panic(fmt.Sprintf("at %v: %s: no feature provider found for feature %s:%s (%T)", pos, objName, kind, name, obj))
}

// CallCodeFeatureFunc looks for feature with given params, tries to assert it to CodeHelperFunc and call; panics if feature not found
func (p *Project) CallCodeFeatureFunc(obj interface{}, kind FeatureKind, name string, args ...interface{}) jen.Code {
	f := p.GetFeatureMust(obj, kind, name)
	if chf, ok := f.(CodeHelperFunc); ok {
		return chf(args...)
	}
	panic(fmt.Sprintf("feature %s:%s is not found or not a feature function: %T", kind, name, f))
}

// CallFeatureFunc looks for feature with given params, tries to assert it to FeatureFunc and call and returns it's result; panics if feature not found
func (p *Project) CallFeatureFunc(obj interface{}, kind FeatureKind, name string, args ...interface{}) (any, error) {
	f := p.GetFeatureMust(obj, kind, name)
	if ff, ok := f.(FeatureFunc); ok {
		return ff(args...)
	}
	panic(fmt.Sprintf("feature %s:%s is not found or not a feature function: %T", kind, name, f))
}

// CurrentStage returns current stage
func (p *Project) CurrentStage() GenerationStage { return p.stage }

// CallFeatureHookFunc looks for feature with given params, tries to assert it to HookFeatureFunc and call
func (p *Project) CallFeatureHookFunc(
	obj interface{},
	kind FeatureKind,
	name string,
	args HookArgsDescriptor,
) jen.Code {
	if f, ok := p.GetFeature(obj, kind, name).(HookFeatureFunc); ok {
		return f(args)
	}
	panic(fmt.Sprintf("feature %s:%s is not a hook function", kind, name))
}

func (p *Project) Generate() (err error) {
	p.Options.OutputDir = filepath.FromSlash(p.Options.OutputDir)
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

	p.stage = StageBeforePrepare
	for _, pckg := range p.packages {
		err = pckg.beforePrepare()
		if err != nil {
			return
		}
	}

	p.stage = StagePrepare
	for _, pckg := range p.packages {
		err = pckg.prepare()
		if err != nil {
			return
		}
	}

	p.stage = StageMetaProcessing
	for _, pckg := range p.packages {
		err := pckg.processMetas()
		if err != nil {
			return err
		}
	}

	p.stage = StageGenerating
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
				Project:    p,
			}

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
			utils.WalkMap(
				bldr.consts,
				func(stmts []*jen.Statement, _ string) error {
					if len(stmts) == 1 {
						bldr.JenFile.Add(jen.Const().Add(stmts[0]))
					} else {
						multiLineConst := jen.Options{
							Close: ")",
							Multi: true,
							Open:  "(",
						}
						bldr.JenFile.Add(
							jen.Const().CustomFunc(
								multiLineConst,
								func(g *jen.Group) {
									for _, stmt := range stmts {
										g.Add(stmt)
									}
								},
							),
						)
					}
					return nil
				},
			)

			utils.WalkMap(
				bldr.vars,
				func(stmts []*jen.Statement, _ string) error {
					if len(stmts) == 1 {
						bldr.JenFile.Add(jen.Var().Add(stmts[0]))
					} else {
						multiLineConst := jen.Options{
							Close: ")",
							Multi: true,
							Open:  "(",
						}
						bldr.JenFile.Add(
							jen.Var().CustomFunc(
								multiLineConst,
								func(g *jen.Group) {
									for _, stmt := range stmts {
										g.Add(stmt)
									}
								},
							),
						)
					}
					return nil
				},
			)

			bldr.JenFile.Add(gen)
			bldr.JenFile.Add(bldr.Functions)
			pckg.builders = append(pckg.builders, bldr)
			pckg.Engine.Initializator.Add(jen.Id(EngineVar).Dot(fname).Params()).Line()
		}
		pckg.generateEngine()
	}
	//for _, pckg := range p.packages {
	//	err = pckg.processMetas()
	//	if err != nil {
	//		return
	//	}
	//}

	return
}

func (p *Project) GetInternalPackage() *Package {
	pckg := p.GetPackage(InternalPackageName)
	if pckg.Engine == nil {
		pckg.initEngine()
	}
	return pckg
}

func (p *Project) ProvideCodeFragment(
	module interface{},
	action interface{},
	point interface{},
	ctx interface{},
	theOnly bool,
) interface{} {
	var atLeastOne interface{}
	for _, cfp := range p.fragmentProviders {
		ret := cfp.ProvideCodeFragment(module, action, point, ctx)
		if theOnly && ret != nil {
			return ret
		}
		if atLeastOne == nil {
			atLeastOne = ret
		}
	}
	return atLeastOne
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
			err = os.MkdirAll(filepath.Join(p.Options.OutputDir, desc.Name), os.ModeDir|os.ModePerm)
			if err != nil {
				return
			}
			fname := filepath.Join(p.Options.OutputDir, desc.Name, bldr.File.Name+".go")
			err = bldr.JenFile.Save(fname)
			if err != nil {
				return
			}
		}
		if !desc.engineless {
			err := desc.Engine.file.Save(filepath.Join(p.Options.OutputDir, desc.Name, "engine.go"))
			if err != nil {
				return err
			}
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

func (p *Project) addExternal(e *Entity, packag *Package) error {
	ann := e.Annotations.Find(AnnotationGo, AnnGoPackage)
	//if ann == nil || len(ann.Values) == 0 {
	//	return fmt.Errorf("at %v: no package for %s type %s", e.Pos, TypeModifierExternal, e.Name)
	//}
	pckg := ""
	pckgAlias := ""
	name := e.Name
	if idx := strings.LastIndex(name, "."); idx != -1 {
		pckgAlias = name[:idx]
		e.Name = name[idx+1:]
	}
	if ann != nil && len(ann.Values) > 0 {
		if ann.Values[0].Value != nil && ann.Values[0].Value.String != nil {
			pckgAlias = ann.Values[0].Key
			pckg = *ann.Values[0].Value.String
		} else {
			pckg = ann.Values[0].Key
			pckgAlias = pckg
			if idx := strings.LastIndex(pckg, "/"); idx != -1 {
				pckgAlias = pckg[idx+1:]
			}
			if idx := strings.LastIndex(pckgAlias, "-"); idx != -1 {
				pckgAlias = pckgAlias[idx+1:]
			}
		}
	} else {
		pckgAlias = packag.Name
		pckg = packag.fullPackage
	}
	if pckgAlias != "" {
		p.extPackages[pckgAlias] = pckg
	}
	p.extTypes[fmt.Sprintf("%s.%s", pckgAlias, e.Name)] = &DefinedType{
		name:        e.Name,
		entry:       e,
		external:    true,
		packagePath: pckg,
		pckg:        pckgAlias,
	}
	return nil
}

// RegisterExternalType registers reference to external type
func (p *Project) RegisterExternalType(pckg, alias string, name string) error {
	if alias != "" {
		if p, ok := p.extPackages[alias]; ok && p != pckg {
			return fmt.Errorf("attpmt to reassign alias '%s' from package '%s' to '%s", alias, p, pckg)
		}
		p.extPackages[alias] = pckg
	}
	fullName := fmt.Sprintf("%s.%s", alias, name)
	if _, ok := p.extTypes[fullName]; !ok {
		p.extTypes[fullName] = &DefinedType{
			name:        name,
			external:    true,
			packagePath: pckg,
			pckg:        alias,
			entry:       &Entity{Name: name, TypeModifers: map[TypeModifier]bool{TypeModifierExternal: true}},
		}
	}
	return nil
}

func (o *Opts) CustomToStruct(name string, to interface{}) (found bool, err error) {
	if opts, ok := o.Custom[name]; ok {
		return true, OptionsAnyToStruct(opts, to)
	}
	return false, nil
}

func OptionsAnyToStruct(opts any, to any) (err error) {
	tov := reflect.ValueOf(to)
	if tov.Kind() != reflect.Ptr {
		return errors.New("only pointer can be used with CustomToStruct")
	}
	vv := reflect.Indirect(reflect.ValueOf(opts))
	tov = reflect.Indirect(tov)
	if vv.Type() == tov.Type() {
		tov.Set(vv)
		return nil
	} else {
		return AnyToReflect(opts, tov)
	}
}

func AnyToReflect(opt any, to reflect.Value) error {
	if to.Type().Kind() == reflect.Pointer {
		return AnyToReflect(opt, reflect.Indirect(to))
	}
	switch val := opt.(type) {
	case map[string]interface{}:
		switch to.Type().Kind() {
		case reflect.Struct:
			for k, v := range val {
				if idx, ok := FindFieldInStruct(to.Type(), k); ok {
					err := AnyToReflect(v, to.Field(idx))
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("cannot unmarshal '%s' from %v to %s", k, opt, to.Type().Name())
				}
			}
		case reflect.Map:
			if to.IsNil() {
				to.Set(reflect.MakeMap(to.Type()))
			}
			for k, v := range val {
				mapVal := reflect.New(to.Type().Elem())
				err := AnyToReflect(v, mapVal)
				if err != nil {
					return err
				}
				to.SetMapIndex(reflect.ValueOf(k), reflect.Indirect(mapVal))
			}
		default:
			if to.CanSet() {
				rv := reflect.ValueOf(val)
				if rv.Type().AssignableTo(to.Type()) {
					to.Set(reflect.ValueOf(val))
					break
				}
			}
			return fmt.Errorf("cannot set value from %v to %s", val, to.Type().Name())
		}
	case []map[string]interface{}:
		switch to.Type().Kind() {
		case reflect.Slice:
			for _, opt := range val {
				elem := reflect.New(to.Type().Elem())
				err := AnyToReflect(opt, elem)
				if err != nil {
					return err
				}
				reflect.Append(to, reflect.Indirect(elem))
			}
		case reflect.Struct, reflect.Map:
			for _, opt := range val {
				err := AnyToReflect(opt, to)
				if err != nil {
					return err
				}
			}
		}
	default:
		if to.CanSet() {
			rv := reflect.ValueOf(val)
			if rv.Type().AssignableTo(to.Type()) {
				to.Set(reflect.ValueOf(val))
				break
			}
		}
		return fmt.Errorf("cannot set value from %v to %s", val, to.Type().Name())
	}
	return nil
}

func FindFieldInStruct(struc reflect.Type, name string) (int, bool) {
	for i := 0; i < struc.NumField(); i++ {
		fld := struc.Field(i)
		if fld.Name == name || strings.ToLower(fld.Name) == name {
			return i, true
		}
	}
	return -1, false
}

func (desc *Package) FindTypes(cb func(*Entity) bool) (ret []*Entity) {
	for _, t := range desc.types {
		if t.entry != nil && cb(t.entry) {
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
MetaLoop:
	for m.Next() {
		ok, err := desc.ParseAnnotationMeta(m)
		if err != nil {
			return err
		}
		if ok {
			continue
		}
		for _, mp := range desc.Project.metaProcs {
			ok, e := mp.ProcessMeta(desc, m)
			if e != nil {
				return e
			}
			if ok {
				continue MetaLoop
			}
		}
		return fmt.Errorf("at %s: %d: meta block can not be understood", m.Pos.Filename, m.Pos.Line+m.start)
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

var metaAnnParser = participle.MustBuild(
	&ampTags{},
	// participle.Lexer(regexLex),
	participle.Lexer(lex),
	participle.Elide("Comment", "Whitespace"),
	participle.Unquote("String"),
	participle.Map(annotationMapper, "AnnotationTag", "HookTag", "MetaLine"),
	participle.UseLookahead(1),
)

// ParseAnnotationMeta tries to parse meta as annotations set
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
