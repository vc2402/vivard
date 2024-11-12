package gen

import (
	"errors"
	"fmt"
	"github.com/alecthomas/participle/lexer"
	"strconv"
	"strings"
	"unicode"

	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
)

type GQLOperationKind int

const GQLGeneratorName = "GraphQL"

const (
	GQLOperationGet GQLOperationKind = iota
	GQLOperationSet
	GQLOperationCreate
	GQLOperationList
	GQLOperationLookup
	GQLOperationDelete
	GQLOperationFind
	GQLOperationBulkCreate
	GQLOperationBulkSet

	GQLOperationLast
)

var gqlOperationsNamesTemplates = [GQLOperationLast]string{
	"get%s",
	"set%s",
	"create%s",
	"list%s",
	"lookup%s",
	"delete%s",
	"find%s",
	"create%ss",
	"set%ss",
}

const gqlSetNullInputTemplate = "%sSetNull"

const gqlPackage = "github.com/graphql-go/graphql"

const (
	gqlDescriptorVarName = "gqlDesc"

	GQLAnnotation              = "gql"
	GQLAnnotationNameTag       = "name"
	GQLAnnotationSkipTag       = "skip"
	GQLAnnotationGetTag        = "get"
	GQLAnnotationSetTag        = "set"
	GQLAnnotationCreateTag     = "create"
	GQLAnnotationListTag       = "list"
	GQLAnnotationLookupTag     = "lookup"
	GQLAnnotationDeleteTag     = "delete"
	GQLAnnotationQueryTag      = "query"
	GQLAnnotationMutationTag   = "mutation"
	GQLAnnotationBulkCreateTag = "bulkCreate"
	GQLAnnotationBulkSetTag    = "bulkSet"
	// GQLAnnotationReadonlyTag do not generate GraphQL mutations for type
	GQLAnnotationReadonlyTag = "readonly"

	gqlTagJSON = "json"
)

const (
	GQLFeatures        = "gql"
	GQLFTypeTag        = "type"
	GQLFUseDefinedType = "use-defined-type"
	GQLFArgTypeTag     = "argType"
	GQLFIdTypeTag      = "idType"
	GQLFInputTypeName  = "inputType"
	GQLFMethodName     = "name"
	// GQLFMethodType - GQLFMethodTypeMutation or GQLFMethodTypeQuery
	GQLFMethodType         = "method-type"
	GQLFMethodTypeMutation = "mutation"
	GQLFMethodTypeQuery    = "query"
	// GQLFIDOnly use only id for embedded type
	GQLFIDOnly   = "id-only"
	GQLFReadonly = "redonly"

	// GQLFMethodResultType - code generator for method's result type
	GQLFMethodResultType = "result-type"
	// GQLFMethodResultTypeName - feature for method's result type name
	GQLFMethodResultTypeName = "result-type-name"
	// GQLGenerateUnionType - feature for Package; returns function that generates union type
	//  and inserts generator in given Package's init procedure
	//  function gets union name as first param and types (*Entity) that should be parts of union
	GQLGenerateUnionType = "generate-union-type"
	// GQLFSetNullInputField name of boolean field in input type for setting null for nullbale field
	GQLFSetNullInputField = "set-null-input-field"
)

var GQLOperationsAnnotationsTags = [GQLOperationLast]string{
	GQLAnnotationGetTag,
	GQLAnnotationSetTag,
	GQLAnnotationCreateTag,
	GQLAnnotationListTag,
	GQLAnnotationLookupTag,
	GQLAnnotationDeleteTag,
	GQLAnnotationBulkCreateTag,
	GQLAnnotationBulkSetTag,
}

type GQLOptions struct {
	UsePackageNameInTypeNames bool
}

type GQLGenerator struct {
	desc    *Package
	b       *Builder
	options GQLOptions
}

func init() {
	RegisterPlugin(&GQLGenerator{options: GQLOptions{UsePackageNameInTypeNames: true}})
}

func (cg *GQLGenerator) Name() string {
	return GQLGeneratorName
}

func (cg *GQLGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	if ann.Name == GQLAnnotation {
		//TODO check annotation format
		return true, nil
	}
	return false, nil
}

func (cg *GQLGenerator) Prepare(desc *Package) error {
	cg.desc = desc
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			an, ok := t.Annotations[GQLAnnotation]
			if !ok {
				t.Annotations.AddTag(GQLAnnotation, GQLAnnotationNameTag, cg.GetGQLEntityTypeName(t.Name))
			}
			an = t.Annotations[GQLAnnotation]
			if s, ok := an.GetBoolTag(GQLAnnotationSkipTag); (!ok || !s) && !t.HasModifier(TypeModifierExternal) {
				if _, ok := an.GetStringTag(GQLAnnotationNameTag); !ok {
					t.Annotations.AddTag(GQLAnnotation, GQLAnnotationNameTag, cg.GetGQLEntityTypeName(t.Name))
				}
				for i := GQLOperationGet; i < GQLOperationLast; i++ {
					if skip, ok := an.GetBoolTag(GQLOperationsAnnotationsTags[i]); !ok && !skip {
						if i == GQLOperationList && !t.IsDictionary() {
							t.Features.Set(GQLFeatures, GQLOperationsAnnotationsTags[i], false)
						} else if _, ok := an.GetStringTag(GQLOperationsAnnotationsTags[i]); !ok {
							t.Features.Set(GQLFeatures, GQLOperationsAnnotationsTags[i], cg.GetGQLOperationName(t, i))
						}
					} else {
						t.Features.Set(GQLFeatures, GQLOperationsAnnotationsTags[i], false)
					}
				}
				t.Features.Set(GQLFeatures, GQLFTypeTag, cg.GetGQLEntityTypeName(t.Name))
				t.Features.Set(GQLFeatures, GQLFInputTypeName, cg.GetGQLInputTypeName(t.Name))
				if t.HasModifier(TypeModifierTransient) || t.HasModifier(TypeModifierEmbeddable) ||
					(!t.HasModifier(TypeModifierConfig) && t.Annotations[AnnotationConfig] != nil) {
					t.Features.Set(FeaturesAPIKind, FAPILevel, FAPILTypes)
				}
				if t.BaseTypeName != "" {
					f := t.GetBaseField()
					f.Annotations.AddTag(GQLAnnotation, GQLAnnotationNameTag, cg.GetGQLFieldName(f))
				}
				if !t.HasModifier(TypeModifierSingleton) {
					for _, f := range t.GetFields(true, true) {
						if s, ok := f.Annotations.GetBoolAnnotation(GQLAnnotation, GQLAnnotationSkipTag); ok && s {
							f.Features.Set(FeaturesAPIKind, FCIgnore, true)
						}
						if f.HasModifier(AttrModifierAuxiliary) {
							f.Features.Set(FeaturesAPIKind, FCIgnore, true)
						}
						if f.FB(FeaturesAPIKind, FCIgnore) {
							continue
						}
						fieldName, ok := f.Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag)
						if !ok {
							fieldName = cg.GetGQLFieldName(f)
							f.Annotations.AddTag(GQLAnnotation, GQLAnnotationNameTag, fieldName)
						}
						tip := cg.GetGQLTypeName(f.Type)
						f.Features.Set(GQLFeatures, GQLFTypeTag, tip)

						if f.Type.Complex {
							if f.Type.Array != nil {
								//TODO: add special handling for lists
							} else if f.Type.Map != nil {
								//TODO: add special handling for maps
							} else {
								ct, ok := f.Parent().File.Pckg.FindType(f.Type.Type)
								if !ok {
									return fmt.Errorf("gql: at %v: undefined type for %s", f.Pos, f.Name)
								}
								if ct.entry != nil {
									if idfld := ct.entry.GetIdField(); idfld != nil {
										tip = cg.GetGQLTypeName(idfld.Type)
										if ct.entry == t {
											//TODO check for recursive types
											f.Features.Set(GQLFeatures, GQLFIDOnly, true)
											f.Features.Set(GQLFeatures, GQLFTypeTag, tip)
										}
									}
								}
							}
						}
						f.Features.Set(GQLFeatures, GQLFArgTypeTag, tip)
						if !f.Type.NonNullable {
							f.Features.Set(GQLFeatures, GQLFSetNullInputField, fmt.Sprintf(gqlSetNullInputTemplate, fieldName))
						}
					}
				}
				for _, m := range t.Methods {
					mname, ok := m.Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag)
					if !ok {
						mname = cg.GetGQLMethodName(t, m)
					}
					m.Features.Set(GQLFeatures, GQLFMethodName, mname)
					mtype := GQLFMethodTypeMutation
					if q, ok := m.Annotations.GetBoolAnnotation(GQLAnnotation, GQLAnnotationQueryTag); ok && q {
						mtype = GQLFMethodTypeQuery
					}
					m.Features.Set(GQLFeatures, GQLFMethodType, mtype)
					for _, p := range m.Params {
						tip := cg.GetGQLTypeName(p.Type, true)
						p.Features.Set(GQLFeatures, GQLFTypeTag, tip)
						p.Features.Set(GQLFeatures, GQLFInputTypeName, tip)
					}
				}
			} else {
				t.Features.Set(FeaturesAPIKind, FAPILevel, FAPILIgnore)
			}
		}
	}
	return nil
}

