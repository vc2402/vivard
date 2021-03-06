package gen

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
)

//Package - descriptor for generating package
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
}

func (desc *Package) postParsed() error {
	desc.types = map[string]*DefinedType{}
	for _, f := range desc.Files {
		f.Pckg = desc
		for _, e := range f.Entries {
			e.File = f
			e.Pckg = desc
			desc.prepareModifiersFields(e)
			typename := e.Name
			if typename == "" {
				return fmt.Errorf("undefined entity found")
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
				if desc.Project.Options.ExtendableTypeDescr != DoNotGenerateFieldForExtandableTypes {
					tn := TipString
					etdf := &Field{
						parent:      e,
						Annotations: Annotations{},
						Features:    Features{},
						Name:        ExtendableTypeDescriptorFieldName,
						Type:        &TypeRef{NonNullable: true, Type: tn},
					}
					e.Fields = append(e.Fields, etdf)
					etdf.Features.Set(FeaturesAPIKind, FCIgnore, true)
					etdf.Features.Set(FeaturesCommonKind, FCReadonly, true)
				}
			}
			desc.types[typename] = &DefinedType{name: typename, external: false, pckg: desc.Name, entry: e, packagePath: desc.fullPackage}
		}
	}
	err := desc.processMetas()
	if err != nil {
		return err
	}
	return nil
}

func (desc *Package) prepare() error {
	for _, f := range desc.Files {
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
				err := desc.Project.addExternal(e)
				if err != nil {
					return err
				}
			}
		}
	}
	for _, f := range desc.Files {
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
			if e.HasModifier(TypeModifierSingleton) {
				e.Features.Set(FeaturesAPIKind, FAPILevel, FAPILIgnore)
			}
		}
	}
	desc.Engine = &EngineDescriptor{
		Fields:        jen.Id(EngineVivard).Op("*").Qual(vivardPackage, "Engine").Line(),
		Initializator: &jen.Statement{},
		Initialized:   &jen.Statement{},
		Start:         &jen.Statement{},
		Functions:     &jen.Statement{},
	}
	for _, gen := range desc.Project.generators {
		err := gen.Prepare(desc)
		if err != nil {
			return err
		}
	}
	return nil
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
			return fmt.Errorf("unknown annotation: %s", ann.Name)
		case UnknownAnnotationWarning:
			desc.AddWarning(fmt.Sprintf("unknown annotation: %s", ann.Name))
		}
	}
	return nil
}

func (desc *Package) checkStandardAnnotation(ann *Annotation, item interface{}) (ok bool, err error) {
	switch ann.Name {
	case AnnotationFind:
		ok = true
	default:
		gopref := fmt.Sprintf("%s:", AnnotationGo)
		if ann.Name == AnnotationGo || (len(ann.Name) > len(gopref) && ann.Name[:len(gopref)] == gopref) {
			ok = true
		}
	}
	return
}

