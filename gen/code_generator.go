package gen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
)

const (
	codeGeneratorName           = "_codeGenerator"
	codeGeneratorAnnotation     = "go"
	codeGeneratorAnnotationTags = "gotags"
	AnnotationDeletable         = "deletable"
	// AnnotationAccess - log access (currently only createdOn and modifiedOn)
	AnnotationAccess = "access"

	//cgaNameTag may be used to set go name of field or type
	cgaNameTag                     = "name"
	codeGenAnnoSingletonEngineAttr = "attrName"

	CodeGeneratorOptionsName = "go"

	// deletableAnnotationWithField - generate Deleted field for entity that can be deleted
	deletableAnnotationWithField = "field"
	// deletableAnnotationIgnore - do not generate Delete operation (overwrites default generation)
	deletableAnnotationIgnore = "ignore"

	accessAnnotationCreated  = "created"
	accessAnnotationModified = "modified"
	// may be set for 'access' annotation to log user id in createdBy and modifiedBy
	accessAnnotationUserID = "userID"

	AnnotationBulk           = "bulk" // special annotation for bulk operations
	AnnotationBulkNew        = "new"
	AnnotationBulkSet        = "set"
	AnnotationBulkFilter     = "filter"
	AnnotationBulkFilterRaw  = "raw"
	AnnotationBulkFilterAll  = "all"
	AnnotationBulkFilterEach = "each"
)

const (
	FeatGoKind FeatureKind = "go"

	FCGSingletonAttrName = "engineAttr"

	// FCGType - jen.Code feature for *Entity and *Field;
	FCGType = "type"
	// FCGAttrType - jen.Code feature for *Field - type of attr in struct (maybe pointer);
	FCGAttrType = "attr-type"
	// FCGName - string; name of field or struct
	FCGName = "name"
	// FCGPointer - bool; true if attr is pointer in the struct
	FCGPointer = "is-pointer"
	// FCGBaseTypeAccessorName - name of function to access base type
	FCGBaseTypeAccessorName = "base-type-accessor"
	// FCGBaseTypeAccessorInterface - name of function to access base type
	FCGBaseTypeAccessorInterface = "base-type-interface"
	// FCGBaseTypeNameType - name of type for Type identifier
	FCGBaseTypeNameType = "base-type-type-name"
	// FCGDerivedTypeNameConst - name of constant for derived type
	FCGDerivedTypeNameConst = "derived-type-name"
	// FCGCalculated - field should not be stored; instead will be resolved on demand
	FCGCalculated = "calculated"
	// FCGScriptingRequired - feature for package - it is neccessary to create reference to scripting engine from package engine
	FCGScriptingRequired = "scripting-required"
	// FCGScriptingCreated - feature for package - it is neccessary to create reference to scripting engine from package engine
	FCGScriptingCreated = "scripting-created"
	// FCGExtEngineVar - for Field - name of engine var in engine for external types
	FCGExtEngineVar = "ext-engine-var"
	// FCGDeletable - the entity is deletable
	FCGDeletable = "deletable"
	// FCGLogCreated - createdOn field and its filling is required
	FCGLogCreated = "created"
	// FCGLogCreatedBy - createdBy field and its filling from context is required
	FCGLogCreatedBy = "created-by"
	// FCGLogModified - modifiedOn field and its filling is required
	FCGLogModified = "modified"
	// FCGLogModifiedBy - modifiedBy field and its filling is required
	FCGLogModifiedBy = "modified-by"
	// FCGDeletedFieldName - for entity; name of the deleted field (empty if not needed)
	FCGDeletedFieldName = "del-fld-name"
	// FCGDeletedField - for field; set to true for DeletedOn field
	FCGDeletedField = "del-fld"
	// FCGLogCreatedField - createdOn field
	FCGLogCreatedField = "created-fld"
	// FCGLogModifiedField - modifiedOn field
	FCGLogModifiedField = "modified-fld"
	// FCGLogCreatedByField - createdBy field
	FCGLogCreatedByField = "created-by"
	// FCGLogModifiedByField - modifiedBy field
	FCGLogModifiedByField = "modified-by"
	// FCGBulkNew - for Entity; true if should be generated bulk New operations
	FCGBulkNew = "bulk-new"
	// FCGBulkSet - for Entity; true if should be generated bulk Set operation
	FCGBulkSet        = "bulk-set"
	FCGBulkFilterRaw  = "bulk-filter-raw"
	FCGBulkFilterAll  = "bulk-filter-all"
	FCGBulkFilterEach = "bulk-filter-each"
)

type EntityTypeSelector string

const (
	EntityTypeAll           EntityTypeSelector = "all"
	EntityTypeNone          EntityTypeSelector = "none"
	EntityTypeNonDictionary EntityTypeSelector = "non-dictionary"
)

// CodeGeneratorOptions describes possible options for CodeGenerator
type CodeGeneratorOptions struct {
	// GenerateFieldsAccessors - generate Get and Set for each field (returns non pointers for basic types)
	GenerateFieldsAccessors bool
	// GenerateNullMethods - generate IsNull and SetNull methods for each nullable field
	GenerateNullMethods bool
	// GenerateFieldsEnums - generate int consts for each field of each type, e.g. 'Type1Field1Field' (maybe used for NullableField, search filters etc.)
	GenerateFieldsEnums bool
	// GenerateRemoveOperation - generate Remove and Delete operations for each entity
	GenerateRemoveOperation bool
	// GenerateDeletedField - generate field DeletedOn *time.Time for every entity that can be deleted
	GenerateDeletedField bool
	// GenerateAccessField - generate fields CreatedOn and ModifiedOn (time.Time) for entities
	GenerateAccessField EntityTypeSelector
	// GenerateCreatedBy - generate field CreatedBy by default (only if corresponding access field is generated)
	GenerateCreatedBy bool
	// GenerateModifiedBy - generate fields ModifiedBy by default (only if corresponding access field is generated)
	GenerateModifiedBy bool
	// AllowEmbeddedArraysForDictionary - allow arrays of embedded types for dictionary
	AllowEmbeddedArraysForDictionary bool
	// GenerateBulkNew - generate bulk new operations for every type (default false)
	GenerateBulkNew bool
	// GenerateBulkSet - generate bulk set operations for every type (default false)
	GenerateBulkSet bool
}

// CodeGenerator generates Go code (structs, methods, Engine object  and other)
type CodeGenerator struct {
	proj    *Project
	desc    *Package
	b       *Builder
	options CodeGeneratorOptions
	// scriptingRequired bool
	// scriptingCreated  bool
}

type CSMethodKind int

const (
	CGSetterMethod CSMethodKind = iota
	CGGetterMethod
	CGIsNullMethod
	CGSetNullMethod
	CGIsChangedMethod

	cgLastMethod
)

var cgMethodsTemplates = [cgLastMethod]string{
	"Set%s",
	"Get%s",
	"Is%sNull",
	"Set%sNull",
	"Is%sChanged",
}

type CSComplexMethodKind int

const (
	CGGetComplexAttrMethod CSComplexMethodKind = iota
	CGSetComplexAttrMethod
	CGAddComplexAttrMethod

	cgLastComplexMethod
)

const (
	scriptingEngineField         = "scriptEng"
	cronEngineVar                = "cronEngine"
	scriptingEnginePackage       = "github.com/vc2402/vivard/scripting"
	extendableTypeDescriptorType = "V_%sType"
	extendedTypeTypeName         = "V_%s_%s"

	deletedFieldName    = "DeletedOn"
	createdFieldName    = "CreatedOn"
	modifiedFieldName   = "ModifiedOn"
	createdByFieldName  = "CreatedBy"
	modifiedByFieldName = "ModifiedBy"
)

var cgComplexMethodsTemplates = [cgLastComplexMethod]string{
	"%sGet%s",
	"%sSet%s",
	"%sAdd%s",
}

// Name returns name of Generator
func (cg *CodeGenerator) Name() string {
	return codeGeneratorName
}

