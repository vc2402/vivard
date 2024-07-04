package js

import (
  "errors"
  "fmt"
  "io"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "text/template"

  "github.com/vc2402/vivard/gen"
)

const (
  GQLClientGeneratorName   = "GraphQL-TS"
  GQLClientOptions         = "gql-ts"
  GQLClientPathOption      = "path"
  GQLClientNamespaceOption = "useNamespace"

  Annotation             = "js"
  AnnotationName         = "name"
  AnnotationInputName    = "inputName"
  AnnotationSkip         = "skip"
  AnnotationTitle        = "title"
  AnnotationIcon         = "icon"
  AnnotationColor        = "color"
  AnnotationForce        = "force"
  AnnotationForceForFind = "findForce"
)

const (
  Features           = "js"
  FName              = "name"
  FType              = "type"
  FInputType         = "input-type"
  FFillInputFuncName = "fill-inp-func"
  FFile              = "file"
  FIDType            = "id_type"
  FFilePath          = "filepath"
  FInstanceGenerator = "instance_generator"
  // FForceLoadForField - force load field when its parent is included in other object (usually only id and title are loaded)
  FForceLoadForField = "field_force_load"
  // FFunctionName returns function that returns name of GQL operation function for Entity (arg - GQLOperationKind)
  FFunctionName = "gql_function_name"
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
    p := cg.getFilePathForName(file.Name)
    for _, t := range file.Entries {
      t.Features.Set(Features, FFilePath, p)
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
        jsInputType := t.Annotations.GetStringAnnotationDef(
          Annotation,
          AnnotationInputName,
          cg.GetJSEntityInputTypeName(t.Name),
        )
        t.Features.Set(Features, FFillInputFuncName, "fill"+strings.ToUpper(jsInputType[:1])+jsInputType[1:])

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
        t.Features.Set(
          Features,
          FInstanceGenerator,
          "New"+t.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")+"Instance",
        )
        var tf *gen.Field
        for _, f := range t.GetFields(true, true) {
          if s, ok := f.Annotations.GetBoolAnnotation(gen.GQLAnnotation, gen.GQLAnnotationSkipTag); ok && s {
            continue
          }
          if fld, ok := f.Annotations.GetStringAnnotation(gen.GQLAnnotation, gen.GQLAnnotationNameTag); ok {
            f.Annotations.AddTag(Annotation, AnnotationName, fld)
            if !f.FB(gen.FeaturesCommonKind, gen.FCReadonly) && !f.HasModifier(gen.AttrModifierCalculated) {
              f.Annotations.AddTag(Annotation, AnnotationInputName, fld)
            }
          }
          // f.Annotations.AddTag(Annotation, AnnotationType, cg.GetJSTypeName(f.Type))
          f.Features.Set(
            Features,
            FType,
            cg.GetJSTypeName(
              f.Type,
              f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly),
            ),
          )
          f.Features.Set(
            Features,
            FInputType,
            cg.GetJSInputTypeName(
              f.Type,
              f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly),
            ),
          )
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
    for _, enum := range file.Enums {
      tname := cg.GetJSEntityTypeName(enum.Name)
      enum.Features.Set(Features, FName, tname)
      enum.Features.Set(Features, FFilePath, p)
      enum.Features.Set(Features, FType, cg.GetJSTypeNameByVivardName(enum.AliasForType))
    }
  }
  return nil
}