func (cg *GQLGenerator) Generate(b *Builder) (err error) {
	cg.desc = b.Descriptor
	cg.b = b
	b.Generator.Id(gqlDescriptorVarName).Op(":=").Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
		jen.Op("*").Qual(
			VivardPackage,
			"GQLEngine",
		),
	).Dot("Descriptor").Params().Line()
	for _, t := range b.File.Entries {
		level, ok := t.Features.GetString(FeaturesAPIKind, FAPILevel)
		if !ok || level != FAPILIgnore {
			err = cg.generateGQLTypes(t)
			if err != nil {
				return err
			}
			// if !t.FB(FeaturesCommonKind, FCReadonly) {
			err = cg.generateInputTypeGenerator(t)
			if err != nil {
				return err
			}
			// }
		}

		if !ok || level == FAPILAll {
			f := t.GetIdField()
			if f != nil {
				err = cg.generateGQLQuery(t)
				if err != nil {
					return err
				}
				err = cg.generateGQLLookupQuery(t)
				if err != nil {
					return err
				}
				if t.IsDictionary() {
					err = cg.generateGQLListQuery(t)
					if err != nil {
						return err
					}
				}
				err = cg.generateGQLFindQuery(t)
				if err != nil {
					return err
				}
				if !t.FB(FeaturesCommonKind, FCReadonly) {
					err = cg.generateGQLSetMutation(t)
					if err != nil {
						return err
					}
					err = cg.generateGQLCreateMutation(t)
					if err != nil {
						return err
					}
					err = cg.generateGQLDeleteMutation(t)
					if err != nil {
						return err
					}
				}
				err = cg.generateGQLBulkMethods(t)
				if err != nil {
					return err
				}
			}

			if t.HasModifier(TypeModifierConfig) {
				err = cg.generateGQLConfigSetMutation(t)
				if err != nil {
					return err
				}
				err = cg.generateGQLConfigQuery(t)
				if err != nil {
					return err
				}
			}
		}
		err = cg.generateGQLMethods(t)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cg *GQLGenerator) generateGQLTypes(e *Entity) error {
	if e.Pckg.engineless {
		return fmt.Errorf(
			"at %v: GraphQL requires Engine for types registration; please remove annotation %s",
			e.Pckg.pos,
			AnnotationEngineless,
		)
	}
	if name, ok := e.Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag); ok {
		goTypeName := e.Name
		fname := fmt.Sprintf("%sTypeGenerator", name)
		gqlFields := jen.Dict{}
		for _, f := range e.GetFields(true, true) {
			fieldName, ok := f.Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag)
			if !ok {
				continue
			}

			var t *jen.Statement
			var err error
			if tn := f.FS(GQLFeatures, GQLFUseDefinedType); tn != "" {
				t = cg.generateTypeLookupStatement(tn, false)
			} else {
				t, err = cg.getGQLType(f.Type, false, f.HasModifier(AttrModifierEmbeddedRef) || f.FB(GQLFeatures, GQLFIDOnly))
				if err != nil {
					return err
				}
			}

			gqlFields[jen.Lit(fieldName)] = jen.Op("&").Qual(gqlPackage, "Field").Values(
				jen.Dict{
					jen.Id("Type"): t,
					jen.Id("Resolve"): jen.Func().Params(
						jen.Id("p").Qual(
							gqlPackage,
							"ResolveParams",
						),
					).Parens(jen.List(jen.Interface(), jen.Error())).
						// Block(resolve),
						BlockFunc(
							func(g *jen.Group) {
								g.Id("obj").Op(":=").Id("p").Dot("Source").Assert(jen.Op("*").Id(goTypeName))
								if f.HasModifier(AttrModifierEmbeddedRef) || f.FB(GQLFeatures, GQLFIDOnly) {
									g.Return(jen.Id("obj").Dot(f.FS(FeatGoKind, FCGName)), jen.Nil())
									return
								}
								if !f.Type.NonNullable && !f.HasModifier(AttrModifierCalculated) {
									g.If(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCIsNullCode, "obj")).Block(
										jen.Return(jen.Nil(), jen.Nil()),
									)
								}
								if f.Type.Map != nil {
									if f.Type.Map.KeyType != TipString {
										cg.desc.AddError(fmt.Errorf("at %v: GQL: only string can be used as Key for Maps", f.Pos))
										return
									}
									var fn string
									switch f.Type.Map.ValueType.Type {
									case TipString:
										fn = "MapStringStringToArrKV"
									case TipInt:
										fn = "MapStringIntToArrKV"
									default:
										cg.desc.AddError(
											fmt.Errorf(
												"at %v: GQL: only string and int can be used as Maps value currently",
												f.Pos,
											),
										)
										return
									}
									g.List(jen.Id("val"), jen.Id("_")).Op(":=").Add(
										cg.desc.CallCodeFeatureFunc(
											f,
											FeaturesCommonKind,
											FCGetterCode,
											"obj",
											jen.Id("p").Dot("Context"),
											true,
										),
									)
									g.Return(
										jen.Qual(VivardPackage, fn).Params(jen.Id("val")),
										jen.Nil(),
									)
								} else {
									var enum *Enum
									var enumArray *Enum
									if f.Type.Array != nil {
										if dt, ok := cg.desc.FindType(f.Type.Array.Type); ok && dt.enum != nil {
											enumArray = dt.enum
										}
									} else if dt, ok := cg.desc.FindType(f.Type.Type); ok && dt.enum != nil {
										enum = dt.enum
									}
									if enumArray != nil {
										g.Id("items").Op(":=").Add(
											cg.desc.CallCodeFeatureFunc(
												f,
												FeaturesCommonKind,
												FCGetterCode,
												"obj",
												jen.Id("p").Dot("Context"),
												false,
											),
										)
										g.Id("result").Op(":=").Make(
											jen.Index().Add(cg.b.GoType(&TypeRef{Type: enumArray.AliasForType})),
											jen.Len(jen.Id("items")),
										)
										g.For(jen.List(jen.Id("i"), jen.Id("item")).Op(":=").Range().Id("items")).Block(
											jen.Id("result").Index(jen.Id("i")).Op("=").Add(cg.b.GoType(&TypeRef{Type: enumArray.AliasForType})).Parens(jen.Id("item")),
										)
										g.Return(jen.List(jen.Id("items")), jen.Nil())
									} else if enum != nil {
										g.Return(
											jen.List(
												cg.b.GoType(&TypeRef{Type: enum.AliasForType}).Parens(
													cg.desc.CallCodeFeatureFunc(
														f,
														FeaturesCommonKind,
														FCGetterCode,
														"obj",
														jen.Id("p").Dot("Context"),
														false,
													),
												),
												jen.Nil(),
											),
										)
									} else {
										g.Return(
											jen.Add(
												cg.desc.CallCodeFeatureFunc(
													f,
													FeaturesCommonKind,
													FCGetterCode,
													"obj",
													jen.Id("p").Dot("Context"),
													true,
												),
											),
										)
									}
								}
							},
						),
				},
			)
			cg.desc.AddTag(f, gqlTagJSON, fieldName)
		}
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Qual(gqlPackage, "Output").Block(
			jen.Return(
				jen.Qual(gqlPackage, "NewObject").Call(
					jen.Qual(gqlPackage, "ObjectConfig").Values(
						jen.Dict{
							jen.Id("Name"):   jen.Lit(name),
							jen.Id("Fields"): jen.Qual(gqlPackage, "Fields").Values(gqlFields),
						},
					),
				),
			),

			// graphql.NewObject(
			// graphql.ObjectConfig{
			// 	Name: "CheckList",
			// 	Fields: graphql.Fields{
			// 		"id": &graphql.Field{
			// 			Type: graphql.NewNonNull(graphql.String),
			// 			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 				r := p.Source.(*checkList)
			// 				return r.ID, nil
			// 			},
			// 		},
			// 	}
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddTypeGenerator").Params(
			jen.Lit(name),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

func (cg *GQLGenerator) generateInputTypeGenerator(e *Entity) error {
	// var positionInputType = graphql.NewInputObject(
	// 	graphql.InputObjectConfig{
	// 		Name: "PositionInput",
	// 		Fields: graphql.InputObjectConfigFieldMap{
	// 			"lat": &graphql.InputObjectFieldConfig{
	// 				Type: graphql.Float,
	// 			},
	// 			"lon": &graphql.InputObjectFieldConfig{
	// 				Type: graphql.Float,
	// 			},
	// 			"timestamp": &graphql.InputObjectFieldConfig{
	// 				Type:         graphql.Int,
	// 				DefaultValue: 0,
	// 			},
	// 		},
	// 	},
	// )
	if _, ok := e.Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag); ok {
		fname := fmt.Sprintf("%sInputGenerator", e.Name)
		typeName := cg.GetGQLInputTypeName(e.Name)
		gqlFields := jen.Dict{}
		for _, f := range e.GetFields(true, true) {
			fieldName, ok := f.Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag)
			// let's leave readonly fields as input for js simplicity
			if !ok /*|| f.FB(FeaturesCommonKind, FCReadonly)*/ || f.HasModifier(AttrModifierCalculated) {
				continue
			}
			t, err := cg.getGQLType(
				f.Type,
				!(f.IsIdField() && f.HasModifier(AttrModifierIDAuto)),
				f.HasModifier(AttrModifierEmbeddedRef) || f.FB(GQLFeatures, GQLFIDOnly),
				true, //input
			)
			if err != nil {
				return err
			}

			gqlFields[jen.Lit(fieldName)] = jen.Op("&").Qual(gqlPackage, "InputObjectFieldConfig").Values(
				jen.Dict{
					jen.Id("Type"): t,
				},
			)
			if setNullField := f.FS(GQLFeatures, GQLFSetNullInputField); setNullField != "" {
				gqlFields[jen.Lit(setNullField)] = jen.Op("&").Qual(gqlPackage, "InputObjectFieldConfig").Values(
					jen.Dict{
						jen.Id("Type"): jen.Qual(gqlPackage, "Boolean"),
					},
				)
			}
		}
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Qual(gqlPackage, "Input").Block(
			jen.Return(
				jen.Qual(gqlPackage, "NewInputObject").Call(
					jen.Qual(gqlPackage, "InputObjectConfig").Values(
						jen.Dict{
							jen.Id("Name"):   jen.Lit(typeName),
							jen.Id("Fields"): jen.Qual(gqlPackage, "InputObjectConfigFieldMap").Values(gqlFields),
						},
					),
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddInputGenerator").Params(
			jen.Lit(typeName),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	cg.generateGQLInputTypeParser(e)
	return nil
}

func (cg *GQLGenerator) generateGQLQuery(t *Entity) error {
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationGet]); ok {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		idField := t.GetIdField()
		id := idField.Annotations.GetStringAnnotationDef(GQLAnnotation, GQLAnnotationNameTag, "id")
		fname := fmt.Sprintf("%sQueryGenerator", name)
		idtype, _ := cg.getGQLType(idField.Type)
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.Dict{
						jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
							jen.Op("*").Qual(
								VivardPackage,
								"GQLEngine",
							),
						).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
						jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(
							jen.Dict{
								jen.Lit(id): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
									jen.Dict{
										jen.Id("Type"): idtype,
									},
								),
							},
						),
						jen.Id("Resolve"): jen.Func().Params(
							jen.Id("p").Qual(
								gqlPackage,
								"ResolveParams",
							),
						).Parens(jen.List(jen.Interface(), jen.Error())).
							Block(
								jen.Id("id").Op(":=").Id("p").Dot("Args").Index(jen.Lit(id)).Assert(cg.b.GoType(idField.Type)).Line().
									Return(
										jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodGet, name)).Params(
											jen.Id("p").Dot("Context"),
											jen.Id("id"),
										),
									),
							),
					},
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddQueryGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
		// 	"config": &graphql.Field{
		// 	Type:        graphql.NewList(clientConfigType),
		// 	Description: "List all the available routes",
		// Args: graphql.FieldConfigArgument{
		// 	"id": &graphql.ArgumentConfig{
		// 		Type:         graphql.Int,
		// 		DefaultValue: -1,
		// 	},
		// },
		// 	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		//    id := p.Args["id"].(int)
		// 		log.Tracef("configQuery: resolve")
		// 		return loader.getConfig(p.Context, -1)
		// 	},
		// },
	}
	return nil
}

