package gen

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

type FeatureArguments struct {
	args    []*jen.Statement
	rest    []*jen.Statement
	idx     map[string]int
	boolArg bool
	desc    *Package
}

func (cg *CodeGenerator) getIDFromObjectFuncFeature(t *Entity) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("obj", "val", "ctx").parse(args)
		idfld := t.GetIdField()
		if idfld == nil {
			cg.proj.AddError(fmt.Errorf("%s: no id field", t.GetName()))
			return &jen.Statement{}
		}
		return a.get("obj").Dot(idfld.Name)
	}
}

func (cg *CodeGenerator) getFieldSetterFuncFeature(f *Field) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("obj", "val", "ctx").parse(args)

		if complex, ok := f.Features.GetBool(FeaturesCommonKind, FCComplexAccessor); ok && complex {
			return jen.Id("eng").Dot(cg.b.GetComplexMethodName(f.parent, f, CGSetComplexAttrMethod)).Params(
				a.get("ctx"),
				a.get("obj"),
				a.get("val"),
			)
		} else {
			return a.get("obj").Dot(cg.b.GetMethodName(f, CGSetterMethod)).Params(a.get("val"))
		}
	}
}

func (cg *CodeGenerator) getFieldGetterFuncFeature(f *Field) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("obj", "ctx").parse(args)

		if complex, ok := f.Features.GetBool(FeaturesCommonKind, FCComplexAccessor); ok && complex {
			return jen.Id("eng").Dot(cg.b.GetComplexMethodName(f.parent, f, CGGetComplexAttrMethod)).Params(a.get("ctx"), a.get("obj"))
		} else {
			return jen.ListFunc(func(g *jen.Group) {
				g.Add(a.get("obj").Dot(cg.b.GetMethodName(f, CGGetterMethod)).Params())
				if a.boolArg {
					g.Error().Parens(jen.Nil())
				}
			})
		}
	}
}

func (cg *CodeGenerator) getIsAttrNullFuncFeature(f *Field) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("obj").parse(args)
		n := strings.ToUpper(f.Name[:1]) + f.Name[1:]
		fname := fmt.Sprintf(cgMethodsTemplates[CGIsNullMethod], n)
		return a.get("obj").Dot(fname).Params()
	}
}

func (cg *CodeGenerator) getAttrIsEmptyFuncFeature(f *Field) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("obj").parse(args)
		isPointer, _ := cg.desc.GetFeature(f, FeaturesCommonKind, FCAttrIsPointer).(bool)
		if isPointer {
			return a.get("obj").Dot(f.Name).Op("==").Nil()
		} else if a.boolArg {
			return nil
		} else {
			return jen.False()
		}
	}
}

func (cg *CodeGenerator) getGetAttrValueFuncFeature(f *Field) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("obj").parse(args)
		isPointer, _ := cg.desc.GetFeature(f, FeaturesCommonKind, FCAttrIsPointer).(bool)
		if isPointer {
			return jen.Op("*").Add(a.get("obj")).Dot(f.Name)
		}
		return a.get("obj").Dot(f.Name)
	}
}

func (cg *CodeGenerator) getSetAttrValueFuncFeature(f *Field) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("obj", "val").parse(args)
		isPointer, _ := cg.desc.GetFeature(f, FeaturesCommonKind, FCAttrIsPointer).(bool)
		if isPointer {
			// TODO may be we need to New...
			// return jen.If(a.get("obj").Op("==").Nil())
			return a.get("obj").Dot(f.Name).Op("=").Op("&").Add(a.get("val"))
		}
		return a.get("obj").Dot(f.Name).Op("=").Add(a.get("val"))
	}
}

func (cg *CodeGenerator) getSingletonGetFuncFeature(e *Entity) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc, boolArg: true}
		a.init("id", "ctx", "eng").parse(args)
		if a.boolArg {
			return jen.List(a.get("eng").Dot(e.FS(FeatGoKind, FCGSingletonAttrName)), jen.Nil())
		}
		return a.get("eng").Dot(e.FS(FeatGoKind, FCGSingletonAttrName))
	}
}

