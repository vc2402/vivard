package gen

import (
	"errors"
	"fmt"
	"github.com/alecthomas/participle/lexer"
	"github.com/vc2402/vivard/utils"
	"strings"

	"github.com/dave/jennifer/jen"
)

// points for CodeFragmentProvider
const (
	CFGEngineEnter         = "engine-enter"
	CFGEngineExit          = "engine-exit"
	CFGEngineMembers       = "engine-members"
	CFGEngineFileGlobals   = "engine-globals"
	CFGEngineFileFunctions = "engine-functions"
)

// Package - descriptor for generating package
type Package struct {
	Name     string
	Files    []*File
	types    map[string]*DefinedType
	builders []*Builder
	Engine   *EngineDescriptor
	*Project
	Features Features
	//extEngines - map package -> engineVar
	extEngines  map[string]string
	fullPackage string
	pos         lexer.Position
	engineless  bool
}

func (desc *Package) postParsed() error {
	desc.types = map[string]*DefinedType{}
	checkName := func(name string) error {
		if name == "" {
			return errors.New("name can not be empty")
		}
		if dt, ok := desc.types[name]; ok {
			return fmt.Errorf("name duplicate (first occurance at %v)", dt.pos)
		}
		return nil
	}
	for _, f := range desc.Files {
		f.Pckg = desc
		desc.prepareFileFields(f)
		for _, e := range f.Enums {
			e.File = f
			e.Pckg = desc
			err := e.postParsed()
			if err != nil {
				return err
			}
			for _, field := range e.Fields {
				field.Parent = e
			}
		}
		for _, e := range f.Entries {
			e.File = f
			e.Pckg = desc
			desc.prepareModifiersFields(e)
			typename := e.Name
			err := checkName(typename)
			if err != nil {
				return fmt.Errorf("at %v: %v", e.Pos, err)
			}
			if !e.HasModifier(TypeModifierExternal) {
				if dt, ok := desc.types[typename]; ok {
					return fmt.Errorf("duplicate entity found: %s: %#v", typename, dt)
				}
			}
			if e.HasModifier(TypeModifierAbstract) && !e.HasModifier(TypeModifierExtendable) {
				e.TypeModifers[TypeModifierExtendable] = true
			}
			if e.HasModifier(TypeModifierExtendable) && e.BaseTypeName == "" {
				if desc.Project.Options.ExtendableTypeDescr != DoNotGenerateFieldForExtendableTypes {
					tn := TipString
					etdf := &Field{
						parent:      e,
						Annotations: Annotations{},
						Features:    Features{},
						Name:        ExtendableTypeDescriptorFieldName,
						Type:        &TypeRef{NonNullable: true, Type: tn},
					}
					e.Fields = append(e.Fields, etdf)
					e.FieldsIndex[ExtendableTypeDescriptorFieldName] = etdf
					etdf.Features.Set(FeaturesAPIKind, FCIgnore, true)
					etdf.Features.Set(FeaturesCommonKind, FCReadonly, true)
				}
			}
			desc.types[typename] = &DefinedType{
				name:        typename,
				external:    false,
				pckg:        desc.Name,
				entry:       e,
				packagePath: desc.fullPackage,
				pos:         e.Pos,
			}
		}
		for _, enum := range f.Enums {
			err := checkName(enum.Name)
			if err != nil {
				return fmt.Errorf("at %v: %v", enum.Pos, err)
			}

			desc.types[enum.Name] = &DefinedType{
				name:        enum.Name,
				external:    false,
				pckg:        desc.Name,
				enum:        enum,
				packagePath: desc.fullPackage,
				pos:         enum.Pos,
			}
		}
	}
	return nil
}

func (desc *Package) initEngine() {
	desc.Engine = &EngineDescriptor{
		Fields:         jen.Id(EngineVivard).Op("*").Qual(VivardPackage, "Engine").Line(),
		Initializator:  &jen.Statement{},
		Initialized:    &jen.Statement{},
		Start:          &jen.Statement{},
		Functions:      &jen.Statement{},
		SingletonInits: map[string]*jen.Statement{},
	}
}