func (cg *GQLGenerator) generateGQLInputTypeParser(t *Entity) error {
	const arg = "arg"
	const obj = "obj"
	const idRequired = "idRequired"
	name := t.Name
	funcName := cg.getInputParserMethodName(name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(funcName).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id(arg).Interface(),
		jen.Id(obj).Op("*").Id(name),
		jen.Id(idRequired).Bool(),
	).Parens(jen.List(jen.Id("ret").Op("*").Id(name), jen.Err().Error())).Block(
		jen.If(
			jen.List(jen.Id("p"), jen.Id("ok")).Op(":=").Id(arg).Assert(jen.Map(jen.String()).Interface()),
			jen.Id("ok"),
		).BlockFunc(
			func(g *jen.Group) {
				g.If(jen.Id(obj).Op("==").Nil()).BlockFunc(
					func(g *jen.Group) {
						//TODO: add option for panic?
						//if t.HasModifier(TypeModifierExtendable) {
						//	g.Panic(jen.Lit("empty object for base type"))
						//} else {
						g.List(jen.Id(obj), jen.Id("err")).Op("=").Id(EngineVar).Dot(
							cg.desc.GetMethodName(
								MethodInit,
								name,
							),
						).Params(jen.Id("ctx"))
						g.Add(returnIfErr())
						//}
					},
				)
				for _, f := range t.GetFields(true, true) {
					fieldName, ok := f.Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag)
					if !ok || f.FB(FeaturesCommonKind, FCReadonly) || f.HasModifier(AttrModifierCalculated) {
						continue
					}
					engVar := cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCEngineVar)
					//TODO: check values for references
					// useParser := f.Annotations.GetInterfaceAnnotation(codeGeneratorAnnotation, AnnotationTagOneToManyType) != nil
					_, oneToMany := f.Features.GetEntity(FeaturesCommonKind, FCOneToManyType)
					_, manyToMany := f.Features.GetEntity(FeaturesCommonKind, FCManyToManyType)
					useParser := f.Type.Complex && !f.HasModifier(AttrModifierEmbeddedRef) && !f.FB(GQLFeatures, GQLFIDOnly)
					var assertion jen.Code
					if useParser {
						if oneToMany || manyToMany || f.Type.Array != nil {
							assertion = jen.Index().Interface()
						} else if f.Type.Map != nil {
							assertion = jen.Index().Interface()
						} else {
							assertion = jen.Interface()
						}
					} else {
						if f.Type.Array != nil {
							assertion = jen.Index().Interface()
						} else {
							tip := f.Type
							if f.Type.Type != "" {
								if tr, ok := cg.desc.FindType(f.Type.Type); ok && tr.Enum() != nil {
									tip = &TypeRef{Type: tr.Enum().AliasForType}
								}
							}
							assertion = cg.b.GoType(tip)

						}
					}
					stmt := jen.If(
						jen.List(jen.Id("val"), jen.Id("ok")).Op(":=").Id("p").Index(jen.Lit(fieldName)).Assert(assertion),
						jen.Id("ok"),
					).BlockFunc(
						func(g *jen.Group) {
							if useParser {
								if oneToMany || manyToMany {
									artype := jen.Index().Op("*").Id(f.Type.Array.Type)
									g.Id("values").Op(":=").Make(artype, jen.Len(jen.Id("val")))
									g.For(jen.List(jen.Id("i"), jen.Id("item")).Op(":=").Range().Id("val")).Block(
										jen.List(
											// jen.Id("obj").Dot(f.Name),
											jen.Id("v"),
											jen.Err(),
										).Op(":=").Add(engVar).Dot(cg.getInputParserMethodName(f.Type.Array.Type)).Params(
											jen.Id("ctx"),
											jen.Id("item"),
											jen.Nil(),
											jen.False(),
										),
										returnIfErrValue(jen.Nil()),
										jen.Id("values").Index(jen.Id("i")).Op("=").Id("v"),
									)
									g.Add(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCSetterCode, "obj", "values"))
								} else if f.Type.Array != nil {
									g.Var().Id("values").Add(f.Features.Stmt(FeatGoKind, FCGType))
									cg.addArrayInputParser(f, f.Type, g, jen.Id("val"), jen.Id("values"), 1)
									//artype := f.Features.Stmt(FeatGoKind, FCGType)
									//assertArrayItemType := cg.b.GoType(f.Type.Array)
									//resultArrayItem := jen.Id("v")
									//if f.Type.Array.Type != "" {
									//  if tr, ok := cg.desc.FindType(f.Type.Array.Type); ok && tr.Enum() != nil {
									//    assertArrayItemType = cg.b.GoType(&TypeRef{Type: tr.Enum().AliasForType})
									//    if cg.desc == tr.Enum().Pckg {
									//      resultArrayItem = jen.Id(tr.Enum().Name).Parens(resultArrayItem)
									//    } else {
									//      resultArrayItem = jen.Qual(tr.Enum().Pckg.fullPackage, tr.Enum().Name).Parens(resultArrayItem)
									//    }
									//  }
									//}
									//g.Id("values").Op(":=").Make(artype, jen.Len(jen.Id("val")))
									//g.For(jen.List(jen.Id("i"), jen.Id("item")).Op(":=").Range().Id("val")).Block(
									//  jen.List(
									//    // jen.Id("obj").Dot(f.Name),
									//    jen.Id("v"),
									//    jen.Id("ok"),
									//  ).Op(":=").Id("item").Assert(assertArrayItemType),
									//  jen.If(jen.Op("!").Id("ok")).Block(
									//    jen.Return(
									//      jen.Nil(),
									//      jen.Qual(
									//        "errors",
									//        "New",
									//      ).Params(
									//        jen.Lit(
									//          fmt.Sprintf(
									//            "problem while converting array item of type '%s'",
									//            f.Type.Array.Type,
									//          ),
									//        ),
									//      ),
									//    ),
									//  ),
									//  jen.Id("values").Index(jen.Id("i")).Op("=").Add(resultArrayItem),
									//)
									g.Add(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCSetterCode, "obj", "values"))
								} else if f.Type.Map != nil {
									if f.Type.Map.KeyType != TipString {
										cg.desc.AddError(fmt.Errorf("at %v: GQL: only string can be used as Key for Maps", f.Pos))
										return
									}
									var fn string
									switch f.Type.Map.ValueType.Type {
									case TipString:
										fn = "GQLArgToMapStringString"
									case TipInt:
										fn = "GQLArgToMapStringInt"
									default:
										cg.desc.AddError(
											fmt.Errorf(
												"at %v: GQL: only string and int can be used as Maps value currently",
												f.Pos,
											),
										)
										return
									}
									g.List(jen.Id("values"), jen.Err()).Op(":=").Qual(VivardPackage, fn).Params(jen.Id("val"))
									g.Add(returnIfErrValue(jen.Nil()))
									g.Add(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCSetterCode, "obj", "values"))
								} else {
									typeName := f.Type.Type
									if f.Type.Map != nil {
										if f.Type.Map.KeyType == TipString && f.Type.Map.ValueType.Type == TipString {
											typeName = vivard.KVStringString
										} else {
											cg.desc.AddError(
												fmt.Errorf(
													"at %v: map[%s]%s is not yet implemented in GQL",
													f.Pos,
													f.Type.Map.KeyType,
													f.Type.Map.ValueType.Type,
												),
											)
										}
									}
									// if f.Type.Array != nil {
									// 	typeName = f.Type.Array.Type
									// }

									g.List(jen.Id("v"), jen.Err()).Op(":=").Add(engVar).Dot(cg.getInputParserMethodName(typeName)).Params(
										jen.Id("ctx"),
										jen.Id("val"),
										jen.Nil(),
										jen.False(),
									)
									g.Add(returnIfErrValue(jen.Nil()))
									g.Add(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCSetterCode, "obj", "v"))

								}
							} else {
								// g.Id("obj").Dot(cg.b.GetMethodName(f, CGSetterMethod)).Parens(jen.Id("val"))
								if f.Type.Array != nil {
									g.Id("attr").Op(":=").Make(cg.b.GoType(f.Type), jen.Len(jen.Id("val")), jen.Len(jen.Id("val")))
									g.For(jen.List(jen.Id("i"), jen.Id("v")).Op(":=").Range().Id("val")).Block(
										jen.Id("attr").Index(jen.Id("i")).Op("=").Id("v").Assert(cg.b.GoType(f.Type.Array)),
									)
									g.Add(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCAttrSetCode, "obj", "attr"))
								} else {
									if f.HasModifier(AttrModifierEmbeddedRef) || f.FB(GQLFeatures, GQLFIDOnly) {
										//TODO: may be problem with changes tracking...
										g.Add(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCAttrSetCode))
									} else {
										var val any = "val"
										if f.Type.Type != "" {
											if tr, ok := cg.desc.FindType(f.Type.Type); ok && tr.Enum() != nil {
												if cg.desc == tr.Enum().Pckg {
													val = jen.Id(tr.Enum().Name).Parens(jen.Id("val"))
												} else {
													val = jen.Qual(tr.Enum().Pckg.fullPackage, tr.Enum().Name).Parens(jen.Id("val"))
												}
											}
										}
										g.Add(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCSetterCode, "obj", val))
									}
								}
							}
						},
					)
					if f.IsIdField() {
						stmt.Else().If(jen.Id(idRequired)).Block(
							jen.Return(jen.Nil(), jen.Qual("errors", "New").Params(jen.Lit("id should be set"))),
						)
					} else if setNullField := f.FS(GQLFeatures, GQLFSetNullInputField); setNullField != "" {
						stmt.Else().If(
							jen.List(
								jen.Id("setNull"),
								jen.Id("ok"),
							).Op(":=").Id("p").Index(jen.Lit(setNullField)).Assert(jen.Bool()),
							jen.Id("ok").Op("&&").Id("setNull"),
						).Block(cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCSetNullCode))
					}

					g.Add(stmt)
				}
			},
		).Else().Block(
			jen.Return(
				jen.Nil(), jen.Qual("fmt", "Errorf").Params(
					jen.Lit(fmt.Sprintf("%s: input should be 'map[string]interface{}' but got %%T", funcName)), jen.Id(arg),
				),
			),
		),
		jen.Return(jen.Id(obj), jen.Err()),
	).Line()
	cg.b.Functions.Add(f)
	// getIdFromInput func
	if idfld := t.GetIdField(); idfld != nil {
		f = jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(cg.getIdFromInputMethodName(name)).Params(jen.Id(arg).Interface()).Parens(
			jen.List(
				jen.Id("id").Add(cg.b.GoType(idfld.Type)),
				jen.Err().Error(),
			),
		).Block(
			jen.List(
				jen.Id("id"),
				jen.Id("ok"),
			).Op(":=").Id(arg).Assert(jen.Map(jen.String()).Interface()).Index(
				jen.Lit(
					idfld.Annotations.GetStringAnnotationDef(
						GQLAnnotation,
						GQLAnnotationNameTag,
						"id",
					),
				),
			).Assert(cg.b.GoType(idfld.Type)),
			jen.If(jen.Op("!").Id("ok")).Block(
				jen.Id("err").Op("=").Qual("errors", "New").Params(jen.Lit("no id provided")),
			),
			jen.Return(),
		).Line()
		cg.b.Functions.Add(f)
	}

	return nil
}

