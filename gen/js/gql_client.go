package js

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/vc2402/vivard/gen"
)

const (
	GQLClientGeneratorName   = "GraphQL-TS"
	GQLClientOptions         = "gql-ts"
	GQLClientPathOption      = "path"
	GQLClientNamespaceOption = "useNamespace"

	Annotation          = "js"
	AnnotationName      = "name"
	AnnotationInputName = "inputName"
	// AnnotationType = "type"
	// jsAnnotationIDType   = "id_type"
	AnnotationSkip = "skip"
	// jsAnnotationFilePath = "filepath"
	AnnotationTitle        = "title"
	AnnotationIcon         = "icon"
	AnnotationColor        = "color"
	AnnotationForce        = "force"
	AnnotationForceForFind = "findForce"
)

const (
	Features           = "js"
	FType              = "type"
	FInputType         = "input-type"
	FFile              = "file"
	FIDType            = "id_type"
	FFilePath          = "filepath"
	FInstanceGenerator = "instance_generator"
	//FForceLoadForField - force load field when its parent is included in other object (usually only id and title are loaded)
	FForceLoadForField = "field_force_load"
	// JSFTitle = "title"
)
const apolloClientInclude = "import { ApolloClient } from 'apollo-client';\n"
const gqlInclude = "import gql from 'graphql-tag';\n"

const CodeFragmentModule = "ts"
const CodeFragmentActionFile = "file"
const CodeFragmentActionImport = "import"

type CodeFragmentContext struct {
	FileName string
	File     *gen.File
	Output   *os.File
	Error    error
	Imports
}

//TODO ref field is int not an object

func init() {
	gen.RegisterPlugin(&GQLCLientGenerator{})
}

type GQLCLientGenerator struct {
	desc            *gen.Package
	vivardGenerated bool
	useNS           bool
	outputPath      string
}

type Imports map[string][]string

func (imp Imports) addImport(fileName string, typeName string) {
	for _, g := range imp[fileName] {
		if g == typeName {
			return
		}
	}
	imp[fileName] = append(imp[fileName], typeName)
}
func (imp Imports) append(imps Imports) {
	for file, types := range imps {
		for _, t := range types {
			imp.addImport(file, t)
		}
	}
}

func (cg *GQLCLientGenerator) Name() string {
	return GQLClientGeneratorName
}

func (cg *GQLCLientGenerator) SetOptions(options any) error {
	if opts, ok := options.(map[string]any); ok {
		if p, ok := opts[GQLClientNamespaceOption].(bool); ok {
			cg.useNS = p
		}
		if p, ok := opts[GQLClientPathOption].(string); ok {
			cg.outputPath = p
		}
	}
	return nil
}

func (cg *GQLCLientGenerator) CheckAnnotation(desc *gen.Package, ann *gen.Annotation, item interface{}) (bool, error) {
	if ann.Name == Annotation {
		return true, nil
	}
	return false, nil
}