func (desc *Package) beforePrepare() error {
	var codeGenerator *CodeGenerator
	for _, b := range desc.generators {
		if g, ok := b.(*CodeGenerator); ok {
			codeGenerator = g
			break
		}
	}
	for _, f := range desc.Files {
		err := desc.processModifiers(f)
		if err != nil {
			return fmt.Errorf("at %v: %w", f.Pos, err)
		}
		for _, e := range f.Entries {
			if e.BaseTypeName != "" {
				bt := e.GetBaseType()
				if bt == nil {
					return fmt.Errorf("at %v: base type not found: %s", e.Pos, e.BaseTypeName)
				}
				if !bt.HasModifier(TypeModifierExtendable) {
					return fmt.Errorf("at %v: base type should be extendable: %s", e.Pos, e.BaseTypeName)
				}
				bt.AddDescendant(e)
			}
			err := desc.processModifiers(e)
			if err != nil {
				return fmt.Errorf("at %v: %w", e.Pos, err)
			}
			for _, f := range e.Fields {
				f.PostProcess()
				err = desc.processModifiers(f)
				if err != nil {
					return fmt.Errorf("at %v: %w", f.Pos, err)
				}
				if f.HasModifier(AttrModifierCalculated) {
					f.Features.Set(FeaturesCommonKind, FCReadonly, true)
				}
			}
			for _, m := range e.Methods {
				err = desc.processModifiers(m)
				if err != nil {
					return fmt.Errorf("at %v: %w", m.Pos, err)
				}
			}
			if e.HasModifier(TypeModifierExternal) {
				err := desc.Project.addExternal(e, desc)
				if err != nil {
					return err
				}
			}
		}
	}
	for _, f := range desc.Files {
		err := desc.processStandardFileAnnotations(f)
		if err != nil {
			return err
		}
		for _, e := range f.Entries {
			err := desc.checkType(e)
			if err != nil {
				return fmt.Errorf("at %v: %w", e.Pos, err)
			}
			err = desc.checkTypeRelations(e)
			if err != nil {
				return fmt.Errorf("at %v: %w", e.Pos, err)
			}
			err = desc.processStandardTypeAnnotations(e)
			if err != nil {
				return err
			}
			if codeGenerator != nil {
				err = codeGenerator.createAdditionalFields(e)
			}
			if e.HasModifier(TypeModifierSingleton) {
				e.Features.Set(FeaturesAPIKind, FAPILevel, FAPILIgnore)
			}
		}
	}

	return nil
}