// SetDescriptor from DescriptorAware
func (cg *CodeGenerator) SetDescriptor(proj *Project) {
	cg.proj = proj
}

// ProcessMeta - implement MetaProcessor
func (cg *CodeGenerator) ProcessMeta(desc *Package, m *Meta) (bool, error) {
	cg.desc = desc
	ok, err := cg.parseHardcoded(m)
	return ok, err
}

// CheckAnnotation checks that annotation may be utilized by CodeGeneration
func (cg *CodeGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	cg.desc = desc
	if ann.Name == codeGeneratorAnnotationTags {
		for _, v := range ann.Values {
			if v.Value != nil && v.Value.String != nil {
				_, ok := item.(*Field)
				if !ok {
					return true, fmt.Errorf("gotags annotation could be used only with field: %s", *v.Value.String)
				}
				return true, nil
			} else {
				return true, fmt.Errorf("at %v: gotags annotation could countain only strings params: %s", ann.Pos, v.Key)
			}
		}
	} else if _, ok := item.(*Entity); ok && ann.Name == AnnotationDeletable {
		return true, nil
	}
	if _, ok := item.(*Method); ok && ann.Name == AnnotationCall {
		return true, nil
	}
	if _, ok := item.(*Entity); ok && ann.Name == AnnotationAccess {
		return true, nil
	}
	if ann.Name == AnnotationBulk {
		if _, ok := item.(*Entity); ok {
			for _, value := range ann.Values {
				switch value.Key {
				case AnnotationBulkNew, AnnotationBulkSet:
					if _, ok := value.GetBool(); !ok {
						return true, fmt.Errorf("at %v: annotation %s:%s accepts only bool values", ann.Pos, ann.Name, value.Key)
					}
				case AnnotationBulkFilter:
					if _, ok := value.GetString(); !ok {
						return true, fmt.Errorf("at %v: annotation %s:%s should use string value", ann.Pos, ann.Name, value.Key)
					}
				default:
					return true, fmt.Errorf(
						"at %v: annotation %s: unknown annotation parameter: %s",
						ann.Pos,
						ann.Name,
						value.Key,
					)
				}
			}
			return true, nil
		}
		return true, fmt.Errorf("at %v: annotation %s should be used for Entity", ann.Pos, ann.Name)
	}
	if ann.Name == AnnotationReadonly {
		if f, ok := item.(*Field); ok {
			readonly := true
			if set, ok := ann.GetNameTag("set"); ok {
				readonly = set != "false"
			}
			f.Features.Set(FeaturesCommonKind, FCReadonly, readonly)
			return true, nil
		}
		return true, errors.New("readonly annotation can be used with fields only")
	}
	return false, nil
}