func (cg *GQLCLientGenerator) Prepare(desc *gen.Package) error {
	cg.desc = desc
	if opts, ok := cg.desc.Options().Custom[GQLClientOptions].(map[string]interface{}); ok {
		err := cg.SetOptions(opts)
		if err != nil {
			return err
		}
	}

	for _, file := range desc.Files {
		for _, t := range file.Entries {
			tname := cg.GetJSEntityTypeName(t.Name)
			an, ok := t.Annotations[Annotation]
			t.Features.Set(Features, FFile, file.Name)

			if !ok {
				t.Annotations.AddTag(Annotation, AnnotationName, tname)
			}
			an = t.Annotations[Annotation]
			if s, ok := an.GetBoolTag(AnnotationSkip); !ok || !s {
				if _, ok := an.GetStringTag(AnnotationName); !ok {
					t.Annotations.AddTag(Annotation, AnnotationName, tname)
				}
				if _, ok := an.GetStringTag(AnnotationInputName); !ok /*&& !t.FB(gen.FeaturesCommonKind, gen.FCReadonly)*/ {
					t.Annotations.AddTag(Annotation, AnnotationInputName, cg.GetJSEntityInputTypeName(t.Name))
				}
				for i := gen.GQLOperationGet; i < gen.GQLOperationLast; i++ {
					if skip, ok := an.GetBoolTag(gen.GQLOperationsAnnotationsTags[i]); !ok && !skip {
						if op, ok := an.GetStringTag(gen.GQLOperationsAnnotationsTags[i]); ok {
							t.Features.Set(Features, gen.GQLOperationsAnnotationsTags[i], op)
						}
					}
				}
				if idfld := t.GetIdField(); idfld != nil {
					t.Features.Set(Features, FIDType, cg.GetJSTypeName(idfld.Type, false))
				}
				t.Features.Set(Features, FInstanceGenerator, "New"+t.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")+"Instance")
				var tf *gen.Field
				for _, f := range t.GetFields(true, true) {
					if s, ok := f.Annotations.GetBoolAnnotation(gen.GQLAnnotation, gen.GQLAnnotationSkipTag); ok && s {
						continue
					}
					if fld, ok := f.Annotations.GetStringAnnotation(gen.GQLAnnotation, gen.GQLAnnotationNameTag); ok {
						f.Annotations.AddTag(Annotation, AnnotationName, fld)
						if !f.FB(gen.FeaturesCommonKind, gen.FCReadonly) {
							f.Annotations.AddTag(Annotation, AnnotationInputName, fld)
						}
					}
					// f.Annotations.AddTag(Annotation, AnnotationType, cg.GetJSTypeName(f.Type))
					f.Features.Set(Features, FType, cg.GetJSTypeName(f.Type, f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly)))
					f.Features.Set(Features, FInputType, cg.GetJSInputTypeName(f.Type, f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly)))
					if title, ok := f.Annotations.GetBoolAnnotation(Annotation, AnnotationTitle); ok && title {
						tf = f
					}
					if tf == nil && f.Type.Type == gen.TipString {
						tf = f
					}
				}
				if tf != nil {
					if title, ok := tf.Annotations.GetBoolAnnotation(Annotation, AnnotationTitle); !ok || !title {
						tf.Annotations.AddTag(Annotation, AnnotationTitle, true)
					}
					an.SetTag(AnnotationTitle, tf.Name)
					// t.Features.Set(JSFeatures, JSFTitle, tf.Name)
				}
				// for _, m := range t.Methods {
				// 	if _, ok := m.Annotations.GetStringAnnotation(gen.GQLAnnotation, gen.GQLAnnotationNameTag); !ok {
				// 		m.Annotations.AddTag(jsAnnotation, jsAnnotationName, cg.GetJSMethodName(m))
				// 	}
				// }
			}
		}
	}
	return nil
}

func (cg *GQLCLientGenerator) Generate(b *gen.Builder) (err error) {
	cg.desc = b.Descriptor
	fileName := b.File.Name + ".ts"
	p := path.Join(cg.getOutputDir(), fileName)
	outFile, err := os.Create(p)
	if err != nil {
		return
	}
	defer outFile.Close()
	outFile.WriteString(fmt.Sprintf("/*Code generated from file %s by vivgen. DO NOT EDIT.*/", b.File.FileName))
	outFile.WriteString(apolloClientInclude)
	outFile.WriteString(gqlInclude)
	imports := Imports{}
	cfc := CodeFragmentContext{
		FileName: fileName,
		File:     b.File,
		Output:   outFile,
		Imports:  imports,
	}
	b.Project.ProvideCodeFragment(CodeFragmentModule, CodeFragmentActionImport, nil, cfc, false)
	for _, t := range b.File.Entries {
		for _, f := range t.GetFields(true, true) {
			if t, tt := cg.getTypeForImport(f.Parent().Pckg, f.Type, false); tt != nil && tt.File != b.File {
				imports.addImport(tt.File.Name, t)
				if f.Type.NonNullable && f.Type.Type != "" {
					ft, ok := cg.desc.FindType(f.Type.Type)
					if ok {
						imports.addImport(tt.File.Name, ft.Entity().FS(Features, FInstanceGenerator))
					}
				}
			}
			if t, tt := cg.getTypeForImport(f.Parent().Pckg, f.Type, true); tt != nil && tt.File != b.File {
				imports.addImport(tt.File.Name, t)
			}
		}
		for _, m := range t.Methods {
			for _, p := range m.Params {
				if t, tt := cg.getTypeForImport(t.Pckg, p.Type, true); tt != nil && tt.File != b.File {
					imports.addImport(tt.File.Name, t)
				}
			}
			if t, tt := cg.getTypeForImport(t.Pckg, m.RetValue, false); tt != nil && tt.File != b.File {
				imports.addImport(tt.File.Name, t)
			}
		}
	}
	for fn, tt := range imports {
		outFile.WriteString(fmt.Sprintf("import {%s} from './%s';\n", strings.Join(tt, ", "), fn))
	}
	if cg.useNS {
		outFile.WriteString(fmt.Sprintf("namespace %s {", b.File.Package))
	}
	for _, t := range b.File.Entries {
		err := cg.generateQueriesFile(outFile, t)
		if err != nil {
			return err
		}
		//t.Annotations.AddTag(jsAnnotation, jsAnnotationFilePath, p)
		t.Features.Set(Features, FFilePath, p)
	}
	outFile.WriteString(cleanInputFunc)
	if cg.useNS {
		outFile.WriteString("}")
	}
	b.Project.ProvideCodeFragment(CodeFragmentModule, CodeFragmentActionFile, nil, cfc, false)
	if cfc.Error != nil {
		return cfc.Error
	}
	return cg.GenerateVivard()
}