func (cg *GQLGenerator) addArrayInputParser(
	field *Field,
	ref *TypeRef,
	g *jen.Group,
	valueToParse *jen.Statement,
	valueToSave *jen.Statement,
	level int,
) {
	artype := cg.b.GoType(ref)
	assertArrayItemType := cg.b.GoType(ref.Array)
	levelString := strconv.Itoa(level)
	resultArrayItemName := "v" + levelString
	resultArrayItem := jen.Id(resultArrayItemName)
	indexVarName := "i" + levelString
	itemVarName := "it" + levelString
	if ref.Array.Type != "" {
		if tr, ok := cg.desc.FindType(ref.Array.Type); ok && tr.Enum() != nil {
			assertArrayItemType = cg.b.GoType(&TypeRef{Type: tr.Enum().AliasForType})
			if cg.desc == tr.Enum().Pckg {
				resultArrayItem = jen.Id(tr.Enum().Name).Parens(resultArrayItem)
			} else {
				resultArrayItem = jen.Qual(tr.Enum().Pckg.fullPackage, tr.Enum().Name).Parens(resultArrayItem)
			}
		}
	}
	g.Add(valueToSave).Op("=").Make(artype, jen.Len(valueToParse))
	g.For(jen.List(jen.Id(indexVarName), jen.Id(itemVarName)).Op(":=").Range().Add(valueToParse)).BlockFunc(
		func(ig *jen.Group) {
			if ref.Array.Array != nil {
				assertArrayItemType = jen.Index().Interface()
			}
			ig.List(
				// jen.Id("obj").Dot(f.Name),
				jen.Id(resultArrayItemName),
				jen.Id("ok"),
			).Op(":=").Id(itemVarName).Assert(assertArrayItemType)
			ig.If(jen.Op("!").Id("ok")).Block(
				jen.Return(
					jen.Nil(),
					jen.Qual(
						"fmt",
						"Errorf",
					).Params(
						jen.Lit(
							fmt.Sprintf(
								"problem while converting array item for field '%s' of type '%s': expecting type '%s', but got '%%T'",
								field.Name,
								field.Parent().Name,
								assertArrayItemType.GoString(),
							),
						),
						jen.Id(itemVarName),
					),
				),
			)
			if ref.Array.Array == nil {
				ig.Add(valueToSave).Index(jen.Id(indexVarName)).Op("=").Add(resultArrayItem)
			} else {
				cg.addArrayInputParser(
					field,
					ref.Array,
					ig,
					resultArrayItem,
					valueToSave.Clone().Index(jen.Id(indexVarName)),
					level+1,
				)
			}
		},
	)
}

func (cg *GQLGenerator) generateGQLLookupQuery(t *Entity) error {
	//TODO: check annotations for lookup fields, possibly necessary create special type (or use map)
	// if opername, ok := t.Annotations.GetStringAnnotation(gqlAnnotation, gqlOperationsAnnotationsTags[GQLOperationLookup]); ok {
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationLookup]); ok {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		fname := fmt.Sprintf("%sLookupQueryGenerator", name)
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.Dict{
						jen.Id("Type"): jen.Qual(
							gqlPackage,
							"NewList",
						).Params(
							jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
								jen.Op("*").Qual(
									VivardPackage,
									"GQLEngine",
								),
							).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
						),
						jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(
							jen.Dict{
								jen.Lit("query"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
									jen.Dict{
										jen.Id("Type"): jen.Qual(gqlPackage, "String"),
									},
								),
							},
						),
						jen.Id("Resolve"): jen.Func().Params(
							jen.Id("p").Qual(
								gqlPackage,
								"ResolveParams",
							),
						).Parens(jen.List(jen.Interface(), jen.Error())).
							Block(
								jen.Id("query").Op(":=").Id("p").Dot("Args").Index(jen.Lit("query")).Assert(jen.String()).Line().
									Return(
										jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodLookup, name)).Params(
											jen.Id("p").Dot("Context"),
											jen.Id("query"),
										),
									),
							),
					},
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddQueryGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

func (cg *GQLGenerator) generateGQLListQuery(t *Entity) error {
	// if opername, ok := t.Annotations.GetStringAnnotation(gqlAnnotation, gqlOperationsAnnotationsTags[GQLOperationList]); ok {
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationList]); ok {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		fname := fmt.Sprintf("%sListQueryGenerator", name)

		listMethod := MethodList
		withQualifier := false
		var qualIdFld *Field
		var qualGQLType string
		var keyType *jen.Statement
		if t.HasModifier(TypeModifierDictionary) /*t.FB(FeaturesCommonKind, FCReadonly)*/ {
			listMethod = MethodGetAll
			if t.FB(FeatureDictKind, FDQualified) {
				withQualifier = true
				qt, _ := t.Features.GetEntity(FeatureDictKind, FDQualifierType)
				qualIdFld = qt.GetIdField()
				switch qualIdFld.Type.Type {
				case TipInt:
					keyType = jen.Int()
					qualGQLType = "Int"
				case TipString:
					keyType = jen.String()
					qualGQLType = "String"
				default:
					return fmt.Errorf("at %s: only dicts with id field of type int and string may be used as qualifier", t.Pos)
				}
			}
		}
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.DictFunc(
						func(d jen.Dict) {
							d[jen.Id("Type")] = jen.Qual(
								gqlPackage,
								"NewList",
							).Params(
								jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
									jen.Op("*").Qual(
										VivardPackage,
										"GQLEngine",
									),
								).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
							)
							d[jen.Id("Resolve")] = jen.Func().Params(
								jen.Id("p").Qual(
									gqlPackage,
									"ResolveParams",
								),
							).Parens(jen.List(jen.Interface(), jen.Error())).
								BlockFunc(
									func(g *jen.Group) {
										if withQualifier {

											g.List(
												jen.Id("quals"),
												jen.Id("ok"),
											).Op(":=").Id("p").Dot("Args").Index(jen.Lit("quals")).Assert(jen.Index().Interface())
											g.If(jen.Id("ok")).Block(
												jen.Id("qs").Op(":=").Make(jen.Index().Add(keyType), jen.Len(jen.Id("quals"))),
												jen.For(jen.List(jen.Id("i"), jen.Id("q")).Op(":=").Range().Id("quals")).Block(
													jen.Id("qs").Index(jen.Id("i")).Op("=").Id("q").Assert(keyType),
												),
												jen.Return(
													jen.Id(EngineVar).Dot(cg.desc.GetMethodName(listMethod, name)).Params(
														jen.Id("p").Dot("Context"),
														jen.Id("qs").Op("..."),
													),
												),
											)

											g.List(
												jen.Id("qual"),
												jen.Id("ok"),
											).Op(":=").Id("p").Dot("Args").Index(jen.Lit("qual")).Assert(keyType)
											g.If(jen.Id("ok")).Block(
												jen.Return(
													jen.Id(EngineVar).Dot(cg.desc.GetMethodName(listMethod, name)).Params(
														jen.Id("p").Dot("Context"),
														jen.Id("qual"),
													),
												),
											)
										}
										g.Return(
											jen.Id(EngineVar).Dot(cg.desc.GetMethodName(listMethod, name)).Params(
												jen.Id("p").Dot("Context"),
											),
										)
									},
								)
							if withQualifier {
								d[jen.Id("Args")] = jen.Qual(gqlPackage, "FieldConfigArgument").Values(
									jen.Dict{
										jen.Lit("qual"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
											jen.Dict{
												jen.Id("Type"): jen.Qual(gqlPackage, qualGQLType),
											},
										),
										jen.Lit("quals"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
											jen.Dict{
												jen.Id("Type"): jen.Qual(gqlPackage, "NewList").Params(jen.Qual(gqlPackage, qualGQLType)),
											},
										),
									},
								)
							}
						},
					),
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddQueryGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

