package gen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

const (
	// AnnQual - may be used for dictionary fields to show that it is depends on another dictionary
	// Field ma be ref to another dict or array to it; may be Nullable or no Null
	AnnQual = "qualifier"
	//AnnQualBy - annotation for field of type ref to dictionary, should contain id of field that is used as qualifier for qualified dictionairy
	AnnQualBy = "qualified-by"
)

const (
	//FeatureDictKind is used for dictionary features
	FeatureDictKind = "dict"

	//FDQualified - bool for Etnity - true if the dictionary has qualifier field
	FDQualified = "qualified"

	//FDQualifier - bool for Field or *Field for Entity
	//  for Field set if it is qualifier for this dict Entity
	//  for Entity refs to qualifier Field
	FDQualifier = "qualifier"

	//FDQualifierType - *Entity for Entity, refs to dictionary type that qualifier field refs to
	FDQualifierType = "qualifier-type"

	//FDQualifiedBy - *Field for Field, refs the field in the same Entity that is qualifier for this dict field
	FDQualifiedBy = "qual-by"

	//FDQualifierFor - *Field for Field, refs to field requires this as qualifier
	FDQualifierFor = "qualifier-for"

	//FDIdxFieldName - string; field name in Engine for cache
	FDCacheFieldName = "cache-field"

	//FDIdxFieldName - string; field name in Engine for index
	FDIdxFieldName = "idx-field"

	//FDQualIdxName - string; qualifier index field name in Engine (if any)
	FDQualIdxName = "qual-idx-name"
)
const (
	DictCacheTempl          = "%sCache"
	DictIndexCacheTempl     = "%sCacheIndex"
	DictQualIndexCacheTempl = "%s%sCacheIndex"
	DictCacheLoaderTempl    = "%sEnsureCache"
	DictCacheMutexSuffix    = "Lock"
)

type DictionariesGenerator struct {
	proj *Project
	desc *Package
	b    *Builder
}

//SetDescriptor from DescriptorAware
func (ncg *DictionariesGenerator) SetDescriptor(proj *Project) {
	ncg.proj = proj
}

//ProvideFeature from FeatureProvider interface
func (ncg *DictionariesGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	switch kind {
	case FeaturesCommonKind:
		switch name {
		case FCListDictByIDCode:
			if e, isEntity := obj.(*Entity); isEntity {
				return ncg.getListDictByIDCode(e), FeatureProvided
			}
		}
	}
	return nil, FeatureNotProvided
}