func (cg *GQLCLientGenerator) getTypeForImport(pckg *gen.Package, ref *gen.TypeRef, forInput bool) (string, *gen.Entity) {
	if ref.Array != nil {
		return cg.getTypeForImport(pckg, ref.Array, forInput)
	} else if ref.Map != nil {
		return "", nil
	} else {
		switch ref.Type {
		case gen.TipBool, gen.TipString, gen.TipInt, gen.TipFloat, gen.TipDate:
			return "", nil
		default:
			dt, _ := pckg.FindType(ref.Type)
			if dt == nil {
				cg.desc.AddError(fmt.Errorf("type not found: %s", ref.Type))
				return "", nil
			}
			e := dt.Entity()
			if e.HasModifier(gen.TypeModifierSingleton) || e.HasModifier(gen.TypeModifierExternal) {
				return "", nil
			}

			if forInput {
				return cg.GetJSEntityInputTypeName(ref.Type), dt.Entity()
			}
			return cg.GetJSEntityTypeName(ref.Type), dt.Entity()
		}
	}
}

type templateHolder struct {
	templ *template.Template
	err   error
}

type ArgDef struct {
	Name         string
	Type         string
	JSType       string
	ExcessFields []string
	NotNull      bool
	Optional     bool
}

type QueryDef struct {
	Request   string
	QueryName string
	VarName   string
	JSArgType string
	JSRetType string
	Args      []ArgDef
	Fields    []string
}

