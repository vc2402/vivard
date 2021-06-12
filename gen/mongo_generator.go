package gen

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard/mongo"
)

const (
	mongoAnnotation                       = "mongo"
	dbAnnotation                          = "db"
	mongoAnnotationTagName                = "name"
	mongoAnnotationTagIgnore              = "ignore"
	mongoAnnotationTagGenerateIDGenerator = "idGenerator"
	mongoAnnotationTagIncapsulate         = "incapsulate"
	mongoAnnotationTagAddBsonTag          = "bsonTag"

	mongoGoTagBSON = "bson"
)

const (
	bsonPackage    = "go.mongodb.org/mongo-driver/bson"
	mongoPackage   = "go.mongodb.org/mongo-driver/mongo"
	optionsPackage = "go.mongodb.org/mongo-driver/mongo/options"
	engineMongo    = "Mongo"

	optionsMongo                      = "mongo"
	optionGenerateIDGenerator         = "idGenerator"
	optionPrefixCollectionName        = "prefixCollectionName"
	optionUseBaseCollectionForDerived = "baseCollectionForDerived"

	optPrefixWithPackage = "<package>"
)

const (
	mongoFeatures FeatureKind = "mongo"
	mfInited                  = "inited"
)

const (
	mdDeletedFieldName = "deleted"
)

type MongoGenerator struct {
	desc                        *Package
	b                           *Builder
	collections                 map[string]string
	usedCollections             map[string]bool
	generateIDGenerators        bool
	prefixCollectionName        string
	useBaseCollectionForDerived bool
}

func (cg *MongoGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	_, fld := item.(*Field)
	_, ent := item.(*Entity)
	cg.init()
	if ann.Name == mongoAnnotation || ann.Name == dbAnnotation {
		if !fld && !ent {
			return true, fmt.Errorf("at %v: mongo annotation may be used only with type and field", ann.Pos)
		}
		for _, v := range ann.Values {
			switch v.Key {
			case mongoAnnotationTagName:
			case mongoAnnotationTagIgnore:
			case mongoAnnotationTagIncapsulate:
			case mongoAnnotationTagAddBsonTag:
			default:
				return true, fmt.Errorf("at %v: unknown mongo annotation parameter: %s", ann.Pos, v.Key)
			}
		}
		return true, nil
	}
	return false, nil
}