func (cg *GQLGenerator) generateGQLFindQuery(t *Entity) error {
	if it, ok := t.Features.GetEntity(FeaturesAPIKind, FAPIFindParamType); ok {
		if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationFind]); ok {
			name := t.GetName()
			gqlType := t.FS(GQLFeatures, GQLFTypeTag)
			fname := fmt.Sprintf("%sFindQueryGenerator", name)
			f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
				gqlPackage,
				"Field",
			).Block(
				jen.Return(
					jen.Op("&").Qual(gqlPackage, "Field").Values(
						jen.Dict{
							jen.Id("Type"): jen.Qual(
								gqlPackage,
								"NewList",
							).Params(
								jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
									jen.Op("*").Qual(
										VivardPackage,
										"GQLEngine",
									),
								).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
							),
							jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(
								jen.Dict{
									jen.Lit("query"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
										jen.Dict{
											jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
												jen.Op("*").Qual(
													VivardPackage,
													"GQLEngine",
												),
											).Dot("Descriptor").Params().Dot("GetInputType").Call(jen.Lit(cg.GetGQLInputTypeName(it.Name))),
										},
									),
								},
							),
							jen.Id("Resolve"): jen.Func().Params(
								jen.Id("p").Qual(
									gqlPackage,
									"ResolveParams",
								),
							).Parens(jen.List(jen.Interface(), jen.Error())).
								Block(
									jen.If(
										jen.List(jen.Id("query"), jen.Id("ok").Op(":=").Id("p").Dot("Args").Index(jen.Lit("query"))),
										jen.Id("ok"),
									).Block(
										jen.List(jen.Id("q"), jen.Err()).Op(":=").Add(
											cg.callInputParserMethod(
												jen.Id("p").Dot("Context"),
												it.Name,
												"query",
												jen.Nil(),
												false,
											),
										),
										jen.Return(
											jen.List(
												jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodFind, name)).Params(
													jen.Id("p").Dot("Context"),
													jen.Id("q"),
												),
											),
										),
									).Else().Block(
										jen.Return(
											jen.Nil(),
											jen.Qual("errors", "New").Params(jen.Lit("find withput query")),
										),
									),
								),
						},
					),
				),
			).Line()

			cg.b.Functions.Add(f)
			cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddQueryGenerator").Params(
				jen.Lit(opername),
				jen.Id(EngineVar).Dot(fname),
			).Line()
		}
	}
	return nil
}

func (cg *GQLGenerator) generateGQLSetMutation(t *Entity) error {
	// if opername, ok := t.Annotations.GetStringAnnotation(gqlAnnotation, gqlOperationsAnnotationsTags[GQLOperationSet]); ok {
	if t.FB(FeaturesCommonKind, FCReadonly) {
		return nil
	}
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationSet]); ok {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		fname := fmt.Sprintf("%sSetMutationGenerator", name)
		args := jen.Dict{
			jen.Lit("val"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
				jen.Dict{
					jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
						jen.Op("*").Qual(
							VivardPackage,
							"GQLEngine",
						),
					).Dot("Descriptor").Params().Dot("GetInputType").Call(jen.Lit(cg.GetGQLInputTypeName(name))),
				},
			),
		}
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.Dict{
						// jen.Id("Type"): jen.Qual(gqlPackage, "Boolean"),
						jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
							jen.Op("*").Qual(
								VivardPackage,
								"GQLEngine",
							),
						).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
						jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(args),
						jen.Id("Resolve"): jen.Func().Params(
							jen.Id("p").Qual(
								gqlPackage,
								"ResolveParams",
							),
						).Parens(jen.List(jen.Interface(), jen.Error())).
							Block(
								// idstmt,
								jen.If(
									jen.List(jen.Id("val"), jen.Id("ok").Op(":=").Id("p").Dot("Args").Index(jen.Lit("val"))),
									jen.Id("ok"),
								).Block(
									jen.List(
										jen.Id("id"),
										jen.Err(),
									).Op(":=").Id(EngineVar).Dot(cg.getIdFromInputMethodName(name)).Params(jen.Id("val")),
									returnIfErrValue(jen.Nil()),
									jen.List(jen.Id("obj"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(
										cg.desc.GetMethodName(
											MethodGet,
											name,
										),
									).Params(
										jen.Id("p").Dot("Context"),
										jen.Id("id"),
									),
									returnIfErrValue(jen.Nil()),
									// assigns,
									jen.List(jen.Id("obj"), jen.Err()).Op("=").Add(
										cg.callInputParserMethod(
											jen.Id("p").Dot("Context"),
											name,
											"val",
											jen.Id("obj"),
											true,
										),
									),
									jen.Return(
										jen.List(
											jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodSet, name)).Params(
												jen.Id("p").Dot("Context"),
												jen.Id("obj"),
											),
										),
									),
								).Else().Block(
									jen.Return(
										jen.Nil(),
										jen.Qual("errors", "New").Params(jen.Lit("set without val")),
									),
								),
							),
					},
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddMutationGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

func (cg *GQLGenerator) generateGQLCreateMutation(t *Entity) error {
	// if opername, ok := t.Annotations.GetStringAnnotation(gqlAnnotation, gqlOperationsAnnotationsTags[GQLOperationCreate]); ok {
	if t.FB(FeaturesCommonKind, FCReadonly) {
		return nil
	}
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationCreate]); ok {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		fname := fmt.Sprintf("%sCreateMutationGenerator", name)
		args := jen.Dict{
			jen.Lit("val"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
				jen.Dict{
					jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
						jen.Op("*").Qual(
							VivardPackage,
							"GQLEngine",
						),
					).Dot("Descriptor").Params().Dot("GetInputType").Call(jen.Lit(cg.GetGQLInputTypeName(name))),
				},
			),
		}
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.Dict{
						jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
							jen.Op("*").Qual(
								VivardPackage,
								"GQLEngine",
							),
						).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
						jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(args),
						jen.Id("Resolve"): jen.Func().Params(
							jen.Id("p").Qual(
								gqlPackage,
								"ResolveParams",
							),
						).Parens(jen.List(jen.Interface(), jen.Error())).Block(
							jen.If(
								jen.List(jen.Id("val"), jen.Id("ok").Op(":=").Id("p").Dot("Args").Index(jen.Lit("val"))),
								jen.Id("ok"),
							).BlockFunc(
								func(g *jen.Group) {
									if t.HasModifier(TypeModifierExtendable) {
										g.List(jen.Id("obj"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(
											cg.desc.GetMethodName(
												MethodInit,
												name,
											),
										).Params(jen.Id("p").Dot("Context"))
										g.Add(returnIfErrValue(jen.Nil()))
									} else {
										g.Var().Id("obj").Op("*").Id(name)
										g.Var().Err().Error()
									}
									g.List(jen.Id("obj"), jen.Err()).Op("=").Add(
										cg.callInputParserMethod(
											jen.Id("p").Dot("Context"),
											name,
											"val",
											jen.Id("obj"),
											false,
										),
									)

									g.Return(
										jen.List(
											jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodNew, name)).Params(
												jen.Id("p").Dot("Context"),
												jen.Id("obj"),
											),
										),
									)
								},
							).Else().Block(
								jen.Return(
									jen.Nil(),
									jen.Qual("errors", "New").Params(jen.Lit("set without val")),
								),
							),
						),
					},
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddMutationGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

func (cg *GQLGenerator) generateGQLDeleteMutation(t *Entity) error {
	if t.FB(FeaturesCommonKind, FCReadonly) {
		return nil
	}
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationDelete]); ok {
		name := t.GetName()
		fname := fmt.Sprintf("%sDeleteMutationGenerator", name)
		idField := t.GetIdField()
		id := idField.Annotations.GetStringAnnotationDef(GQLAnnotation, GQLAnnotationNameTag, "id")
		idtype, _ := cg.getGQLType(idField.Type)
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.Dict{
						jen.Id("Type"): jen.Qual(gqlPackage, "NewNonNull").Params(jen.Qual(gqlPackage, "Boolean")),
						jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(
							jen.Dict{
								jen.Lit(id): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
									jen.Dict{
										jen.Id("Type"): idtype,
									},
								),
							},
						),
						jen.Id("Resolve"): jen.Func().Params(
							jen.Id("p").Qual(
								gqlPackage,
								"ResolveParams",
							),
						).Parens(jen.List(jen.Interface(), jen.Error())).
							Block(
								jen.Id("id").Op(":=").Id("p").Dot("Args").Index(jen.Lit(id)).Assert(cg.b.GoType(idField.Type)).Line().
									Return(
										jen.List(
											jen.True(),
											jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodDelete, name)).Params(
												jen.Id("p").Dot("Context"),
												jen.Id("id"),
											),
										),
									),
							),
					},
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddMutationGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

