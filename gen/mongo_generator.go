package gen

import (
	"fmt"
	"strings"

	"github.com/vc2402/vivard/utils"

	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard/mongo"
)

const (
	mongoGeneratorName     = "Mongo"
	mongoAnnotation        = "mongo"
	dbAnnotation           = "db"
	mongoAnnotationTagName = "name"
	//mongoAnnotationTagNameMutable - bool; if true - store collection name in var instead of const
	mongoAnnotationTagNameMutable         = "nameMutable"
	mongoAnnotationTagIgnore              = "ignore"
	mongoAnnotationTagGenerateIDGenerator = "idGenerator"
	mongoAnnotationTagEncapsulate         = "encapsulate"
	mongoAnnotationTagAddBsonTag          = "bsonTag"
	mongoAnnotationOrder                  = "sort"
	mongoAnnotationDeleteMethod           = "deleteMethod"
	madmUpdate                            = "update"
	madmDelete                            = "delete"
	madmCustomQueryGenerator              = "customQueryGenerator"
	// madmProcessQuery tells to call method with query before call find
	madmProcessQuery = "postProcessQuery"

	mongoGoTagBSON = "bson"
)

const (
	bsonPackage     = "go.mongodb.org/mongo-driver/bson"
	mongoPackage    = "go.mongodb.org/mongo-driver/mongo"
	optionsPackage  = "go.mongodb.org/mongo-driver/mongo/options"
	vivMongoPackage = "github.com/vc2402/vivard/mongo"
	engineMongo     = "Mongo"

	optionsMongo                      = "mongo"
	optionGenerateIDGenerator         = "idGenerator"
	optionPrefixCollectionName        = "prefixCollectionName"
	optionUseBaseCollectionForDerived = "baseCollectionForDerived"
	optionDeleteMethod                = "deleteMethod"

	optPrefixWithPackage = "<package>"
)

const (
	mongoFeatures           FeatureKind = "mongo"
	mfInited                            = "inited"
	mfDelete                            = "delete"
	mfSortField                         = "sort-field"
	mfSortDesc                          = "sort-desc"
	mfCollectionConst                   = "collection-const"
	mfSkipQueryGenerator                = "skip-q-gen"
	mfQueryPostProcessor                = "q-p-processor"
	mfCollectionNameMutable             = "collection-name-mutable"
)

const (
	mdDeletedFieldName = "deleted"

	queryGeneratorFuncNameTemplate = "generateQueryFrom%s"
	postProcessorFuncNameTemplate  = "postProcessQuery%s"
)

type MongoGenerator struct {
	desc                        *Package
	b                           *Builder
	collections                 map[string]string
	usedCollections             map[string]bool
	generateIDGenerators        bool
	prefixCollectionName        string
	useBaseCollectionForDerived bool
	deleteMethod                string
	inited                      bool
}

func init() {
	plugin := &MongoGenerator{}
	plugin.init()
	RegisterPlugin(plugin)
}

func (cg *MongoGenerator) Name() string {
	return mongoGeneratorName
}

func (cg *MongoGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	f, fld := item.(*Field)
	e, ent := item.(*Entity)
	cg.init()
	if ann.Name == mongoAnnotation || ann.Name == dbAnnotation {
		if !fld && !ent {
			return true, fmt.Errorf("at %v: mongo annotation may be used only with type and field", ann.Pos)
		}
		for _, v := range ann.Values {
			switch v.Key {
			case mongoAnnotationTagName:
			case mongoAnnotationTagIgnore:
			case mongoAnnotationTagEncapsulate:
			case mongoAnnotationTagAddBsonTag:
			case mongoAnnotationDeleteMethod:
				if m, ok := v.GetString(); !ok || m != madmUpdate && m != madmDelete {
					return true, fmt.Errorf("at %v: invalid value for %s annotation: %v", ann.Pos, v.Key, v.Value)
				}
			case mongoAnnotationOrder:
				if fld {
					f.parent.Features.Set(mongoFeatures, mfSortField, item)
				}
			case madmCustomQueryGenerator:
				if ent && e.Annotations[AnnotationFind] != nil {
					if val, ok := v.GetBool(); ok {
						e.Features.Set(mongoFeatures, mfSkipQueryGenerator, val)
					} else {
						return true, fmt.Errorf("at %v: mongo annotation parameter '%s' should be true or false", ann.Pos, v.Key)
					}
				} else {
					return true, fmt.Errorf(
						"at %v: mongo annotation parameter '%s' can be used for find annotated type only",
						ann.Pos,
						v.Key,
					)
				}
			case madmProcessQuery:
				if ent && e.Annotations[AnnotationFind] != nil {
					if val, ok := v.GetBool(); ok && val {
						e.Features.Set(mongoFeatures, mfQueryPostProcessor, fmt.Sprintf(postProcessorFuncNameTemplate, e.Name))
					} else if name, ok := v.GetString(); ok && name != "" {
						e.Features.Set(mongoFeatures, mfQueryPostProcessor, name)
					} else {
						return true, fmt.Errorf("at %v: mongo annotation parameter '%s' should be true or false", ann.Pos, v.Key)
					}
				} else {
					return true, fmt.Errorf(
						"at %v: mongo annotation parameter '%s' can be used for find annotated type only",
						ann.Pos,
						v.Key,
					)
				}
			case mongoAnnotationTagNameMutable:
				if ent {
					val, ok := v.GetBool()
					if !ok {
						return true, fmt.Errorf(
							"at %v: mongo annotation parameter '%s' should be of bool type",
							ann.Pos,
							v.Key,
						)
					}
					e.Features.Set(mongoFeatures, mfCollectionNameMutable, val)
				} else {
					return true, fmt.Errorf(
						"at %v: mongo annotation parameter '%s' can be used for type only",
						ann.Pos,
						v.Key,
					)
				}
			default:
				return true, fmt.Errorf("at %v: unknown mongo annotation parameter: %s", ann.Pos, v.Key)
			}
		}
		return true, nil
	} else if ann.Name == AnnotationConfig && fld {
		if _, ok := ann.GetBoolTag(AnnCfgMutable); ok {
			return true, nil
		}
	} else if ann.Name == AnnotationSort && fld {
		f.parent.Features.Set(mongoFeatures, mfSortField, f)
		if ann.GetBool(AnnSortDescending, false) {
			f.Features.Set(mongoFeatures, mfSortDesc, true)
		}
		return true, nil
	} else if ann.Name == AnnotationLookup && fld {
		return true, nil
	}
	return false, nil
}