func (cg *MongoGenerator) Prepare(desc *Package) error {
	cg.init()
	cg.desc = desc
	if opts, ok := desc.Options().Custom[optionsMongo].(map[string]interface{}); ok {
		if gengen, ok := opts[optionGenerateIDGenerator]; ok {
			switch v := gengen.(type) {
			case bool:
				cg.generateIDGenerators = v
			case string:
				v = strings.ToLower(v)
				cg.generateIDGenerators = v == "true" || v == "on"
			}
		}
		if pref, ok := opts[optionPrefixCollectionName]; ok {
			switch v := pref.(type) {
			case bool:
				if !v {
					cg.prefixCollectionName = ""
				}
			case string:
				cg.prefixCollectionName = v
			}
		}
		if bc, ok := opts[optionUseBaseCollectionForDerived]; ok {
			switch v := bc.(type) {
			case bool:
				cg.useBaseCollectionForDerived = v
			case string:
				v = strings.ToLower(v)
				cg.useBaseCollectionForDerived = v == "true" || v == "on"
			}
		}
	}

	desc.Engine.Fields.Add(jen.Id(engineMongo).Op("*").Qual(mongoPackage, "Database")).Line()
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if t.HasModifier(TypeModifierTransient) || //t.HasModifier(TypeModifierEmbeddable) ||
				t.HasModifier(TypeModifierSingleton) || t.HasModifier(TypeModifierExternal) {
				t.Features.Set(FeaturesDBKind, FCIgnore, true)
				if !t.Annotations.GetBoolAnnotationDef(mongoAnnotation, mongoAnnotationTagAddBsonTag, false) {
					continue
				}
			}
			if t.HasModifier(TypeModifierEmbeddable) {
				// generate bson tag for embeddable
				t.Features.Set(FeaturesDBKind, FCIgnore, true)
			}
			if t.BaseTypeName != "" {
				desc.AddBaseFieldTag(t, mongoGoTagBSON, ",inline")
			}
			if t.HasModifier(TypeModifierConfig) &&
				(t.Annotations.GetBoolAnnotationDef(AnnotationConfig, AnnCfgValue, false) || t.Annotations.GetBoolAnnotationDef(AnnotationConfig, AnnCfgGroup, false)) {
				t.Features.Set(FeaturesDBKind, FCIgnore, true)
			}
			if ann, ok := t.Annotations[mongoAnnotation]; ok {
				if ig, ok := ann.GetBoolTag(mongoAnnotationTagIgnore); ok || ig {
					t.Features.Set(FeaturesDBKind, FCIgnore, true)
					continue
				}
			}
			for _, f := range t.Fields {
				if f.HasModifier(AttrModifierAuxiliary) {
					f.Features.Set(FeaturesDBKind, FCIgnore, true)
				}
				if f.Features.Bool(FeaturesDBKind, FCIgnore) {
					continue
				}
				if ann, ok := f.Annotations[mongoAnnotation]; ok {
					if ig, ok := ann.GetBoolTag(mongoAnnotationTagIgnore); ok && ig {
						continue
					}
				}
				if f.HasModifier(AttrModifierOneToMany) {
					inc := true
					if in, ok := f.Annotations.GetBoolAnnotation(mongoAnnotation, mongoAnnotationTagIncapsulate); ok {
						inc = in
					} else if in, ok := f.Annotations.GetBoolAnnotation(dbAnnotation, mongoAnnotationTagIncapsulate); ok {
						inc = in
					}
					f.Features.Set(FeaturesDBKind, FDBIncapsulate, inc)
					if inc {
						ft, _ := f.Features.GetEntity(FeaturesCommonKind, FCOneToManyType)
						ft.Features.Set(FeaturesDBKind, FCIgnore, true)
						ft.Features.Set(FeaturesAPIKind, FAPILevel, FAPILTypes)
					}
					// 	desc.GetFeature(f, FeaturesCommonPrefix, FCModifiedFieldName)
				}
				tag := cg.fieldName(f)
				if f.IsIdField() {
					tag = "_id"
				}
				desc.AddTag(f, mongoGoTagBSON, tag)
			}
		}
	}
	return nil
}

func (cg *MongoGenerator) Generate(bldr *Builder) (err error) {
	cg.desc = bldr.Descriptor
	cg.b = bldr
	if !cg.desc.Features.Bool(mongoFeatures, mfInited) {
		bldr.Descriptor.Engine.Initializator.Add(jen.Id(EngineVar).Dot(engineMongo).Op("=").Id("v").Dot("GetService").Params(jen.Lit(mongo.ServiceMongo)).
			Assert(jen.Op("*").Qual(vivardPackage, "MongoService")).Dot("DB")).Line()
		cg.desc.Features.Set(mongoFeatures, mfInited, true)
	}
	for _, t := range bldr.File.Entries {
		if ignore, ok := t.Features.GetBool(FeaturesDBKind, FCIgnore); !ok || !ignore {
			if !t.HasModifier(TypeModifierConfig) {
				err = cg.generateLoadFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}
				err = cg.generateSaveFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}
				err = cg.generateCreateFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}

				err = cg.generateRemoveFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}

				if t.IsDictionary() {
					err = cg.generateListFunc(t)
					if err != nil {
						err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
						return
					}
				}
				err = cg.generateLookupFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}
				err = cg.generateFindFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}
				if _, ok := t.Features.GetEntity(FeaturesCommonKind, FCForeignKey); ok {
					err = cg.generateListFKFunc(t)
					if err != nil {
						err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
						return
					}
					err = cg.generateRemoveFKFunc(t)
					if err != nil {
						err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
						return
					}
					err = cg.generateReplaceFKFunc(t)
					if err != nil {
						err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
						return
					}
				}
				gengen := cg.generateIDGenerators
				if !gengen {
					if gengenann, ok := t.Annotations[mongoAnnotation]; ok {
						if gen, ok := gengenann.GetBoolTag(mongoAnnotationTagGenerateIDGenerator); ok {
							gengen = gen
						}
					}
				}
				if gengen {
					err = cg.generateMongoIDGeneratorFunc(t)
					if err != nil {
						err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
						return
					}
				}
			} else {
				err = cg.generateCongifSaveFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}
				err = cg.generateCongifLoadFunc(t)
				if err != nil {
					err = fmt.Errorf("while generating %s (%s): %w", t.Name, bldr.File.FileName, err)
					return
				}
			}
		}
	}
	return nil
}