func (cg *GQLGenerator) generateGQLBulkMethods(t *Entity) error {
	if !t.FB(FeaturesCommonKind, FCReadonly) &&
		(t.FB(FeatGoKind, FCGBulkNew) || t.FB(FeatGoKind, FCGBulkSet)) {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		var fname string
		args := jen.Dict{
			jen.Lit("val"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
				jen.Dict{
					jen.Id("Type"): jen.Qual(
						gqlPackage,
						"NewList",
					).Params(
						jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
							jen.Op("*").Qual(
								VivardPackage,
								"GQLEngine",
							),
						).Dot("Descriptor").Params().Dot("GetInputType").Call(jen.Lit(cg.GetGQLInputTypeName(name))),
					),
				},
			),
		}
		createFunction := func(method MethodKind) *jen.Statement {
			return jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
				gqlPackage,
				"Field",
			).Block(
				jen.Return(
					jen.Op("&").Qual(gqlPackage, "Field").Values(
						jen.Dict{
							jen.Id("Type"): jen.Qual(
								gqlPackage,
								"NewList",
							).Params(
								jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
									jen.Op("*").Qual(
										VivardPackage,
										"GQLEngine",
									),
								).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
							),
							jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(args),
							jen.Id("Resolve"): jen.Func().Params(
								jen.Id("p").Qual(
									gqlPackage,
									"ResolveParams",
								),
							).Parens(jen.List(jen.Interface(), jen.Error())).Block(
								jen.If(
									jen.List(
										jen.Id("vals"),
										jen.Id("ok").Op(":=").Id("p").Dot("Args").Index(jen.Lit("val")).Assert(jen.Index().Any()),
									),
									jen.Id("ok"),
								).BlockFunc(
									func(g *jen.Group) {
										g.Var().Err().Error()
										if t.FB(FeatGoKind, FCGBulkFilterRaw) {
											g.List(
												jen.Id("vals"),
												jen.Err(),
											).Op("=").Id(EngineVar).Dot(fmt.Sprintf("%sBulkFilterRaw", name)).Call(
												jen.Id("p").Dot("Context"),
												jen.Id("vals"),
											)
											g.Add(returnIfErrValue(jen.Nil()))
											g.If(jen.Len(jen.Id("vals")).Op("==").Lit(0)).Block(
												jen.Return(jen.Nil(), jen.Nil()),
											)
										}
										g.Id("objs").Op(":=").Make(jen.Index().Op("*").Id(name), jen.Lit(0), jen.Len(jen.Id("vals")))
										g.For(jen.List(jen.Id("idx"), jen.Id("param")).Op(":=").Range().Id("vals")).BlockFunc(
											func(fg *jen.Group) {
												fg.Var().Id("obj").Op("*").Id(name)
												fg.List(
													jen.Id("obj"),
													jen.Err(),
												).Op("=").Add(
													cg.callInputParserMethod(
														jen.Id("p").Dot("Context"),
														name,
														"param",
														jen.Id("obj"),
														false,
													),
												)
												fg.Add(returnIfErrValue(jen.Nil()))
												if t.FB(FeatGoKind, FCGBulkFilterEach) {
													fg.List(
														jen.Id("obj"),
														jen.Err(),
													).Op("=").Id(EngineVar).Dot(fmt.Sprintf("%sBulkFilterEach", name)).Call(
														jen.Id("p").Dot("Context"),
														jen.Id("obj"),
														jen.Id("idx"),
														jen.Len(jen.Id("vals")),
													)
													fg.Add(returnIfErrValue(jen.Nil()))
													fg.If(jen.Id("obj").Op("==").Nil()).Block(
														jen.Continue(),
													)
												}
												fg.Id("objs").Op("=").Append(jen.Id("objs"), jen.Id("obj"))
											},
										)
										if t.FB(FeatGoKind, FCGBulkFilterAll) {
											g.List(
												jen.Id("objs"),
												jen.Err(),
											).Op("=").Id(EngineVar).Dot(fmt.Sprintf("%sBulkFilterAll", name)).Call(
												jen.Id("p").Dot("Context"),
												jen.Id("objs"),
											)
											g.Add(returnIfErrValue(jen.Nil()))
											g.If(jen.Len(jen.Id("objs")).Op("==").Lit(0)).Block(
												jen.Return(jen.Nil(), jen.Nil()),
											)
										}
										g.Return(
											jen.List(
												jen.Id(EngineVar).Dot(cg.desc.GetMethodName(method, name)).Params(
													jen.Id("p").Dot("Context"),
													jen.Id("objs"),
												),
											),
										)
									},
								).Else().Block(
									jen.Return(
										jen.Nil(),
										jen.Qual("errors", "New").Params(jen.Lit("set without val")),
									),
								),
							),
						},
					),
				),
			).Line()
		}
		if t.FB(FeatGoKind, FCGBulkNew) {
			fname = fmt.Sprintf("%sBulkCreateMutationGenerator", name)
			if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationBulkCreate]); ok {
				f := createFunction(MethodNewBulk)
				cg.b.Functions.Add(f)
				cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddMutationGenerator").Params(
					jen.Lit(opername),
					jen.Id(EngineVar).Dot(fname),
				).Line()
			}
		}
		if t.FB(FeatGoKind, FCGBulkSet) {
			fname = fmt.Sprintf("%sBulkSetMutationGenerator", name)
			if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationBulkSet]); ok {
				f := createFunction(MethodSetBulk)
				cg.b.Functions.Add(f)
				cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddMutationGenerator").Params(
					jen.Lit(opername),
					jen.Id(EngineVar).Dot(fname),
				).Line()
			}
		}
	}
	return nil
}

func (cg *GQLGenerator) generateGQLMethods(t *Entity) error {
	for _, m := range t.Methods {
		switch m.FS(GQLFeatures, GQLFMethodType) {
		case GQLFMethodTypeQuery:
			cg.desc.AddWarning(
				fmt.Sprintf(
					"at %v: query generation for method is not yet supported; falling throw to mutation",
					m.Pos,
				),
			)
			fallthrough
		case GQLFMethodTypeMutation:
			err := cg.generateGQLMethodMutation(m)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (cg *GQLGenerator) generateGQLMethodMutation(m *Method) (err error) {
	opername := m.FS(GQLFeatures, GQLFMethodName)
	// t := jen.Qual(gqlPackage, "Boolean")
	// if m.RetValue != nil {
	// 	t, err = cg.getGQLType(m.RetValue)
	// 	if err != nil {
	// 		return fmt.Errorf("at %v: type not find: %v", m.Pos, m.RetValue)
	// 	}
	// }
	fname := fmt.Sprintf("%sMutationGenerator", opername)
	idField := m.parent.GetIdField()

	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
		gqlPackage,
		"Field",
	).Block(
		jen.Return(
			jen.Op("&").Qual(gqlPackage, "Field").Values(
				jen.Dict{
					jen.Id("Type"): cg.desc.CallCodeFeatureFunc(m, GQLFeatures, GQLFMethodResultType),
					jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").ValuesFunc(
						func(g *jen.Group) {
							args := jen.Dict{}
							if idField != nil {
								idtype, _ := cg.getGQLType(idField.Type)
								args[jen.Lit("id")] = jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
									jen.Dict{
										jen.Id("Type"): idtype,
									},
								)
							}
							for _, p := range m.Params {
								pt, err := cg.getGQLType(p.Type, false, false, true)
								if err != nil {
									cg.desc.AddError(fmt.Errorf("at %v: type not find: %s", p.Pos, p.Type.Type))
									return
								}
								args[jen.Lit(p.Name)] = jen.Op("&").Qual(
									gqlPackage,
									"ArgumentConfig",
								).Values(jen.Dict{jen.Id("Type"): pt})
							}
							g.Add(args)
						},
					),
					jen.Id("Resolve"): jen.Func().Params(
						jen.Id("p").Qual(
							gqlPackage,
							"ResolveParams",
						),
					).Parens(jen.List(jen.Interface(), jen.Error())).BlockFunc(
						func(g *jen.Group) {
							for _, p := range m.Params {
								g.Var().Id(p.Name + "_Arg").Add(cg.b.GoType(p.Type, true))
								g.Add(cg.inputParserCodeGenerator(p.Type, p.Name, jen.Id(p.Name+"_Arg"), p.Pos))
							}
							id := jen.Nil()
							if idField != nil {
								id = jen.Id("p").Dot("Args").Index(jen.Lit("id")).Assert(cg.b.GoType(idField.Type))
							}
							g.Var().Err().Error()
							g.List(jen.Id("obj"), jen.Err()).Op(":=").Add(
								cg.desc.CallCodeFeatureFunc(
									m.parent, FeaturesCommonKind, FCGetterCode,
									id, jen.Id("p").Dot("Context"), true,
								),
							)
							g.Add(returnIfErrValue(jen.Nil()))
							g.ReturnFunc(
								func(g *jen.Group) {
									args := make([]HookArgParam, len(m.Params))
									for i, p := range m.Params {
										args[i] = HookArgParam{p.Name, jen.Id(p.Name + "_Arg")}
									}
									descr := HookArgsDescriptor{
										Str: m.Name,
										// Obj: jen.Id(EngineVar).Dot(m.parent.Name),
										Ctx:    jen.Id("p").Dot("Context"),
										Params: args,
									}
									g.Add(cg.desc.CallFeatureHookFunc(m, FeaturesHookCodeKind, TypeHookMethod, descr))
								},
							)
						},
					),
				},
			),
		),
	).Line()

	cg.b.Functions.Add(f)
	cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddMutationGenerator").Params(
		jen.Lit(opername),
		jen.Id(EngineVar).Dot(fname),
	).Line()
	return nil
}

func (cg *GQLGenerator) generateGQLConfigQuery(t *Entity) error {
	// if opername, ok := t.Annotations.GetStringAnnotation(gqlAnnotation, gqlOperationsAnnotationsTags[GQLOperationGet]); ok {
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationGet]); ok {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		fname := fmt.Sprintf("%sQueryGenerator", name)
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.Dict{
						jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
							jen.Op("*").Qual(
								VivardPackage,
								"GQLEngine",
							),
						).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
						jen.Id("Resolve"): jen.Func().Params(
							jen.Id("p").Qual(
								gqlPackage,
								"ResolveParams",
							),
						).Parens(jen.List(jen.Interface(), jen.Error())).
							Block(
								jen.Return(
									jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodGet, name)).Params(
										jen.Id("p").Dot("Context"),
									),
								),
							),
					},
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddQueryGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