func (cg *GQLCLientGenerator) generateQueriesFile(wr io.Writer, e *gen.Entity) (err error) {
	// wr := os.Stdout

	tip := template.New("TYPE").
		Funcs(cg.getFuncsMap())
	if !e.HasModifier(gen.TypeModifierSingleton) && !e.HasModifier(gen.TypeModifierExternal) {
		tip, err = tip.Parse(typeTemplate)
		if err != nil {
			return err
		}
		err = tip.Execute(wr, e)
		if err != nil {
			return err
		}
		tip, err = tip.Parse(inputTypeTemplate)
		if err != nil {
			return err
		}
		err = tip.Execute(wr, e)
		if err != nil {
			return err
		}
	}

	if !e.HasModifier(gen.TypeModifierTransient) && !e.HasModifier(gen.TypeModifierEmbeddable) &&
		!e.HasModifier(gen.TypeModifierSingleton) && !e.HasModifier(gen.TypeModifierExternal) {
		isCfg := e.HasModifier(gen.TypeModifierConfig)
		var idfld *gen.Field
		if !isCfg {
			idfld = e.GetIdField()
		}
		fields := e.GetFields(true, true)
		for i := gen.GQLOperationGet; i < gen.GQLOperationLast; i++ {
			if qn, ok := e.Features.GetString(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[i]); ok {
				ad := []ArgDef{}
				req := "query"
				jsarg := ""
				rt := e.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")
				switch i {
				case gen.GQLOperationGet:
					if !isCfg {
						idt, ok := idfld.Features.GetString(gen.GQLFeatures, gen.GQLFTypeTag)
						if !ok {
							return fmt.Errorf("no type found for %s", idfld.Name)
						}
						ad = []ArgDef{{Name: idfld.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, ""),
							Type:    idt,
							JSType:  e.FS(Features, FIDType), //cg.GetJSTypeName(idfld.Type),
							NotNull: true}}
					}
				case gen.GQLOperationCreate, gen.GQLOperationSet:
					if e.FB(gen.FeaturesCommonKind, gen.FCReadonly) {
						continue
					}
					if i == gen.GQLOperationSet || !isCfg {
						req = "mutation"
						jstype := cg.GetJSEntityTypeName(e.Name)
						if i == gen.GQLOperationSet {
							jstype = cg.GetJSEntityInputTypeName(e.Name)
						}
						ad = []ArgDef{{
							Name:    "val",
							Type:    e.Features.String(gen.GQLFeatures, gen.GQLFInputTypeName),
							JSType:  jstype,
							NotNull: true,
						}}

					}
				case gen.GQLOperationList:
					if !isCfg {
						rt += "[]"
						if e.FB(gen.FeatureDictKind, gen.FDQualified) {
							qt, _ := e.Features.GetEntity(gen.FeatureDictKind, gen.FDQualifierType)
							qualIDFld := qt.GetIdField()
							var qualGQLType string
							var jstype string
							switch qualIDFld.Type.Type {
							case gen.TipInt:
								jstype = "number[]"
								qualGQLType = "[Int]"
							case gen.TipString:
								jstype = "string[]"
								qualGQLType = "[String]"
							}
							ad = []ArgDef{{Name: "quals",
								Type:     qualGQLType,
								JSType:   jstype,
								NotNull:  false,
								Optional: true,
							}}
						}
					} else {
						continue
					}
				case gen.GQLOperationLookup:
					if !isCfg {
						ad = []ArgDef{{Name: "query",
							Type:    "String!",
							JSType:  "string",
							NotNull: true}}
						rt += "[]"
					} else {
						continue
					}
				case gen.GQLOperationDelete:
					if !isCfg {
						idt, ok := idfld.Features.GetString(gen.GQLFeatures, gen.GQLFTypeTag)
						if !ok {
							return fmt.Errorf("no type found for %s", idfld.Name)
						}
						ad = []ArgDef{{Name: idfld.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, ""),
							Type:    idt,
							JSType:  e.FS(Features, FIDType), //cg.GetJSTypeName(idfld.Type),
							NotNull: true}}
						req = "mutation"
						rt = "boolean"
					} else {
						continue
					}
				case gen.GQLOperationFind:
					if it, ok := e.Features.GetEntity(gen.FeaturesAPIKind, gen.FAPIFindParamType); ok {
						ad = []ArgDef{{
							Name:    "query",
							Type:    it.Features.String(gen.GQLFeatures, gen.GQLFInputTypeName),
							JSType:  cg.GetJSEntityTypeName(it.Name),
							NotNull: true,
						}}
						rt += "[]"
					} else {
						continue
					}
				}
				params := QueryDef{
					Request:   req,
					QueryName: qn,
					JSArgType: jsarg,
					VarName:   qn + "Request",
					JSRetType: rt,
					Args:      ad,
					Fields:    make([]string, len(fields)),
				}
				if i == gen.GQLOperationDelete {
					params.Fields = []string{}
				} else {
					j := 0
					for _, f := range fields {
						//TODO exclude dictionaries for config
						if s, ok := f.Annotations.GetBoolAnnotation(gen.GQLAnnotation, gen.GQLAnnotationSkipTag); ok && s {
							continue
						}
						if n, ok := f.Annotations.GetStringAnnotation(Annotation, AnnotationName); ok {
							if f.Type.Complex {
								if f.Type.Array != nil {
									if i == gen.GQLOperationGet ||
										(i == gen.GQLOperationFind &&
											f.Annotations.GetBoolAnnotationDef(Annotation, AnnotationForceForFind, false)) {
										n, err = cg.getQueryForEmbeddedType(n, f, e)
										if err != nil {
											cg.desc.AddWarning(fmt.Sprintf("at %v: %v", f.Pos, err))
											continue
										}
									} else if i == gen.GQLOperationFind &&
										(f.HasModifier(gen.AttrModifierEmbeddedRef) || !f.Type.Array.Complex) {
										// nothing to do - it just n and should be
									} else {
										n = ""
									}
								} else {
									n, err = cg.getQueryForEmbeddedType(n, f, e)
									if err != nil {
										cg.desc.AddWarning(fmt.Sprintf("at %v: %v", f.Pos, err))
										continue
									}
								}
							}
							params.Fields[j] = n
							if n != "" {
								j++
							}
						}
					}
					if len(params.Fields) > j {
						params.Fields = params.Fields[:j]
					}
				}
				th := &templateHolder{templ: template.New("QUERY_VAR")}
				th.parse(queryTemplate).
					parse(queryFunctionTemplate).
					parse(queryTemplateVar)
				if th.err != nil {
					fmt.Printf("Error while parsing template: %v\n", th.err)
					return nil
				}
				err = th.templ.Execute(wr, params)
			}
		}
	}
	return cg.processMethods(wr, e)
}

