package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
)

type Validator struct {
	proj    *Project
	desc    *Package
	b       *Builder
	options ValidatorOptions
}

type ValidatorOptions struct {
	// ValidateDictionaries defines behaviour of checking existing dictionary's value on instance create/set
	ValidateDictionaries bool
}

const (
	FeaturesValidator = "validator"

	fvValidateFunc       = "validate-func"
	fvValidationRequired = "required"
)

const ValidatorFuncNameTemplate = "Validate%s"

const ValidatorOptionsName = "validatorGO"

func (cg *Validator) SetDescriptor(proj *Project) {
	cg.proj = proj
	cg.options = ValidatorOptions{ValidateDictionaries: true}
	cg.proj.Options.CustomToStruct(ValidatorOptionsName, &cg.options)
}

func (cg *Validator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	return false, nil
}

func (cg *Validator) Prepare(desc *Package) error {
	cg.desc = desc
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			for _, f := range t.Fields {
				var refType *DefinedType
				var ok bool
				if f.Type.Array != nil {
					refType, ok = desc.FindType(f.Type.Array.Type)
				}
				if f.Type.Complex {
					refType, ok = desc.FindType(f.Type.Type)
				}
				if ok && refType.Entity().HasModifier(TypeModifierDictionary) {
					f.Features.Set(FeaturesValidator, fvValidationRequired, true)
					t.Features.Set(FeaturesValidator, fvValidationRequired, true)
					t.Features.Set(FeaturesValidator, fvValidateFunc, fmt.Sprintf(ValidatorFuncNameTemplate, t.Name))
				}
			}
		}
	}
	return nil
}

func (cg *Validator) Generate(b *Builder) (err error) {
	cg.b = b
	for _, file := range b.Files {
		for _, e := range file.Entries {
			if e.FB(FeaturesValidator, fvValidationRequired) {
				err = cg.generateValidator(e)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (cg *Validator) generateValidator(e *Entity) error {
	fname := e.FS(FeaturesValidator, fvValidateFunc)

	fun := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("obj").Op("*").Id(e.Name),
	).Error().BlockFunc(func(g *jen.Group) {
		//TODO call base class validator if any
		for _, f := range e.Fields {
			var typeName string
			if f.Type.Array != nil {
				typeName = f.Type.Array.Type
			} else {
				typeName = f.Type.Type
			}
			if f.FB(FeaturesValidator, fvValidationRequired) {
				engVar := cg.desc.CallFeatureFunc(f, FeaturesCommonKind, FCEngineVar)
				valueChecker := func(id jen.Code) *jen.Statement {
					return jen.If(jen.List(jen.Id("v"), jen.Id("_")).Op(":=").Add(engVar).
						Dot(
							cg.desc.GetMethodName(MethodGet, f.Type.Type),
						).
						Call(
							jen.Id("ctx"),
							id,
						),
						jen.Id("v").Op("==").Nil(),
					).Block(
						jen.Return(
							jen.Qual("fmt", "Errorf").
								Params(jen.Lit(fmt.Sprintf("invalid value for type %s: %%v", typeName)), id),
						),
					)
				}
				if f.Type.NonNullable {
					g.Add(valueChecker(jen.Id("obj").Dot(f.Name)))
				} else {
					g.If(jen.Id("obj").Dot(f.Name).Op("!=").Nil()).Block(
						valueChecker(jen.Op("*").Id("obj").Dot(f.Name)),
					)
				}
			}
		}
		g.Return(jen.Nil())
	})
	cg.b.Functions.Add(fun)
	return nil
}

func (cg *Validator) ProvideCodeFragment(module interface{}, action interface{}, point interface{}, ctx interface{}) interface{} {
	return nil
}