func (desc *Package) processStandardTypeAnnotations(e *Entity) (err error) {
	for _, a := range e.Annotations {
		switch a.Name {
		case AnnotationFind:
			for _, at := range a.Values {
				if at.Value == nil {
					t, ok := desc.FindType(at.Key)
					if !ok {
						return fmt.Errorf("at %v: type %s not found for %s annotation", at.Pos, at.Key, AnnotationFind)
					}
					if _, ok := t.entry.Features.GetEntity(FeaturesAPIKind, FAPIFindParamType); ok {
						desc.AddWarning(fmt.Sprintf("at %v: %s annotation: too many param types for type %s; skipping", at.Pos, AnnotationFind, e.Name))
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
						if searchField == nil {
							return fmt.Errorf("at %v: can not find field %s for find attr %s", f.Pos, fldName, f.Name)
						}
						f.Features.Set(FeaturesAPIKind, FAPIFindParam, op)
						f.Features.Set(FeaturesAPIKind, FAPIFindFor, searchField)

					}
				}
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
					return fmt.Errorf("can not auto generate %s field for type %s:field with this name already exists", autoGeneratedIDFieldName, t.Name)
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
				return fmt.Errorf("there is no id field defined for type %s (use AutoGenerateIDField option for automatic generating)", t.Name)
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
			if tt, ok := desc.FindType(f.Type.Array.Type); ok {
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
				tt.entry.Features.Set(FeaturesCommonKind, FCSkipAccessors, true)

				f.Features.Set(FeaturesCommonKind, FCIgnore, !f.HasModifier(AttrModifierEmbeeded))
				f.Features.Set(FeaturesCommonKind, FCOneToManyType, tt.entry)
				f.Features.Set(FeaturesCommonKind, FCOneToManyField, fkField)
			} else {
				return fmt.Errorf("undefined type %s for foreign-key field ", f.Type.Array.Type)
			}
		} else if f.Type.Array != nil {
			if refT, ok := desc.FindType(f.Type.Array.Type); ok {
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
	desc.generateEngine()
	return nil
}

func (desc *Package) generateEngine() error {
	extInit := &jen.Statement{}
	for pckg, varname := range desc.extEngines {
		pn := desc.Project.GetFullPackage(pckg)
		desc.Engine.Fields.Add(
			jen.Id(varname).Op("*").Qual(pn, "Engine").Line(),
		)
		extInit.Add(
			jen.Id(EngineVar).Dot(varname).Op("=").Id("v").Dot("Engine").Params(jen.Lit(pckg)).Assert(jen.Op("*").Qual(pn, "Engine")).Line(),
		)
	}
	desc.Engine.file = jen.NewFile(desc.Name)
	desc.Engine.file.Add(
		jen.Type().Id("Engine").Struct(desc.Engine.Fields).Line(),
		jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id("Name").Params().Parens(jen.String()).Block(
			jen.Return(
				jen.Lit(desc.Name),
			),
		).Line(),
		jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id("Prepare").Params(jen.Id("v").Op("*").Qual(vivardPackage, "Engine")).Parens(jen.Error()).BlockFunc(func(g *jen.Group) {
			g.Var().Id("err").Id("error")
			g.Id("eng").Dot(EngineVivard).Op("=").Id("v")
			if desc.Features.Bool(FeatGoKind, FCGCronRequired) {
				g.Id(cronEngineVar).Op(":=").Id("v").Dot("GetService").Params(jen.Lit(vivard.ServiceCRON)).
					Assert(jen.Op("*").Qual(vivardPackage, "CRONService")).Dot("Cron").Params()
			}
			g.Add(desc.Engine.Initializator)
			g.Add(desc.Engine.Initialized)
			g.Add(extInit)
			g.Return(
				jen.Id("err"),
			)
		}).Line(),
		jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id("Start").Params().Parens(jen.Error()).Block(
			jen.Var().Id("err").Id("error"),
			desc.Engine.Start,
			jen.Return(
				jen.Id("err"),
			),
		).Line(),
		desc.Engine.Functions,
	)
	return nil
}

func returnIfErr() *jen.Statement {
	return jen.If(jen.Id("err").Op("!=").Nil().Block(
		jen.Return(),
	))
}
func returnIfErrValue(prefix ...*jen.Statement) *jen.Statement {
	return jen.If(jen.Id("err").Op("!=").Nil().Block(
		jen.Return(jen.ListFunc(
			func(g *jen.Group) {
				for i := 0; i < len(prefix); i++ {
					g.Add(prefix[i])
				}
				g.Id("err")
			}),
		),
	))
}

//FindType looks for type descriptor and returns it
func (desc *Package) FindType(name string) (dt *DefinedType, ok bool) {
	dt, ok = desc.types[name]
	if !ok {
		dt, ok = desc.Project.FindType(name)
	}
	return
}

//RegisterType looks for type descriptor and returns it
func (desc *Package) RegisterType(e *Entity) {
	desc.types[e.Name] = &DefinedType{name: e.Name, external: false, pckg: desc.Name, entry: e, packagePath: desc.fullPackage}
}

//Entity returnsunderlaying type
func (dt *DefinedType) Entity() *Entity {
	return dt.entry
}

//GetName returns name of entity
func (e *Entity) GetName() string {
	return e.Name
}

//GetMethodName returns name for method of given kind
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

//AddTag adds tag to Go struct
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

//AddBaseFieldTag adds tag to Go struct
func (desc *Package) AddBaseFieldTag(e *Entity, key string, value string) {
	f := e.GetBaseField()
	desc.AddTag(f, key, value)
}

//GetMethodName returns name for method of given kind
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

//GetHookName returns name for method of given kind
func (desc *Package) GetHookName(hookKind string, f *Field) string {
	if f != nil {
		return fmt.Sprintf(hookFuncsTemmplates[hookKind], f.parent.Name, f.Name)
	} else {
		return hookFuncsTemmplates[hookKind]
	}
}

//HasModifier checks whether type that given TypeRef refers to has modifier
func (desc *Package) HasModifier(tr *TypeRef, modifier TypeModifier) bool {
	if t, ok := desc.FindType(tr.Type); ok {
		return t.Entity().HasModifier(modifier)
	}
	return false
}

//GetExtEngineRef returns name of property in Engine for external engine with name pckgName
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