func (cg *GQLCLientGenerator) processMethods(wr io.Writer, e *gen.Entity) (err error) {
	for _, m := range e.Methods {
		ad := []ArgDef{}
		req := m.FS(gen.GQLFeatures, gen.GQLFMethodType)
		jsarg := ""
		rt := cg.GetJSTypeName(m.RetValue, false)
		idfld := e.GetIdField()
		if idfld != nil {
			if idt, ok := idfld.Features.GetString(gen.GQLFeatures, gen.GQLFTypeTag); ok {
				ad = append(ad, ArgDef{Name: "id",
					Type:    idt,
					JSType:  e.FS(Features, FIDType),
					NotNull: true})
			}
		}
		for _, a := range m.Params {
			var gqlType string
			if a.Type.Complex {
				if t, ok := cg.desc.FindType(a.Type.Type); ok {
					gqlType = t.Entity().Features.String(gen.GQLFeatures, gen.GQLFInputTypeName)
				} else {
					cg.desc.AddWarning(fmt.Sprintf("at %v: type %s not found for parameter; skipping", a.Pos, a.Type.Type))
				}
			} else {
				gqlType = a.Features.String(gen.GQLFeatures, gen.GQLFTypeTag)

			}
			ad = append(ad,
				ArgDef{Name: a.Name,
					Type:    gqlType,
					JSType:  cg.GetJSInputTypeName(a.Type, false),
					NotNull: a.Type.NonNullable})
		}
		qn := m.FS(gen.GQLFeatures, gen.GQLFMethodName)
		var retval string
		fields := []*gen.Field{}
		if m.RetValue != nil && m.RetValue.Complex {
			//TODO process ret value of array of arrays
			retval = m.RetValue.Type
			if m.RetValue.Array != nil {
				retval = m.RetValue.Array.Type
			}
			rettype, ok := cg.desc.FindType(retval)
			if !ok {
				return fmt.Errorf("at %v: type '%s' not found", m.Pos, retval)
			}
			fields = rettype.Entity().GetFields(true, true)
		}
		params := QueryDef{
			Request:   req,
			QueryName: qn,
			JSArgType: jsarg,
			VarName:   qn + "Request",
			JSRetType: rt,
			Args:      ad,
			Fields:    make([]string, len(fields)),
		}
		j := 0
		if m.RetValue != nil && m.RetValue.Complex {
			for _, f := range fields {
				if s, ok := f.Annotations.GetBoolAnnotation(gen.GQLAnnotation, gen.GQLAnnotationSkipTag); ok && s {
					continue
				}
				if n, ok := f.Annotations.GetStringAnnotation(Annotation, AnnotationName); ok {
					if f.Type.Complex {
						// if f.Type.Array != nil {
						n, err = cg.getQueryForEmbeddedType(n, f, e)
						if err != nil {
							cg.desc.AddWarning(fmt.Sprintf("at %v: %v", f.Pos, err))
							continue
						}
						// }
					}
					params.Fields[j] = n
					if n != "" {
						j++
					}
				}
			}
		}
		if len(params.Fields) > j {
			params.Fields = params.Fields[:j]
		}
		th := &templateHolder{templ: template.New("QUERY_VAR")}
		th.parse(queryTemplate).
			parse(queryFunctionTemplate).
			parse(queryTemplateVar)
		if th.err != nil {
			fmt.Printf("Error while parsing template: %v\n", th.err)
			return nil
		}
		err = th.templ.Execute(wr, params)
	}
	return nil
}
func (th *templateHolder) parse(str string) *templateHolder {
	if th.err != nil {
		return th
	}
	th.templ, th.err = th.templ.Parse(str)
	return th
}

func (cg *GQLCLientGenerator) GetJSTypeName(ref *gen.TypeRef, asRef bool) (ret string) {
	if ref.Array != nil {
		params := cg.GetJSTypeName(ref.Array, asRef)
		ret = fmt.Sprintf("%s[]", params)
	} else if ref.Map != nil {
		valtype := "string"
		if ref.Map.ValueType.Type == gen.TipInt {
			valtype = "number"
		}
		ret = fmt.Sprintf("{key: string, val: %s}[]", valtype)
	} else {
		if asRef {
			if t, ok := cg.desc.FindType(ref.Type); ok {
				if idfld := t.Entity().GetIdField(); idfld != nil {
					ref = idfld.Type
				}
			}
		}
		if ret == "" {
			switch ref.Type {
			case gen.TipBool:
				ret = "boolean"
			case gen.TipString:
				ret = "string"
			case gen.TipInt, gen.TipFloat:
				ret = "number"
			case gen.TipDate:
				ret = "string"
			default:
				ret = cg.GetJSEntityTypeName(ref.Type)
			}
		}
	}
	return ret
}