func (cg *GQLGenerator) generateGQLConfigSetMutation(t *Entity) error {
	// if opername, ok := t.Annotations.GetStringAnnotation(gqlAnnotation, gqlOperationsAnnotationsTags[GQLOperationSet]); ok {
	if opername, ok := t.Features.GetString(GQLFeatures, GQLOperationsAnnotationsTags[GQLOperationSet]); ok {
		name := t.GetName()
		gqlType := t.FS(GQLFeatures, GQLFTypeTag)
		fname := fmt.Sprintf("%sSetMutationGenerator", name)
		args := jen.Dict{
			jen.Lit("val"): jen.Op("&").Qual(gqlPackage, "ArgumentConfig").Values(
				jen.Dict{
					jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
						jen.Op("*").Qual(
							VivardPackage,
							"GQLEngine",
						),
					).Dot("Descriptor").Params().Dot("GetInputType").Call(jen.Lit(cg.GetGQLInputTypeName(name))),
				},
			),
		}
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Op("*").Qual(
			gqlPackage,
			"Field",
		).Block(
			jen.Return(
				jen.Op("&").Qual(gqlPackage, "Field").Values(
					jen.Dict{
						// jen.Id("Type"): jen.Qual(gqlPackage, "Boolean"),
						jen.Id("Type"): jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
							jen.Op("*").Qual(
								VivardPackage,
								"GQLEngine",
							),
						).Dot("Descriptor").Params().Dot("GetType").Call(jen.Lit(gqlType)),
						jen.Id("Args"): jen.Qual(gqlPackage, "FieldConfigArgument").Values(args),
						jen.Id("Resolve"): jen.Func().Params(
							jen.Id("p").Qual(
								gqlPackage,
								"ResolveParams",
							),
						).Parens(jen.List(jen.Interface(), jen.Error())).
							Block(
								// idstmt,
								jen.If(
									jen.List(jen.Id("val"), jen.Id("ok").Op(":=").Id("p").Dot("Args").Index(jen.Lit("val"))),
									jen.Id("ok"),
								).Block(
									jen.List(jen.Id("obj"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(
										cg.desc.GetMethodName(
											MethodGet,
											name,
										),
									).Params(
										jen.Id("p").Dot("Context"),
									),
									returnIfErrValue(jen.Nil()),
									// assigns,
									jen.List(jen.Id("obj"), jen.Err()).Op("=").Add(
										cg.callInputParserMethod(
											jen.Id("p").Dot("Context"),
											name,
											"val",
											jen.Id("obj"),
											true,
										),
									),
									jen.Return(
										jen.List(
											jen.Id(EngineVar).Dot(cg.desc.GetMethodName(MethodSet, name)).Params(
												jen.Id("p").Dot("Context"),
												jen.Id("obj"),
											),
										),
									),
								).Else().Block(
									jen.Return(
										jen.Nil(),
										jen.Qual("errors", "New").Params(jen.Lit("set without val")),
									),
								),
							),
					},
				),
			),
		).Line()

		cg.b.Functions.Add(f)
		cg.b.Generator.Id(gqlDescriptorVarName).Dot("AddMutationGenerator").Params(
			jen.Lit(opername),
			jen.Id(EngineVar).Dot(fname),
		).Line()
	}
	return nil
}

// getGQLType returns type;
// skipNotNull returns only type if true (without NotNull wrapper for NotNull fields)
// next argument force to return key type for complex types
// third - true if requires InputType
func (cg *GQLGenerator) getGQLType(ref *TypeRef, skipNotNull ...bool) (ret *jen.Statement, err error) {
	if ref.Array != nil {
		params, err := cg.getGQLType(ref.Array, skipNotNull...)
		if err != nil {
			return ret, err
		}
		ret = jen.Qual(gqlPackage, "NewList").Call(params)
	} else {
		switch ref.Type {
		case TipBool:
			ret = jen.Qual(gqlPackage, "Boolean")
		case TipString:
			ret = jen.Qual(gqlPackage, "String")
		case TipInt:
			ret = jen.Qual(gqlPackage, "Int")
		case TipDate:
			ret = jen.Qual(gqlPackage, "DateTime")
		case TipFloat:
			ret = jen.Qual(gqlPackage, "Float")
		default:
			if len(skipNotNull) > 1 && skipNotNull[1] {
				if t, ok := cg.desc.FindType(ref.Type); ok && t.entry != nil {
					if idfld := t.entry.GetIdField(); idfld != nil {
						ret, err = cg.getGQLType(idfld.Type, true)
						if err != nil {
							return
						}
					}
				}
			}
			if ret == nil {
				var typeName string
				isInput := len(skipNotNull) > 2 && skipNotNull[2]
				if ref.Map != nil {
					typeName, err = cg.getMapTypeName(ref, isInput)
					if err != nil {
						return
					}
				} else {
					e, ok := cg.desc.FindType(ref.Type)
					if !ok {
						return nil, fmt.Errorf("type not found: %s", ref.Type)
					}
					if e.enum != nil {
						//TODO create enum type for enums
						return cg.getGQLType(&TypeRef{Type: e.enum.AliasForType}, skipNotNull...)
					}
					if e.entry == nil {
						return nil, fmt.Errorf("external type can not be used here: %s", ref.Type)
					}
					typeName = e.entry.FS(GQLFeatures, GQLFTypeTag)
					if isInput {
						typeName = e.entry.FS(GQLFeatures, GQLFInputTypeName)
					}
				}
				ret = cg.generateTypeLookupStatement(typeName, isInput)
				if ref.Map != nil {
					ret = jen.Qual(gqlPackage, "NewList").Params(ret)
				}
			}
		}
	}
	if ret == nil {
		return nil, fmt.Errorf("unknown GQL type for field: %v", *ref)
	}
	if ref.NonNullable &&
		(len(skipNotNull) == 0 || !skipNotNull[0]) {
		ret = jen.Qual(gqlPackage, "NewNonNull").Call(ret)
	}
	return ret, nil
}

func (cg *GQLGenerator) generateTypeLookupStatement(typeName string, isInput bool) *jen.Statement {
	typeLookuper := "GetType"
	if isInput {
		typeLookuper = "GetInputType"
	}
	return jen.Id(EngineVar).Dot(EngineVivard).Dot("GetService").Params(jen.Lit("gql")).Assert(
		jen.Op("*").Qual(
			VivardPackage,
			"GQLEngine",
		),
	).Dot("Descriptor").Params().Dot(typeLookuper).Call(jen.Lit(typeName))
}
func (cg *GQLGenerator) getMapTypeName(ref *TypeRef, forInput bool) (string, error) {
	if ref.Map != nil {
		if ref.Map.KeyType != TipString {
			return "", fmt.Errorf(
				"map type: only string keys may be used for GQL at the moment, but found: %s ",
				ref.Map.KeyType,
			)
		}
		switch ref.Map.ValueType.Type {
		case TipString:
			if forInput {
				return vivard.KVStringStringInput, nil
			} else {
				return vivard.KVStringString, nil
			}
		case TipInt:
			if forInput {
				return vivard.KVStringIntInput, nil
			} else {
				return vivard.KVStringInt, nil
			}
		default:
			return "", fmt.Errorf(
				"map type: only string and int values may be used for GQL at the moment, but found: %s ",
				ref.Map.ValueType.Type,
			)
		}
	}
	return "", errors.New("not a map")
}

// inputParserGenerator returns function that returns code for parsing input;
func (cg *GQLGenerator) inputParserCodeGenerator(
	t *TypeRef,
	name string,
	assignTo jen.Code,
	pos lexer.Position,
) jen.Code {
	ret := jen.If(jen.Id("p").Dot("Args").Index(jen.Lit(name)).Op("==").Nil()).Block(
		jen.Add(assignTo).Op("=").Add(cg.b.goEmptyValue(t)),
	).Else().BlockFunc(
		func(g *jen.Group) {
			if t.Array != nil {
				g.If(
					jen.List(
						jen.Id("val"),
						jen.Id("ok"),
					).Op(":=").Id("p").Dot("Args").Index(jen.Lit(name)).Assert(jen.Index().Interface()),
					jen.Id("ok"),
				).Block(
					jen.Add(assignTo).Op("=").Make(
						jen.Index(). /*Op("*").*/ Add(cg.GetInputGoType(t.Array)),
						jen.Len(jen.Id("val")),
					),
					jen.For(jen.List(jen.Id("i"), jen.Id("item")).Op(":=").Range().Id("val")).BlockFunc(
						func(g *jen.Group) {
							if !t.Array.Complex {
								g.Add(assignTo).Index(jen.Id("i")).Op("=").Id("item").Assert(cg.GetInputGoType(t.Array))
							} else {
								g.Var().Err().Error()
								g.List(
									jen.Add(assignTo).Index(jen.Id("i")),
									jen.Err(),
								).Op("=").Add(
									cg.callInputParserMethod(
										jen.Id("p").Dot("Context"),
										t.Array.Type,
										"item",
										jen.Nil(),
										false,
									),
								)
								g.Add(returnIfErrValue(jen.Nil()))
							}
						},
					),
				).Else().Block(
					jen.Return(
						jen.Nil(),
						jen.Qual("errors", "New").Params(jen.Lit("invalid type for array")),
					),
				)
			} else if t.Map != nil {
				if t.Map.KeyType != TipString {
					cg.desc.AddError(fmt.Errorf("at %v: GQL: only string can be used as Key for Maps", pos))
					return
				}
				var fn string
				switch t.Map.ValueType.Type {
				case TipString:
					fn = "GQLArgToMapStringString"
				case TipInt:
					fn = "GQLArgToMapStringInt"
				default:
					cg.desc.AddError(fmt.Errorf("at %v: GQL: only string and int can be used as Maps value currently", pos))
					return
				}
				g.List(jen.Id("values"), jen.Err()).Op(":=").Qual(
					VivardPackage,
					fn,
				).Params(jen.Id("p").Dot("Args").Index(jen.Lit(name)).Assert(cg.GetInputGoType(t)))
				g.Add(returnIfErrValue(jen.Nil()))
				g.Add(assignTo).Op("=").Id("values")
			} else {
				if !t.Complex {
					g.Add(assignTo).Op("=").Id("p").Dot("Args").Index(jen.Lit(name)).Assert(cg.GetInputGoType(t))
				} else {
					g.Var().Err().Error()
					g.Id("item").Op(":=").Id("p").Dot("Args").Index(jen.Lit(name))
					g.List(assignTo, jen.Err()).Op("=").Add(
						cg.callInputParserMethod(
							jen.Id("p").Dot("Context"),
							t.Type,
							"item",
							jen.Nil(),
							false,
						),
					)
					g.Add(returnIfErrValue(jen.Nil()))
				}
			}
		},
	)
	return ret
}