// Prepare from Generator interface
func (cg *CodeGenerator) Prepare(desc *Package) error {
	cg.desc = desc
	if _, err := desc.Options().CustomToStruct(CodeGeneratorOptionsName, &cg.options); err != nil {
		desc.AddWarning(fmt.Sprintf("problem while setting custom options for code generator: %v", err))
	}

	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if t.HasModifier(TypeModifierExtendable) && t.BaseTypeName == "" {
				tn := fmt.Sprintf(extendableTypeDescriptorType, t.Name)
				t.Features.Set(FeatGoKind, FCGBaseTypeNameType, tn)
			}
			if t.BaseTypeName != "" ||
				t.HasModifier(TypeModifierExtendable) && !t.HasModifier(TypeModifierAbstract) {
				name := t.Name
				if t.BaseTypeName != "" && t.GetBaseType().Package() != t.Package() {
					name = fmt.Sprintf("%s_%s", t.Package(), t.Name)
				}
				cn := fmt.Sprintf(extendedTypeTypeName, cg.desc.GetRealTypeName(t.BaseTypeName), name)
				t.Features.Set(FeatGoKind, FCGDerivedTypeNameConst, cn)
			}
			if ann, ok := t.Annotations[AnnotationDeletable]; ok || cg.options.GenerateRemoveOperation {
				if !ann.GetBool(deletableAnnotationIgnore, false) {
					t.Features.Set(FeatGoKind, FCGDeletable, true)
					if tag := ann.GetTag(deletableAnnotationWithField); tag != nil || cg.options.GenerateDeletedField {
						dfn := deletedFieldName
						if tag != nil {
							if n, ok := tag.GetString(); ok {
								dfn = n
							} else if b, ok := tag.GetBool(); ok && !b {
								dfn = ""
							}
						}
						if dfn != "" {
							t.Features.Set(FeatGoKind, FCGDeletedFieldName, dfn)
						}
					}
				}
			}
			if ann, ok := t.Annotations[AnnotationBulk]; ok || cg.options.GenerateBulkNew || cg.options.GenerateBulkSet {
				generate := ann.GetBool(AnnotationBulkNew, cg.options.GenerateBulkNew)
				t.Features.Set(FeatGoKind, FCGBulkNew, generate)
				generate = ann.GetBool(AnnotationBulkSet, cg.options.GenerateBulkSet)
				t.Features.Set(FeatGoKind, FCGBulkSet, generate)
				filter := ann.GetString(AnnotationBulkFilter, "")
				if filter != "" {
					filterParts := strings.Split(filter, ",")
					for _, part := range filterParts {
						trimmed := strings.TrimSpace(part)
						switch trimmed {
						case AnnotationBulkFilterRaw:
							t.Features.Set(FeatGoKind, FCGBulkFilterRaw, true)
						case AnnotationBulkFilterAll:
							t.Features.Set(FeatGoKind, FCGBulkFilterAll, true)
						case AnnotationBulkFilterEach:
							t.Features.Set(FeatGoKind, FCGBulkFilterEach, true)
						default:
							return fmt.Errorf("at %v: unknown filter parameter: %s", ann.Pos, trimmed)
						}
					}
				}
			}
			err := cg.prepareFields(t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Generate from generator interface
func (cg *CodeGenerator) Generate(bldr *Builder) (err error) {
	cg.desc = bldr.Descriptor
	cg.b = bldr
	for _, e := range bldr.File.Enums {
		err = cg.generateEnum(e)
		if err != nil {
			err = fmt.Errorf("while generating enum %s (%s): %w", e.Name, bldr.File.FileName, err)
			return
		}
	}
	for _, t := range bldr.File.Entries {
		if t.HasModifier(TypeModifierExternal) {
			continue
		}
		err = cg.generateEntity(t)
		if err != nil {
			err = fmt.Errorf("while generating entity %s (%s): %w", t.Name, bldr.File.FileName, err)
			return
		}
		err = cg.generateInitializer(t)
		if err != nil {
			err = fmt.Errorf("while generating initializer %s (%s): %w", t.Name, bldr.File.FileName, err)
			return
		}
		err = cg.generateSingleton(t)
		if err != nil {
			err = fmt.Errorf("while generating singleton for %s (%s): %w", t.Name, bldr.File.FileName, err)
			return
		}
		err = cg.generateInterface(t)
		if err != nil {
			err = fmt.Errorf("while generating interface for %s (%s): %w", t.Name, bldr.File.FileName, err)
			return
		}

	}
	if cg.desc.Features.Bool(FeatGoKind, FCGScriptingRequired) &&
		!cg.desc.Features.Bool(FeatGoKind, FCGScriptingCreated) {
		cg.desc.Features.Set(FeatGoKind, FCGScriptingCreated, true)
		bldr.Descriptor.Engine.Fields.Add(
			jen.Id(scriptingEngineField).Op("*").Qual(
				scriptingEnginePackage,
				"Service",
			),
		).Line()
		bldr.Descriptor.Engine.Initializator.Add(
			jen.Id(EngineVar).Dot(scriptingEngineField).Op("=").Id("v").Dot("GetService").Params(jen.Lit(vivard.ServiceScripting)).
				Assert(jen.Op("*").Qual(scriptingEnginePackage, "Service")),
		).Line()
		bldr.Descriptor.Engine.Initializator.Add(
			jen.Id(EngineVar).Dot(scriptingEngineField).Dot("SetContext").Params(
				jen.Map(jen.String()).Interface().Values(
					jen.Dict{
						// jen.Lit("SequenceProvider"): jen.Id(EngineVar).Dot("seqProv"),
						jen.Lit("eng"): jen.Id(EngineVar),
					},
				),
			),
		).Line()
	}
	return nil
}

// ProvideFeature from FeatureProvider interface
func (cg *CodeGenerator) ProvideFeature(
	kind FeatureKind,
	name string,
	obj interface{},
) (feature interface{}, ok ProvideFeatureResult) {
	switch kind {
	case FeaturesCommonKind:
		switch name {
		case FCModifiedFieldName:
			if f, isField := obj.(*Field); isField {
				mfname := fmt.Sprintf("%sWasModified", f.Name)
				t := f.parent
				modifiedField := &Field{
					Pos:      f.Pos,
					Name:     mfname,
					Type:     &TypeRef{Type: TipBool, NonNullable: true},
					Features: Features{},
					parent:   t,
				}
				modifiedField.Features.Set(FeaturesDBKind, FCIgnore, true)
				t.Fields = append(t.Fields, modifiedField)
				t.FieldsIndex[mfname] = modifiedField
				return mfname, FeatureProvided
			}
		case FCObjIDCode:
			if t, ok := obj.(*Entity); ok {
				return cg.getIDFromObjectFuncFeature(t), FeatureProvided
			}
		case FCSetterCode:
			if f, ok := obj.(*Field); ok {
				return cg.getFieldSetterFuncFeature(f), FeatureProvided
			}
		case FCGetterCode:
			if f, ok := obj.(*Field); ok {
				return cg.getFieldGetterFuncFeature(f), FeatureProvided
			}
			if t, ok := obj.(*Entity); ok && t.HasModifier(TypeModifierSingleton) {
				return cg.getSingletonGetFuncFeature(t), FeatureProvided
			}
		case FCIsNullCode:
			if f, ok := obj.(*Field); ok {
				return cg.getIsAttrNullFuncFeature(f), FeatureProvided
			}
		case FCSetNullCode:
			if f, ok := obj.(*Field); ok {
				return cg.getAttrSetNullFuncFeature(f), FeatureProvided
			}
		case FCAttrIsPointer:
			if f, ok := obj.(*Field); ok {
				return cg.getAttrIsPointerFeature(f), FeatureProvided
			}
		case FCAttrValueCode:
			if f, ok := obj.(*Field); ok {
				return cg.getGetAttrValueFuncFeature(f), FeatureProvided
			}
		case FCAttrSetCode:
			if f, ok := obj.(*Field); ok {
				return cg.getSetAttrValueFuncFeature(f), FeatureProvided
			}
		case FCEngineVar:
			if f, ok := obj.(*Field); ok {
				return cg.getEngineVarFuncFeature(f), FeatureProvided
			}
		}
	case FeaturesHookCodeKind:
		var value string
		var spec string
		found := false
		if name == AttrHookSet || name == AttrHookCalculate {
			if f, ok := obj.(*Field); ok {
				if hook, hok := f.HaveHook(name); hok {
					value = hook.Value
					spec = hook.Spec
					found = true
				}
			}
		} else if name == TypeHookChange || name == TypeHookChanged || name == TypeHookCreate || name == TypeHookStart || name == TypeHookDelete {
			if t, ok := obj.(*Entity); ok {
				if hook, hok := t.HaveHook(name); hok {
					value = hook.Value
					spec = hook.Spec
					found = true
				}
			}
		} else if name == TypeHookMethod {
			if m, ok := obj.(*Method); ok {
				value = m.Name
				spec = HookGoPrefix
				if ann, ok := m.Annotations[AnnotationCall]; ok {
					value = ann.GetString(AnnCallName, value)
					if js, ok := ann.GetBoolTag(AnnCallJS); ok && js {
						spec = HookJSPrefix
					}
				}
				found = true
			}
		}
		if found {
			if spec == "" || spec == HookGoPrefix {
				return cg.getGoHookFuncFeature(value), FeatureProvidedNonCacheable
			}
			if spec == HookJSPrefix {
				return cg.getJSHookFuncFeature(value), FeatureProvidedNonCacheable
			}
		}
	}
	return
}

func (cg *CodeGenerator) generateEnum(e *Enum) error {
	cg.b.Types.Add(jen.Type().Id(e.Name).Add(cg.goType(&TypeRef{Type: e.AliasForType})).Line())

	if len(e.Fields) > 0 {
		constSection := "const_" + e.Name
		withIota := e.Fields[0].FloatVal == nil && e.Fields[0].IntVal == nil && e.Fields[0].StringVal == nil
		for i, field := range e.Fields {
			expr := jen.Id(field.Name)

			if withIota {
				if i == 0 {
					expr.Id(e.Name).Op("=").Iota()
				}
			} else {
				expr.Id(e.Name)
				var val jen.Code
				if field.IntVal != nil {
					val = jen.Lit(*field.IntVal)
				} else if field.FloatVal != nil {
					val = jen.Lit(*field.FloatVal)
				} else if field.StringVal != nil {
					val = jen.Lit(*field.StringVal)
				}
				//TODO add warning
				if val != nil {
					expr.Op("=").Add(val)
				}
			}
			cg.b.AddConst(constSection, expr)
		}
	}

	return nil
}

func (cg *CodeGenerator) generateEntity(ent *Entity) error {
	fields := jen.Statement{}
	typeName := ent.Name
	if ent.HasModifier(TypeModifierSingleton) {
		fields.Add(jen.Id(EngineVar).Op("*").Id("Engine"))
	}
	if ent.BaseTypeName != "" {
		bc := ent.GetBaseType()
		f := jen.Op("*").Id(bc.FS(FeatGoKind, FCGName))
		bf := ent.GetBaseField()
		f.Tag(bf.Tags)
		fields.Add(f)

	}
	if ent.BaseTypeName != "" || ent.HasModifier(TypeModifierExtendable) && !ent.HasModifier(TypeModifierAbstract) {
		tn := ent.FS(FeatGoKind, FCGBaseTypeNameType)
		if tn == "" {
			bc := ent.GetBaseType()
			tn = bc.FS(FeatGoKind, FCGBaseTypeNameType)
		}
		cn := ent.FS(FeatGoKind, FCGDerivedTypeNameConst)
		constSection := ent.BaseTypeName + " types names"
		if ent.BaseTypeName == "" {
			constSection = ent.Name + " types names"
		}
		cg.b.AddConst(constSection, jen.Id(cn).Id(tn).Op("=").Lit(typeName))
	}
	for _, d := range ent.Fields {
		if ent.IsDictionary() &&
			// at the moment not more than 2-dimension arrays are allowed for dictionaries
			(d.Type.Array != nil && d.Type.Array.Complex && (d.Type.Array.Array == nil || d.Type.Array.Array.Complex) &&
				!d.HasModifier(AttrModifierEmbeddedRef)) && !d.HasModifier(AttrModifierEmbedded) ||
			!cg.options.AllowEmbeddedArraysForDictionary {
			return fmt.Errorf("at %v: %s: only simple types allowed for Dictionary", d.Pos, d.Name)
		}
		fieldName := d.FS(FeatGoKind, FCGName)
		otm, itsOneToMany := d.Features.GetEntity(FeaturesCommonKind, FCOneToManyType)
		mtm, itsManyToMany := d.Features.GetEntity(FeaturesCommonKind, FCManyToManyType)
		if compl, ok := d.Features.GetBool(FeaturesCommonKind, FCComplexAccessor); ok && compl {
			var paramType jen.Code
			codeForType := func(t string) jen.Code {
				parts := strings.SplitN(t, ".", 2)
				if len(parts) == 1 {
					return jen.Id(t)
				} else {
					return jen.Qual(ent.Pckg.GetFullPackage(parts[0]), parts[1])
				}
			}
			if itsOneToMany {
				paramType = jen.Index().Op("*").Add(codeForType(d.Type.Array.Type))
			} else {
				if itsManyToMany {
					paramType = jen.Index().Op("*").Add(codeForType(d.Type.Array.Type))
				} else {
					paramType, _ = cg.b.addType(&jen.Statement{}, d.Type, true) //jen.Op("*").Id(d.Type.Type)
				}
			}
			if !ent.HasModifier(TypeModifierSingleton) && !ent.HasModifier(TypeModifierConfig) {
				tn := jen.Id("obj").Op("*").Id(typeName)
				if ent.HasModifier(TypeModifierExtendable) {
					tn = jen.Id("o").Id(ent.FS(FeatGoKind, FCGBaseTypeAccessorInterface))
				}
				cg.b.Functions.Add(
					jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(
						cg.b.GetComplexMethodName(
							ent,
							d,
							CGGetComplexAttrMethod,
						),
					).
						Params(
							jen.Id("ctx").Qual("context", "Context"),
							tn,
						).Parens(jen.List(paramType, jen.Error())).BlockFunc(
						func(g *jen.Group) {
							if ent.HasModifier(TypeModifierExtendable) {
								g.Id("obj").Op(":=").Id("o").Dot(ent.FS(FeatGoKind, FCGBaseTypeAccessorName)).Params()
							}
							if itsOneToMany {
								if inc, ok := d.Features.GetBool(FeaturesDBKind, FDBIncapsulate); !ok || !inc {
									if d.HasModifier(AttrModifierEmbedded) {
										g.If(jen.Id("obj").Dot(fieldName).Op("!=").Nil()).Block(
											jen.Return(jen.Id("obj").Dot(fieldName), jen.Nil()),
										)
									}
									g.Return(
										jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodListFK, otm.Name)).Params(
											jen.Id("ctx"),
											cg.getIDFromObjectFuncFeature(ent)(),
										),
									)

								} else {
									g.Return(jen.Id("obj").Dot(fieldName), jen.Nil())
								}
							} else if itsManyToMany {
								if mtm.IsDictionary() {
									if mtm.Pckg.Name == ent.Pckg.Name {
										if code := cg.desc.CallCodeFeatureFunc(
											mtm,
											FeaturesCommonKind,
											FCListDictByIDCode,
											jen.Id("obj").Dot(fieldName),
										); code != nil {
											g.Add(code)
											return
										}
									} else {
										foreignEngine := ent.Pckg.GetExtEngineRef(mtm.Pckg.Name)
										getterName := mtm.Pckg.GetMethodName(MethodGet, mtm.Name)
										g.Id("ret").Op(":=").Make(paramType, jen.Len(jen.Id("obj").Dot(fieldName)))
										g.Var().Id("err").Error()
										g.For(jen.List(jen.Id("i"), jen.Id("v")).Op(":=").Range().Id("obj").Dot(fieldName)).Block(
											jen.List(jen.Id("ret").Index(jen.Id("i")), jen.Id("err")).Op("=").
												Id(EngineVar).Dot(foreignEngine).Dot(getterName).Params(
												jen.Id("ctx"),
												jen.Id("v"),
											),
											jen.If(jen.Id("err").Op("!=").Nil()).Block(
												jen.Return(jen.Nil(), jen.Id("err")),
											),
										)
										g.Return(jen.Id("ret"), jen.Nil())
										return
									}
								}
								if code := cg.desc.CallCodeFeatureFunc(
									mtm,
									FeaturesCommonKind,
									FCListByIDCode,
									jen.Id("obj").Dot(fieldName),
								); code != nil {
									g.Add(code)
									return
								}
								cg.desc.AddError(fmt.Errorf("at %v: can not find provider for %s", d.Pos, FCListByIDCode))
							} else if d.Type.Array != nil || d.Type.Map != nil {
								// TODO: check type
								g.Return(jen.Id("obj").Dot(fieldName), jen.Nil())
							} else if d.FB(FeatGoKind, FCGCalculated) {
								g.ReturnFunc(
									func(g *jen.Group) {
										descr := HookArgsDescriptor{
											Str: cg.desc.GetHookName(AttrHookCalculate, d),
										}
										g.Add(cg.desc.CallFeatureHookFunc(d, FeaturesHookCodeKind, AttrHookCalculate, descr))
									},
								)
							} else {
								engVar := cg.desc.CallCodeFeatureFunc(d, FeaturesCommonKind, FCEngineVar)
								g.Return(
									jen.List(
										jen.Add(engVar).Dot(cg.desc.GetMethodName(MethodGet, d.Type.Type)).Call(
											jen.Id("ctx"),
											cg.getGetAttrValueFuncFeature(d)(),
											// jen.Id("obj").Dot(d.Name),
										),
									),
								)
							}
						},
					).Line(),
				)
				//TODO maybe correct to use readonly feature here?
				if !d.FB(FeatGoKind, FCGCalculated) {
					cg.b.Functions.Add(
						jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(
							cg.b.GetComplexMethodName(
								ent,
								d,
								CGSetComplexAttrMethod,
							),
						).
							Params(
								jen.Id("ctx").Qual("context", "Context"),
								tn,
								jen.Id("val").Add(paramType),
							).Parens(jen.Error()).BlockFunc(
							func(g *jen.Group) {
								//TODO add OnHook(HookSet, HMStart,...)
								// g.Add(cg.proj.OnHook(HookSet, HMStart, d, NewHookVars("newValue", cg.getIDFromObjectFuncFeature(ref.entry)("val"))))

								if ent.HasModifier(TypeModifierExtendable) {
									g.Id("obj").Op(":=").Id("o").Dot(ent.FS(FeatGoKind, FCGBaseTypeAccessorName)).Params()
								}
								if _, hok := d.HaveHook(AttrHookSet); hok {
									g.Add(
										cg.desc.CallFeatureHookFunc(
											d, FeaturesHookCodeKind, AttrHookSet, HookArgsDescriptor{
												Str: cg.desc.GetHookName(AttrHookSet, d),
												Params: []HookArgParam{
													{"val", jen.Op("&").Id("val")},
												},
											},
										),
									)
								}
								if itsOneToMany {
									if inc, ok := d.Features.GetBool(FeaturesDBKind, FDBIncapsulate); !ok || !inc {
										g.Return(
											jen.Id(EngineVar).Dot(
												cg.desc.GetMethodName(
													MethodReplaceFK,
													otm.Name,
												),
											).Params(jen.Id("ctx"), cg.getIDFromObjectFuncFeature(ent)(), jen.Id("val")),
										)
									} else {
										g.Id("obj").Dot(fieldName).Op("=").Id("val").Line()
									}
								} else if itsManyToMany {
									//TODO: change string to id type
									if idfld, ok := d.Features.GetField(FeaturesCommonKind, FCManyToManyIDField); ok {
										g.Id("obj").Dot(fieldName).Op("=").Make(jen.Index().String(), jen.Len(jen.Id("val")))
										g.For(jen.List(jen.Id("i"), jen.Id("v")).Op(":=").Range().Id("val")).Block(
											jen.Id("obj").Dot(fieldName).Index(jen.Id("i")).Op("=").Id("v").Dot(idfld.Name),
										)
									} else {
										cg.desc.AddError(
											fmt.Errorf(
												"at %v: no %s feature found for field %s",
												d.Pos,
												FCManyToManyIDField,
												fieldName,
											),
										)
										return
									}
								} else if d.Type.Array != nil {
									tr := d.Type.Array
									if d.HasModifier(AttrModifierEmbeddedRef) {
										if t, ok := cg.desc.FindType(tr.Type); ok {
											if idt, err := t.IdType(); err == nil {
												tr = &TypeRef{Type: idt}
											}
										}
									}
									cg.createArraySetter(g, tr, "a", "val", "i")
									g.Id("obj").Dot(fieldName).Op("=").Id("a")
								} else if d.Type.Map != nil {
									cg.createMapSetter(g, d.Type.Map, "m", "val", "i")
									g.Id("obj").Dot(fieldName).Op("=").Id("m")
								} else {
									ref, ok := cg.desc.FindType(d.Type.Type)
									if !ok {
										if d.Type.Array != nil {
											if ref, ok = cg.desc.FindType(d.Type.Array.Type); !ok {
												cg.desc.AddError(fmt.Errorf("at %v: type not found", d.Pos))
												return
											}
										} else {
											cg.desc.AddError(fmt.Errorf("at %v: type not found", d.Pos))
											return
										}
									}
									if ref.entry != nil {
										g.Add(
											cg.proj.OnHook(
												HookSet,
												HMStart,
												d,
												NewHookVars("newValue", cg.getIDFromObjectFuncFeature(ref.entry)("val")),
											),
										)
										g.Add(cg.getSetAttrValueFuncFeature(d)("obj", cg.getIDFromObjectFuncFeature(ref.entry)("val")))
									} else {
										g.Add(cg.proj.OnHook(HookSet, HMStart, d, NewHookVars("newValue", "val")))
										g.Add(cg.getSetAttrValueFuncFeature(d)("obj", "val"))
									}
									//g.Id("obj").Dot(d.Name).Op("=").Add(cg.getIDFromObjectFuncFeature(ref.entry)("val"))
								}
								g.Add(cg.proj.OnHook(HookSet, HMExit, d, nil))
								g.Return(jen.Nil())
							},
						).Line(),
					)
				}
			}
		}
		if ignore, ok := d.Features.GetBool(FeaturesCommonKind, FCIgnore); ok && ignore {
			continue
		}
		if d.FB(FeatGoKind, FCGCalculated) {
			continue
		}
		t := jen.Id(fieldName).Add(d.Features.Stmt(FeatGoKind, FCGAttrType))
		if d.Tags != nil {
			t = t.Tag(d.Tags)
		}
		fields.Add(t)
		if d.HasModifier(AttrModifierAuxiliary) {
			continue
		}
		if !ent.HasModifier(TypeModifierSingleton) {
			pointer := d.FB(FeatGoKind, FCGPointer)
			// n := jen.Id(fieldName)
			setter := jen.Id("o").Dot(fieldName).Op("=")
			getter := jen.Return(jen.Id("o").Dot(fieldName))
			nullFuncs := &jen.Statement{}
			if pointer && !d.HasModifier(AttrModifierEmbedded) {
				getter = jen.If(jen.Id("o").Dot(fieldName).Op("==").Nil()).
					Block(jen.Return(cg.b.goEmptyValue(d.Type, true))).
					Else().Block(jen.Return(jen.Op("*").Id("o").Dot(fieldName)))

				// n = n.Op("*")
				setter.Add(jen.Op("&"))
			}
			if pointer || d.HasModifier(AttrModifierOneToMany) || d.HasModifier(AttrModifierEmbedded) || itsManyToMany || d.Type.Array != nil || d.Type.Map != nil {
				nullFuncs = jen.Func().Parens(jen.Id("o").Op("*").Id(typeName)).Id(
					cg.b.GetMethodName(
						d,
						CGIsNullMethod,
					),
				).Params().Bool().Block(
					jen.Return(jen.Id("o").Dot(fieldName).Op("==").Nil()),
				).Line().
					Func().Parens(jen.Id("o").Op("*").Id(typeName)).Id(cg.b.GetMethodName(d, CGSetNullMethod)).Params().Block(
					jen.Id("o").Dot(fieldName).Op("=").Nil(),
				).
					Line()
			}
			setter.Add(jen.Id("arg"))

			cg.b.Functions.Add(
				jen.Func().Parens(jen.Id("o").Op("*").Id(typeName)).Id(
					cg.b.GetMethodName(
						d,
						CGSetterMethod,
					),
				).Params(jen.Id("arg").Add(d.Features.Stmt(FeatGoKind, FCGType))).Block(
					cg.proj.OnHook(HookSet, HMStart, d, NewHookVars(GHVContext, false, GHVObject, "o", "newValue", "arg")),
					setter,
					cg.proj.OnHook(HookSet, HMExit, d, &GeneratorHookVars{Ctx: false, Obj: "o"}),
				).Line(),
				jen.Func().Parens(jen.Id("o").Op("*").Id(typeName)).Id(
					cg.b.GetMethodName(
						d,
						CGGetterMethod,
					),
				).Params().Add(d.Features.Stmt(FeatGoKind, FCGType)).Block(getter).Line(),
				nullFuncs,
			)
		}
	}
	cg.b.Types.Add(jen.Type().Id(typeName).Struct(fields...).Line())
	if ent.HasModifier(TypeModifierExtendable) && ent.BaseTypeName == "" {
		tn := ent.FS(FeatGoKind, FCGBaseTypeNameType)
		cg.b.Types.Add(jen.Type().Id(tn).String()).Line()
	}
	return nil
}