func (cg *GQLCLientGenerator) GetJSInputTypeName(ref *gen.TypeRef, asRef bool) (ret string) {
	if ref.Array != nil {
		params := cg.GetJSInputTypeName(ref.Array, asRef)
		ret = fmt.Sprintf("%s[]", params)
	} else if ref.Map != nil {
		valtype := "string"
		if ref.Map.ValueType.Type == gen.TipInt {
			valtype = "number"
		}
		ret = fmt.Sprintf("{key: string, val: %s}[]", valtype)
	} else {
		if asRef {
			if t, ok := cg.desc.FindType(ref.Type); ok {
				if idfld := t.Entity().GetIdField(); idfld != nil {
					ref = idfld.Type
				}
			}
		}
		if ret == "" {
			switch ref.Type {
			case gen.TipBool:
				ret = "boolean"
			case gen.TipString:
				ret = "string"
			case gen.TipInt, gen.TipFloat:
				ret = "number"
			case gen.TipDate:
				ret = "string"
			default:
				ret = cg.GetJSEntityInputTypeName(ref.Type)
			}
		}
	}
	return ret
}

func (cg *GQLCLientGenerator) GetJSEmptyVal(ref *gen.TypeRef) (ret string) {
	if ref.Array != nil || ref.Map != nil {
		ret = "[]"
	} else {
		if ref.NonNullable {
			initType := ref.Type
			if ref.Complex {
				t, ok := cg.desc.FindType(initType)
				if ok {
					return t.Entity().FS(Features, FInstanceGenerator) + "()"
				}
				// in case of only id will send to client:
				// t, ok := cg.desc.FindType(initType)
				// if ok {
				// 	fld := t.Entity().GetIdField()
				// 	if fld != nil {
				// 		initType = fld.Type.Type
				// 	}

				// }
			}
			switch initType {
			case gen.TipBool:
				ret = "false"
			case gen.TipString:
				ret = "\"\""
			case gen.TipInt, gen.TipFloat:
				ret = "0"
			case gen.TipDate:
				ret = "new Date().toISOString()" //"\"1970-01-01 00:00:00\""
			default:
				ret = "null"
			}
		} else {
			ret = "null"
		}
	}
	return ret
}

func (cg *GQLCLientGenerator) GetJSEntityTypeName(name string) string {
	if strings.Index(name, ".") != -1 {
		parts := strings.SplitN(name, ".", 2)
		name = parts[1]
	}
	return name + "Type"
}

func (cg *GQLCLientGenerator) GetJSEntityInputTypeName(name string) string {
	if strings.Index(name, ".") != -1 {
		parts := strings.SplitN(name, ".", 2)
		name = parts[1]
	}
	return name + "__InputType"
}