func (cg *MongoGenerator) generateLoadFunc(e *Entity) error {
	if e.HasModifier(TypeModifierConfig) {
		return nil
	}
	name := e.Name
	fname := cg.desc.GetMethodName(MethodLoad, name)
	params, err := cg.b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), e.GetIdField().Type)

	if err != nil {
		return err
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(params).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.Id("ret").Op(":=").Op("&").Id(name).Values(jen.Dict{}),
		jen.Id("err").Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("FindOne").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Id("id")}),
		).Dot("Decode").Params(jen.Id("ret")),
		returnIfErrValue(jen.Nil()),
		jen.Return(
			jen.List(jen.Id("ret"), jen.Nil()),
		),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateSaveFunc(e *Entity) error {
	if e.FB(FeaturesCommonKind, FCReadonly) {
		return nil
	}
	name := e.Name
	fname := cg.desc.GetMethodName(MethodSave, name)
	//TODO: hook
	var f jen.Code
	if e.HasModifier(TypeModifierConfig) {

	} else {
		resultName := "_" //"ur"
		f = jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).
			Params(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name))).
			Parens(jen.List(jen.Op("*").Id(name), jen.Error())).
			Block(
				cg.desc.Project.OnHook(HookSave, HMStart, e, &GeneratorHookVars{Obj: "o"}),
				jen.List(jen.Id(resultName), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("ReplaceOne").Params(
					jen.Id("ctx"),
					jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Id("o").Dot(e.GetIdField().Name)}),
					jen.Id("o"),
				),
				returnIfErrValue(jen.Nil()),
				cg.desc.Project.OnHook(HookSave, HMExit, e, &GeneratorHookVars{Obj: "o"}),
				jen.Return(
					jen.List(jen.Id("o"), jen.Nil()),
				),
			).Line()
	}
	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateCreateFunc(e *Entity) error {
	if e.FB(FeaturesCommonKind, FCReadonly) {
		return nil
	}
	if e.HasModifier(TypeModifierConfig) {
		return nil
	}
	name := e.Name
	fname := cg.desc.GetMethodName(MethodCreate, name)
	//TODO: hook
	resultName := "_" //"ir"

	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name))).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.List(jen.Id(resultName), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("InsertOne").Params(
			jen.Id("ctx"),
			jen.Id("o"),
		),
		returnIfErrValue(jen.Nil()),
		jen.Return(
			jen.List(jen.Id("o"), jen.Nil()),
		),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateRemoveFunc(e *Entity) error {
	if e.HasModifier(TypeModifierConfig) {
		return nil
	}
	name := e.Name
	fname := cg.desc.GetMethodName(MethodRemove, name)
	params, err := cg.b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), e.GetIdField().Type)

	if err != nil {
		return err
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(params).Parens(jen.Error()).Block(
		jen.Id("now").Op(":=").Qual("time", "Now").Params(),
		jen.List(jen.Id("_"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("UpdateOne").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Id("id")}),
			jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("$set"): jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit(mdDeletedFieldName): jen.Id("now")})}),
		),
		jen.Return(
			jen.Id("err"),
		),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateListFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodList, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context")).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).BlockFunc(func(g *jen.Group) {
		g.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{})
		g.Id("query").Op(":=").Qual(bsonPackage, "M").Values(jen.Dict{
			jen.Lit(mdDeletedFieldName): jen.Qual(bsonPackage, "M").Values(jen.Dict{
				jen.Lit("$exists"): jen.Lit(0),
			}),
		})
		if e.BaseTypeName != "" {
			g.Add(cg.addDescendantsToQuery(e, "query"))
		}
		g.List(jen.Id("curr"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("Find").Params(
			jen.Id("ctx"),
			jen.Id("query"),
		)
		g.Add(returnIfErrValue(jen.Nil()))
		g.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx"))
		g.Id("err").Op("=").Op("curr").Dot("All").Params(jen.Id("ctx"), jen.Op("&").Id("ret"))
		g.Return(
			jen.List(jen.Id("ret"), jen.Id("err")),
		)
	}).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateListFKFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodListFK, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("parentID").Int()).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).Block(
		jen.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{}),
		jen.List(jen.Id("curr"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("Find").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(jen.DictFunc(func(d jen.Dict) {
				// if ff, ok := e.Annotations.GetInterfaceAnnotation(codeGeneratorAnnotation, AnnotationTagForeignKeyField).(*Field); ok {
				if ff, ok := e.Features.GetField(FeaturesCommonKind, FCForeignKeyField); ok {
					d[jen.Lit(cg.fieldName(ff))] = jen.Id("parentID")
				} else {
					d[jen.Lit("InCaseIfAnythingWrong")] = jen.Lit("ThisShouldntHappen")
				}
			})),
		),
		returnIfErrValue(jen.Nil()),
		jen.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx")),
		jen.Id("err").Op("=").Op("curr").Dot("All").Params(jen.Id("ctx"), jen.Op("&").Id("ret")),
		jen.Return(
			jen.List(jen.Id("ret"), jen.Id("err")),
		),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateRemoveFKFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodRemoveFK, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("parentID").Int()).
		Parens(jen.Id("err").Error()).Block(
		jen.List(jen.Id("_"), jen.Id("err")).Op("=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("DeleteMany").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(jen.DictFunc(func(d jen.Dict) {
				// if ff, ok := e.Annotations.GetInterfaceAnnotation(codeGeneratorAnnotation, AnnotationTagForeignKeyField).(*Field); ok {
				if ff, ok := e.Features.GetField(FeaturesCommonKind, FCForeignKeyField); ok {
					d[jen.Lit(cg.fieldName(ff))] = jen.Id("parentID")
				} else {
					d[jen.Lit("InCaseIfAnythingWrong")] = jen.Lit("ThisShouldntHappen")
				}
			})),
		),
		jen.Return(),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateReplaceFKFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodReplaceFK, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual("context", "Context"), jen.Id("parentID").Int(), jen.Id("vals").Index().Op("*").Id(e.Name)).
		Parens(jen.Id("err").Error()).Block(
		jen.Id("err").Op("=").Id("eng").Dot(cg.desc.GetMethodName(MethodRemoveFK, name)).Params(jen.Id("ctx"), jen.Id("parentID")),
		returnIfErr(),
		jen.Id("items").Op(":=").Make(jen.Index().Interface(), jen.Len(jen.Id("vals"))),
		jen.For(jen.List(jen.Id("i"), jen.Id("v").Op(":=").Range().Id("vals"))).Block(
			jen.Id("items").Index(jen.Id("i")).Op("=").Id("v"),
		),
		jen.List(jen.Id("_"), jen.Id("err")).Op("=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("InsertMany").Params(
			jen.Id("ctx"),
			jen.Id("items"),
		),
		jen.Return(),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}