func (desc *Package) prepare() error {
	desc.initEngine()
	for _, gen := range desc.Project.generators {
		err := gen.Prepare(desc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (desc *Package) prepareFileFields(f *File) {
	f.Annotations = Annotations{}
}

func (desc *Package) prepareModifiersFields(t *Entity) {
	t.Annotations = Annotations{}
	for _, f := range t.Fields {
		f.Annotations = Annotations{}
	}
	for _, m := range t.Methods {
		m.Annotations = Annotations{}
	}
}

func (desc *Package) processModifiers(item interface{}) error {
	if t, ok := item.(*Entity); ok {
		for _, m := range t.Modifiers {
			if m.Annotation != nil {
				err := desc.checkAnnotation(m.Annotation, item)
				if err != nil {
					return err
				}
				t.Annotations[m.Annotation.Name] = m.Annotation
			}
		}
	} else if f, ok := item.(*File); ok {
		for _, m := range f.Modifiers {
			if m.Annotation != nil {
				err := desc.checkAnnotation(m.Annotation, item)
				if err != nil {
					return err
				}
				f.Annotations[m.Annotation.Name] = m.Annotation
			}
		}
	} else {
		var m []*EntryModifier
		// ann := Annotations{}
		var ann Annotations
		switch t := item.(type) {
		case *Field:
			// t.Annotations = ann
			ann = t.Annotations
			m = t.Modifiers
		case *Method:
			// t.Annotations = ann
			ann = t.Annotations
			m = t.Modifiers
		}
		for _, m := range m {
			if m.Annotation != nil {
				err := desc.checkAnnotation(m.Annotation, item)
				if err != nil {
					return err
				}
				ann[m.Annotation.Name] = m.Annotation
			} else if m.AttrModifier != "" {
				if m.AttrModifier == string(AttrModifierID) || m.AttrModifier == string(AttrModifierIDAuto) {
					item.(*Field).Type.NonNullable = true
				}
			}
		}
	}

	return nil
}

func (desc *Package) checkAnnotation(ann *Annotation, item interface{}) error {
	found, err := desc.checkStandardAnnotation(ann, item)
	if err != nil {
		return err
	}

	for _, gen := range desc.generators {
		ok, err := gen.CheckAnnotation(desc, ann, item)
		if err != nil {
			return err
		}
		found = found || ok
	}
	if !found {
		switch desc.Options().UnknownAnnotation {
		case UnknownAnnotationError:
			return fmt.Errorf("at %v: unknown annotation: %s", ann.Pos, ann.Name)
		case UnknownAnnotationWarning:
			desc.AddWarning(fmt.Sprintf("at %v: unknown annotation: %s", ann.Pos, ann.Name))
		}
	}
	return nil
}

func (desc *Package) checkStandardAnnotation(ann *Annotation, item interface{}) (ok bool, err error) {
	switch ann.Name {
	case AnnotationFind:
		ok = true
	case AnnotationRefPackage, AnnotationEngineless:
		if _, isFile := item.(*File); isFile {
			ok = true
		}
	default:
		gopref := fmt.Sprintf("%s:", AnnotationGo)
		if ann.Name == AnnotationGo || (len(ann.Name) > len(gopref) && ann.Name[:len(gopref)] == gopref) {
			ok = true
		}
	}
	return
}

func (desc *Package) processStandardFileAnnotations(f *File) (err error) {
	for _, a := range f.Annotations {
		switch a.Name {
		case AnnotationRefPackage:
			found := false
			for _, value := range a.Values {
				if value.Value == nil {
					desc.GetExtEngineRef(value.Key)
					continue
				}

				switch value.Key {
				case ARFPackageName:
					if value.Value.String == nil {
						return fmt.Errorf("at %v: package name should be given for annotation %s", a.Pos, AnnotationRefPackage)
					}
					desc.GetExtEngineRef(*value.Value.String)
					found = true
				case ARFPackageNames:
					if value.Value.String == nil {
						return fmt.Errorf("at %v: packages names should be given for annotation %s", a.Pos, AnnotationRefPackage)
					}
					packages := strings.Fields(*value.Value.String)
					for _, p := range packages {
						desc.GetExtEngineRef(p)
					}
					found = true
				}
			}
			if !found {
				// we should not be here actually...
				packageName, ok := a.GetNameTag(ARFPackageName)
				if !ok {
					return fmt.Errorf("at %v: package name should be given for annotation %s", a.Pos, AnnotationRefPackage)
				}
				desc.GetExtEngineRef(packageName)
			}
		case AnnotationEngineless:
			if len(a.Values) == 0 || a.Values[0].Key == "false" {
				desc.engineless = true
				desc.pos = a.Pos
				break
			}
		}
	}
	return nil
}

func (desc *Package) processStandardTypeAnnotations(e *Entity) (err error) {
	for _, a := range e.Annotations {
		switch a.Name {
		case AnnotationFind:
			for _, at := range a.Values {
				if at.Value == nil {
					t, ok := desc.FindType(at.Key)
					if !ok || t.entry == nil {
						return fmt.Errorf("at %v: type %s not found for %s annotation", at.Pos, at.Key, AnnotationFind)
					}
					if _, ok := t.entry.Features.GetEntity(FeaturesAPIKind, FAPIFindParamType); ok {
						desc.AddWarning(
							fmt.Sprintf(
								"at %v: %s annotation: too many param types for type %s; skipping",
								at.Pos,
								AnnotationFind,
								e.Name,
							),
						)
						continue
					}
					t.entry.Features.Set(FeaturesAPIKind, FAPIFindParamType, e)
					e.Features.Set(FeaturesAPIKind, FAPIFindFor, t.entry)
					for _, f := range e.Fields {
						fldName := f.Name
						op := AFTEqual
						an, ok := f.Annotations[AnnotationFind]
						if ok {
							if len(an.Values) > 0 && an.Values[0].Value == nil {
								fldName = an.Values[0].Key
							} else if fn, ok := an.GetStringTag(AnnFndFieldTag); ok {
								fldName = fn
							}
							op = an.GetString(AnnFndTypeTag, AFTEqual)
						}
						searchField := t.entry.GetField(fldName)
						f.Features.Set(FeaturesAPIKind, FAPIFindParam, op)
						f.Features.Set(FeaturesAPIKind, FAPIFindForName, fldName)
						if searchField == nil {
							if fldName != AFFDeleted {
								if strings.Index(fldName, ".") != -1 {
									fields, err := desc.FindFieldsForComplexName(t.entry, fldName)
									if err != nil {
										return fmt.Errorf("at %v: %v", an.Pos, err)
									}
									f.Features.Set(FeaturesAPIKind, FAPIFindFor, fields[0])
									f.Features.Set(FeaturesAPIKind, FAPIFindForEmbedded, fields)
								} else {
									return fmt.Errorf("at %v: can not find field %s for find attr %s", f.Pos, fldName, f.Name)
								}
							}
						} else {
							f.Features.Set(FeaturesAPIKind, FAPIFindFor, searchField)
						}

					}
				}
			}
		case AnnotationDeletable:
			if !a.GetBool(deletableAnnotationIgnore, false) {
				e.Features.Set(FeatGoKind, FCGDeletable, true)
				if tag := a.GetTag(deletableAnnotationWithField); tag != nil /* || cg.options.GenerateDeletedField  */ {
					dfn := deletedFieldName
					//if tag != nil {
					if n, ok := tag.GetString(); ok {
						dfn = n
					} else if b, ok := tag.GetBool(); ok && !b {
						dfn = ""
					}
					//}
					if dfn != "" {
						e.Features.Set(FeatGoKind, FCGDeletedFieldName, dfn)
					}
				}
			}
		case AnnotationAccess:
			logCreated := a.GetBool(accessAnnotationCreated, true)
			logModified := a.GetBool(accessAnnotationModified, true)
			e.Features.Set(FeatGoKind, FCGLogCreated, logCreated)
			e.Features.Set(FeatGoKind, FCGLogModified, logModified)
			if val, ok := a.GetBoolTag(accessAnnotationUserID); ok {
				e.Features.Set(FeatGoKind, FCGLogCreatedBy, val && logCreated)
				e.Features.Set(FeatGoKind, FCGLogModifiedBy, val && logModified)
			}
		}
	}
	return
}

func (desc *Package) checkType(t *Entity) error {
	if idfld := t.GetIdField(); idfld == nil {
		if !t.HasModifier(TypeModifierEmbeddable) && !t.HasModifier(TypeModifierTransient) &&
			!t.HasModifier(TypeModifierSingleton) &&
			!t.HasModifier(TypeModifierConfig) &&
			!(t.HasModifier(TypeModifierExternal) && t.Incomplete) &&
			t.BaseTypeName == "" {
			if desc.Options().AutoGenerateIDField {
				if _, ok := t.FieldsIndex[autoGeneratedIDFieldName]; ok {
					return fmt.Errorf(
						"can not auto generate %s field for type %s:field with this name already exists",
						autoGeneratedIDFieldName,
						t.Name,
					)
				}
				idfld := &Field{
					Pos:         t.Pos,
					Annotations: Annotations{},
					Modifiers:   []*EntryModifier{{AttrModifier: string(AttrModifierIDAuto)}},
					Name:        autoGeneratedIDFieldName,
					Type:        &TypeRef{Type: TipInt, NonNullable: true},
					Features:    Features{},
					parent:      t,
				}
				t.Fields = append(t.Fields, idfld)
				t.FieldsIndex[autoGeneratedIDFieldName] = idfld
			} else {
				return fmt.Errorf(
					"there is no id field defined for type %s (use AutoGenerateIDField option for automatic generating)",
					t.Name,
				)
			}
		}
	}
	return nil
}

func (desc *Package) checkTypeRelations(t *Entity) error {
	idfld := t.GetIdField()
	for _, f := range t.FieldsIndex {
		if f.HasModifier(AttrModifierOneToMany) {
			if !f.Type.Complex || f.Type.Array == nil {
				return fmt.Errorf("one-to-many modifier can't be used with type %s", f.Type.Type)
			}
			if tt, ok := desc.FindType(f.Type.Array.Type); ok && tt.entry != nil {
				//TODO make possible to change foreign-key field
				fldname := t.Name + "ID"
				var fkField *Field
				if f := tt.entry.GetField(fldname); f != nil {
					if f.Type.Type != idfld.Type.Type {
						return fmt.Errorf("foreign-key field with name %s already exists in type %s", fldname, f.Type.Type)
					}
					fkField = f
				} else {
					fkField = &Field{
						Pos:         t.Pos,
						Name:        fldname,
						Type:        &TypeRef{Type: idfld.Type.Type, NonNullable: true},
						Features:    Features{},
						parent:      tt.entry,
						Annotations: Annotations{},
					}
					tt.entry.Fields = append(tt.entry.Fields, fkField)
					tt.entry.FieldsIndex[autoGeneratedIDFieldName] = fkField
				}
				fkField.Modifiers = []*EntryModifier{{AttrModifier: string(AttrModifierForeignKey)}}
				tt.entry.Features.Set(FeaturesCommonKind, FCForeignKey, t)
				tt.entry.Features.Set(FeaturesCommonKind, FCForeignKeyField, fkField)
				// tt.entry.Features.Set(FeaturesCommonKind, FCSkipAccessors, true)

				f.Features.Set(FeaturesCommonKind, FCIgnore, !f.HasModifier(AttrModifierEmbedded))
				f.Features.Set(FeaturesCommonKind, FCOneToManyType, tt.entry)
				f.Features.Set(FeaturesCommonKind, FCOneToManyField, fkField)
			} else {
				return fmt.Errorf("undefined type %s for foreign-key field ", f.Type.Array.Type)
			}
		} else if f.Type.Array != nil {
			if refT, ok := desc.FindType(f.Type.Array.Type); ok && !f.HasModifier(AttrModifierEmbeddedRef) && refT.entry != nil {
				f.Features.Set(FeaturesCommonKind, FCManyToManyType, refT.entry)
				f.Features.Set(FeaturesCommonKind, FCManyToManyIDField, refT.entry.GetIdField())
				refT.entry.Features.Set(FeaturesCommonKind, FCRefsAsManyToMany, true)
			}

		}
	}
	return nil
}

func (desc *Package) doGenerate(bldr *Builder) error {

	for _, gen := range desc.generators {
		err := gen.Generate(bldr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (desc *Package) generateEngine() error {

	if desc.engineless {
		if len(desc.extEngines) > 0 {
			return fmt.Errorf("at %v: engineless can not be used: there are references to another packages", desc.pos)
		}
		//TODO check another requirements for Engine
		return nil
	}
	extInit := &jen.Statement{}
	utils.WalkMap(
		desc.extEngines,
		func(varname string, pckg string) error {
			pn := desc.Project.GetFullPackage(pckg)
			desc.Engine.Fields.Add(
				jen.Id(varname).Op("*").Qual(pn, "Engine").Line(),
			)
			extInit.Add(
				jen.Id(EngineVar).Dot(varname).Op("=").Id("v").Dot("Engine").Params(jen.Lit(pckg)).Assert(
					jen.Op("*").Qual(
						pn,
						"Engine",
					),
				).Line(),
			)
			return nil
		},
	)
	cf := CodeFragmentContext{
		Package:    desc,
		MethodKind: EngineNotAMethod,
	}
	if desc.Project.ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGEngineMembers, &cf, false) != nil {
		desc.Engine.Fields.Add(cf.body).Line()
	}
	desc.Engine.file = jen.NewFile(desc.Name)
	desc.Engine.file.HeaderComment(fmt.Sprintf("Code generated for package %s by vivgen. DO NOT EDIT.", desc.Name))
	desc.Engine.file.Add(
		jen.Type().Id("Engine").Struct(desc.Engine.Fields).Line(),
		jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id("Name").Params().Parens(jen.String()).Block(
			jen.Return(
				jen.Lit(desc.Name),
			),
		).Line(),
	)
	cf = CodeFragmentContext{
		Package:    desc,
		MethodKind: EngineNotAMethod,
	}
	if desc.Project.ProvideCodeFragment(
		CodeFragmentModuleGeneral,
		cf.MethodKind,
		CFGEngineFileGlobals,
		&cf,
		false,
	) != nil {
		desc.Engine.file.Add(cf.body).Line()
	}
	cf = CodeFragmentContext{
		Package:           desc,
		MethodName:        "Start",
		MethodKind:        MethodEngineRegisterService,
		EngineAvailable:   true,
		ErrorRet:          []jen.Code{jen.Id("err")},
		BeforeReturnError: func() {},
		ErrVar:            "err",
		Params:            map[string]string{ParamVivardEngine: "v"},
	}
	if desc.Project.ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGEngineEnter, &cf, false) != nil {
		desc.Engine.file.Add(
			jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id("ProvideServices").Params(
				jen.Id("v").Op("*").Qual(VivardPackage, "Engine"),
			).Block(cf.body).Line(),
		)
	}
	cf = CodeFragmentContext{
		Package:           desc,
		MethodName:        "Prepare",
		MethodKind:        MethodEnginePrepare,
		EngineAvailable:   true,
		ErrorRet:          []jen.Code{jen.Id("err")},
		BeforeReturnError: func() {},
		ErrVar:            "err",
	}
	desc.Engine.file.Add(
		jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id("Prepare").
			Params(
				jen.Id("v").Op("*").Qual(VivardPackage, "Engine"),
				//jen.Id("_").Op("*").Qual(dependenciesPackage, "Provider"),
			).Parens(jen.Error()).BlockFunc(
			func(g *jen.Group) {
				g.Var().Id("err").Id("error")
				g.Id("eng").Dot(EngineVivard).Op("=").Id("v")
				cf.Push(g)
				desc.Project.ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGEngineEnter, &cf, false)
				g.Add(desc.Engine.Initializator)
				utils.WalkMap(
					desc.Engine.SingletonInits,
					func(statement *jen.Statement, _ string) error {
						g.Add(statement)
						return nil
					},
				)

				g.Add(desc.Engine.Initialized)
				g.Add(extInit)
				if desc.Engine.prepAdd != nil {
					g.Add(desc.Engine.prepAdd)
				}
				desc.Project.ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGEngineExit, &cf, false)
				g.Return(
					jen.Id("err"),
				)
				cf.Pop()
			},
		).Line(),
	)
	cf = CodeFragmentContext{
		Package:           desc,
		MethodName:        "Start",
		MethodKind:        MethodEngineStart,
		EngineAvailable:   true,
		ErrorRet:          []jen.Code{jen.Id("err")},
		BeforeReturnError: func() {},
		ErrVar:            "err",
	}
	desc.Engine.file.Add(
		jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id("Start").Params().Parens(jen.Error()).BlockFunc(
			func(g *jen.Group) {
				g.Var().Id("err").Id("error")
				cf.Push(g)
				desc.Project.ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGEngineEnter, &cf, false)
				g.Add(desc.Engine.Start)
				if desc.Engine.startAdd != nil {
					g.Add(desc.Engine.startAdd)
				}
				cf.Push(g)
				desc.Project.ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGEngineExit, &cf, false)
				g.Return(
					jen.Id("err"),
				)
				cf.Pop()
			},
		).Line(),
	)
	desc.Engine.file.Add(
		desc.Engine.Functions,
	)

	cf = CodeFragmentContext{
		Package:    desc,
		MethodKind: EngineNotAMethod,
	}
	if desc.Project.ProvideCodeFragment(
		CodeFragmentModuleGeneral,
		cf.MethodKind,
		CFGEngineFileFunctions,
		&cf,
		false,
	) != nil {
		desc.Engine.file.Add(cf.body).Line()
	}
	return nil
}