func (cg *CodeGenerator) getEngineVarFuncFeature(f *Field) CodeHelperFunc {
	parts := strings.Split(f.Type.Type, ".")
	if len(parts) > 1 {
		if f.FS(FeatGoKind, FCGExtEngineVar) == "" {
			engvar := cg.desc.GetExtEngineRef(parts[0])
			f.Features.Set(FeatGoKind, FCGExtEngineVar, engvar)
		}
	}
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: cg.desc}
		a.init("eng").parse(args)
		if len(parts) > 1 {
			return a.get("eng").Dot(f.FS(FeatGoKind, FCGExtEngineVar))
		}
		return a.get("eng")
	}
}
func (cg *CodeGenerator) getGoHookFuncFeature(name string) HookFeatureFunc {
	return func(args HookArgsDescriptor) jen.Code {
		fname := args.Str
		skipEng := false
		if name != "" {
			// TODO: possibility to call function from another package
			fname = name
			if strings.HasSuffix(fname, WithoutEngSuffix) {
				fname = strings.TrimSuffix(fname, WithoutEngSuffix)
				skipEng = true
			}
		}
		return stmtFromInterfaceDef(args.Obj, "obj").Dot(fname).ParamsFunc(func(g *jen.Group) {
			g.Add(stmtFromInterfaceDef(args.Ctx, "ctx"))
			if !skipEng {
				g.Add(stmtFromInterfaceDef(args.Eng, "eng"))
			}
			for _, a := range args.Params {
				g.Add(stmtFromInterfaceDef(a.Param, nil))
			}
		})
	}
}

func (cg *CodeGenerator) getJSHookFuncFeature(name string) HookFeatureFunc {
	return func(args HookArgsDescriptor) jen.Code {
		cg.desc.Features.Set(FeatGoKind, FCGScriptingRequired, true)
		return jen.Id(EngineVar).Dot(scriptingEngineField).Dot("ProcessSingleRet").Params(
			stmtFromInterfaceDef(args.Ctx, "ctx"),
			jen.Lit(name),
			jen.Map(jen.String()).Interface().ValuesFunc(func(g *jen.Group) {
				m := jen.Dict{
					jen.Line().Lit("obj"): stmtFromInterfaceDef(args.Obj, "obj"),
				}
				for _, p := range args.Params {
					m[jen.Lit(p.Name)] = stmtFromInterfaceDef(p.Param, nil)
				}
				g.Add(jen.Dict{jen.Line().Lit("params"): jen.Map(jen.String()).Interface().Values(m)})
			}),
		)
	}
}

func (cg *CodeGenerator) getAttrIsPointerFeature(f *Field) bool {
	return cg.desc.Options().NullsHandling == NullablePointers && /* !d.Type.complex && */ !f.Type.NonNullable
}

func stmtFromInterface(val interface{}) (ret *jen.Statement, err string) {
	switch v := val.(type) {
	case string:
		ret = jen.Id(v)
	case *jen.Statement:
		ret = v
	case *jen.Group:
		ret = v.Add()
	case bool:
		return nil, ""
	default:
		err = fmt.Sprintf("undefined type for arg for FeatureFunc: %T; using default", val)
	}
	return
}

func stmtFromInterfaceDef(val interface{}, def interface{}) *jen.Statement {
	if val == nil {
		return stmtFromInterfaceDef(def, nil)
	}
	c, e := stmtFromInterface(val)
	if e == "" {
		return c
	}
	if def != nil {
		return stmtFromInterfaceDef(def, nil)
	}
	panic(fmt.Sprintf("invalid hook function args: %s", e))
}

func (a *FeatureArguments) init(vars ...string) *FeatureArguments {
	a.idx = map[string]int{}
	a.args = make([]*jen.Statement, len(vars))
	for i, v := range vars {
		a.args[i] = jen.Id(v)
		a.idx[v] = i
	}
	return a
}

func (a *FeatureArguments) parse(args []interface{}) {
	i := 0
	writeTo := &a.args
	for _, arg := range args {
		if i == len(*writeTo) {
			if writeTo == &a.rest {
				return
			}
			a.rest = make([]*jen.Statement, len(args)-len(a.args))
			writeTo = &a.rest
			i = 0
		}
		if arg == nil {
			a.args[i] = nil
		} else {
			switch v := arg.(type) {
			case string:
				(*writeTo)[i] = jen.Id(v)
				i++
			case *jen.Statement:
				(*writeTo)[i] = v
				i++
			case *jen.Group:
				(*writeTo)[i] = v.Add()
				i++
			case bool:
				a.boolArg = v
			default:
				a.desc.AddWarning(fmt.Sprintf("undefined type for arg for FeatureFunc: %T; using default", arg))
			}
		}
	}
}

func (a *FeatureArguments) get(arg string) *jen.Statement {
	return a.args[a.idx[arg]]
}