func (ncg *DictionariesGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	if fld, ok := item.(*Field); ok {
		switch ann.Name {
		case AnnQual:
			if fld.Parent().HasModifier(TypeModifierDictionary) {
				if fld.Parent().FB(FeatureDictKind, FDQualified) {
					return true, fmt.Errorf("at %v: only one qualifier is allowed for dictionary", fld.Pos)
				}
				qt := fld.Type.Type
				if fld.Type.Array != nil {
					qt = fld.Type.Array.Type
				}
				qr, ok := desc.FindType(qt)
				if !ok {
					return true, fmt.Errorf("at %v: not found type %s for qualifier", fld.Pos, qt)
				}
				if !qr.Entity().HasModifier(TypeModifierDictionary) {
					return true, fmt.Errorf("at %v: only dictionary can be used for qualifier", fld.Pos)
				}
				fld.Parent().Features.Set(FeatureDictKind, FDQualifierType, qr.Entity())
				fld.Parent().Features.Set(FeatureDictKind, FDQualified, true)
				fld.Parent().Features.Set(FeatureDictKind, FDQualifier, fld)
				fld.Features.Set(FeatureDictKind, FDQualifier, true)
				return true, nil
			}
		case AnnQualBy:
			if t, ok := desc.FindType(fld.Type.Type); ok {
				dict := t.Entity()
				if dict.HasModifier(TypeModifierDictionary) {
					if ann.Values == nil {
						return true, fmt.Errorf("at %v: qualified-by should reference to field", fld.Pos)
					}
					name := ann.Values[0].Key
					qb := fld.Parent().GetField(name)
					if qb == nil {
						return true, fmt.Errorf("at %v: qualified-by should reference to existing field: %s not found in type %s", fld.Pos, name, dict.Name)
					}
					fld.Features.Set(FeatureDictKind, FDQualifiedBy, qb)
					qb.Features.Set(FeatureDictKind, FDQualifierFor, fld)
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (ncg *DictionariesGenerator) Prepare(desc *Package) error {
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if t.HasModifier(TypeModifierDictionary) {
				name := t.Name
				fname := fmt.Sprintf(DictCacheTempl, name)
				idxname := fmt.Sprintf(DictIndexCacheTempl, name)
				t.Features.Set(FeatureDictKind, FDCacheFieldName, fname)
				t.Features.Set(FeatureDictKind, FDIdxFieldName, idxname)
				if t.FB(FeatureDictKind, FDQualified) {
					qf, _ := t.Features.GetField(FeatureDictKind, FDQualifier)
					idxname = fmt.Sprintf(DictQualIndexCacheTempl, name, qf.Name)
					t.Features.Set(FeatureDictKind, FDQualIdxName, idxname)
				}
			}
			// let's check qualified-by annotations
			for _, f := range t.Fields {
				if qb, ok := f.Features.GetField(FeatureDictKind, FDQualifiedBy); ok {
					dt, _ := desc.FindType(f.Type.Type)
					dict := dt.Entity()
					qual, ok := dict.Features.GetField(FeatureDictKind, FDQualifier)
					if !ok {
						return fmt.Errorf("at %v: qualified-by for not qualified dictionary", f.Pos)
					}
					dqt := qual.Type.Type
					if qual.Type.Array != nil {
						dqt = qual.Type.Array.Type
					}
					if qb.Type.Type != dqt {
						return fmt.Errorf("at %v: qualified-by should reference to type '%s' but refs to '%s'", f.Pos, dqt, qb.Type.Type)
					}
				}
			}
		}
	}
	return nil
}

func (ncg *DictionariesGenerator) Generate(b *Builder) (err error) {
	ncg.b = b
	ncg.desc = b.Descriptor
	for _, t := range b.File.Entries {
		if t.IsDictionary() {
			f := t.GetIdField()
			ncg.addDictionaryCache(t, f.Type)
			ncg.generateDictGetter(t, f.Type)
			ncg.generateAllDictGetter(t, f.Type)
			ncg.generateDictCacheLoader(t, f)
			if !t.FB(FeaturesCommonKind, FCReadonly) {
				ncg.generateDictSetter(t, f)
				ncg.generateDictNew(t, f)
				ncg.generateDictDelete(t, f)
			}
		} else if t.HasModifier(TypeModifierConfig) {
			ncg.addConfigAttr(t.Name)
			ncg.generateConfigSetter(t.Name)
			ncg.generateConfigGetter(t.Name)
			ncg.generateConfigProvider(t)
		} else if ann := t.Annotations[AnnotationConfig]; ann != nil {
			ncg.generateConfigProvider(t)
		}
	}
	return nil
}

func (ncg *DictionariesGenerator) addDictionaryCache(t *Entity, idType *TypeRef) error {
	idt, err := ncg.b.addType(&jen.Statement{}, idType)
	if err != nil {
		return err
	}
	name := t.Name
	fname := t.FS(FeatureDictKind, FDCacheFieldName)
	idxname := t.FS(FeatureDictKind, FDIdxFieldName)
	ncg.desc.Engine.Fields.Add(jen.Id(idxname).Map(idt).Int().Line())
	ncg.desc.Engine.Fields.Add(jen.Id(fname).Index().Op("*").Id(name).Line())
	if t.FB(FeatureDictKind, FDQualified) {
		qf, _ := t.Features.GetField(FeatureDictKind, FDQualifier)
		idxname = t.FS(FeatureDictKind, FDQualIdxName)
		var keyType *jen.Statement

		qt, _ := t.Features.GetEntity(FeatureDictKind, FDQualifierType)
		idfld := qt.GetIdField()
		switch idfld.Type.Type {
		case TipInt:
			keyType = jen.Int()
		case TipString:
			keyType = jen.String()
		default:
			return fmt.Errorf("at %s: only dicts with id field of type int and string may be used as qualifier", qf.Pos)
		}
		ncg.desc.Engine.Fields.Add(jen.Id(idxname).Map(keyType).Index().Int().Line())
	}
	ncg.desc.Engine.Fields.Add(jen.Id(fname+DictCacheMutexSuffix).Qual("sync", "RWMutex").Line())
	// b.descriptor.Engine.Initializator.Id(EngineVar).Dot(fname).Op("=").Map(idt).Op("*").Id(name).Values().Line()
	return nil
}

func (ncg *DictionariesGenerator) generateDictGetter(t *Entity, idType *TypeRef) error {
	name := t.Name
	fname := ncg.desc.GetMethodName(MethodGet, name)
	params, err := ncg.b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), idType)
	if err != nil {
		return err
	}
	fldname := t.FS(FeatureDictKind, FDCacheFieldName)
	idxname := t.FS(FeatureDictKind, FDIdxFieldName)
	loadername := fmt.Sprintf(DictCacheLoaderTempl, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(params).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.Id(EngineVar).Dot(loadername).Params(jen.Id("ctx")).Line().
			Add(ncg.generateLockUnlockStmt(name, true)).
			Return(
				// TODO: check that item exists
				jen.List(jen.Id(EngineVar).Dot(fldname).Index(jen.Id(EngineVar).Dot(idxname).Index(jen.Id("id"))), jen.Nil()),
			),
	).Line()

	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateDictSetter(t *Entity, idField *Field) error {
	name := t.Name
	fname := ncg.desc.GetMethodName(MethodSet, name)
	fldname := t.FS(FeatureDictKind, FDCacheFieldName)
	idxname := t.FS(FeatureDictKind, FDIdxFieldName)
	loadername := fmt.Sprintf(DictCacheLoaderTempl, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)).
		Parens(jen.List(jen.Id("ret").Op("*").Id(name), jen.Id("err").Error())).Block(
		jen.Id(EngineVar).Dot(loadername).Params(jen.Id("ctx")).Line().
			Add(ncg.generateLockUnlockStmt(name, false)).
			If(
				jen.List(jen.Id("idx"), jen.Id("ok")).Op(":=").Id(EngineVar).Dot(idxname).Index(jen.Id("o").Dot(idField.Name)),
				jen.Id("ok"),
			).
			BlockFunc(func(g *jen.Group) {
				//TODO rebuild qualified index if required
				g.Id(EngineVar).Dot(fldname).Index(jen.Id("idx")).Op("=").Id("o")
				g.Return(jen.Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodSave, name)).Params(jen.List(jen.Id("ctx"), jen.Id("o"))))
			}).
			Else().Block(
			jen.Return(
				jen.Nil(),
				jen.Qual("fmt", "Errorf").Params(jen.Lit(fmt.Sprintf("no %s found with id %%v", name)), jen.Id("o").Dot(idField.Name)),
			),
		),
	).Line()

	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateDictDelete(t *Entity, idField *Field) error {
	name := t.Name
	fname := ncg.desc.GetMethodName(MethodDelete, name)
	fldname := t.FS(FeatureDictKind, FDCacheFieldName)
	idxname := t.FS(FeatureDictKind, FDIdxFieldName)
	loadername := fmt.Sprintf(DictCacheLoaderTempl, name)
	params, err := ncg.b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), idField.Type)
	if err != nil {
		return err
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(params).Parens(jen.Error()).Block(
		jen.Id(EngineVar).Dot(loadername).Params(jen.Id("ctx")),
		jen.Add(ncg.generateLockUnlockStmt(name, true)),
		jen.If(jen.List(jen.Id( /*"idx"*/ "_"), jen.Id("ok")).Op(":=").Id(EngineVar).Dot(idxname).Index(jen.Id("id")), jen.Id("ok")).Block(
			jen.Id("err").Op(":=").Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodRemove, name)).Params(jen.List(jen.Id("ctx"), jen.Id("id"))),
			jen.Add(returnIfErrValue()),
			//TODO remove from cache
			//jen.Delete(jen.Id(EngineVar).Dot(idxname), jen.Id("id")),
			//jen.Id(EngineVar).Dot(fldname).Op("=").Id(EngineVar).Dot(fldname).Index(jen.Id("idx")),
			jen.Id(EngineVar).Dot(fldname).Op("=").Nil(),
			jen.Id(EngineVar).Dot(idxname).Op("=").Nil(),
			jen.Return(jen.Nil()),
		),
		jen.Return(
			// TODO: return error?
			jen.Nil(),
		),
	).Line()

	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateDictNew(t *Entity, idField *Field) error {
	name := t.Name
	fname := ncg.desc.GetMethodName(MethodNew, name)
	idxname := t.FS(FeatureDictKind, FDIdxFieldName)
	fldname := t.FS(FeatureDictKind, FDCacheFieldName)
	loadername := fmt.Sprintf(DictCacheLoaderTempl, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)).
		Parens(jen.List(jen.Id("ret").Op("*").Id(name), jen.Id("err").Error())).BlockFunc(func(c *jen.Group) {

		c.Id(EngineVar).Dot(loadername).Params(jen.Id("ctx"))
		c.Add(ncg.generateLockUnlockStmt(name, false))

		if idField.HasModifier(AttrModifierIDAuto) {
			c.If(
				ncg.b.checkIfEmptyValue(jen.Id("o").Dot(idField.Name), idField.Type, true),
			).Block(
				jen.Id("err").Op("=").Qual("errors", "New").Params(jen.Lit("dict: not empty id for New")),
				jen.Return(),
			)
			c.List(jen.Id("id"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodGenerateID, name)).Params(jen.Id("ctx"))
			c.Add(returnIfErr())
			c.Id("o").Dot(idField.Name).Op("=").Id("id")
		} else {
			c.If(
				ncg.b.checkIfEmptyValue(jen.Id("o").Dot(idField.Name), idField.Type, false),
			).Block(
				jen.Id("err").Op("=").Qual("errors", "New").Params(jen.Lit("id should not be empty for New")),
				jen.Return(),
			)
			c.If(
				jen.List(jen.Id("_"), jen.Id("ok")).Op(":=").Id(EngineVar).Dot(idxname).Index(jen.Id("o").Dot(idField.Name)),
				jen.Id("ok"),
			).Block(
				jen.Id("err").Op("=").Qual("errors", "New").Params(jen.Lit("duplicate id found")),
				jen.Return(),
			)
		}
		c.List(jen.Id("ret"), jen.Id("err")).Op("=").Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodCreate, name)).Params(jen.List(jen.Id("ctx"), jen.Id("o"))).Line()
		c.Add(returnIfErr())
		c.Id(EngineVar).Dot(fldname).Op("=").Append(jen.Id(EngineVar).Dot(fldname), jen.Id("ret"))
		c.Id(EngineVar).Dot(idxname).Index(jen.Id("o").Dot(idField.Name)).Op("=").Len(jen.Id(EngineVar).Dot(fldname)).Op("-").Lit(1)
		qualIdxName := t.FS(FeatureDictKind, FDQualIdxName)
		if qualIdxName != "" {
			qf, _ := t.Features.GetField(FeatureDictKind, FDQualifier)
			derefOp := ""
			if !qf.Type.NonNullable {
				derefOp = "*"
			}
			c.Id(EngineVar).Dot(qualIdxName).Index(jen.Op(derefOp).Id("o").Dot(qf.Name)).Op("=").Append(
				jen.Id(EngineVar).Dot(qualIdxName).Index(jen.Op(derefOp).Id("o").Dot(qf.Name)),
				jen.Len(jen.Id(EngineVar).Dot(fldname)).Op("-").Lit(1),
			)
		}
		c.Return()
	},
	).Line()

	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateAllDictGetter(t *Entity, idType *TypeRef) error {
	name := t.Name
	fname := ncg.desc.GetMethodName(MethodGetAll, name)
	fldname := t.FS(FeatureDictKind, FDCacheFieldName)
	qualIdxName := t.FS(FeatureDictKind, FDQualIdxName)
	loadername := fmt.Sprintf(DictCacheLoaderTempl, name)
	var qualIDFld *Field
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).ParamsFunc(func(g *jen.Group) {
		g.Id("ctx").Qual("context", "Context")
		if qualIdxName != "" {
			qt, _ := t.Features.GetEntity(FeatureDictKind, FDQualifierType)
			qualIDFld = qt.GetIdField()
			switch qualIDFld.Type.Type {
			case TipInt:
				g.Id("qual").Op("...").Int()
			case TipString:
				g.Id("qual").Op("...").String()
			}
		}
	}).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).BlockFunc(func(g *jen.Group) {
		g.Id(EngineVar).Dot(loadername).Params(jen.Id("ctx"))
		g.Add(ncg.generateLockUnlockStmt(name, true))
		if qualIdxName != "" {
			g.If(jen.Len(jen.Id("qual")).Op(">").Lit(0)).Block(
				jen.Id("ret").Op(":=").Make(jen.Index().Op("*").Id(name), jen.Lit(0), jen.Len(jen.Id(EngineVar).Dot(qualIdxName).Index(jen.Id("qual").Index(jen.Lit(0))))),
				jen.For(jen.List(jen.Id("_"), jen.Id("q")).Op(":=").Range().Id("qual")).Block(
					jen.For(jen.List(jen.Id("_"), jen.Id("idx")).Op(":=").Range().Id(EngineVar).Dot(qualIdxName).Index(jen.Id("q"))).Block(
						jen.Id("ret").Op("=").Append(jen.Id("ret"), jen.Id(EngineVar).Dot(fldname).Index(jen.Id("idx"))),
					),
				),
				jen.Return(jen.Id("ret"), jen.Nil()),
			)
		}
		g.Return(
			// TODO: check that item exists
			jen.List(jen.Id(EngineVar).Dot(fldname), jen.Nil()),
		)
	}).Line()

	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateDictCacheLoader(t *Entity, idField *Field) error {
	name := t.Name
	fname := fmt.Sprintf(DictCacheLoaderTempl, name)
	fldname := t.FS(FeatureDictKind, FDCacheFieldName)
	qualIdxName := t.FS(FeatureDictKind, FDQualIdxName)
	idxname := t.FS(FeatureDictKind, FDIdxFieldName)
	idt, _ := ncg.b.addType(&jen.Statement{}, idField.Type)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context")).Error().Block(
		jen.If(jen.Id(EngineVar).Dot(fldname).Op("==").Nil()).Block(
			ncg.generateLockUnlockStmt(name, false).
				If(jen.Id(EngineVar).Dot(fldname).Op("==").Nil()).BlockFunc(func(g *jen.Group) {
				var itemsCode jen.Code
				if c, ok := t.Features.Get(FeatGoKind, FCDictGetter); ok {
					if code, ok := c.(jen.Code); ok {
						// there is getter code
						itemsCode = jen.Id("items").Op(":=").Add(code)
					}
				}
				if itemsCode == nil {
					// loading the cache
					g.List(jen.Id("items"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodList, name)).Params(jen.Id("ctx"))
					g.Add(returnIfErrValue())
				} else {
					g.Add(itemsCode)
				}
				g.Id(EngineVar).Dot(fldname).Op("=").Id("items")
				g.Id(EngineVar).Dot(idxname).Op("=").Map(idt).Int().Values()
				if qualIdxName != "" {
					var keyType *jen.Statement
					qt, _ := t.Features.GetEntity(FeatureDictKind, FDQualifierType)
					idfld := qt.GetIdField()
					switch idfld.Type.Type {
					case TipInt:
						keyType = jen.Int()
					case TipString:
						keyType = jen.String()
					default:
						ncg.desc.AddError(fmt.Errorf("at %s: only dicts with id field of type int and string may be used as qualifier", qt.Pos))
						return
					}
					g.Id(EngineVar).Dot(qualIdxName).Op("=").Map(keyType).Index().Int().Values()
				}
				g.For(jen.List(jen.Id("idx"), jen.Id("val")).Op(":=").Range().Id("items")).BlockFunc(func(g *jen.Group) {
					g.Id(EngineVar).Dot(idxname).Index(jen.Id("val").Dot(idField.Name)).Op("=").Id("idx")
					if qualIdxName != "" {
						qf, _ := t.Features.GetField(FeatureDictKind, FDQualifier)
						derefOp := ""
						if !qf.Type.NonNullable {
							derefOp = "*"
						}
						g.Id(EngineVar).Dot(qualIdxName).Index(jen.Op(derefOp).Id("val").Dot(qf.Name)).Op("=").Append(
							jen.Id(EngineVar).Dot(qualIdxName).Index(jen.Op(derefOp).Id("val").Dot(qf.Name)),
							jen.Id("idx"),
						)
					}
				})
			}),
		),
		jen.Return(jen.Nil()),
	).Line()
	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateLockUnlockStmt(name string, read bool) *jen.Statement {
	fldname := fmt.Sprintf(DictCacheTempl, name)
	lockname := fldname + DictCacheMutexSuffix
	lock := "RLock"
	unlock := "RUnlock"
	if !read {
		lock = "Lock"
		unlock = "Unlock"
	}
	return jen.Id(EngineVar).Dot(lockname).Dot(lock).Params().Line().
		Defer().Id(EngineVar).Dot(lockname).Dot(unlock).Params().Line()
}