// ProvideFeature from FeatureProvider interface
func (cg *MongoGenerator) ProvideFeature(
	kind FeatureKind,
	name string,
	obj interface{},
) (feature interface{}, ok ProvideFeatureResult) {
	switch kind {
	case FeaturesDBKind:
		switch name {
		case FDBFlushDict:
			if t, ok := obj.(*Entity); ok && t.HasModifier(TypeModifierDictionary) {
				return func(args ...interface{}) jen.Code {
					obj := jen.Id("items")
					if len(args) > 0 {
						n, ok := args[0].(string)
						if ok {
							obj = jen.Id(n)
						}
					}

					stmt := &jen.Statement{}
					idField := t.GetIdField()
					if idField.HasModifier(AttrModifierIDAuto) {
						stmt = jen.Id("maxId").Op(":=").Lit(0).Line()
					}
					stmt.For(jen.List(jen.Id("_"), jen.Id("o")).Op(":=").Range().Add(obj)).BlockFunc(
						func(g *jen.Group) {
							if idField.HasModifier(AttrModifierIDAuto) {
								g.If(jen.Id("maxId").Op("<=").Id("o").Dot(idField.Name)).Block(
									jen.Id("maxId").Op("=").Id("o").Dot(idField.Name),
								)
							}
							g.Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
								jen.Id(
									t.FS(
										mongoFeatures,
										mfCollectionConst,
									),
								),
							).Dot("InsertOne").Params(
								jen.Id("ctx"),
								jen.Id("o"),
							)

						},
					).Line()
					if idField.HasModifier(AttrModifierIDAuto) {
						if f := cg.desc.GetFeature(t, SequenceFeatures, SFSetCurrentValue); f != nil {
							fun, ok := f.(func(args ...interface{}) jen.Code)
							if ok {
								//stmt.Id("maxId").Op("++")
								stmt.Add(fun("maxId"))
							}
						}
					}
					return stmt
				}, FeatureProvided
			}
		}
	}
	return nil, FeatureNotProvided
}

func (cg *MongoGenerator) SetOptions(options any) error {
	if opts, ok := options.(map[string]interface{}); ok {
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
		if dm, ok := opts[optionDeleteMethod].(string); ok {
			if dm == madmUpdate || dm == madmDelete {
				cg.deleteMethod = dm
			} else {
				cg.desc.AddWarning(
					fmt.Sprintf(
						"ignoring invalid value for mongo option %s: %s (allowed '%s' and '%s')",
						optionDeleteMethod,
						dm,
						madmUpdate,
						madmDelete,
					),
				)
			}
		}
	}
	return nil
}

