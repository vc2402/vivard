package gen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type NoCacheOptions struct {
	GenerateForDictionaries bool
}
type NoCacheGenerator struct {
	Options NoCacheOptions
	desc    *Package
	b       *Builder
}

const (
	nocacheGeneratorName = "Go"
	nocacheAnnotation    = "nocache"
)

// points for CodeFragmentProvider
const (
	CFGPointEnterBeforeHooks = "enter-before-hooks"
	CFGPointEnterAfterHooks  = "enter-after-hooks"
	CFGPointMainAction       = "main-action"
	CFGPointExitBeforeHooks  = "exit-before-hooks"
	CFGPointExitAfterHooks   = "exit-after-hooks"
	CFGPointExitError        = "exit-error"
)

func init() {
	RegisterPlugin(&NoCacheGenerator{})
}

func (ncg *NoCacheGenerator) Name() string {
	return nocacheGeneratorName
}

func (ncg *NoCacheGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	ncg.desc = desc
	if ann.Name == nocacheAnnotation {
		return true, nil
	}
	return false, nil
}

func (ncg *NoCacheGenerator) Prepare(desc *Package) error {
	ncg.desc = desc
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if _, hok := t.HaveHook(TypeHookCreate); hok {
				// just to set it is required
				if t.HasModifier(TypeModifierSingleton) {
					desc.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookCreate, HookArgsDescriptor{
						Str: desc.GetHookName(TypeHookCreate, nil),
						Obj: "o",
					})
				} else {
					ncg.desc.AddWarning(fmt.Sprintf("at: %v: hook %s can be used only with singletons", t.Pos, TypeHookCreate))
				}
			}
			if _, hok := t.HaveHook(TypeHookChange); hok {
				desc.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChange, HookArgsDescriptor{
					Str:    desc.GetHookName(TypeHookChange, nil),
					Obj:    "o",
					ErrVar: "err",
				})
			}
			if _, hok := t.HaveHook(TypeHookChanged); hok {
				desc.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChanged, HookArgsDescriptor{
					Str:    desc.GetHookName(TypeHookChanged, nil),
					Obj:    "o",
					ErrVar: "err",
				})
			}
			if _, hok := t.HaveHook(TypeHookDelete); hok {
				desc.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookDelete, HookArgsDescriptor{
					Str:    desc.GetHookName(TypeHookDelete, nil),
					Obj:    "o",
					ErrVar: "err",
				})
			}
		}
	}
	return nil
}

func (ncg *NoCacheGenerator) Generate(b *Builder) (err error) {
	ncg.desc = b.Descriptor
	ncg.b = b
	for _, t := range b.File.Entries {
		name := t.Name
		f := t.GetIdField()
		if t.HasModifier(TypeModifierTransient) || t.HasModifier(TypeModifierEmbeddable) ||
			t.HasModifier(TypeModifierSingleton) || t.HasModifier(TypeModifierExternal) ||
			t.HasModifier(TypeModifierConfig) {
			continue
		}
		if f == nil {
			b.Descriptor.AddWarning(fmt.Sprintf("skipping getter and setter generation for type %s: no id field defined", name))
		} else {
			if skip, ok := t.Features.GetBool(FeaturesCommonKind, FCSkipAccessors); !ok || !skip {
				err = b.generateGetter(name, t, f.Type)
				if err != nil {
					return fmt.Errorf("while generating getter for %s: %w", name, err)
				}
				if !t.FB(FeaturesCommonKind, FCReadonly) {
					err = b.generateSetter(t)
					if err != nil {
						return fmt.Errorf("while generating setter for %s: %w", name, err)
					}
					err = b.generateNew(t)
					if err != nil {
						return fmt.Errorf("while generating setter for %s: %w", name, err)
					}
					err = b.generateDelete(t)
					if err != nil {
						return fmt.Errorf("while generating delete for %s: %w", name, err)
					}
				}
				err = ncg.generateBulk(t)
				if err != nil {
					err = fmt.Errorf("while generating bulk methods for %s: %w", name, err)
					return
				}
			}
		}
	}
	return nil
}