func (cg *GQLCLientGenerator) Generate(b *gen.Builder) (err error) {
  cg.desc = b.Descriptor
  fileName := b.File.Name + ".ts"
  outFile, err := os.Create(cg.getFilePathForName(b.File.Name))
  if err != nil {
    return
  }
  defer outFile.Close()
  outFile.WriteString(fmt.Sprintf("/*Code generated from file %s by vivgen. DO NOT EDIT.*/\n\n", b.File.FileName))
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
  for _, e := range b.File.Enums {
    err = cg.generateEnum(outFile, e)
    if err != nil {
      return err
    }
  }
  for _, t := range b.File.Entries {
    for _, f := range t.GetFields(true, true) {
      if t, tf := cg.getTypeForImport(f.Parent().Pckg, f.Type, false); tf != nil && tf != b.File {
        imports.addImport(tf.Name, t)
        if f.Type.NonNullable && f.Type.Type != "" {
          ft, ok := cg.desc.FindType(f.Type.Type)
          if ok && ft.Entity() != nil {
            imports.addImport(tf.Name, ft.Entity().FS(Features, FInstanceGenerator))
          }
        }
      }
      if t, tf := cg.getTypeForImport(f.Parent().Pckg, f.Type, true); tf != nil && tf != b.File {
        imports.addImport(tf.Name, t)
      }
    }
    for _, m := range t.Methods {
      for _, p := range m.Params {
        if t, tf := cg.getTypeForImport(t.Pckg, p.Type, true); tf != nil && tf != b.File {
          imports.addImport(tf.Name, t)
        }
      }
      if t, tf := cg.getTypeForImport(t.Pckg, m.RetValue, false); tf != nil && tf != b.File {
        imports.addImport(tf.Name, t)
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

func (cg *GQLCLientGenerator) ProvideFeature(
  kind gen.FeatureKind,
  name string,
  obj interface{},
) (feature interface{}, ok gen.ProvideFeatureResult) {
  switch kind {
  case Features:
    switch name {
    case FFunctionName:
      if e, isEntity := obj.(*gen.Entity); isEntity {
        return gen.FeatureFunc(
            func(args ...interface{}) (any, error) {
              if len(args) > 0 {
                if tip, ok := args[0].(gen.GQLOperationKind); ok {
                  return cg.getOperationFunctionName(tip, e), nil
                }
              }
              return "", fmt.Errorf("feature %s:%s expects GQLOperationKind arg", Features, FFunctionName)
            },
          ),
          gen.FeatureProvided
      }
    }
  }
  return nil, gen.FeatureNotProvided
}

func (cg *GQLCLientGenerator) getTypeForImport(pckg *gen.Package, ref *gen.TypeRef, forInput bool) (string, *gen.File) {
  if ref.Array != nil {
    return cg.getTypeForImport(pckg, ref.Array, forInput)
  } else if ref.Map != nil {
    return "", nil
  } else {
    switch ref.Type {
    case gen.TipBool, gen.TipString, gen.TipInt, gen.TipFloat, gen.TipDate, gen.TipAny:
      return "", nil
    default:
      dt, _ := pckg.FindType(ref.Type)
      if dt == nil {
        cg.desc.AddError(fmt.Errorf("type not found: %s", ref.Type))
        return "", nil
      }
      e := dt.Entity()
      if e != nil {
        if e.HasModifier(gen.TypeModifierSingleton) || e.HasModifier(gen.TypeModifierExternal) {
          return "", nil
        }

        if forInput {
          return cg.GetJSEntityInputTypeName(ref.Type), e.File
        }
        return cg.GetJSEntityTypeName(ref.Type), e.File
      } else if e := dt.Enum(); e != nil {
        if forInput {
          return cg.GetJSEntityInputTypeName(ref.Type), e.File
        }
        return cg.GetJSEntityTypeName(ref.Type), e.File
      } else {
        cg.desc.AddError(fmt.Errorf("at %v: type is not an Entity", dt.Position()))
        return "", nil
      }
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
  Request       string
  QueryName     string
  FuncName      string
  VarName       string
  JSArgType     string
  JSRetType     string
  Args          []ArgDef
  Fields        []string
  FillInputName string
}

var gqlFunctionsNamesTemplates = [...]string{
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
    tip, err = tip.Parse(fillInputFuncTemplate)
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
        inputInit := ""
        rt := e.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")
        switch i {
        case gen.GQLOperationGet:
          if !isCfg {
            idt, ok := idfld.Features.GetString(gen.GQLFeatures, gen.GQLFTypeTag)
            if !ok {
              return fmt.Errorf("no type found for %s", idfld.Name)
            }
            ad = []ArgDef{
              {
                Name:    idfld.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, ""),
                Type:    idt,
                JSType:  e.FS(Features, FIDType), //cg.GetJSTypeName(idfld.Type),
                NotNull: true,
              },
            }
          }
        case gen.GQLOperationCreate, gen.GQLOperationSet:
          if e.FB(gen.FeaturesCommonKind, gen.FCReadonly) {
            continue
          }
          if i == gen.GQLOperationSet || !isCfg {
            req = "mutation"
            jstype := cg.GetJSEntityTypeName(e.Name)
            inputInit = e.FS(Features, FFillInputFuncName)
            if i == gen.GQLOperationSet {
              jstype = cg.GetJSEntityInputTypeName(e.Name)
            }
            ad = []ArgDef{
              {
                Name:    "val",
                Type:    e.Features.String(gen.GQLFeatures, gen.GQLFInputTypeName),
                JSType:  jstype,
                NotNull: true,
              },
            }
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
              ad = []ArgDef{
                {
                  Name:     "quals",
                  Type:     qualGQLType,
                  JSType:   jstype,
                  NotNull:  false,
                  Optional: true,
                },
              }
            }
          } else {
            continue
          }
        case gen.GQLOperationLookup:
          if !isCfg {
            ad = []ArgDef{
              {
                Name:    "query",
                Type:    "String!",
                JSType:  "string",
                NotNull: true,
              },
            }
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
            ad = []ArgDef{
              {
                Name:    idfld.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, ""),
                Type:    idt,
                JSType:  e.FS(Features, FIDType), //cg.GetJSTypeName(idfld.Type),
                NotNull: true,
              },
            }
            req = "mutation"
            rt = "boolean"
          } else {
            continue
          }
        case gen.GQLOperationFind:
          if it, ok := e.Features.GetEntity(gen.FeaturesAPIKind, gen.FAPIFindParamType); ok {
            ad = []ArgDef{
              {
                Name:    "query",
                Type:    it.Features.String(gen.GQLFeatures, gen.GQLFInputTypeName),
                JSType:  cg.GetJSEntityTypeName(it.Name),
                NotNull: true,
              },
            }
            rt += "[]"
          } else {
            continue
          }
        case gen.GQLOperationBulkCreate:
          if !e.FB(gen.FeatGoKind, gen.FCGBulkNew) {
            continue
          }
          req = "mutation"
          rt += "[]"
          ad = []ArgDef{
            {
              Name:    "val",
              Type:    "[" + e.Features.String(gen.GQLFeatures, gen.GQLFInputTypeName) + "]",
              JSType:  rt, //cg.GetJSTypeName(idfld.Type),
              NotNull: true,
            },
          }
        case gen.GQLOperationBulkSet:
          if !e.FB(gen.FeatGoKind, gen.FCGBulkSet) {
            continue
          }
          req = "mutation"
          rt += "[]"
          ad = []ArgDef{
            {
              Name:    "val",
              Type:    "[" + e.Features.String(gen.GQLFeatures, gen.GQLFInputTypeName) + "]",
              JSType:  rt, //cg.GetJSTypeName(idfld.Type),
              NotNull: true,
            },
          }
        }
        params := QueryDef{
          Request:       req,
          QueryName:     qn,
          FuncName:      cg.getOperationFunctionName(i, e),
          JSArgType:     jsarg,
          VarName:       qn + "Request",
          JSRetType:     rt,
          Args:          ad,
          Fields:        make([]string, len(fields)),
          FillInputName: inputInit,
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
                      f.Annotations.GetBoolAnnotationDef(Annotation, AnnotationForceForFind, false)) ||
                    f.Annotations.GetBoolAnnotationDef(Annotation, AnnotationForce, false) {
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
          return fmt.Errorf("while parsing template for %s: %v", params.FuncName, th.err)
        }
        err = th.templ.Execute(wr, params)
        if err != nil {
          return fmt.Errorf("while executing template for %s: %v\n", params.FuncName, err)
        }
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
        ad = append(
          ad, ArgDef{
            Name:    "id",
            Type:    idt,
            JSType:  e.FS(Features, FIDType),
            NotNull: true,
          },
        )
      }
    }
    for _, a := range m.Params {
      var gqlType string
      //if a.Type.Complex {
      //  if t, ok := cg.desc.FindType(a.Type.Type); ok {
      //    gqlType = t.Entity().Features.String(gen.GQLFeatures, gen.GQLFInputTypeName)
      //  } else {
      //    cg.desc.AddWarning(fmt.Sprintf("at %v: type %s not found for parameter; skipping", a.Pos, a.Type.Type))
      //  }
      //} else {
      gqlType = a.Features.String(gen.GQLFeatures, gen.GQLFInputTypeName)
      //}
      //if a.Type.NonNullable {
      //  gqlType += "!"
      //}
      ad = append(
        ad,
        ArgDef{
          Name:    a.Name,
          Type:    gqlType,
          JSType:  cg.GetJSInputTypeName(a.Type, false),
          NotNull: a.Type.NonNullable,
        },
      )
    }
    qn := m.FS(gen.GQLFeatures, gen.GQLFMethodName)

    params := QueryDef{
      Request:   req,
      QueryName: qn,
      FuncName:  qn,
      JSArgType: jsarg,
      VarName:   qn + "Request",
      JSRetType: rt,
      Args:      ad,
    }
    j := 0
    if m.RetValue != nil && m.RetValue.Complex {
      var retval string
      fields := []*gen.Field{}
      ret := m.RetValue
      for ret.Array != nil {
        ret = ret.Array
      }
      if !ret.Embedded {
        retval = ret.Type
        if m.RetValue.Array != nil {
          retval = m.RetValue.Array.Type
        }
        rettype, ok := cg.desc.FindType(retval)
        if !ok || rettype.Entity() == nil && rettype.Enum() == nil {
          return fmt.Errorf("at %v: type '%s' not found or is not an Entity", m.Pos, retval)
        }
        if rettype.Entity() != nil {
          fields = rettype.Entity().GetFields(true, true)
          params.Fields = make([]string, len(fields))
        }
      }
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
        tip, err := t.IdType()
        if err != nil {
          return "any"
        }
        ref = &gen.TypeRef{Type: tip}
        //if t.Entity() != nil {
        //	if idfld := t.Entity().GetIdField(); idfld != nil {
        //		ref = idfld.Type
        //	}
        //} else if t.Enum() != nil {
        //
        //}
      }
    }
    ret = cg.GetJSTypeNameByVivardName(ref.Type)
  }
  return ret
}