func (cg *GQLCLientGenerator) getOutputDir() (ret string) {
	ret = "./gql-ts"
	if cg.outputPath != "" {
		ret = cg.outputPath
	}

	ret = path.Join(ret, "types")
	os.MkdirAll(ret, os.ModeDir|os.ModePerm)
	return
}
func (cg *GQLCLientGenerator) getQueryForEmbeddedType(field string, f *gen.Field, baseType *gen.Entity) (ret string, err error) {
	var t *gen.TypeRef
	isConfig := baseType.HasModifier(gen.TypeModifierConfig)
	if f.Type.Array != nil {
		t = f.Type.Array
	} else if f.Type.Map != nil {
		//TODO: add val specs for complex types
		return fmt.Sprintf("%s { key val }", field), nil
	} else {
		t = f.Type
	}
	if tt, ok := f.Parent().Pckg.FindType(t.Type); ok || !t.Complex {
		id := ""
		title := ""
		if isConfig && tt.Entity().HasModifier(gen.TypeModifierDictionary) {
			return
		}
		if f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly) {
			ret = field
			return
		}
		full := f.Type.Embedded || f.Annotations.GetBoolAnnotationDef(Annotation, AnnotationForce, false) /* && f.Features.Bool(gen.FeaturesDBKind, gen.FDBIncapsulate) */
		if ok {
			for _, ff := range tt.Entity().GetFields(true, true) {
				if s, ok := f.Annotations.GetBoolAnnotation(gen.GQLAnnotation, gen.GQLAnnotationSkipTag); ok && s {
					continue
				}
				if ff.Type.Complex {
					if n, ok := ff.Annotations.GetStringAnnotation(Annotation, AnnotationName); ok {
						r, e := cg.getQueryForEmbeddedType(n, ff, baseType)
						if e != nil {
							return "", e
						}
						id = fmt.Sprintf("%s %s", id, r)
						continue
					}
				}
				if full || ff.FB(Features, FForceLoadForField) || ff.Annotations.GetBoolAnnotationDef(Annotation, AnnotationForce, false) {
					if skip, ok := ff.Features.GetBool(gen.FeaturesAPIKind, gen.FCIgnore); !ok || !skip {
						id = fmt.Sprintf("%s %s", id, ff.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, ""))
					}
				} else if ff.IsIdField() {
					id = ff.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "id")
				} else if tit, ok := ff.Annotations.GetBoolAnnotation(Annotation, AnnotationTitle); ok && tit {
					title = fmt.Sprintf("%s %s ", title, ff.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "id"))
				} else if tit, ok := ff.Annotations.GetBoolAnnotation(Annotation, AnnotationIcon); ok && tit {
					title = fmt.Sprintf("%s %s ", title, ff.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "id"))
				} else if tit, ok := ff.Annotations.GetBoolAnnotation(Annotation, AnnotationColor); ok && tit {
					title = fmt.Sprintf("%s %s ", title, ff.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "id"))
				}
			}
		}
		if strings.Trim(id, " \t") != "" || strings.Trim(title, " \t") != "" {
			ret = fmt.Sprintf("%s { %s %s }", field, id, title)
		} else {
			if !isConfig {
				ret = field
			}
		}
	} else {
		err = fmt.Errorf("type %s not found for %s", t.Type, field)
	}
	return
}

// Add adds str to file and append \n
func (cfc CodeFragmentContext) Add(str string) error {
	if cfc.Output != nil {
		_, err := cfc.Output.WriteString(str)
		if err == nil {
			_, err = cfc.Output.WriteString("\n")
		}
		if err != nil {
			cfc.Error = err
		}
		return err
	}
	return errors.New("output is not initialized")
}

func (cfc CodeFragmentContext) AddImport(file string, tip string) {
	cfc.Imports.addImport(file, tip)
}

func (cg *GQLCLientGenerator) getFuncsMap() template.FuncMap {
	return template.FuncMap{
		"TypeName": func(e *gen.Entity) string {
			return e.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")
		},
		"InstanceGenerator": func(e *gen.Entity) string {
			return e.FS(Features, FInstanceGenerator)
		},
		"FieldName": func(f *gen.Field) string {
			return f.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")
		},
		"Nullable": func(f *gen.Field) bool { return !f.Type.NonNullable },
		"FieldType": func(f *gen.Field) string {
			// return f.Annotations.GetStringAnnotationDef(Annotation, AnnotationType, "")
			return f.FS(Features, FType)
		},
		"EmptyVal": func(f *gen.Field) string { return cg.GetJSEmptyVal(f.Type) },
		"NeedInit": func(f *gen.Field) bool {
			return !f.Annotations.GetBoolAnnotationDef(gen.GQLAnnotation, gen.GQLAnnotationSkipTag, false) &&
				(f.Type.NonNullable && f.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "") != "" ||
					f.HasModifier(gen.AttrModifierEmbedded))
		},
		"Init": func(f *gen.Field) string {
			if f.HasModifier(gen.AttrModifierEmbedded) {
				if f.Type.Array != nil || f.Type.Map != nil {
					return "[]"
				}
				t, ok := cg.desc.FindType(f.Type.Type)
				if ok {
					return t.Entity().FS(Features, FInstanceGenerator) + "()"
				}
				return "null"
			} else if f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly) {
				if f.Type.Array != nil || f.Type.Map != nil {
					return "[]"
				}
				t := f.Parent()
				if idfld := t.GetIdField(); idfld != nil {
					return cg.GetJSEmptyVal(idfld.Type)
				}
			}
			return cg.GetJSEmptyVal(f.Type)
		},
		"GetFields": func(e *gen.Entity) []*gen.Field {
			return e.GetFields(true, true)
		},
		"RequiresInput": func(o interface{}) bool {
			switch e := o.(type) {
			case *gen.Entity:
				return e.Annotations.GetStringAnnotationDef(Annotation, AnnotationInputName, "") != ""
			case *gen.Field:
				return e.Annotations.GetStringAnnotationDef(Annotation, AnnotationInputName, "") != ""
			}
			return false
		},
		"InputTypeName": func(e *gen.Entity) string {
			return e.Annotations.GetStringAnnotationDef(Annotation, AnnotationInputName, "")
		},
		"InputFieldName": func(f *gen.Field) string {
			return f.Annotations.GetStringAnnotationDef(Annotation, AnnotationInputName, "")
		},
		"InputFieldType": func(f *gen.Field) string {
			return f.FS(Features, FInputType)
		},
		"SetNullFields": func(e *gen.Entity) []string {
			var ret []string
			for _, field := range e.GetFields(true, false) {
				if setNullField := field.FS(gen.GQLFeatures, gen.GQLFSetNullInputField); setNullField != "" {
					ret = append(ret, setNullField)
				}
			}
			return ret
		},
	}
}

