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
}

const (
	nocacheAnnotation = "nocache"
)

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

				desc.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookCreate, HookArgsDescriptor{
					Str: desc.GetHookName(TypeHookCreate, nil),
					Obj: "o",
				})
			}
			if _, hok := t.HaveHook(TypeHookChange); hok {
				desc.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChange, HookArgsDescriptor{
					Str: desc.GetHookName(TypeHookChange, nil),
					Obj: "o",
				})
			}
		}
	}
	return nil
}

func (ncg *NoCacheGenerator) Generate(b *Builder) (err error) {
	ncg.desc = b.Descriptor
	for _, t := range b.File.Entries {
		name := t.Name
		f := t.GetIdField()
		if t.IsDictionary() {
			if !ncg.Options.GenerateForDictionaries {
				continue
			}
		}
		if t.HasModifier(TypeModifierTransient) || t.HasModifier(TypeModifierEmbeddable) ||
			t.HasModifier(TypeModifierSingleton) || t.HasModifier(TypeModifierExternal) ||
			t.HasModifier(TypeModifierConfig) {
			continue
		}
		if f == nil {
			b.Descriptor.AddWarning(fmt.Sprintf("skipping getter and setter generation for type %s: no id field defined", name))
		} else {
			if skip, ok := t.Features.GetBool(FeaturesCommonKind, FCSkipAccessors); !ok || !skip {
				err = b.generateGetter(name, f.Type)
				if err != nil {
					return fmt.Errorf("while generating getter for %s: %w", name, err)
				}
				err = b.generateSetter(t)
				if err != nil {
					return fmt.Errorf("while generating setter for %s: %w", name, err)
				}
				err = b.generateNew(t)
				if err != nil {
					return fmt.Errorf("while generating setter for %s: %w", name, err)
				}
				err = b.generateDelete(name, f.Type)
				if err != nil {
					return fmt.Errorf("while generating delete for %s: %w", name, err)
				}
			}
		}
	}
	return nil
}

func (b *Builder) generateGetter(name string, idType *TypeRef) error {
	fname := b.Descriptor.GetMethodName(MethodGet, name)

	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).ParamsFunc(func(g *jen.Group) {
		params, err := b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), idType)
		if err != nil {
			b.Descriptor.AddError(err)
		} else {
			g.Add(params)
		}
	}).Parens(jen.List(jen.Op("*").Id(name), jen.Error())).Block(
		jen.Return(
			jen.Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodLoad, name)).Params(jen.List(jen.Id("ctx"), jen.Id("id"))),
		),
	).Line()

	b.Functions.Add(f)
	return nil
}

func (b *Builder) generateSetter(t *Entity) error {
	name := t.Name
	fname := b.Descriptor.GetMethodName(MethodSet, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)).
		Parens(jen.List(jen.Op("*").Id(name), jen.Error())).BlockFunc(func(g *jen.Group) {
		if _, hok := t.HaveHook(TypeHookChange); hok {
			g.Add(b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookChange, HookArgsDescriptor{
				Str: b.Descriptor.GetHookName(TypeHookChange, nil),
				Obj: "o",
			}))
		}
		g.Return(
			jen.Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodSave, name)).Params(jen.List(jen.Id("ctx"), jen.Id("o"))),
		)
	}).Line()

	b.Functions.Add(f)
	return nil
}

func (b *Builder) generateNew(t *Entity) error {
	name := t.Name
	idField := t.GetIdField()
	fname := b.Descriptor.GetMethodName(MethodNew, name)
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context"), jen.Id("o").Op("*").Id(name)).
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

		c.List(jen.Id("ret"), jen.Id("err")).Op("=").Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodCreate, name)).Params(jen.List(jen.Id("ctx"), jen.Id("o")))
		if _, hok := t.HaveHook(TypeHookCreate); hok {
			c.If(jen.Err().Op("==").Nil()).Block(
				b.Descriptor.CallFeatureHookFunc(t, FeaturesHookCodeKind, TypeHookCreate, HookArgsDescriptor{
					Str: b.Descriptor.GetHookName(TypeHookCreate, nil),
					Obj: "o",
					// Args: map[string]interface{}{},
				}),
			)
		}
		c.Return()
	}).Line()

	b.Functions.Add(f)
	return nil
}

func (b *Builder) generateDelete(name string, idType *TypeRef) error {
	fname := b.Descriptor.GetMethodName(MethodDelete, name)

	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).ParamsFunc(func(g *jen.Group) {
		params, err := b.addType(jen.List(jen.Id("ctx").Qual("context", "Context"), jen.Id("id")), idType)
		if err != nil {
			b.Descriptor.AddError(err)
		} else {
			g.Add(params)
		}
	}).Parens(jen.List(jen.Error())).Block(
		jen.Return(
			jen.Id(EngineVar).Dot(b.Descriptor.GetMethodName(MethodRemove, name)).Params(jen.List(jen.Id("ctx"), jen.Id("id"))),
		),
	).Line()

	b.Functions.Add(f)
	return nil
}

//ProvideFeature from FeatureProvider interface
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