func returnIfErr() *jen.Statement {
	return jen.If(
		jen.Id("err").Op("!=").Nil().Block(
			jen.Return(),
		),
	)
}
func returnIfErrValue(prefix ...*jen.Statement) *jen.Statement {
	return jen.If(
		jen.Id("err").Op("!=").Nil().Block(
			jen.Return(
				jen.ListFunc(
					func(g *jen.Group) {
						for i := 0; i < len(prefix); i++ {
							g.Add(prefix[i])
						}
						g.Id("err")
					},
				),
			),
		),
	)
}

// FindType looks for type descriptor and returns it
func (desc *Package) FindType(name string) (dt *DefinedType, ok bool) {
	dt, ok = desc.types[name]
	if !ok {
		dt, ok = desc.Project.FindType(name)
	}
	return
}

func (desc *Package) AddType(name string) (*Entity, error) {
	if _, ok := desc.FindType(name); ok {
		return nil, fmt.Errorf("type '%s' already exists in package '%s", name, desc.Name)
	}
	tip := &Entity{
		Modifiers:       []*EntityModifier{},
		Name:            name,
		BaseTypeName:    "",
		Entries:         nil,
		Incomplete:      false,
		Fields:          []*Field{},
		Methods:         []*Method{},
		FieldsIndex:     map[string]*Field{},
		MethodsIndex:    map[string]*Method{},
		Annotations:     Annotations{},
		Features:        Features{},
		TypeModifers:    map[TypeModifier]bool{},
		Pckg:            desc,
		BaseField:       nil,
		File:            nil,
		FullAnnotations: nil,
	}
	desc.types[tip.Name] = &DefinedType{
		name:        tip.Name,
		pckg:        desc.Name,
		external:    false,
		entry:       tip,
		packagePath: desc.fullPackage,
	}

	file := &File{
		Name:     strings.ToLower(name),
		FileName: name,
		Package:  desc.Name,
		Meta:     nil,
		Entries:  []*Entity{tip},
		Pckg:     desc,
	}
	tip.File = file
	desc.Files = append(desc.Files, file)
	return tip, nil
}