func (cg *MongoGenerator) generateLookupFunc(e *Entity) error {
	if e.HasModifier(TypeModifierConfig) {
		return nil
	}
	name := e.Name
	fname := cg.desc.GetMethodName(MethodLookup, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual("context", "Context"), jen.Id("query").String()).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).BlockFunc(func(g *jen.Group) {
		g.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{})
		g.Id("q").Op(":=").Qual(bsonPackage, "M").Values(jen.Dict{
			jen.Lit(mdDeletedFieldName): jen.Qual(bsonPackage, "M").Values(jen.Dict{
				jen.Lit("$exists"): jen.Lit(0),
			}),
		})
		if e.BaseTypeName != "" {
			g.Add(cg.addDescendantsToQuery(e, "q"))
		}

		g.List(jen.Id("curr"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("Find").Params(
			jen.Id("ctx"),
			//TODO use query
			jen.Id("q"),
		)
		g.Add(returnIfErrValue(jen.Nil()))
		g.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx"))
		g.Id("err").Op("=").Op("curr").Dot("All").Params(jen.Id("ctx"), jen.Op("&").Id("ret"))
		g.Return(
			jen.List(jen.Id("ret"), jen.Id("err")),
		)
	}).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateFindFunc(e *Entity) error {
	if it, ok := e.Features.GetEntity(FeaturesAPIKind, FAPIFindParamType); ok {
		name := e.Name
		fname := cg.desc.GetMethodName(MethodFind, name)
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("query").Op("*").Id(it.Name),
		).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).BlockFunc(func(g *jen.Group) {
			g.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{})
			g.Id("q").Op(":=").Qual(bsonPackage, "M").Values(jen.Dict{
				jen.Lit(mdDeletedFieldName): jen.Qual(bsonPackage, "M").Values(jen.Dict{
					jen.Lit("$exists"): jen.Lit(0),
				}),
			})
			possiblyFilledFields := map[string]struct{}{}
			for _, f := range it.Fields {
				// fldName := f.Name
				// op := AFTEqual
				// an, ok := f.Annotations[AnnotationFind]
				// if ok {
				// 	if len(an.Values) > 0 && an.Values[0].Value == nil {
				// 		fldName = an.Values[0].Key
				// 	} else if fn, ok := an.GetStringTag(AnnFndFieldTag); ok {
				// 		fldName = fn
				// 	}
				// 	op = an.GetString(AnnFndTypeTag, AFTEqual)
				// }
				// searchField := e.GetField(fldName)
				// if searchField == nil {
				// 	cg.desc.AddError(fmt.Errorf("at %v: can not find field %s in type %s", f.Pos, fldName, e.Name))
				// 	return
				// }
				searchField, _ := f.Features.GetField(FeaturesAPIKind, FAPIFindFor)
				op := f.FS(FeaturesAPIKind, FAPIFindParam)
				mngFldName := cg.fieldName(searchField)
				g.IfFunc(func(g *jen.Group) {
					if f.Type.Array == nil {
						g.Id("query").Dot(f.Name).Op("!=").Nil()
					} else {
						g.Len(jen.Id("query").Dot(f.Name)).Op("!=").Lit(0)
					}
				}).BlockFunc(func(g *jen.Group) {
					if f.Type.Array != nil {
						// g.Id("values").Op(":=").Make(jen.Index().Qual(bsonPackage, "A"), jen.Len(jen.Id("query").Dot(f.Name)))
						// g.For(jen.List(jen.Id("i"), jen.Id("v")).Op(":=").Range().Id("query").Dot(f.Name)).Block(
						// 	jen.Id("values").Index(jen.Id("i")).Op("=").Id("v"),
						// )
						g.Id("q").Index(jen.Lit(mngFldName)).Op("=").Qual(bsonPackage, "M").Values(jen.Dict{
							jen.Lit("$in"): jen.Id("query").Dot(f.Name),
						})
					} else if f.Type.Map != nil {
						g.Id("arr").Op(":=").Make(jen.Qual(bsonPackage, "A"), jen.Len(jen.Id("query").Dot(f.Name)))
						g.Id("i").Op(":=").Lit(0)
						g.For(jen.List(jen.Id("k"), jen.Id("val")).Op(":=").Range().Id("query").Dot(f.Name)).Block(
							jen.Id("arr").Index(jen.Id("i")).Op("=").Qual(bsonPackage, "M").Values(
								jen.Dict{jen.Lit(mngFldName).Op("+").Lit(".").Op("+").Id("k"): jen.Id("val")},
							),
						)
						g.Id("q").Index(jen.Lit("$and")).Op("=").Id("arr")
					} else {
						pref := jen.Id("q").Index(jen.Lit(mngFldName))
						if _, ok := possiblyFilledFields[mngFldName]; ok {
							g.Var().Id("op").Qual(bsonPackage, "M")
							g.If(
								jen.List(jen.Id("o"), jen.Id("ok")).Op(":=").Id("q").Index(jen.Lit(mngFldName)).Assert(jen.Qual(bsonPackage, "M")),
								jen.Id("ok"),
							).Block(jen.Id("op").Op("=").Id("o")).Else().Block(
								jen.Id("op").Op("=").Qual(bsonPackage, "M").Values(),
								jen.Id("q").Index(jen.Lit(mngFldName)).Op("=").Id("op"),
							)
							pref = jen.Id("op")
						}
						switch op {
						case AFTEqual:
							g.Add(pref).Op("=").Id("query").Dot(f.Name)
						case AFTNotEqual:
							g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(jen.Dict{
								jen.Lit("$ne"): jen.Id("query").Dot(f.Name),
							})
						case AFTGreaterThan:
							g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(jen.Dict{
								jen.Lit("$gt"): jen.Id("query").Dot(f.Name),
							})
						case AFTGreaterThanOrEqual:
							g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(jen.Dict{
								jen.Lit("$gte"): jen.Id("query").Dot(f.Name),
							})
						case AFTLessThan:
							g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(jen.Dict{
								jen.Lit("$lt"): jen.Id("query").Dot(f.Name),
							})
						case AFTLessThanOrEqual:
							g.Add(pref).Index(jen.Lit("$lte")).Op("=").Id("query").Dot(f.Name)
						case AFTStartsWith:
							g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(jen.Dict{
								jen.Lit("$regex"): jen.Lit("^").Op("+").Op("*").Id("query").Dot(f.Name),
							})
						case AFTContains:
							g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(jen.Dict{
								jen.Lit("$regex"): jen.Op("*").Id("query").Dot(f.Name),
							})
						default:
							cg.desc.AddError(fmt.Errorf("at %v: undefined comparision type: %s", f.Pos, op))
							return
						}
						possiblyFilledFields[mngFldName] = struct{}{}
					}
				})
			}
			g.List(jen.Id("curr"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("Find").Params(
				jen.Id("ctx"),
				jen.Id("q"),
			)
			g.Add(returnIfErrValue(jen.Nil()))
			g.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx"))
			g.Id("err").Op("=").Op("curr").Dot("All").Params(jen.Id("ctx"), jen.Op("&").Id("ret"))
			g.Return(
				jen.List(jen.Id("ret"), jen.Id("err")),
			)
		}).Line()

		cg.b.Functions.Add(f)
	}
	return nil
}