func (cg *MongoGenerator) Prepare(desc *Package) error {
	cg.init()
	cg.desc = desc
	cg.deleteMethod = madmUpdate
	if opts, ok := desc.Options().Custom[optionsMongo]; ok {
		_ = cg.SetOptions(opts)
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
				(t.Annotations.GetBoolAnnotationDef(
					AnnotationConfig,
					AnnCfgValue,
					false,
				) || t.Annotations.GetBoolAnnotationDef(AnnotationConfig, AnnCfgGroup, false)) {
				t.Features.Set(FeaturesDBKind, FCIgnore, true)
			}
			if ann, ok := t.Annotations[mongoAnnotation]; ok {
				if ig, ok := ann.GetBoolTag(mongoAnnotationTagIgnore); ok || ig {
					t.Features.Set(FeaturesDBKind, FCIgnore, true)
					continue
				}
				if dm, ok := ann.GetStringTag(mongoAnnotationDeleteMethod); ok {
					t.Features.Set(mongoFeatures, mfDelete, dm)
				}
			}
			collConstName := fmt.Sprintf("Col%s%s", strings.ToUpper(t.Name)[:1], t.Name[1:])
			t.Features.Set(mongoFeatures, mfCollectionConst, collConstName)
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
					inc := false
					if in, ok := f.Annotations.GetBoolAnnotation(mongoAnnotation, mongoAnnotationTagEncapsulate); ok {
						inc = in
					} else if in, ok := f.Annotations.GetBoolAnnotation(dbAnnotation, mongoAnnotationTagEncapsulate); ok {
						inc = in
					}
					f.Features.Set(FeaturesDBKind, FDBEncapsulate, inc)
					if inc {
						ft, _ := f.Features.GetEntity(FeaturesCommonKind, FCOneToManyType)
						ft.Features.Set(FeaturesDBKind, FCIgnore, true)
						ft.Features.Set(FeaturesAPIKind, FAPILevel, FAPILTypes)
					}
					// 	desc.GetFeature(f, FeaturesCommonPrefix, FCModifiedFieldName)
				}
				tag := cg.fieldName(f)
				if f.Features.Bool(FeaturesCommonKind, FCGDeletedField) {
					tag += ",omitempty"
				}
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
		bldr.Descriptor.Engine.Initializator.Add(
			jen.List(
				jen.Id(EngineVar).Dot(engineMongo),
				jen.Id("err"),
			).Op("=").Id("v").Dot("GetService").Params(jen.Lit(mongo.ServiceMongo)).
				Assert(jen.Op("*").Qual(vivMongoPackage, "Service")).Dot("GetDefaultMongo").
				Params(jen.Qual("context", "TODO").Params()),
		).
			Line()
		cg.desc.Features.Set(mongoFeatures, mfInited, true)
	}
	for _, t := range bldr.File.Entries {
		if ignore, ok := t.Features.GetBool(FeaturesDBKind, FCIgnore); !ok || !ignore {
			cg.generateConst(t)
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
				err = cg.generateConfigSaveFunc(t)
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

func (cg *MongoGenerator) generateConst(e *Entity) {
	cn := cg.collectionName(e)
	constName := e.FS(mongoFeatures, mfCollectionConst)
	if e.FB(mongoFeatures, mfCollectionNameMutable) {
		cg.b.vars["mongo_collections"] = append(cg.b.vars["mongo_collections"], jen.Id(constName).Op("=").Lit(cn))
	} else {
		cg.b.consts["mongo_collections"] = append(cg.b.consts["mongo_collections"], jen.Id(constName).Op("=").Lit(cn))
	}
	return
}

func (cg *MongoGenerator) generateLoadFunc(e *Entity) error {
	if e.HasModifier(TypeModifierConfig) {
		return nil
	}
	name := e.Name
	fname := cg.desc.GetMethodName(MethodLoad, name)
	idField := e.GetIdField()
	if idField == nil {
		return fmt.Errorf("at %v: Mongo:Load: no id field found for type %s", e.Pos, e.Name)
	}
	params, err := cg.b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), idField.Type)

	if err != nil {
		return err
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(params).Parens(
		jen.List(
			jen.Op("*").Id(name),
			jen.Error(),
		),
	).Block(
		jen.Id("ret").Op(":=").Op("&").Id(name).Values(jen.Dict{}),
		jen.Id("err").Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
			jen.Id(
				e.FS(
					mongoFeatures,
					mfCollectionConst,
				),
			),
		).Dot("FindOne").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Id("id")}),
		).Dot("Decode").Params(jen.Id("ret")),
		//returnIfErrValue(jen.Nil()),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.If(jen.Id("err").Op("==").Qual(mongoPackage, "ErrNoDocuments")).Block(
				jen.Return(jen.Nil(), jen.Nil()),
			),
			jen.Return(jen.Nil(), jen.Id("err")),
		),
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
				jen.List(
					jen.Id(resultName),
					jen.Id("err"),
				).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
					jen.Id(
						e.FS(
							mongoFeatures,
							mfCollectionConst,
						),
					),
				).Dot("ReplaceOne").Params(
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
		jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)),
	).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.List(
			jen.Id(resultName),
			jen.Id("err"),
		).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
			jen.Id(
				e.FS(
					mongoFeatures,
					mfCollectionConst,
				),
			),
		).Dot("InsertOne").Params(
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
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(params).Parens(jen.Error()).BlockFunc(
		func(g *jen.Group) {
			dm := cg.deleteMethod
			if m, ok := e.Features.GetString(mongoFeatures, mfDelete); ok {
				dm = m
			}
			if dm == madmDelete {
				g.List(
					jen.Id("_"),
					jen.Id("err"),
				).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
					jen.Id(
						e.FS(
							mongoFeatures,
							mfCollectionConst,
						),
					),
				).Dot("DeleteOne").Params(
					jen.Id("ctx"),
					jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Id("id")}),
				)
			} else {
				g.Id("now").Op(":=").Qual("time", "Now").Params()
				g.List(
					jen.Id("_"),
					jen.Id("err"),
				).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
					jen.Id(
						e.FS(
							mongoFeatures,
							mfCollectionConst,
						),
					),
				).Dot("UpdateOne").Params(
					jen.Id("ctx"),
					jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Id("id")}),
					jen.Qual(bsonPackage, "M").Values(
						jen.Dict{
							jen.Lit("$set"): jen.Qual(
								bsonPackage,
								"M",
							).Values(jen.Dict{jen.Lit(mdDeletedFieldName): jen.Id("now")}),
						},
					),
				)
			}
			g.Return(
				jen.Id("err"),
			)
		},
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateListFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodList, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual(
			"context",
			"Context",
		),
	).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).BlockFunc(
		func(g *jen.Group) {
			g.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{})
			g.Id("query").Op(":=").Qual(bsonPackage, "M").Values(
				jen.Dict{
					jen.Lit(mdDeletedFieldName): jen.Qual(bsonPackage, "M").Values(
						jen.Dict{
							jen.Lit("$exists"): jen.Lit(0),
						},
					),
				},
			)
			if e.BaseTypeName != "" {
				g.Add(cg.addDescendantsToQuery(e, "query"))
			}
			if f, ok := e.Features.Get(mongoFeatures, mfSortField); ok {
				fld, ok := f.(*Field)
				if !ok {
					cg.desc.AddError(
						fmt.Errorf(
							"at %v: internal error in mongo:listAll: sort feature is not a field: %T",
							e.Pos,
							f,
						),
					)
					return
				}
				order := 1
				if fld.FB(mongoFeatures, mfSortDesc) {
					order = -1
				}

				g.Id("op").Op(":=").Qual(optionsPackage, "Find").Call().Dot("SetSort").Params(
					jen.Qual(bsonPackage, "M").Values(
						jen.Dict{
							jen.Lit(cg.fieldName(fld)): jen.Lit(order),
						},
					),
				)
			} else {
				g.Id("op").Op(":=").Qual(optionsPackage, "Find").Call()
			}
			g.List(
				jen.Id("curr"),
				jen.Id("err"),
			).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
				jen.Id(
					e.FS(
						mongoFeatures,
						mfCollectionConst,
					),
				),
			).Dot("Find").Params(
				jen.Id("ctx"),
				jen.Id("query"),
				jen.Id("op"),
			)
			g.Add(returnIfErrValue(jen.Nil()))
			g.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx"))
			g.Id("err").Op("=").Op("curr").Dot("All").Params(jen.Id("ctx"), jen.Op("&").Id("ret"))
			g.Return(
				jen.List(jen.Id("ret"), jen.Id("err")),
			)
		},
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateListFKFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodListFK, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual(
			"context",
			"Context",
		), jen.Id("parentID").Int(),
	).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).Block(
		jen.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{}),
		jen.List(
			jen.Id("curr"),
			jen.Id("err"),
		).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
			jen.Id(
				e.FS(
					mongoFeatures,
					mfCollectionConst,
				),
			),
		).Dot("Find").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(
				jen.DictFunc(
					func(d jen.Dict) {
						// if ff, ok := e.Annotations.GetInterfaceAnnotation(codeGeneratorAnnotation, AnnotationTagForeignKeyField).(*Field); ok {
						if ff, ok := e.Features.GetField(FeaturesCommonKind, FCForeignKeyField); ok {
							d[jen.Lit(cg.fieldName(ff))] = jen.Id("parentID")
						} else {
							d[jen.Lit("InCaseIfAnythingWrong")] = jen.Lit("ThisShouldntHappen")
						}
					},
				),
			),
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
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual(
			"context",
			"Context",
		), jen.Id("parentID").Int(),
	).
		Parens(jen.Id("err").Error()).Block(
		jen.List(
			jen.Id("_"),
			jen.Id("err"),
		).Op("=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
			jen.Id(
				e.FS(
					mongoFeatures,
					mfCollectionConst,
				),
			),
		).Dot("DeleteMany").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(
				jen.DictFunc(
					func(d jen.Dict) {
						// if ff, ok := e.Annotations.GetInterfaceAnnotation(codeGeneratorAnnotation, AnnotationTagForeignKeyField).(*Field); ok {
						if ff, ok := e.Features.GetField(FeaturesCommonKind, FCForeignKeyField); ok {
							d[jen.Lit(cg.fieldName(ff))] = jen.Id("parentID")
						} else {
							d[jen.Lit("InCaseIfAnythingWrong")] = jen.Lit("ThisShouldntHappen")
						}
					},
				),
			),
		),
		jen.Return(),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateReplaceFKFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodReplaceFK, name)
	foreignKeyField, ok := e.Features.GetField(FeaturesCommonKind, FCForeignKeyField)
	if !ok {
		return fmt.Errorf("at %v: no foreign key field found")
	}
	foreignKeyFieldName := foreignKeyField.Name
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual("context", "Context"), jen.Id("parentID").Int(), jen.Id("vals").Index().Op("*").Id(e.Name),
	).
		Parens(jen.Id("err").Error()).Block(
		jen.Id("err").Op("=").Id("eng").Dot(cg.desc.GetMethodName(MethodRemoveFK, name)).Params(
			jen.Id("ctx"),
			jen.Id("parentID"),
		),
		returnIfErr(),
		jen.Id("items").Op(":=").Make(jen.Index().Interface(), jen.Len(jen.Id("vals"))),
		jen.For(jen.List(jen.Id("i"), jen.Id("v").Op(":=").Range().Id("vals"))).Block(
			jen.Id("v").Dot(foreignKeyFieldName).Op("=").Id("parentID"),
			jen.Id("items").Index(jen.Id("i")).Op("=").Id("v"),
		),
		jen.List(
			jen.Id("_"),
			jen.Id("err"),
		).Op("=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
			jen.Id(
				e.FS(
					mongoFeatures,
					mfCollectionConst,
				),
			),
		).Dot("InsertMany").Params(
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
		jen.Id("ctx").Qual("context", "Context"), jen.Id("query").String(),
	).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).BlockFunc(
		func(g *jen.Group) {
			g.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{})
			g.Id("q").Op(":=").Qual(bsonPackage, "M").Values(
				jen.Dict{
					jen.Lit(mdDeletedFieldName): jen.Qual(bsonPackage, "M").Values(
						jen.Dict{
							jen.Lit("$exists"): jen.Lit(0),
						},
					),
				},
			)
			if e.BaseTypeName != "" {
				g.Add(cg.addDescendantsToQuery(e, "q"))
			}
			searchFields := map[string]string{}
			otherFields := map[string]jen.Code{}
			for _, field := range e.GetFields(true, true) {
				if ann, ok := field.Annotations[AnnotationLookup]; ok {
					for _, tag := range ann.Values {
						var value jen.Code
						if v, ok := tag.GetString(); ok {
							value = jen.Lit(v)
						} else if v, ok := tag.GetBool(); ok {
							value = jen.Lit(v)
						} else if v, ok := tag.GetInt(); ok {
							value = jen.Lit(v)
						} else if v, ok := tag.GetFloat(); ok {
							value = jen.Lit(v)
						}
						name := cg.fieldName(field)
						switch tag.Key {
						case AFTEqual:
							otherFields[name] = value
						case AFTNotEqual:
							otherFields[name] = jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("$ne"): value})
						case ALStartsWith, ALStartsWithIgnoreCase, ALContains, ALContainsIgnoreCase:
							searchFields[name] = tag.Key
						}
					}

				}
			}
			if len(searchFields) > 0 || len(otherFields) > 0 {
				makeRegexp := func(op string) jen.Code {
					regexp := jen.Id("query")
					if op == ALStartsWith || op == ALStartsWithIgnoreCase {
						regexp = jen.Qual("fmt", "Sprintf").Params(jen.Lit("^%s"), regexp)
					}
					values := jen.Dict{
						jen.Lit("$regex"): regexp,
					}
					if op == ALStartsWithIgnoreCase || op == ALContainsIgnoreCase {
						values[jen.Lit("$options")] = jen.Lit("i")
					}
					return jen.Qual(bsonPackage, "M").Values(values)
				}
				if len(searchFields) > 0 {
					var or []jen.Code
					utils.WalkMap(
						searchFields,
						func(val string, key string) error {
							if len(searchFields) == 1 {
								g.Id("q").Index(jen.Lit(key)).Op("=").Add(makeRegexp(val))
							} else {
								or = append(or, jen.Line().Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit(key): makeRegexp(val)}))
							}
							return nil
						},
					)
					if len(searchFields) > 1 {
						g.Id("q").Index(jen.Lit("$or")).Op("=").Qual(bsonPackage, "A").Values(or...)
					}
					//for name, fn := range searchFields {
					//	if len(searchFields) == 1 {
					//		g.Id("q").Index(jen.Lit(name)).Op("=").Add(makeRegexp(fn))
					//	} else {
					//		or = append(or, jen.Line().Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit(name): makeRegexp(fn)}))
					//	}
					//}
					//if len(searchFields) > 1 {
					//	g.Id("q").Index(jen.Lit("$or")).Op("=").Qual(bsonPackage, "A").Values(or...)
					//}
				}
				if len(otherFields) > 0 {
					utils.WalkMap(
						otherFields,
						func(val jen.Code, key string) error {
							g.Id("q").Index(jen.Lit(key)).Op("=").Add(val)
							return nil
						},
					)
				}
				//for name, fld := range otherFields {
				//	g.Id("q").Index(jen.Lit(name)).Op("=").Add(fld)
				//}
			} else {
				// allow to send query as json
				g.Qual("encoding/json", "Unmarshal").Params(jen.Op("[]").Byte().Parens(jen.Id("query")), jen.Op("&").Id("q"))
			}
			g.List(
				jen.Id("curr"),
				jen.Id("err"),
			).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
				jen.Id(
					e.FS(
						mongoFeatures,
						mfCollectionConst,
					),
				),
			).Dot("Find").Params(
				jen.Id("ctx"),
				jen.Id("q"),
			)
			g.Add(returnIfErrValue(jen.Nil()))
			g.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx"))
			g.Id("err").Op("=").Op("curr").Dot("All").Params(jen.Id("ctx"), jen.Op("&").Id("ret"))
			g.Return(
				jen.List(jen.Id("ret"), jen.Id("err")),
			)
		},
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateFindFunc(e *Entity) error {
	if it, ok := e.Features.GetEntity(FeaturesAPIKind, FAPIFindParamType); ok {
		name := e.Name
		generatorFuncName := fmt.Sprintf(queryGeneratorFuncNameTemplate, it.Name)
		postprocessorMethod := it.FS(mongoFeatures, mfQueryPostProcessor)
		fname := cg.desc.GetMethodName(MethodFind, name)
		deletedFound := false
		if !it.FB(mongoFeatures, mfSkipQueryGenerator) {
			f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(generatorFuncName).Params(
				jen.Id("query").Op("*").Id(it.Name),
			).Parens(jen.List(jen.Interface(), jen.Error())).BlockFunc(
				func(g *jen.Group) {
					addSkipDeletedToQuery := func() *jen.Statement {
						return jen.Id("q").Index(jen.Lit(mdDeletedFieldName)).Op("=").Qual(
							bsonPackage,
							"M",
						).Values(jen.Dict{jen.Lit("$exists"): jen.Lit(0)})
					}
					g.Id("q").Op(":=").Qual(bsonPackage, "M").Values()
					possiblyFilledFields := map[string]struct{}{}
					for _, f := range it.Fields {
						searchField, ok := f.Features.GetField(FeaturesAPIKind, FAPIFindFor)
						op := f.FS(FeaturesAPIKind, FAPIFindParam)
						var mngFldName string
						var elseStmt *jen.Statement = jen.Empty()
						if ok {
							if fields, ok := f.Features.Get(FeaturesAPIKind, FAPIFindForEmbedded); ok {
								for i, field := range fields.([]*Field) {
									if i > 0 {
										mngFldName += "."
									}
									mngFldName += cg.fieldName(field)
								}
							} else {
								//TODO save mongo name in feature...
								if searchField.IsIdField() {
									mngFldName = "_id"
								} else {
									mngFldName = cg.fieldName(searchField)
								}
							}
						} else {
							if n := f.FS(FeaturesAPIKind, FAPIFindForName); n == AFFDeleted {
								mngFldName = mdDeletedFieldName
								deletedFound = true
								elseStmt = jen.Else().BlockFunc(
									func(g *jen.Group) {
										if mngFldName == mdDeletedFieldName {
											g.Add(addSkipDeletedToQuery())
										}
									},
								)
							}
						}

						g.IfFunc(
							func(g *jen.Group) {
								if f.Type.Array == nil {
									//TODO add option to compare with empty string
									if !f.Type.NonNullable {
										if f.Type.Type == TipString {
											g.Id("query").Dot(f.Name).Op("!=").Nil().Op("&&").Op("*").Id("query").Dot(f.Name).Op("!=").Lit("")
										} else {
											g.Id("query").Dot(f.Name).Op("!=").Nil()
										}
									} else {
										switch f.Type.Type {
										case TipString:
											g.Id("query").Dot(f.Name).Op("!=").Lit("")
										default:
											g.True()
										}
									}
								} else {
									g.Len(jen.Id("query").Dot(f.Name)).Op("!=").Lit(0)
								}
							},
						).BlockFunc(
							func(g *jen.Group) {
								if f.Type.Array != nil {
									g.Id("q").Index(jen.Lit(mngFldName)).Op("=").Qual(bsonPackage, "M").Values(
										jen.Dict{
											jen.Lit("$in"): jen.Id("query").Dot(f.Name),
										},
									)
								} else if f.Type.Map != nil {
									g.Id("arr").Op(":=").Make(jen.Qual(bsonPackage, "A"), jen.Len(jen.Id("query").Dot(f.Name)))
									g.Id("i").Op(":=").Lit(0)
									g.For(jen.List(jen.Id("k"), jen.Id("val")).Op(":=").Range().Id("query").Dot(f.Name)).Block(
										jen.Id("arr").Index(jen.Id("i")).Op("=").Qual(bsonPackage, "M").Values(
											jen.Dict{jen.Lit(mngFldName).Op("+").Lit(".").Op("+").Id("k"): jen.Id("val")},
										),
										jen.Id("i").Op("++"),
									)
									g.Id("q").Index(jen.Lit("$and")).Op("=").Id("arr")
								} else {
									pref := jen.Id("q").Index(jen.Lit(mngFldName))
									if td, ok := cg.desc.FindType(f.Type.Type); ok && td.entry != nil {
										for _, field := range td.entry.GetFields(true, true) {
											if !field.HasModifier(AttrModifierCalculated) {
												g.Id("q").Index(
													jen.Lit(
														fmt.Sprintf(
															"%s.%s",
															mngFldName,
															cg.fieldName(field),
														),
													),
												).Op("=").Id("query").Dot(f.Name).Dot(field.Name)
											}
										}
									} else {
										addToExisting := false
										if _, ok := possiblyFilledFields[mngFldName]; ok {
											g.Var().Id("op").Qual(bsonPackage, "M")
											g.If(
												jen.List(
													jen.Id("o"),
													jen.Id("ok"),
												).Op(":=").Id("q").Index(jen.Lit(mngFldName)).Assert(jen.Qual(bsonPackage, "M")),
												jen.Id("ok"),
											).Block(jen.Id("op").Op("=").Id("o")).Else().Block(
												jen.Id("op").Op("=").Qual(bsonPackage, "M").Values(),
												jen.Id("q").Index(jen.Lit(mngFldName)).Op("=").Id("op"),
											)
											pref = jen.Id("op")
											addToExisting = true
										}
										switch op {
										case AFTEqual:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$eq")).Op("=").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Id("query").Dot(f.Name)
											}
										case AFTNotEqual:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$ne")).Op("=").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$ne"): jen.Id("query").Dot(f.Name),
													},
												)
											}
										case AFTGreaterThan:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$gt")).Op("=").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$gt"): jen.Id("query").Dot(f.Name),
													},
												)
											}
										case AFTGreaterThanOrEqual:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$gte")).Op("=").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$gte"): jen.Id("query").Dot(f.Name),
													},
												)
											}
										case AFTLessThan:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$lt")).Op("=").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$lt"): jen.Id("query").Dot(f.Name),
													},
												)
											}
										case AFTLessThanOrEqual:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$lte")).Op("=").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$lte"): jen.Id("query").Dot(f.Name),
													},
												)
											}
										case AFTStartsWith:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$regex")).Op("=").Lit("^").Op("+").Op("*").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$regex"): jen.Lit("^").Op("+").Op("*").Id("query").Dot(f.Name),
													},
												)
											}
										case AFTStartsWithIgnoreCase:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$regex")).Op("=").Lit("^").Op("+").Op("*").Id("query").Dot(f.Name)
												g.Add(pref).Index(jen.Lit("$options")).Op("=").Lit("i")
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$regex"):   jen.Lit("^").Op("+").Op("*").Id("query").Dot(f.Name),
														jen.Lit("$options"): jen.Lit("i"),
													},
												)
											}
										case AFTContains:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$regex")).Op("=").Op("*").Id("query").Dot(f.Name)
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$regex"): jen.Op("*").Id("query").Dot(f.Name),
													},
												)
											}
										case AFTContainsIgnoreCase:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$regex")).Op("=").Op("*").Id("query").Dot(f.Name)
												g.Add(pref).Index(jen.Lit("$options")).Op("=").Lit("i")
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$regex"):   jen.Op("*").Id("query").Dot(f.Name),
														jen.Lit("$options"): jen.Lit("i"),
													},
												)
											}
										case AFTIsNull, AFTIsNotNull:
											ops := [...]string{"$eq", "$ne"}
											idx := 0
											if op == AFTIsNotNull {
												idx = 1
											}
											if addToExisting {
												g.If(jen.Op("*").Id("query").Dot(f.Name)).
													Block(jen.Add(pref).Index(jen.Lit(ops[idx])).Op("=").Nil()).
													Else().Block(jen.Add(pref).Index(jen.Lit(ops[1-idx])).Op("=").Nil())
											} else {
												g.If(jen.Op("*").Id("query").Dot(f.Name)).
													Block(
														jen.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
															jen.Dict{
																jen.Lit(ops[idx]): jen.Nil(),
															},
														),
													).Else().
													Block(
														jen.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
															jen.Dict{
																jen.Lit(ops[1-idx]): jen.Nil(),
															},
														),
													)
											}
										case AFTExists, AFTNotExists:
											if addToExisting {
												g.Add(pref).Index(jen.Lit("$exists").Op("=").Op("*").Id("query").Dot(f.Name))
											} else {
												g.Add(pref).Op("=").Qual(bsonPackage, "M").Values(
													jen.Dict{
														jen.Lit("$exists"): jen.Op("*").Id("query").Dot(f.Name),
													},
												)
											}
										case AFTIgnore:
											if mngFldName == mdDeletedFieldName && f.Type.Type == TipBool {
												arg := jen.Id("query").Dot(f.Name)
												if !f.Type.NonNullable {
													arg = jen.Op("*").Id("query").Dot(f.Name)
												}
												g.If(jen.Op("!").Add(arg)).Block(addSkipDeletedToQuery())
											} else {
												cg.desc.AddError(
													fmt.Errorf(
														"at %v: comparision type %s can be used only with bool type for _deleted_ field",
														f.Pos,
														op,
													),
												)
												return
											}
										default:
											cg.desc.AddError(fmt.Errorf("at %v: undefined comparision type: %s", f.Pos, op))
											return
										}
										possiblyFilledFields[mngFldName] = struct{}{}
									}
								}
							},
						).Add(elseStmt)
					}
					if e.BaseTypeName != "" {
						g.Add(cg.addDescendantsToQuery(e, "q"))
					}
					if !deletedFound {
						g.Add(addSkipDeletedToQuery())
					}
					g.Return(
						jen.List(jen.Id("q"), jen.Nil()),
					)
				},
			).Line()

			cg.b.Functions.Add(f)
		}
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("query").Op("*").Id(it.Name),
		).Parens(jen.List(jen.Index().Op("*").Id(name), jen.Error())).BlockFunc(
			func(g *jen.Group) {
				g.Id("ret").Op(":=").Index().Op("*").Id(name).Values(jen.Dict{})
				g.List(jen.Id("q"), jen.Id("_")).Op(":=").Id(EngineVar).Dot(generatorFuncName).Params(jen.Id("query"))
				if postprocessorMethod != "" {
					g.List(jen.Id("q"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(postprocessorMethod).Params(
						jen.Id("ctx"),
						jen.Id("query"),
						jen.Id("q"),
					)
					g.Add(returnIfErrValue(jen.Nil()))
				}
				if f, ok := e.Features.Get(mongoFeatures, mfSortField); ok {
					fld, ok := f.(*Field)
					if !ok {
						cg.desc.AddError(
							fmt.Errorf(
								"at %v: internal error in mongo:listAll: sort feature is not a field: %T",
								e.Pos,
								f,
							),
						)
						return
					}
					order := 1
					if fld.FB(mongoFeatures, mfSortDesc) {
						order = -1
					}

					g.Id("op").Op(":=").Qual(optionsPackage, "Find").Call().Dot("SetSort").Params(
						jen.Qual(bsonPackage, "M").Values(
							jen.Dict{
								jen.Lit(cg.fieldName(fld)): jen.Lit(order),
							},
						),
					)
				} else {
					g.Id("op").Op(":=").Qual(optionsPackage, "Find").Call()
				}
				g.List(
					jen.Id("curr"),
					jen.Id("err"),
				).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
					jen.Id(
						e.FS(
							mongoFeatures,
							mfCollectionConst,
						),
					),
				).Dot("Find").Params(
					jen.Id("ctx"),
					jen.Id("q"),
					jen.Id("op"),
				)
				g.Add(returnIfErrValue(jen.Nil()))
				g.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx"))
				g.Id("err").Op("=").Op("curr").Dot("All").Params(jen.Id("ctx"), jen.Op("&").Id("ret"))
				g.Return(
					jen.List(jen.Id("ret"), jen.Id("err")),
				)
			},
		).Line()

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
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual(
			"context",
			"Context",
		),
	).Parens(
		jen.List(ret, jen.Id("err").Error()),
	).Block(
		jen.Id("id").Op("=").Add(cg.b.goEmptyValue(e.GetIdField().Type)).Line().
			Return(),
	).Line()
	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateCongifLoadFunc(e *Entity) error {
	//TODO: add check that object is nil and initialize it in this case
	name := e.Name
	fname := cg.desc.GetMethodName(MethodLoad, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual(
			"context",
			"Context",
		),
	).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.List(jen.Id("ret"), jen.Id("_")).Op(":=").Id(EngineVar).Dot(
			cg.b.Descriptor.GetMethodName(
				MethodInit,
				e.Name,
			),
		).Params(jen.Id("ctx")),
		jen.List(
			jen.Id("curr"),
			jen.Id("err"),
		).Op(":=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
			jen.Id(
				e.FS(
					mongoFeatures,
					mfCollectionConst,
				),
			),
		).Dot("Find").Params(
			jen.Id("ctx"),
			jen.Qual(bsonPackage, "M").Values(),
		),
		jen.Add(returnIfErrValue(jen.Nil())),
		jen.Defer().Id("curr").Dot("Close").Params(jen.Id("ctx")),
		jen.For(jen.Id("curr").Dot("Next").Call(jen.Id("ctx"))).Block(
			jen.Id("idVal").Op(":=").Id("curr").Dot("Current").Dot("Lookup").Call(jen.Lit("_id")).Dot("StringValue").Call(),
			jen.Switch(jen.Id("idVal")).BlockFunc(
				func(sg *jen.Group) {
					for _, f := range e.Fields {
						// isPointer := f.FB(FeatGoKind, FCGPointer)
						ft, complex := cg.desc.FindType(f.Type.Type)
						sg.Case(jen.Lit(f.Name)).BlockFunc(
							func(g *jen.Group) {
								g.Id("curr").Dot("Current").Dot("Lookup").Call(jen.Lit("value")).Dot("Unmarshal").Call(jen.Op("&").Id("ret").Dot(f.Name))
								if /*isPointer &&*/ complex {
									engVar := cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCEngineVar)
									ent := ft.Entity()
									if ent == nil {
										cg.desc.AddError(fmt.Errorf("at %v: only Entity can be used here", ft.pos))
									} else {
										g.If(jen.Id("ret").Dot(f.Name).Op("==").Nil()).Block(
											jen.List(
												jen.Id("ret").Dot(f.Name),
												jen.Id("_"),
											).Op("=").Add(engVar).Dot(cg.b.Descriptor.GetMethodName(MethodInit, ent.Name)).Params(
												jen.Id("ctx"),
											),
										)
									}
								}
							},
						)
					}
				},
			),
		),
		jen.Return(
			jen.List(jen.Id("ret"), jen.Id("err")),
		),
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