// RegisterType looks for type descriptor and returns it
func (desc *Package) RegisterType(e *Entity) {
	desc.types[e.Name] = &DefinedType{
		name:        e.Name,
		external:    false,
		pckg:        desc.Name,
		entry:       e,
		packagePath: desc.fullPackage,
	}
}

// Entity returns underlying type
func (dt *DefinedType) Entity() *Entity {
	return dt.entry
}

// Enum returns underlying type
func (dt *DefinedType) Enum() *Enum {
	return dt.enum
}

func (dt *DefinedType) IdType() (string, error) {
	if dt.entry != nil {
		if idField := dt.entry.GetIdField(); idField != nil {
			return idField.Type.Type, nil
		}
		return "", errors.New("id field is not defined")
	} else if dt.enum != nil {
		return dt.enum.AliasForType, nil
	}
	return "", errors.New("type is not defined")
}

func (dt *DefinedType) Position() lexer.Position {
	return dt.pos
}

// GetName returns name of entity
func (e *Entity) GetName() string {
	return e.Name
}

// GetMethodName returns name for method of given kind
func (e *Entity) GetMethodName(mk MethodKind) string {
	if mk > methodMax {
		//TODO: lookup for custom methods
		panic(fmt.Sprintf("cannot find method kind: %d", int(mk)))
	}
	templ := MethodsNamesTemplates[mk]
	if templ == "" {
		panic(fmt.Sprintf("name template is not given for method kind: %d", int(mk)))
	}
	name := e.GetName()
	return fmt.Sprintf(templ, name)
}