func (cg *MongoGenerator) generateMongoIDGeneratorFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodGenerateID, name)
	ret, err := cg.b.addType(jen.Id("id"), e.GetIdField().Type)
	if err != nil {
		return err
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context")).Parens(
		jen.List(ret, jen.Id("err").Error()),
	).Block(
		jen.Id("id").Op("=").Add(cg.b.goEmptyValue(e.GetIdField().Type)).Line().
			Return(),
	).Line()
	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateCongifLoadFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodLoad, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context")).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.List(jen.Id("ret"), jen.Id("_")).Op(":=").Id(EngineVar).Dot(cg.b.Descriptor.GetMethodName(MethodInit, e.Name)).Params(jen.Id("ctx")),
		jen.List(jen.Id("curr"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("Find").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(),
		),
		jen.Add(returnIfErrValue(jen.Nil())),
		jen.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx")),
		jen.For(jen.Id("curr").Dot("Next").Call(jen.Id("ctx"))).Block(
			jen.Id("idVal").Op(":=").Id("curr").Dot("Current").Dot("Lookup").Call(jen.Lit("_id")).Dot("StringValue").Call(),
			jen.Switch(jen.Id("idVal")).BlockFunc(func(sg *jen.Group) {
				for _, f := range e.Fields {
					sg.Case(jen.Lit(f.Name)).Block(
						jen.Id("curr").Dot("Current").Dot("Lookup").Call(jen.Lit("value")).Dot("Unmarshal").Call(jen.Op("&").Id("ret").Dot(f.Name)),
					)
				}
			}),
		),
		jen.Return(
			jen.List(jen.Id("ret"), jen.Id("err")),
		),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateCongifSaveFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodSave, name)
	resultName := "_" //"ur"

	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("o").Op("*").Id(name)).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).BlockFunc(func(g *jen.Group) {
		// g.Var().Id(resultName).Op("*").Qual(mongoPackage, UpdateResult)
		g.Var().Err().Error()
		g.Id("opts").Op(":=").Qual(optionsPackage, "Update").Call().Dot("SetUpsert").Call(jen.Lit(true))
		for _, f := range e.Fields {
			g.List(jen.Id(resultName), jen.Id("err")).Op("=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(jen.Lit(cg.collectionName(e))).Dot("UpdateOne").Params(
				jen.Id("ctx"),
				jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Lit(f.Name)}),
				jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("$set"): jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("value"): jen.Id("o").Dot(f.Name)})}),
				jen.Id("opts"),
			)

		}

		g.Add(returnIfErrValue(jen.Nil()))

		g.Return(
			jen.List(jen.Id("o"), jen.Id("err")),
		)
	}).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) collectionName(e *Entity) string {
	cn, ok := cg.collections[e.Name]
	if !ok {
		pref := ""
		if cg.prefixCollectionName != "" {
			if cg.prefixCollectionName == optPrefixWithPackage {
				pref = e.Pckg.Name + "_"
			} else {
				pref = cg.prefixCollectionName
			}
		}
		name := e.Name
		t := e
		for t.BaseTypeName != "" {
			t = t.GetBaseType()
			name = t.Name
		}
		cn = pref + ToSnakeCase(name)

		if ann, ok := e.Annotations[mongoAnnotation]; ok {
			if t, ok := ann.GetStringTag(mongoAnnotationTagName); ok {
				cn = t
			}
		}
		if cg.usedCollections[cn] && t.BaseTypeName == "" && !t.HasModifier(TypeModifierExtendable) {
			cg.desc.AddWarning(fmt.Sprintf("mongo: collection duplicate: %s", cn))
		}
		cg.collections[e.Name] = cn
		cg.usedCollections[cn] = true
	}
	return cn
}