const cleanInputFunc = `
function cleanInput(inp: any): any {
  if(!inp)
    return inp;
  if(Array.isArray(inp)) {
    inp.forEach(o => cleanInput(o));
  } else {
    if(typeof inp == "object") {
      if (inp.__typename !== undefined)
        delete inp.__typename;
      for(let k in inp) {
        if (typeof inp[k] == "object")
          cleanInput(inp[k]);
      }
    }
  }
  return inp;
}
`

const queryTemplate = `
{{define "QUERY"}}
  {{.Request}} {{.QueryName}}{{if gt (len .Args) 0}}({{range $idx, $arg := .Args}}{{if gt $idx 0}}, {{end}}${{$arg.Name}}:{{$arg.Type}}{{end}}){{end}} {
    {{.QueryName}}{{if gt (len .Args) 0}}({{range $idx, $arg := .Args}}{{if gt $idx 0}},{{end}}{{$arg.Name}}:${{$arg.Name}}{{end}}){{end}} {{if gt (len .Fields) 0}}{ {{range .Fields}}
        {{.}}{{end}}
    } {{end}}
  }
{{end}}
`
const queryFunctionTemplate = `
{{define "FUNCTION"}}
export async function {{.QueryName}}(apollo: ApolloClient<any>, {{if .JSArgType}}arg: {{.JSArgType}}{{else}}{{range $idx, $arg := .Args}}{{if gt $idx 0}}, {{end}}{{$arg.Name}}{{if $arg.Optional}}?{{end}}:{{$arg.JSType}}{{end}}{{end}}): Promise<{{.JSRetType}}> {
  let res = await apollo.query({
      query: {{.VarName}},
      fetchPolicy: "no-cache",
      variables: {{if .JSArgType}}arg{{else}} { {{range $idx, $arg := .Args}}{{if gt $idx 0}}, {{end}}{{$arg.Name}}:cleanInput({{$arg.Name}}){{end}} } {{end}}
    });
  if(res.data.{{.QueryName}})
    return res.data.{{.QueryName}};
  else
    throw res.errors;
}
{{end}}
`
const typeTemplate = `
export type {{TypeName .}} = { {{range GetFields .}} {{if ne (FieldName .) "" }}{{FieldName .}}{{if Nullable .}}?{{end}}: {{FieldType .}},{{end}}{{end}} };
export function {{InstanceGenerator .}}(): {{TypeName .}} {
  return {
    {{range GetFields .}}{{if NeedInit . }}{{FieldName .}}: {{Init .}},{{end}}
    {{end}}
  }
}
`
const inputTypeTemplate = `
{{if RequiresInput .}}
export type {{InputTypeName .}} = { {{range GetFields .}} {{if RequiresInput . }}{{InputFieldName .}}{{if not .IsIdField}}?{{end}}: {{InputFieldType .}},{{end}}{{end}} {{range SetNullFields .}} {{.}}?: boolean, {{end}}};
export function New{{InputTypeName .}}Instance(): {{InputTypeName .}} {
  return {
    {{range GetFields .}}{{if .IsIdField}}{{FieldName .}}: {{Init .}},{{end}}{{end}}
  }
}
{{end}}
`
const queryTemplateVar = "export const {{.VarName}} = gql`{{template \"QUERY\" .}}`\n {{template \"FUNCTION\" .}}\n"