// AddTag adds tag to Go struct
func (desc *Package) AddTag(f *Field, key string, value string) {
	if old, ok := f.Tags[key]; ok && old != value {
		desc.AddWarning(fmt.Sprintf("at %v: gotags annotation rewrites tag's '%s' value: %s => %s", f.Pos, key, old, value))
	}
	f.Tags[key] = value
}

func (desc *Package) GetTypeEngineAccessor(t *Entity) jen.Code {
	if t.Pckg != desc {
		//desc.AddError(fmt.Errorf("at %v: cross package engine access not implemented yet", t.Pos))
		engvar := desc.GetExtEngineRef(t.Pckg.Name)
		return jen.Id(EngineVar).Dot(engvar)
	}
	return jen.Id(EngineVar)
}

// AddBaseFieldTag adds tag to Go struct
func (desc *Package) AddBaseFieldTag(e *Entity, key string, value string) {
	f := e.GetBaseField()
	desc.AddTag(f, key, value)
}

// GetMethodName returns name for method of given kind
func (desc *Package) GetMethodName(mk MethodKind, name string) string {
	if mk > methodMax {
		//TODO: lookup for custom methods
		panic(fmt.Sprintf("cannot find method kind: %d", int(mk)))
	}
	templ := MethodsNamesTemplates[mk]
	if templ == "" {
		panic(fmt.Sprintf("name template is not given for method kind: %d", int(mk)))
	}
	parts := strings.SplitN(name, ".", 2)
	return fmt.Sprintf(templ, parts[len(parts)-1])
}