func (cg *CodeGenerator) generateInitializer(ent *Entity) (err error) {
	name := ent.Name
	fname := cg.b.Descriptor.GetMethodName(MethodInit, name)
	fields := jen.Dict{}
	idstmt := &jen.Statement{}
	initstmt := &jen.Statement{}
	initParams := []jen.Code{jen.Id("ctx").Qual("context", "Context")}
	_, isFind := ent.Features.Get(FeaturesAPIKind, FAPIFindFor)

	if ent.HasModifier(TypeModifierAbstract) {
		tt := ent.FS(FeatGoKind, FCGBaseTypeNameType)
		initParams = append(initParams, jen.Id("tip").Id(tt))
	} else if ent.HasModifier(TypeModifierExtendable) {
		tt := ent.FS(FeatGoKind, FCGBaseTypeNameType)
		if tt == "" {
			if ent.BaseTypeName != "" {
				bt := ent.GetBaseType()
				tt = bt.FS(FeatGoKind, FCGBaseTypeNameType)
			} else {
				return fmt.Errorf(
					"at %v: internal error: FCGBaseTypeNameType feature not defined for extendable type %s",
					ent.Pos,
					ent.Name,
				)
			}
		}
		initParams = append(initParams, jen.Id("tip").Op("...").Id(tt))
	}
	if ent.BaseTypeName != "" {
		bt := ent.GetBaseType()
		tn := ent.FS(FeatGoKind, FCGDerivedTypeNameConst)
		params := []jen.Code{jen.Id("ctx"), jen.Id(tn)}
		if ent.HasModifier(TypeModifierExtendable) {
			initstmt.Add(jen.Id("t").Op(":=").Id(tn)).Line()
			initstmt.Add(
				jen.If(jen.Len(jen.Id("tip")).Op(">").Lit(0)).Block(
					jen.Id("t").Op("=").Id("tip").Index(jen.Lit(0)),
				),
			).Line()
			params[1] = jen.Id("t")
		}
		initstmt.Add(
			jen.List(jen.Id("base"), jen.Id("_")).Op(":=").
				Add(cg.desc.GetTypeEngineAccessor(bt)).Dot(
				cg.b.Descriptor.GetMethodName(
					MethodInit,
					bt.Name,
				),
			).Params(params...),
		)
		fields[jen.Id(bt.Name)] = jen.Id("base")
	}
	for _, d := range ent.Fields {
		if d.HasModifier(AttrModifierAuxiliary) {
			continue
		}
		if ignore, ok := d.Features.GetBool(FeaturesCommonKind, FCIgnore); ok && ignore {
			continue
		}
		if d.FB(FeatGoKind, FCGCalculated) {
			continue
		}
		if d.Name == ExtendableTypeDescriptorFieldName && !isFind {
			if ent.HasModifier(TypeModifierAbstract) {
				fields[jen.Id(ExtendableTypeDescriptorFieldName)] = jen.String().Parens(jen.Id("tip"))
			} else {
				tn := ent.FS(FeatGoKind, FCGDerivedTypeNameConst)
				initstmt.Add(jen.Id("t").Op(":=").Id(tn)).Line()
				initstmt.Add(
					jen.If(jen.Len(jen.Id("tip")).Op(">").Lit(0)).Block(
						jen.Id("t").Op("=").Id("tip").Index(jen.Lit(0)),
					),
				)
				fields[jen.Id(ExtendableTypeDescriptorFieldName)] = jen.String().Parens(jen.Id("t"))
			}
			continue
		}
		fieldName := d.FS(FeatGoKind, FCGName)
		if d.IsIdField() && d.HasModifier(AttrModifierIDAuto) {
			idstmt.Add(
				jen.List(jen.Id("id"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(
					cg.b.Descriptor.GetMethodName(
						MethodGenerateID,
						name,
					),
				).Params(
					jen.Id("ctx"),
				).Line().Add(returnIfErrValue(jen.Nil())),
			).Line()
			fields[jen.Id(fieldName)] = jen.Id("id")
		} else if d.Type.NonNullable {
			//TODO: initialize if set default annotation
			if d.Type.Complex && d.HasModifier(AttrModifierEmbedded) && d.Type.Array == nil && d.Type.Map == nil {
				ft, ok := cg.desc.FindType(d.Type.Type)
				if !ok {
					return fmt.Errorf("at %v: type not found for embedded field: %s", d.Pos, d.Type.Type)
				}
				engVar := cg.desc.CallCodeFeatureFunc(d, FeaturesCommonKind, FCEngineVar)
				//TODO: check that it is not necessary add params to initializer
				fields[jen.Id(fieldName)] = jen.Id(d.Name)
				if ft.Entity() != nil {
					idstmt.Add(
						jen.List(jen.Id(d.Name), jen.Id("_")).Op(":=").Add(engVar).Dot(
							cg.b.Descriptor.GetMethodName(
								MethodInit,
								ft.Entity().Name,
							),
						).Params(jen.Id("ctx")),
					).Line()
				} else if ft.Enum() != nil {
					jen.Id(d.Name).Op(":=").Add(ft.Enum().GetDefaultValue(cg.desc))
				}
			} else {
				fields[jen.Id(fieldName)] = cg.b.goEmptyValue(d.Type, !d.HasModifier(AttrModifierEmbedded))
			}
		}
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(initParams...).Parens(
		jen.List(
			jen.Op("*").Id(name),
			jen.Error(),
		),
	).Block(
		idstmt,
		initstmt,
		jen.Return(
			jen.Op("&").Id(name).Values(fields),
			jen.Nil(),
		),
	).Line()

	cg.b.Functions.Add(f)
	return
}

func (cg *CodeGenerator) generateSingleton(ent *Entity) (err error) {
	if ent.HasModifier(TypeModifierSingleton) {
		name := ent.Name
		if n, ok := ent.Annotations.GetStringAnnotation(codeGeneratorAnnotation, codeGenAnnoSingletonEngineAttr); ok {
			name = n
		}
		ent.Features.Set(FeatGoKind, FCGSingletonAttrName, name)
		cg.desc.Engine.Fields.Add(jen.Id(name).Op("*").Id(ent.Name)).Line()
		cg.desc.Engine.SingletonInits[name] = jen.Id(EngineVar).Dot(name).Op("=").Op("&").Id(ent.Name).Values(
			jen.Dict{jen.Id(EngineVar): jen.Id(EngineVar)},
		).Line()
		if _, hok := ent.HaveHook(TypeHookCreate); hok {
			cg.desc.Engine.Initialized.Add(
				cg.desc.CallFeatureHookFunc(
					ent, FeaturesHookCodeKind, TypeHookCreate, HookArgsDescriptor{
						Str: cg.desc.GetHookName(TypeHookCreate, nil),
						Obj: jen.Id(EngineVar).Dot(name),
						Ctx: jen.Qual("context", "TODO").Params(),
					},
				),
			).Line()
		}
		if _, hok := ent.HaveHook(TypeHookStart); hok {
			cg.desc.Engine.Start.Add(
				cg.desc.CallFeatureHookFunc(
					ent, FeaturesHookCodeKind, TypeHookStart, HookArgsDescriptor{
						Str: cg.desc.GetHookName(TypeHookStart, nil),
						Obj: jen.Id(EngineVar).Dot(name),
						Ctx: jen.Qual("context", "TODO").Params(),
					},
				),
			).Line()
		}
	}
	return nil
}

func (cg *CodeGenerator) generateInterface(ent *Entity) (err error) {
	if ent.HasModifier(TypeModifierExtendable) {
		iname := ent.FS(FeatGoKind, FCGBaseTypeAccessorInterface)
		fname := ent.FS(FeatGoKind, FCGBaseTypeAccessorName)

		cg.b.Types.Add(jen.Type().Id(iname).Interface(jen.Id(fname).Params().Op("*").Id(ent.Name)).Line())
		cg.b.Functions.Add(
			jen.Func().Parens(jen.Id("o").Op("*").Id(ent.Name)).Id(fname).Params().Op("*").Id(ent.Name).Block(
				jen.Return(jen.Id("o")),
			).Line(),
		)
	}
	if ent.BaseTypeName != "" {
		// bt := ent.GetBaseType()
		// fname := bt.FS(FeatGoKind, FCGBaseTypeAccessorName)
		// cg.b.Functions.Add(
		// 	jen.Func().Parens(jen.Id("o").Op("*").Id(ent.Name)).Id(fname).Params().Op("*").Id(bt.Name).Block(
		// 		jen.Return(jen.Id("o").Dot(bt.Name))).Line(),
		// )
		bt := ent
		accessor := jen.Id("o") //.Dot(bt.Name)
		for bt != nil && bt.BaseTypeName != "" {
			bt = bt.GetBaseType()
			fname := bt.FS(FeatGoKind, FCGBaseTypeAccessorName)
			cg.b.Functions.Add(
				jen.Func().Parens(jen.Id("o").Op("*").Id(ent.Name)).Id(fname).Params().Op("*").Id(bt.Name).Block(
					jen.Return(jen.Add(accessor).Dot(bt.Name)),
				).Line(),
			)
			accessor = jen.Add(accessor).Dot(bt.Name)
		}
	}
	return nil
}

func (cg *CodeGenerator) createArraySetter(g *jen.Group, ref *TypeRef, goalVar string, array string, idx string) {
	g.Id(goalVar).Op(":=").Make(jen.Index().Add(cg.b.GoType(ref)), jen.Len(jen.Id(array)))
	g.For(jen.List(jen.Id(idx), jen.Id("v")).Op(":=").Range().Id(array)).BlockFunc(
		func(gg *jen.Group) {
			if ref.Array != nil {
				cg.createArraySetter(gg, ref.Array, goalVar+"a", "v", idx+"i")
				// gg.Id(goalVar).Index(jen.Id(idx)).Op("=").Id(goalVar + "a")
			} else if ref.Complex {
				if t, ok := cg.desc.FindType(ref.Type); ok && t.entry != nil {
					if idfld := t.entry.GetIdField(); idfld != nil {
						gg.Id(goalVar).Index(jen.Id(idx)).Op("=").Id("v").Dot(idfld.Name)
						return
					}
				}
			}
			gg.Id(goalVar).Index(jen.Id(idx)).Op("=").Id("v")

		},
	)
}

func (cg *CodeGenerator) createMapSetter(g *jen.Group, mapType *MapType, goalVar string, from string, idx string) {
	g.Id(goalVar).Op(":=").Make(
		jen.Map(jen.Id(mapType.KeyType)).Add(cg.b.GoType(mapType.ValueType)),
		jen.Len(jen.Id(from)),
	)
	g.For(jen.List(jen.Id(idx), jen.Id("v")).Op(":=").Range().Id(from)).BlockFunc(
		func(gg *jen.Group) {
			if mapType.ValueType.Array != nil {
				cg.createArraySetter(gg, mapType.ValueType.Array, goalVar+"a", "v", idx+"i")
			} else if mapType.ValueType.Complex {
				if t, ok := cg.desc.FindType(mapType.ValueType.Type); ok && t.entry != nil {
					if idfld := t.entry.GetIdField(); idfld != nil {
						gg.Id(goalVar).Index(jen.Id(idx)).Op("=").Id("v").Dot(idfld.Name)
						return
					}
				}
			}
			gg.Id(goalVar).Index(jen.Id(idx)).Op("=").Id("v")

		},
	)
}
func (b *Builder) addType(stmt *jen.Statement, ref *TypeRef, embedded ...bool) (f *jen.Statement, err error) {
	if ref.Array != nil {
		return b.addType(stmt.Index(), ref.Array, ref.Embedded)
	}
	if ref.Map != nil {
		return b.addType(stmt.Map(jen.Id(ref.Map.KeyType)), ref.Map.ValueType, ref.Map.ValueType.Embedded)
	}
	switch ref.Type {
	case TipString:
		f = stmt.String()
	case TipInt:
		f = stmt.Int()
	case TipBool:
		f = stmt.Bool()
	case TipDate:
		f = stmt.Qual("time", "Time")
	case TipFloat:
		f = stmt.Float64()
	case TipAny:
		f = stmt.Any()
	case TipAuto:
		err = fmt.Errorf("'auto' type can be used only with annotation, changing it")
		return
	default:
		ref.Complex = true
		if dt, ok := b.Descriptor.FindType(ref.Type); ok {
			if dt.enum != nil {
				ref.Complex = false
				if b.File.Package != dt.pckg {
					f = stmt.Qual(dt.packagePath, dt.enum.Name)
				} else {
					f = stmt.Id(dt.enum.Name)
				}
			} else if dt.entry != nil &&
				!ref.Embedded && (len(embedded) == 0 || !embedded[0]) &&
				!dt.entry.HasModifier(TypeModifierEmbeddable) &&
				!dt.entry.HasModifier(TypeModifierTransient) &&
				!dt.entry.HasModifier(TypeModifierExternal) &&
				!dt.entry.HasModifier(TypeModifierConfig) {
				it := dt.entry.GetIdField()
				if it == nil {
					err = fmt.Errorf("there is no id field for type: %s", dt.name)
					return
				}
				f, err = b.addType(stmt, it.Type)
			} else {
				f = stmt.Op("*")
				if b.File.Package != dt.pckg {
					f = f.Qual(dt.packagePath, dt.name)
				} else {
					f = f.Id(dt.name)
				}
			}
		} else {
			err = fmt.Errorf("%s: undefined type: %s", b.File.FileName, ref.Type)
		}
	}
	return
}

func (b *Builder) mustAddType(stmt *jen.Statement, ref *TypeRef) *jen.Statement {
	s, e := b.addType(stmt, ref)
	if e != nil {
		panic(e)
	}
	return s
}

// GoType returns statement with Go type for ref
func (b *Builder) GoType(ref *TypeRef, embedded ...bool) *jen.Statement {
	ret, err := b.addType(&jen.Statement{}, ref, embedded...)
	if err != nil {
		b.Descriptor.AddError(err)
		ret = &jen.Statement{}
	}
	return ret
}

func (b *Builder) checkIfEmptyValue(stmt *jen.Statement, ref *TypeRef, inverse bool) (f *jen.Statement) {
	var v interface{}
	switch ref.Type {
	case TipString:
		v = ""
	case TipInt:
		v = 0
	case TipBool:
		v = false

	case TipAny:
		f = stmt.Op("==").Nil()
		return
	case TipDate:
		f = stmt.Dot("Zero").Params()
		if inverse {
			f.Op("==").Lit(false)
		}
		return
	case TipFloat:
		v = 0.0
	}
	op := "=="
	if inverse {
		op = "!="
	}
	f = stmt.Op(op).Lit(v)
	return
}

func (b *Builder) goEmptyValue(ref *TypeRef, idForRef ...bool) (f *jen.Statement) {
	var v interface{}
	if ref.Array != nil || ref.Map != nil {
		return jen.Nil()
	}
	switch ref.Type {
	case TipString:
		v = ""
	case TipInt:
		v = 0
	case TipBool:
		v = false
	case TipDate:
		f = jen.Qual("time", "Time").Values()
		return
	case TipAny:
		f = jen.Nil()
		return
	case TipFloat:
		v = 0.0
	default:
		if dt, ok := b.Descriptor.FindType(ref.Type); ok {
			if dt.enum != nil {
				return dt.enum.GetDefaultValue(b.Pckg)
			} else if len(idForRef) > 0 && idForRef[0] && ref.Type != "" {
				if dt.entry != nil {
					it := dt.entry.GetIdField()
					if it == nil {
						if !dt.external {
							b.Descriptor.AddWarning(fmt.Sprintf("there is no id field for type: %s", dt.name))
						}
						return
					}
					return b.goEmptyValue(it.Type)
				} else {

				}
			}
		} else {
			b.Descriptor.AddWarning(fmt.Sprintf("%s: undefined type: %s (ev)", b.File.FileName, ref.Type))
		}
		return jen.Nil()
	}
	f = jen.Lit(v)
	return
}

func (b *Builder) GetMethodName(f *Field, method CSMethodKind) string {
	n := strings.ToUpper(f.Name[:1]) + f.Name[1:]
	return fmt.Sprintf(cgMethodsTemplates[method], n)
}

func (b *Builder) GetComplexMethodName(t *Entity, f *Field, method CSComplexMethodKind) string {
	fn := strings.ToUpper(f.Name[:1]) + f.Name[1:]
	tn := strings.ToUpper(t.Name[:1]) + t.Name[1:]
	return fmt.Sprintf(cgComplexMethodsTemplates[method], tn, fn)
}

func (cg *CodeGenerator) prepareFields(ent *Entity) error {
	tname := ent.Name
	if n, ok := ent.Annotations.GetStringAnnotation(codeGeneratorAnnotation, cgaNameTag); ok {
		tname = n
	}
	ent.Features.Set(FeatGoKind, FCGName, tname)
	ent.Features.Set(FeatGoKind, FCGType, jen.Id(tname))
	if ent.HasModifier(TypeModifierExtendable) {
		ent.Features.Set(FeatGoKind, FCGBaseTypeAccessorInterface, ent.Name+"er")
		ent.Features.Set(FeatGoKind, FCGBaseTypeAccessorName, "Get"+ent.Name)
	}
	for _, f := range ent.Fields {
		fname := f.Name
		if n, ok := f.Annotations.GetStringAnnotation(codeGeneratorAnnotation, cgaNameTag); ok {
			fname = n
		}
		if f.HasModifier(AttrModifierCalculated) {
			f.Features.Set(FeatGoKind, FCGCalculated, true)
		}
		f.Features.Set(FeatGoKind, FCGName, fname)
		_, itsOneToMany := f.Features.GetEntity(FeaturesCommonKind, FCOneToManyType)
		_, itsManyToMany := f.Features.GetEntity(FeaturesCommonKind, FCManyToManyType)

		isPointer := cg.proj.Options.NullsHandling == NullablePointers &&
			!f.Type.NonNullable &&
			!itsOneToMany &&
			!itsManyToMany &&
			f.Type.Array == nil &&
			f.Type.Map == nil &&
			!f.HasModifier(AttrModifierEmbedded)

		f.Features.Set(FeatGoKind, FCGPointer, isPointer)
		ftype := cg.goType(f.Type)
		f.Features.Set(FeatGoKind, FCGType, ftype)
		if isPointer {
			ftype = jen.Op("*").Add(ftype)
		}
		f.Features.Set(FeatGoKind, FCGAttrType, ftype)

		f.Tags = map[string]string{}
		if tags, ok := f.Annotations[codeGeneratorAnnotationTags]; ok {
			for _, v := range tags.Values {
				cg.desc.AddTag(f, v.Key, *v.Value.String)
			}
		}
		if f.Type.Complex && !f.Type.Embedded && (f.Type.Array == nil || f.Type.Array.Complex) || f.FB(
			FeatGoKind,
			FCGCalculated,
		) {
			f.Features.Set(FeaturesCommonKind, FCComplexAccessor, true)
		} else if _, ok := f.Features.GetEntity(FeaturesCommonKind, FCOneToManyType); ok {
			f.Features.Set(FeaturesCommonKind, FCComplexAccessor, true)
		}
		// just to set it is required
		if _, hok := f.HaveHook(AttrHookSet); hok {
			cg.desc.CallFeatureHookFunc(
				f, FeaturesHookCodeKind, AttrHookSet, HookArgsDescriptor{
					Str: cg.desc.GetHookName(AttrHookSet, f),
					Params: []HookArgParam{
						{"val", jen.Op("&").Id("val")},
					},
				},
			)
		}
	}
	return nil
}

func (cg *CodeGenerator) goType(ref *TypeRef, embedded ...bool) (f *jen.Statement) {
	if ref.Array != nil {
		emb := embedded
		if len(embedded) == 0 && ref.Embedded {
			emb = []bool{true}
		}
		return jen.Index().Add(cg.goType(ref.Array, emb...))
	}
	if ref.Map != nil {
		return jen.Map(jen.Id(ref.Map.KeyType)).Add(cg.goType(ref.Map.ValueType))
	}
	switch ref.Type {
	case TipString:
		f = jen.String()
	case TipInt:
		f = jen.Int()
	case TipBool:
		f = jen.Bool()
	case TipDate:
		f = jen.Qual("time", "Time")
	case TipFloat:
		f = jen.Float64()
	case TipAny:
		f = jen.Any()
	default:
		f = &jen.Statement{}
		if dt, ok := cg.desc.FindType(ref.Type); ok {
			if dt.enum != nil {
				if dt.pckg != cg.desc.Name {
					f = f.Qual(dt.packagePath, dt.enum.Name)
				} else {
					f = f.Id(dt.enum.Name)
				}
			} else if dt.entry != nil && dt.entry.HasModifier(TypeModifierExternal) {
				if cg.desc.fullPackage != dt.packagePath {
					f = jen.Qual(dt.packagePath, dt.name)
				} else {
					f = jen.Id(dt.name)
				}
			} else if !ref.Embedded &&
				(len(embedded) == 0 || !embedded[0]) &&
				dt.entry != nil &&
				!dt.entry.HasModifier(TypeModifierEmbeddable) &&
				!dt.entry.HasModifier(TypeModifierTransient) &&
				!dt.entry.HasModifier(TypeModifierConfig) {
				it := dt.entry.GetIdField()
				if it == nil {
					cg.desc.AddError(fmt.Errorf("there is no id field for type: %s", dt.name))
					return
				}
				idTypeName := it.Type
				if !IsPrimitiveType(idTypeName.Type) &&
					strings.Index(idTypeName.Type, ".") == -1 &&
					cg.desc.Name != it.Parent().Pckg.Name {
					// not very elegant...
					lookFor := &Field{}
					*lookFor = *it
					lookFor.Type = &TypeRef{}
					*lookFor.Type = *it.Type
					lookFor.Type.Type = it.Parent().Pckg.Name + "." + idTypeName.Type
					it = lookFor
				}
				f = cg.goType(it.Type)
			} else {
				f = jen.Op("*")
				if dt.pckg != cg.desc.Name {
					f = f.Qual(dt.packagePath, dt.name)
				} else {
					f = f.Id(dt.name)
				}
			}
		} else {
			cg.desc.AddError(fmt.Errorf("%s: undefined type: %s (gt)", cg.desc.Name, ref.Type))
		}
	}
	return
}

func (cg *CodeGenerator) createAdditionalFields(e *Entity) (err error) {
	createTimeField := func(name string, nonNullable bool) *Field {
		field := &Field{
			Pos:         e.Pos,
			Name:        name,
			Type:        &TypeRef{Type: TipDate, NonNullable: nonNullable},
			Features:    Features{},
			Annotations: Annotations{},
			parent:      e,
		}
		e.Fields = append(e.Fields, field)
		e.FieldsIndex[name] = field
		return field
	}
	createIntField := func(name string, nonNullable bool) *Field {
		field := &Field{
			Pos:         e.Pos,
			Name:        name,
			Type:        &TypeRef{Type: TipInt, NonNullable: nonNullable},
			Features:    Features{},
			Annotations: Annotations{},
			parent:      e,
		}
		e.Fields = append(e.Fields, field)
		e.FieldsIndex[name] = field
		return field
	}
	if e.FB(FeatGoKind, FCGDeletable) {
		if dfn := e.FS(FeatGoKind, FCGDeletedFieldName); dfn != "" ||
			cg.options.GenerateDeletedField {
			if dfn == "" {
				dfn = deletedFieldName
			}
			deletedField := createTimeField(dfn, false)
			deletedField.Features.Set(FeaturesDBKind, FCGName, mdDeletedFieldName)
			deletedField.Features.Set(FeaturesCommonKind, FCGDeletedField, true)
		}
	}
	isDictionary := e.HasModifier(TypeModifierDictionary)
	generateAccessFields := cg.options.GenerateAccessField == EntityTypeAll ||
		cg.options.GenerateAccessField == EntityTypeNonDictionary && !isDictionary

	generateCreated, set := e.Features.GetBool(FeatGoKind, FCGLogCreated)
	if !set {
		generateCreated = generateAccessFields
	}
	if generateCreated {
		e.Features.Set(FeatGoKind, FCGLogCreated, true)
		cf := createTimeField(createdFieldName, true)
		e.Features.Set(FeatGoKind, FCGLogCreatedField, cf)
		by, ok := e.Features.GetBool(FeatGoKind, FCGLogCreatedBy)
		if !ok {
			by = cg.options.GenerateCreatedBy
		}
		if by {
			cf := createIntField(createdByFieldName, true)
			e.Features.Set(FeatGoKind, FCGLogCreatedByField, cf)
		}
	}

	generateModified, set := e.Features.GetBool(FeatGoKind, FCGLogModified)
	if !set {
		generateModified = generateAccessFields
	}
	if generateModified {
		e.Features.Set(FeatGoKind, FCGLogModified, true)
		mf := createTimeField(modifiedFieldName, true)
		e.Features.Set(FeatGoKind, FCGLogModifiedField, mf)
		by, ok := e.Features.GetBool(FeatGoKind, FCGLogModifiedBy)
		if !ok {
			by = cg.options.GenerateModifiedBy
		}
		if by {
			cf := createIntField(modifiedByFieldName, true)
			e.Features.Set(FeatGoKind, FCGLogModifiedByField, cf)
		}
	}
	return nil
}

func (cg *CodeGenerator) ProvideCodeFragment(
	module interface{},
	action interface{},
	point interface{},
	ctx interface{},
) interface{} {
	if module == CodeFragmentModuleGeneral {
		if cf, ok := ctx.(*CodeFragmentContext); ok {
			if point == CFGPointEnterAfterHooks && cf.Entity != nil {
				if action == MethodNew && (cf.Entity.FB(FeatGoKind, FCGLogCreated) || cf.Entity.FB(
					FeatGoKind,
					FCGLogModified,
				)) ||
					action == MethodSet && cf.Entity.FB(FeatGoKind, FCGLogModified) {
					if mf, ok := cf.Entity.Features.GetField(FeatGoKind, FCGLogModifiedField); ok {
						cf.Add(
							cf.GetParam(ParamObject).Dot(cg.b.GetMethodName(mf, CGSetterMethod)).Params(
								jen.Qual(
									"time",
									"Now",
								).Params(),
							),
						)
					}
					if action == MethodNew {
						if mf, ok := cf.Entity.Features.GetField(FeatGoKind, FCGLogCreatedField); ok {
							cf.Add(
								cf.GetParam(ParamObject).Dot(cg.b.GetMethodName(mf, CGSetterMethod)).Params(
									jen.Qual(
										"time",
										"Now",
									).Params(),
								),
							)
						}
					}
				}
			}
		}
	}
	return nil
}