func (cg *MongoGenerator) fieldName(f *Field) string {
	if an, ok := f.Annotations[mongoAnnotation]; ok {
		if t, ok := an.GetStringTag(mongoAnnotationTagName); ok {
			return t
		}
	}
	return ToSnakeCase(f.Name)
}

func (cg *MongoGenerator) addDescendantsToQuery(e *Entity, queryVar string) *jen.Statement {
	name := ToSnakeCase(ExtendableTypeDescriptorFieldName)
	if desc, ok := e.Features.Get(FeaturesCommonKind, FCDescendants); ok {
		return jen.Id(queryVar).Index(jen.Lit(name)).Op("=").Qual(bsonPackage, "M").Values(
			jen.Dict{jen.Lit("$in"): jen.Qual(bsonPackage, "A").ValuesFunc(func(g *jen.Group) {
				g.Lit(e.FS(FeatGoKind, FCGDerivedTypeNameConst))
				for _, d := range desc.([]*Entity) {
					g.Lit(d.FS(FeatGoKind, FCGDerivedTypeNameConst))
				}
			})},
		)
	}
	return jen.Id(queryVar).Index(jen.Lit(name)).Op("=").Lit(e.FS(FeatGoKind, FCGDerivedTypeNameConst))
}

func (cg *MongoGenerator) init() {
	if cg.collections == nil {
		cg.collections = map[string]string{}
	}
	if cg.usedCollections == nil {
		cg.usedCollections = map[string]bool{}
	}
	cg.prefixCollectionName = optPrefixWithPackage
	cg.useBaseCollectionForDerived = true
}