func (cg *MongoGenerator) generateConfigSaveFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodSave, name)
	resultName := "_" //"ur"

	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("o").Op("*").Id(name),
	).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).BlockFunc(
		func(g *jen.Group) {
			// g.Var().Id(resultName).Op("*").Qual(mongoPackage, UpdateResult)
			g.Var().Err().Error()
			g.Id("opts").Op(":=").Qual(optionsPackage, "Update").Call().Dot("SetUpsert").Call(jen.Lit(true))
			for _, f := range e.Fields {
				g.List(
					jen.Id(resultName),
					jen.Id("err"),
				).Op("=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
					jen.Id(
						e.FS(
							mongoFeatures,
							mfCollectionConst,
						),
					),
				).Dot("UpdateOne").Params(
					jen.Id("ctx"),
					jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Lit(f.Name)}),
					jen.Qual(bsonPackage, "M").Values(
						jen.Dict{
							jen.Lit("$set"): jen.Qual(
								bsonPackage,
								"M",
							).Values(jen.Dict{jen.Lit("value"): jen.Id("o").Dot(f.Name)}),
						},
					),
					jen.Id("opts"),
				)

			}

			g.Add(returnIfErrValue(jen.Nil()))

			g.Return(
				jen.List(jen.Id("o"), jen.Id("err")),
			)
		},
	).Line()

	cg.b.Functions.Add(f)
	for _, fld := range e.Fields {
		if fld.Annotations.GetBoolAnnotationDef(AnnotationConfig, AnnCfgMutable, false) {
			fname = cg.desc.GetMethodName(MethodSave, name) + fld.Name
			f = jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
				jen.Id("ctx").Qual("context", "Context"),
				jen.Id("o").Add(fld.Features.Stmt(FeatGoKind, FCGAttrType)),
			).Parens(jen.Err().Error()).BlockFunc(
				func(g *jen.Group) {
					g.List(
						jen.Id("_"),
						jen.Id("err"),
					).Op("=").Id(EngineVar).Dot(engineMongo).Dot("Collection").Params(
						jen.Id(
							e.FS(
								mongoFeatures,
								mfCollectionConst,
							),
						),
					).Dot("UpdateOne").Params(
						jen.Id("ctx"),
						jen.Qual(bsonPackage, "M").Values(jen.Dict{jen.Lit("_id"): jen.Lit(fld.Name)}),
						jen.Qual(bsonPackage, "M").Values(
							jen.Dict{
								jen.Lit("$set"): jen.Qual(
									bsonPackage,
									"M",
								).Values(jen.Dict{jen.Lit("value"): jen.Id("o")}),
							},
						),
						jen.Qual(optionsPackage, "Update").Call().Dot("SetUpsert").Call(jen.Lit(true)),
					)
					g.Return(jen.Id("err"))
				},
			).Line()
			cg.b.Functions.Add(f)
		}
	}

	return nil
}