func (b *Builder) generateGetter(name string, ent *Entity, idType *TypeRef) error {
	cf := CodeFragmentContext{
		Builder:         b,
		MethodName:      b.Descriptor.GetMethodName(MethodGet, name),
		MethodKind:      MethodGet,
		TypeName:        ent.Name,
		EngineAvailable: true,
		Entity:          ent,
		Params: map[string]string{
			ParamContext: "ctx",
			ParamID:      "id",
		},
	}

	cf.body = jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(cf.MethodName).ParamsFunc(func(g *jen.Group) {
		params, err := b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), idType)
		if err != nil {
			b.Descriptor.AddError(err)
		} else {
			g.Add(params)
		}
	}).Parens(jen.List(jen.Id("obj").Op("*").Id(name), jen.Err().Error())).BlockFunc(func(g *jen.Group) {
		cf.Push(g)
		cf.ErrVar = "err"
		cf.ObjVar = "obj"
		cf.Enter(true)
		cf.Enter(false)

		//TODO: add error feature calls
		provided := cf.MainAction()
		if !provided {
			cf.Add(
				jen.List(cf.GetObjVar(), cf.GetErr()).Op("=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodLoad, name)).Params(jen.List(jen.Id("ctx"), jen.Id("id"))),
			)
		}

		cf.Exit(true)
		cf.Exit(false)
		cf.Add(jen.Return())
		cf.Pop()
	}).Line()

	b.Functions.Add(cf.body)
	return nil
}

func (b *Builder) generateSetter(t *Entity) error {
	name := t.Name
	cf := CodeFragmentContext{
		Builder:         b,
		MethodName:      b.Descriptor.GetMethodName(MethodSet, name),
		MethodKind:      MethodSet,
		TypeName:        name,
		EngineAvailable: true,
		Entity:          t,
		Params: map[string]string{
			ParamContext: "ctx",
			ParamObject:  "o",
		},
	}
	idFld := t.GetIdField()
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(cf.MethodName).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)).
		Parens(jen.List(jen.Id("obj").Op("*").Id(name), jen.Err().Error())).BlockFunc(func(g *jen.Group) {
		cf.Push(g)
		cf.ErrVar = "err"
		cf.ObjVar = "obj"
		g.List(jen.Id("obj"), jen.Id("err")).Op("=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodGet, name)).Params(jen.List(jen.Id("ctx"), jen.Id("o").Dot(idFld.Name)))
		cf.AddCheckError()
		g.If(jen.Id("obj").Op("==").Nil()).BlockFunc(func(g *jen.Group) {
			cf.Push(g)
			g.Id("err").Op("=").Qual("errors", "New").Params(jen.Lit("not found"))
			cf.AddOnErrorReturnStatement()
			cf.Pop()
		})
		cf.Enter(true)
		if _, hok := t.HaveHook(TypeHookChange); hok {
			g.Add(b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChange, HookArgsDescriptor{
				Str: b.Descriptor.GetHookName(TypeHookChange, nil),
				Obj: "obj",
				Params: []HookArgParam{
					//{"oldValue", jen.Id("o")},
					{"newValue", jen.Id("o")},
				},
				ErrVar: "err",
			}))
			cf.AddCheckError()
			//g.Add(returnIfErrValue(jen.Nil()))
		}
		cf.Enter(false)

		provided := cf.MainAction()
		if !provided {
			cf.Add(jen.List(cf.GetObjVar(), cf.GetErr()).Op("=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodSave, name)).Params(jen.List(jen.Id("ctx"), jen.Id("o"))))
			cf.AddCheckError()
			//cf.Add(returnIfErrValue(jen.Nil()))
		}
		cf.Exit(true)
		if _, hok := t.HaveHook(TypeHookChanged); hok {
			g.Add(b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChanged, HookArgsDescriptor{
				Str: b.Descriptor.GetHookName(TypeHookChanged, nil),
				Obj: "obj",
				Params: []HookArgParam{
					{"newValue", jen.Id("o")},
				},
				ErrVar: "err",
			}))
			cf.AddCheckError()
		}
		cf.Exit(false)
		g.Return()
		cf.Pop()
	}).Line()

	b.Functions.Add(f)
	return nil
}