func (cg *GQLGenerator) GetGQLOperationName(e *Entity, tip GQLOperationKind) string {
	name := cg.GetGQLEntityTypeName(e.Name)
	return fmt.Sprintf(gqlOperationsNamesTemplates[tip], strings.ToUpper(name[:1])+name[1:])
}

func (cg *GQLGenerator) GetGQLEntityTypeName(name string) (ret string) {
	if entity, ok := cg.desc.FindType(name); ok {
		if entity.Entity() != nil {
			if name, ok := entity.Entity().Annotations.GetStringAnnotation(GQLAnnotation, GQLAnnotationNameTag); ok {
				return name
			}
		}
	}

	names := strings.SplitN(name, ".", 2)
	packageName := cg.desc.Name
	if len(names) == 2 {
		packageName = names[0]
		name = names[1]
	}
	if cg.options.UsePackageNameInTypeNames && name != "" {
		return fmt.Sprintf("%s%s%s", packageName, strings.ToUpper(name[:1]), name[1:])
	}
	return name
}

func (cg *GQLGenerator) GetGQLInputTypeName(name string) (ret string) {
	names := strings.SplitN(name, ".", 2)
	packageName := cg.desc.Name
	if len(names) == 2 {
		packageName = names[0]
		name = names[1]
	}
	if cg.options.UsePackageNameInTypeNames && name != "" {
		return fmt.Sprintf("%s%s%sInput", packageName, strings.ToUpper(name[:1]), name[1:])
	}
	return name + "Input"
}
func (cg *GQLGenerator) GetGQLTypeName(ref *TypeRef, forInput ...bool) (ret string) {
	if ref.Array != nil {
		params := cg.GetGQLTypeName(ref.Array, forInput...)
		ret = fmt.Sprintf("[%s]", params)
	} else if ref.Map != nil {
		var err error
		ret, err = cg.getMapTypeName(ref, len(forInput) > 0 && forInput[0])
		if err != nil {
			cg.desc.AddError(err)
		}
		ret = fmt.Sprintf("[%s]", ret)
		return
	} else {
		switch ref.Type {
		case TipBool:
			ret = "Boolean"
		case TipString:
			ret = "String"
		case TipInt:
			ret = "Int"
		case TipDate:
			ret = "DateTime"
		case TipFloat:
			ret = "Float"
		default:
			if dt, ok := cg.desc.FindType(ref.Type); ok {
				if dt.Entity() != nil {
					if len(forInput) > 0 && forInput[0] {
						ret = dt.Entity().FS(GQLFeatures, GQLFInputTypeName)
					} else {
						ret = dt.Entity().FS(GQLFeatures, GQLFTypeTag)
					}
				} else if dt.Enum() != nil {
					if len(forInput) > 0 && forInput[0] {
						ret, _ = dt.Enum().Features.GetString(GQLFeatures, GQLFInputTypeName)
					} else {
						ret, _ = dt.Enum().Features.GetString(GQLFeatures, GQLFTypeTag)
					}
				}
			}
			if ret == "" {
				if len(forInput) > 0 && forInput[0] {
					ret = cg.GetGQLInputTypeName(ref.Type)
				} else {
					ret = cg.GetGQLEntityTypeName(ref.Type)
				}
			}
		}
	}
	if ref.NonNullable {
		ret += "!"
	}
	return ret
}
func (cg *GQLGenerator) GetInputGoType(ref *TypeRef) *jen.Statement {
	if ref.Map != nil {
		return jen.Index().Interface()
	}
	return cg.b.GoType(ref)
}
func (cg *GQLGenerator) GetGQLFieldName(f *Field) string {
	return toCamelCase(f.Name)
}

func toCamelCase(s string) string {
	ret := make([]rune, len(s))
	needConvert := true
	runes := []rune(s)
	for i, c := range runes {
		if needConvert && unicode.IsUpper(c) && (i < 2 || i >= len(runes)-2 || unicode.IsUpper(runes[i+1])) {
			c = unicode.ToLower(c)
		} else {
			needConvert = false
		}
		ret[i] = c
	}
	return string(ret)
}

func (cg *GQLGenerator) GetGQLMethodName(e *Entity, m *Method) string {
	return fmt.Sprintf("%s%s_%s", strings.ToLower(e.Name[:1]), e.Name[1:], m.Name)
}

func (cg *GQLGenerator) getInputParserMethodName(name string) string {
	if name == "" {
		return name
	}
	n := cg.desc.GetRealTypeName(name)
	return fmt.Sprintf("Parse%s%sInput", strings.ToUpper(n[:1]), n[1:])
}

func (cg *GQLGenerator) getIdFromInputMethodName(name string) string {
	return fmt.Sprintf("IDFrom%s%sInput", strings.ToUpper(name[:1]), name[1:])
}

func (cg *GQLGenerator) callInputParserMethod(ctx jen.Code, name, argVar string, obj jen.Code, checkId bool) jen.Code {
	ret := &jen.Group{}
	ret.Add(jen.Id(EngineVar))
	if dot := strings.Index(name, "."); dot != -1 {
		packageName := name[:dot]
		ret.Add(jen.Dot(cg.desc.GetExtEngineRef(packageName)))
	}
	ret.Add(
		jen.Dot(cg.getInputParserMethodName(name)).Params(ctx, jen.Id(argVar), obj, jen.Lit(checkId)).Line(),
		returnIfErrValue(jen.Nil()),
	)
	return ret
}

// ProvideFeature from FeatureProvider interface
func (cg *GQLGenerator) ProvideFeature(
	kind FeatureKind,
	name string,
	obj interface{},
) (feature interface{}, ok ProvideFeatureResult) {
	switch kind {
	case GQLFeatures:
		switch name {
		case GQLFMethodResultType:
			if m, ok := obj.(*Method); ok {
				return cg.getMethodResultTypeFunc(m), FeatureProvided
			}
		case GQLFMethodResultTypeName:
			if m, ok := obj.(*Method); ok {
				t := "Boolean"
				if m.RetValue != nil {
					t = cg.GetGQLTypeName(m.RetValue)
				}
				return t, FeatureProvided
			}
		case GQLGenerateUnionType:
			if p, ok := obj.(*Builder); ok {
				return cg.getUnionGeneratorFunc(p), FeatureProvided
			}
		}
	}
	return
}

func (cg *GQLGenerator) getMethodResultTypeFunc(m *Method) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		t := jen.Qual(gqlPackage, "Boolean")
		if m.RetValue != nil {
			var err error
			t, err = cg.getGQLType(m.RetValue)
			if err != nil {
				cg.desc.AddError(fmt.Errorf("at %v: type not find: %v", m.Pos, m.RetValue))
			}
		}
		return t
	}
}

func (cg *GQLGenerator) getUnionGeneratorFunc(builder *Builder) FeatureFunc {
	return func(args ...interface{}) (any, error) {
		//graphql.NewUnion(
		//	graphql.UnionConfig{
		//		Name:        "checkListUnion",
		//		Description: "interface for all the checklist items types",
		//		Types:       []*graphql.Object{checkListComboType, checkListIntegerType, checkListPhotosType},
		//		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
		//			switch p.Value.(*checkListItem).Type {
		//			case clitCheckBox:
		//				return checkListComboType
		//			case clitInteger:
		//				return checkListIntegerType
		//			default:
		//				return checkListPhotosType
		//			}
		//		},
		//	},
		//)
		if len(args) < 2 {
			return nil, errors.New("GQLGenerateUnionType feature requires type name and entities as parameters")
		}

		name, ok := args[0].(string)
		if !ok {
			return nil, errors.New("GQLGenerateUnionType feature: first parameter should be a string")
		}
		var types []jen.Code
		var resolves []jen.Code
		addEntity := func(e *Entity) {
			//typeRef := e.TypeRef()
			//typeName := builder.Pckg.GetRealTypeName(typeRef.Type)
			typeLookup := cg.generateTypeLookupStatement(
				e.FS(GQLFeatures, GQLFTypeTag),
				false,
			).Assert(jen.Op("*").Qual(gqlPackage, "Object"))
			types = append(types, typeLookup)
			resolves = append(resolves, jen.Case(builder.Pckg.TypeStmt(e)).Block(jen.Return(typeLookup)))
		}
		for _, a := range args[1:] {
			if e, ok := a.(*Entity); ok {
				addEntity(e)
			} else if ents, ok := a.([]*Entity); ok {
				for _, e := range ents {
					addEntity(e)
				}
			} else {
				return nil, errors.New("GQLGenerateUnionType feature: parameters from idx 1 should be *Entity")
			}
		}
		fname := fmt.Sprintf("%sTypeGenerator", name)
		f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params().Qual(gqlPackage, "Output").Block(
			jen.Return(
				jen.Qual(gqlPackage, "NewUnion").Call(
					jen.Qual(gqlPackage, "UnionConfig").Values(
						jen.Dict{
							jen.Id("Name"):  jen.Lit(name),
							jen.Id("Types"): jen.Index().Op("*").Qual(gqlPackage, "Object").Values(types...),
							jen.Id("ResolveType"): jen.Func().Params(
								jen.Id("p").Qual(
									gqlPackage,
									"ResolveTypeParams",
								),
							).Op("*").Qual(gqlPackage, "Object").Block(
								jen.Switch(jen.Id("p").Dot("Value").Assert(jen.Type())).Block(resolves...),
								jen.Return(jen.Nil()),
							),
						},
					),
				),
			),
		).Line()

		builder.Functions.Add(f)
		builder.Generator.Id(gqlDescriptorVarName).Dot("AddTypeGenerator").Params(
			jen.Lit(name),
			jen.Id(EngineVar).Dot(fname),
		).Line()
		return nil, nil
	}
}