func (ncg *DictionariesGenerator) getListDictByIDCode(e *Entity) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: ncg.desc}
		a.init("ids", "ctx").parse(args)

		name := e.Name
		fldname := e.FS(FeatureDictKind, FDCacheFieldName)
		idxname := e.FS(FeatureDictKind, FDIdxFieldName)
		loadername := fmt.Sprintf(DictCacheLoaderTempl, name)
		s := jen.Id(EngineVar).Dot(loadername).Params(jen.Id("ctx")).Line().
			Add(ncg.generateLockUnlockStmt(name, true)).
			Id("ret").Op(":=").Make(jen.Index().Op("*").Id(name), jen.Len(a.get("ids"))).Line().
			For(jen.List(jen.Id("i"), jen.Id("v")).Op(":=").Range().Add(a.get("ids"))).Block(
			jen.If(
				jen.List(jen.Id("idx"), jen.Id("ok")).Op(":=").Id(EngineVar).Dot(idxname).Index(jen.Id("v")),
				jen.Id("ok"),
			).Block(
				jen.Id("ret").Index(jen.Id("i")).Op("=").Id(EngineVar).Dot(fldname).Index(jen.Id("idx")),
			).Else().Block(
				jen.Return(
					jen.Nil(),
					jen.Qual("fmt", "Errorf").Params(jen.Lit("item not found: %v"), jen.Id("v")),
				),
			).Line(),
		).Line().
			Return(
				jen.List(jen.Id("ret"), jen.Nil()),
			).Line()
		return s
	}
}