func (b *Builder) generateNew(t *Entity) error {
	name := t.Name
	idField := t.GetIdField()
	cf := CodeFragmentContext{
		Builder:         b,
		MethodName:      b.Descriptor.GetMethodName(MethodNew, name),
		MethodKind:      MethodNew,
		TypeName:        name,
		EngineAvailable: true,
		Entity:          t,
		Params: map[string]string{
			ParamContext: "ctx",
			ParamObject:  "o",
		},
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(cf.MethodName).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)).
		Parens(jen.List(jen.Id("ret").Op("*").Id(name), jen.Err().Error())).BlockFunc(func(c *jen.Group) {
		if idField.HasModifier(AttrModifierIDAuto) {
			c.Add(
				jen.If(jen.Id("o").Dot(idField.Name).Op("==").Add(b.goEmptyValue(idField.Type))).Block(
					jen.List(jen.Id("o").Dot(idField.Name), jen.Id("err")).Op("=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodGenerateID, name)).Params(jen.Id("ctx")).Line(),
					returnIfErr().Line(),
					// jen.Id("o").Dot(idField.Name).Op("=").Id("id").Line(),
				),
			)
		}
		if t.BaseTypeName != "" || t.HasModifier(TypeModifierExtendable) {
			tn := t.FS(FeatGoKind, FCGDerivedTypeNameConst)
			c.If(jen.Id("o").Dot(ExtendableTypeDescriptorFieldName).Op("==").Lit("")).Block(
				jen.Id("o").Dot(ExtendableTypeDescriptorFieldName).Op("=").String().Parens(jen.Id(tn)),
			)
		}
		cf.Push(c)
		cf.ErrVar = "err"
		cf.ObjVar = "ret"
		cf.Enter(true)
		if _, hok := t.HaveHook(TypeHookChange); hok {
			cf.Add(
				b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChange, HookArgsDescriptor{
					Str: b.Descriptor.GetHookName(TypeHookChange, nil),
					Obj: jen.Id("ret"),
					Params: []HookArgParam{
						{"newValue", jen.Id("o")},
					},
					ErrVar: "err",
				}))
			cf.AddCheckError()
		}
		cf.Enter(false)

		provided := cf.MainAction()
		if !provided {
			cf.Add(jen.List(cf.GetObjVar(), cf.GetErr()).Op("=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodCreate, name)).Params(jen.List(jen.Id("ctx"), jen.Id("o"))))
			cf.AddCheckError()
		}
		cf.Exit(true)
		if _, hok := t.HaveHook(TypeHookChanged); hok {
			cf.Add(b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChanged, HookArgsDescriptor{
				Str: b.Descriptor.GetHookName(TypeHookChanged, nil),
				Obj: jen.Id("ret"),
				Params: []HookArgParam{
					{"newValue", jen.Id("o")},
				},
				ErrVar: "err",
			}))
			cf.AddCheckError()
		}
		cf.Exit(false)
		c.Return()
		cf.Pop()
	}).Line()

	b.Functions.Add(f)
	return nil
}

func (ncg *NoCacheGenerator) generateBulk(t *Entity) error {
	if !t.FB(FeaturesCommonKind, FCReadonly) {
		if t.FB(FeatGoKind, FCGBulkNew) {
			name := t.Name
			methodName := ncg.b.Descriptor.GetMethodName(MethodNewBulk, name)
			newMethodName := ncg.b.Descriptor.GetMethodName(MethodNew, name)
			f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(methodName).
				Params(
					jen.Id("ctx").Qual("context", "Context"),
					jen.Id("objs").Index().Op("*").Id(name),
				).Parens(jen.List(jen.Id("ret").Index().Op("*").Id(name), jen.Err().Error())).Block(
				jen.For(jen.List(jen.Id("idx"), jen.Id("o")).Op(":=").Range().Id("objs")).Block(
					jen.List(jen.Id("objs").Index(jen.Id("idx")), jen.Err()).Op("=").Id(EngineVar).Dot(newMethodName).Params(
						jen.Id("ctx"),
						jen.Id("o"),
					),
					jen.Add(returnIfErrValue(jen.Id("objs"))),
				),
				jen.Return(jen.List(jen.Id("objs"), jen.Nil())),
			).Line()
			ncg.b.Functions.Add(f)
		}
	}
	return nil
}