func (cg *GQLCLientGenerator) GetJSTypeNameByVivardName(tip string) string {
  var ret string
  switch tip {
  case gen.TipBool:
    ret = "boolean"
  case gen.TipString:
    ret = "string"
  case gen.TipInt, gen.TipFloat:
    ret = "number"
  case gen.TipDate:
    ret = "string"
  case gen.TipAny:
    ret = "any"
  default:
    //if t, ok := cg.desc.FindType(ref.Type); ok {
    ret = cg.GetJSEntityTypeName(tip)
    //}
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
      if t, ok := cg.desc.FindType(ref.Type); ok && t.Entity() != nil {
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
        if ok && t.Entity() != nil {
          return t.Entity().FS(Features, FInstanceGenerator) + "()"
        }
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
        t, ok := cg.desc.FindType(initType)
        if ok && t.Enum() != nil {
          if len(t.Enum().Fields) > 0 {
            return cg.GetEnumFieldJSValue(t.Enum().Fields[0])
          }
          return "null"
        }

        ret = "null"
      }
    } else {
      ret = "null"
    }
  }
  return ret
}

func (cg *GQLCLientGenerator) GetEnumFieldJSValue(ef *gen.EnumField) string {
  if ef.IntVal != nil {
    return strconv.Itoa(*ef.IntVal)
  } else if ef.FloatVal != nil {
    return strconv.FormatFloat(*ef.FloatVal, 'G', -1, 32)
  } else if ef.StringVal != nil {
    return `"` + *ef.StringVal + `"`
  } else {
    // ordinal from 0
    for i, field := range ef.Parent.Fields {
      if field == ef {
        return strconv.Itoa(i)
      }
    }
  }
  return "0"
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

func (cg *GQLCLientGenerator) getFilePathForName(name string) string {
  fileName := name + ".ts"
  return filepath.Join(cg.getOutputDir(), fileName)
}

func (cg *GQLCLientGenerator) getOutputDir() (ret string) {
  ret = "./gql-ts"
  if opt := cg.desc.Options().ClientOutputDir; opt != "" {
    ret = opt
  }
  if cg.outputPath != "" {
    ret = cg.outputPath
  }

  ret = filepath.FromSlash(filepath.Join(ret, "types"))
  os.MkdirAll(ret, os.ModeDir|os.ModePerm)
  return
}
func (cg *GQLCLientGenerator) getQueryForEmbeddedType(
  field string,
  f *gen.Field,
  baseType *gen.Entity,
) (ret string, err error) {
  var t *gen.TypeRef
  isConfig := baseType.HasModifier(gen.TypeModifierConfig)
  if f.Type.Array != nil {
    t = f.Type.Array
    for t.Array != nil {
      t = t.Array
    }
  } else if f.Type.Map != nil {
    //TODO: add val specs for complex types
    return fmt.Sprintf("%s { key val }", field), nil
  } else {
    t = f.Type
  }
  if tt, ok := f.Parent().Pckg.FindType(t.Type); ok || !t.Complex {
    if !t.Complex {
      ret = field
      return
    }
    id := ""
    title := ""
    if isConfig && tt.Entity() != nil && tt.Entity().HasModifier(gen.TypeModifierDictionary) {
      return
    }
    if f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly) {
      ret = field
      return
    }
    full := f.Type.Embedded || f.Annotations.GetBoolAnnotationDef(
      Annotation,
      AnnotationForce,
      false,
    ) /* && f.Features.Bool(gen.FeaturesDBKind, gen.FDBIncapsulate) */
    if ok && tt.Entity() != nil {
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
        if full || ff.FB(Features, FForceLoadForField) || ff.Annotations.GetBoolAnnotationDef(
          Annotation,
          AnnotationForce,
          false,
        ) {
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

func (cg *GQLCLientGenerator) getOperationFunctionName(operation gen.GQLOperationKind, e *gen.Entity) string {
  return fmt.Sprintf(gqlFunctionsNamesTemplates[operation], e.Name[:1]+e.Name[1:])
}

func (cg *GQLCLientGenerator) getFuncsMap() template.FuncMap {
  return template.FuncMap{
    "TypeName": func(e *gen.Entity) string {
      return e.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")
    },
    "InstanceGenerator": func(e *gen.Entity) string {
      return e.FS(Features, FInstanceGenerator)
    },
    "FillInputName": func(e *gen.Entity) string {
      return e.FS(Features, FFillInputFuncName)
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
    "NeedFill": func(f *gen.Field) bool {
      return !f.Annotations.GetBoolAnnotationDef(gen.GQLAnnotation, gen.GQLAnnotationSkipTag, false) &&
        !f.HasModifier(gen.AttrModifierAuxiliary) &&
        !f.HasModifier(gen.AttrModifierCalculated) &&
        !f.FB(gen.FeaturesCommonKind, gen.FCReadonly)
    },
    "Init": func(f *gen.Field) string {
      if f.HasModifier(gen.AttrModifierEmbedded) {
        if f.Type.Array != nil || f.Type.Map != nil {
          return "[]"
        } else if !f.Type.NonNullable {
          return "undefined"
        }
        t, ok := cg.desc.FindType(f.Type.Type)
        if ok && t.Entity() != nil {
          return t.Entity().FS(Features, FInstanceGenerator) + "()"
        }
        return "null"
      } else if f.HasModifier(gen.AttrModifierEmbeddedRef) || f.FB(gen.GQLFeatures, gen.GQLFIDOnly) {
        if f.Type.Array != nil || f.Type.Map != nil {
          return "[]"
        }
        if t, ok := cg.desc.FindType(f.Type.Type); ok && t.Entity() != nil {
          if idfld := t.Entity().GetIdField(); idfld != nil {
            return cg.GetJSEmptyVal(idfld.Type)
          }
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
    "CallFill": func(f *gen.Field) bool {
      return false
    },
    "GetFillName": func(f *gen.Field) string {
      if t, ok := cg.desc.FindType(f.Type.Type); ok && t.Entity() != nil {
        return t.Entity().FS(Features, FFillInputFuncName)
      }
      return ""
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
    "EnumName": func(e *gen.Enum) string {
      return cg.GetJSEntityTypeName(e.Name)
    },
    "EnumInputName": func(e *gen.Enum) string {
      return cg.GetJSEntityInputTypeName(e.Name)
    },
    "EnumType": func(e *gen.Enum) string {
      return cg.GetJSTypeName(&gen.TypeRef{Type: e.AliasForType}, false)
    },
    "EnumFieldName": func(ef *gen.EnumField) string {
      return ef.Name
    },
    "EnumFieldValue": func(ef *gen.EnumField) string {
      return cg.GetEnumFieldJSValue(ef)
      //if ef.IntVal != nil {
      //	return strconv.Itoa(*ef.IntVal)
      //} else if ef.FloatVal != nil {
      //	return strconv.FormatFloat(*ef.FloatVal, 'G', -1, 32)
      //} else if ef.StringVal != nil {
      //	return `"` + *ef.StringVal + `"`
      //} else {
      //	// ordinal from 0
      //	for i, field := range ef.Parent.Fields {
      //		if field == ef {
      //			return strconv.Itoa(i)
      //		}
      //	}
      //}
      //return "0"
    },
  }
}

const cleanInputFunc = `
export function cleanInput(inp: any): any {
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

// TODO do not use cleanInput but call fill for complex fields and arrays
const fillInputFuncTemplate = `
export function {{FillInputName .}}(arg: {{InputTypeName .}}): {{InputTypeName .}} {
  return {
    {{range GetFields .}}{{if NeedFill . }}{{FieldName .}}: {{if CallFill .}}{{GetFillName .}}(arg.{{FieldName .}}){{else}}cleanInput(arg.{{FieldName .}}){{end}},
		{{end}}{{end}}{{range SetNullFields .}}{{.}}: arg.{{.}},
		{{end}}}
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
export async function {{.FuncName}}(apollo: ApolloClient<any>, {{if .JSArgType}}arg: {{.JSArgType}}{{else}}{{range $idx, $arg := .Args}}{{if gt $idx 0}}, {{end}}{{$arg.Name}}{{if $arg.Optional}}?{{end}}:{{$arg.JSType}}{{end}}{{end}}): Promise<{{.JSRetType}}> {
  let res = await apollo.query({
      query: {{.VarName}},
      fetchPolicy: "no-cache",
      variables: {{if .JSArgType}}arg{{else}} { {{range $idx, $arg := .Args}}{{if gt $idx 0}}, {{end}}{{$arg.Name}}:{{if ne $.FillInputName ""}} {{$.FillInputName}}({{$arg.Name}}){{else}}cleanInput({{$arg.Name}}){{end}}{{end}} } {{end}}
    });
  if(res.data.{{.QueryName}} !== undefined)
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