func (ncg *DictionariesGenerator) addConfigAttr(name string) error {
	ncg.desc.Engine.Fields.Add(jen.Id(name).Op("*").Id(name).Line())
	return nil
}

func (ncg *DictionariesGenerator) generateConfigGetter(name string) error {
	fname := ncg.desc.GetMethodName(MethodGet, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context")).
		Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.If(jen.Id(EngineVar).Dot(name).Op("==").Nil()).Block(
			jen.List(jen.Id("res"), jen.Err()).Op(":=").Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodLoad, name)).Call(jen.Id("ctx")),
			returnIfErrValue(jen.Nil()),
			jen.Id(EngineVar).Dot(name).Op("=").Id("res"),
		),
		jen.Return(
			// TODO: check that item exists
			jen.List(jen.Id(EngineVar).Dot(name), jen.Nil()),
		),
	).Line()

	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateConfigSetter(name string) error {
	fname := ncg.desc.GetMethodName(MethodSet, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)).
		Parens(jen.List(jen.Id("ret").Op("*").Id(name), jen.Id("err").Error())).Block(
		jen.Id(EngineVar).Dot(name).Op("=").Id("o"),
		jen.List(jen.Id("ret"), jen.Err()).Op("=").Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodSave, name)).Call(jen.Id("ctx"), jen.Id("o")),
		jen.Return(),
	).Line()

	ncg.b.Functions.Add(f)
	return nil
}