func (b *Builder) generateDelete(t *Entity) error {
	name := t.Name
	idType := t.GetIdField().Type
	cf := CodeFragmentContext{
		Builder:         b,
		MethodName:      b.Descriptor.GetMethodName(MethodDelete, name),
		MethodKind:      MethodDelete,
		TypeName:        name,
		EngineAvailable: true,
		Entity:          t,
		Params: map[string]string{
			ParamContext: "ctx",
			ParamID:      "id",
		},
	}

	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(cf.MethodName).ParamsFunc(func(g *jen.Group) {
		params, err := b.addType(jen.List(cf.GetParam(ParamContext).Qual("context", "Context"), cf.GetParam(ParamID)), idType)
		if err != nil {
			b.Descriptor.AddError(err)
		} else {
			g.Add(params)
		}
	}).Parens(jen.List(jen.Err().Error())).BlockFunc(func(g *jen.Group) {
		cf.Push(g)
		cf.ErrVar = "err"
		cf.ObjVar = "o"
		g.List(jen.Id("o"), jen.Err()).Op(":=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodGet, name)).Params(jen.List(jen.Id("ctx"), jen.Id("id")))
		cf.AddCheckError()
		g.If(jen.Id("o").Op("==").Nil()).BlockFunc(func(g *jen.Group) {
			cf.Push(g)
			g.Err().Op("=").Qual("errors", "New").Params(jen.Lit("not found"))
			cf.AddOnErrorReturnStatement()
			cf.Pop()
		})
		cf.Enter(true)
		if _, hok := t.HaveHook(TypeHookDelete); hok {
			g. /*Id("err").Op(":=").*/ Add(b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookDelete, HookArgsDescriptor{
				Str:    b.Descriptor.GetHookName(TypeHookDelete, nil),
				Obj:    "o",
				ErrVar: "err",
			}))
			cf.AddCheckError()
			//g.Add(returnIfErrValue(jen.Nil()))
		}
		if _, hok := t.HaveHook(TypeHookChange); hok {
			cf.Add(
				b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChange, HookArgsDescriptor{
					Str: b.Descriptor.GetHookName(TypeHookChange, nil),
					Obj: "o",
					Params: []HookArgParam{
						//{"oldValue", jen.Id("o")},
						{"newValue", jen.Nil()},
					},
				}),
			)
		}
		cf.Enter(false)

		provided := cf.MainAction()
		if !provided {
			cf.Add(cf.GetErr().Op("=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodRemove, name)).Params(jen.List(jen.Id("ctx"), jen.Id("id"))))
			cf.AddCheckError()
		}
		cf.Exit(true)
		if _, hok := t.HaveHook(TypeHookChanged); hok {
			cf.Add(
				b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChanged, HookArgsDescriptor{
					Str: b.Descriptor.GetHookName(TypeHookChanged, nil),
					Obj: "o",
					Params: []HookArgParam{
						{"newValue", jen.Nil()},
					},
				}),
			)
		}

		cf.Exit(false)
		g.Return()
		cf.Pop()
	},
	).Line()

	b.Functions.Add(f)
	return nil
}

// ProvideFeature from FeatureProvider interface
func (ncg *NoCacheGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	switch kind {
	case FeaturesCommonKind:
		switch name {
		case FCGetterCode:
			if t, ok := obj.(*Entity); ok && !t.HasModifier(TypeModifierSingleton) && !t.HasModifier(TypeModifierEmbeddable) && !t.HasModifier(TypeModifierTransient) {
				return ncg.getObjectGetFuncFeature(t), FeatureProvided
			}
		}
	}
	return
}

func (ncg *NoCacheGenerator) getObjectGetFuncFeature(e *Entity) CodeHelperFunc {
	return func(args ...interface{}) jen.Code {
		a := &FeatureArguments{desc: ncg.desc, boolArg: true}
		fname := ncg.desc.GetMethodName(MethodGet, e.Name)
		a.init("id", "ctx", "eng").parse(args)
		if !a.boolArg {
			ncg.desc.AddWarning("getterFeatureFunc supports only two return values variant")
		}
		return a.get("eng").Dot(fname).Params(a.get("ctx"), a.get("id"))
	}
}