func (cg *MongoGenerator) collectionName(e *Entity) string {
	key := e.Pckg.Name + e.Name
	cn, ok := cg.collections[key]
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

		if ann, ok := t.Annotations[mongoAnnotation]; ok {
			if n, ok := ann.GetStringTag(mongoAnnotationTagName); ok {
				cn = n
			}
		}
		if cg.usedCollections[cn] && t.BaseTypeName == "" && !t.HasModifier(TypeModifierExtendable) {
			cg.desc.AddWarning(fmt.Sprintf("mongo: collection duplicate: %s", cn))
		}
		cg.collections[key] = cn
		cg.usedCollections[cn] = true
	}
	return cn
}

func (cg *MongoGenerator) fieldName(f *Field) string {
	if n, ok := f.Features.GetString(FeaturesDBKind, FCGName); ok {
		return n
	}
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
			jen.Dict{
				jen.Lit("$in"): jen.Qual(bsonPackage, "A").ValuesFunc(
					func(g *jen.Group) {
						g.Id(e.FS(FeatGoKind, FCGDerivedTypeNameConst))
						for _, d := range desc.([]*Entity) {
							g.Id(d.FS(FeatGoKind, FCGDerivedTypeNameConst))
						}
					},
				),
			},
		)
	}
	return jen.Id(queryVar).Index(jen.Lit(name)).Op("=").Id(e.FS(FeatGoKind, FCGDerivedTypeNameConst))
}

func (cg *MongoGenerator) init() {
	if cg.inited {
		return
	}
	if cg.collections == nil {
		cg.collections = map[string]string{}
	}
	if cg.usedCollections == nil {
		cg.usedCollections = map[string]bool{}
	}
	cg.prefixCollectionName = optPrefixWithPackage
	cg.useBaseCollectionForDerived = true

	cg.inited = true
}