func (ncg *DictionariesGenerator) generateConfigProvider(e *Entity) error {
	f := jen.Func().Parens(jen.Id("o").Op("*").Id(e.Name)).Id("GetConfigValue").Params(jen.Id("key").String()).Interface().BlockFunc(func(g *jen.Group) {
		g.Id("parts").Op(":=").Qual("strings", "SplitN").Params(jen.Id("key"), jen.Lit("."), jen.Lit(2))
		g.Switch(jen.Id("parts").Index(jen.Lit(0))).BlockFunc(func(g *jen.Group) {
			ca := e.Annotations[AnnotationConfig]
			var name string
			var ft *Entity
			for _, f := range e.Fields {
				isArray := false
				if ca == nil || !ca.GetBool(AnnCfgValue, false) {
					name = f.Type.Type
					if f.Type.Array != nil {
						isArray = true
						name = f.Type.Array.Type
					}
					if name == "" {
						continue
					}
					if dt, ok := ncg.desc.FindType(name); ok {
						ft = dt.Entity()
						ca := ft.Annotations[AnnotationConfig]
						if ca == nil || (!ca.GetBool(AnnCfgGroup, false) && !ca.GetBool(AnnCfgValue, false)) {
							continue
						}
					} else {
						continue
					}
				}
				fname := f.Name
				g.Case(jen.Lit(fname)).BlockFunc(func(g *jen.Group) {
					//TODO return array elements?
					if ca != nil && ca.GetBool(AnnCfgValue, false) || isArray {
						g.Return(jen.Id("o").Dot(f.Name))
					} else {
						g.If(jen.Len(jen.Id("parts")).Op("==").Lit(1)).Block(
							jen.Return(jen.Id("o").Dot(f.Name)),
						)
						g.Return(jen.Id("o").Dot(f.Name).Dot("GetConfigValue").Params(jen.Id("parts").Index(jen.Lit(1))))
					}
				})
			}
		})
		g.Return(jen.Nil())
	})

	ncg.b.Functions.Add(f.Line())

	if e.HasModifier(TypeModifierConfig) {
		ncg.desc.Engine.Initialized.Add(
			jen.Id(EngineVar).Dot(ncg.desc.GetMethodName(MethodGet, e.Name)).Params(jen.Qual("context", "TODO").Params()).Line(),
			jen.If(jen.Id(EngineVar).Dot(e.Name).Op("!=").Nil()).Block(
				jen.Id("v").Dot("RegisterConfigProvider").Params(jen.Id(EngineVar).Dot(e.Name), jen.Lit(10)),
			).Line(),
		)
	}
	return nil
}