// GetHookName returns name for method of given kind
func (desc *Package) GetHookName(hookKind string, f *Field) string {
	if f != nil {
		return fmt.Sprintf(hookFuncsTemmplates[hookKind], f.parent.Name, f.Name)
	} else {
		return hookFuncsTemmplates[hookKind]
	}
}

// HasModifier checks whether type that given TypeRef refers to has modifier
func (desc *Package) HasModifier(tr *TypeRef, modifier TypeModifier) bool {
	if t, ok := desc.FindType(tr.Type); ok {
		return t.Entity() != nil && t.Entity().HasModifier(modifier)
	}
	return false
}

// GetExtEngineRef returns name of property in Engine for external engine with name pckgName
func (desc *Package) GetExtEngineRef(pckgName string) string {
	if desc.extEngines[pckgName] == "" {
		desc.extEngines[pckgName] = pckgName + "Eng"
	}
	return desc.extEngines[pckgName]
}

func (desc *Package) GetFieldTypePackage(f *Field) string {
	return desc.GetTypeRefPackage(f.Type)
}

func (desc *Package) GetTypeRefPackage(tr *TypeRef) string {
	return desc.GetTypePackage(tr.Type)
}

func (desc *Package) GetTypePackage(tip string) string {
	alias := desc.GetTypePackageAlias(tip)
	return desc.Project.GetFullPackage(alias)
}

func (desc *Package) GetTypePackageAlias(tip string) string {
	parts := strings.SplitN(tip, ".", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

func (desc *Package) GetRealTypeName(tip string) string {
	parts := strings.SplitN(tip, ".", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return tip
}

func (desc *Package) AddToEnginePrepare(stmt *jen.Statement) {
	if desc.Engine.prepAdd == nil {
		desc.Engine.prepAdd = stmt
	} else {
		desc.Engine.prepAdd.Add(stmt)
	}
}

func (desc *Package) AddToEngineStart(stmt *jen.Statement) {
	if desc.Engine.startAdd == nil {
		desc.Engine.startAdd = stmt
	} else {
		desc.Engine.startAdd.Add(stmt)
	}
}

func (desc *Package) TypeStmt(e *Entity) *jen.Statement {
	if e.Pckg == desc {
		return e.Features.Stmt(FeatGoKind, FCGType)
	}
	return jen.Qual(e.Pckg.fullPackage, e.FS(FeatGoKind, FCGName))
}

func (desc *Package) FindFieldsForComplexName(e *Entity, name string) ([]*Field, error) {
	return desc.findFieldsForParts(e, strings.Split(name, "."), nil)
}

func (desc *Package) findFieldsForParts(e *Entity, parts []string, fields []*Field) ([]*Field, error) {
	if fields == nil {
		fields = make([]*Field, 0, len(parts))
	}
	if len(parts) == 0 {
		return fields, nil
	}
	f := e.GetField(parts[0])
	if f == nil {
		return nil, fmt.Errorf("field '%s' not found in type '%s'", parts[0], e.Name)
	}
	if f.Type.Map != nil {
		return nil, fmt.Errorf("field '%s': maps are not supported", parts[0])
	}
	if f.Type.Complex && len(parts) > 1 && !f.HasModifier(AttrModifierEmbedded) {
		return nil, fmt.Errorf("field '%s': only embedded fields are supported", parts[0])
	}

	typeName := f.Type.Type

	if f.Type.Array != nil {
		typeName = f.Type.Array.Type
	}
	if typeName != TipInt && typeName != TipString && typeName != TipBool && typeName != TipDate && typeName != TipFloat {
		dt, ok := desc.FindType(typeName)
		if !ok {
			return nil, fmt.Errorf("field '%s': type '%s' not found", parts[0], f.Type.Type)
		}
		if dt.external || dt.entry == nil {
			return nil, fmt.Errorf("field '%s': external types are not supported", parts[0])
		}
		fields = append(fields, f)
		return desc.findFieldsForParts(dt.entry, parts[1:], fields)
	}
	if len(parts) != 1 {
		return nil, fmt.Errorf("field '%s': type %s is not complex", parts[0], typeName)
	}
	fields = append(fields, f)
	return fields, nil
}
