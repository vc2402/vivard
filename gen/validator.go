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
	ValidateDictionaries bool `json:"validate_dictionaries"`
}

const (
	ValidatorGeneratorName = "Validator"
	FeaturesValidator      = "validator"

	FVValidateFunc       = "validate-func"
	FVValidationRequired = "required"
)

const ValidatorFuncNameTemplate = "Validate%s"

const ValidatorOptionsName = "validatorGO"

func init() {
	RegisterPlugin(&Validator{options: ValidatorOptions{ValidateDictionaries: true}})
}

func (cg *Validator) Name() string {
	return ValidatorGeneratorName
}

func (cg *Validator) SetOptions(opts any) error {
	return OptionsAnyToStruct(opts, &cg.options)
}

func (cg *Validator) SetDescriptor(proj *Project) {
	cg.proj = proj
	//cg.options = ValidatorOptions{ValidateDictionaries: true}
	cg.proj.Options.CustomToStruct(ValidatorOptionsName, &cg.options)
}

func (cg *Validator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	return false, nil
}

func (cg *Validator) Prepare(desc *Package) error {
	cg.desc = desc
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if t.HasModifier(TypeModifierTransient) {
				continue
			}
			if t.Annotations[AnnotationConfig] != nil {
				continue
			}
			for _, f := range t.Fields {
				var refType *DefinedType
				var ok bool
				if f.Type.Array != nil {
					refType, ok = desc.FindType(f.Type.Array.Type)
				}
				if f.Type.Complex {
					refType, ok = desc.FindType(f.Type.Type)
				}
				//TODO add enums validation
				if ok && refType.Entity() != nil && refType.Entity().HasModifier(TypeModifierDictionary) {
					f.Features.Set(FeaturesValidator, FVValidationRequired, true)
					t.Features.Set(FeaturesValidator, FVValidationRequired, true)
					t.Features.Set(FeaturesValidator, FVValidateFunc, fmt.Sprintf(ValidatorFuncNameTemplate, t.Name))
				}
			}
		}
	}
	return nil
}

func (cg *Validator) Generate(b *Builder) (err error) {
	cg.b = b
	for _, e := range b.File.Entries {
		if e.FB(FeaturesValidator, FVValidationRequired) {
			err = cg.generateValidator(e)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cg *Validator) generateValidator(e *Entity) error {
	fname := e.FS(FeaturesValidator, FVValidateFunc)

	fun := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("obj").Op("*").Id(e.Name),
	).Error().BlockFunc(func(g *jen.Group) {
		//TODO call base class validator if any
		for _, f := range e.Fields {
			if f.FB(FeaturesValidator, FVValidationRequired) {
				engVar := cg.desc.CallCodeFeatureFunc(f, FeaturesCommonKind, FCEngineVar)
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
								Params(jen.Lit(fmt.Sprintf("validate: %s: invalid value: %%v", f.Name)), id),
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
	}).Line()

	cg.b.Functions.Add(fun)
	return nil
}

func (cg *Validator) ProvideCodeFragment(module interface{}, action interface{}, point interface{}, ctx interface{}) interface{} {
	if module == CodeFragmentModuleGeneral {
		if cf, ok := ctx.(*CodeFragmentContext); ok {
			if point == CFGPointEnterAfterHooks && cf.Entity != nil {
				idField := cf.Entity.GetIdField()
				if action == MethodNew &&
					!idField.HasModifier(AttrModifierIDAuto) {
					//do not check enums
					if df, ok := cf.Entity.Pckg.FindType(idField.Type.Type); !ok || df.Enum() == nil {
						cf.Add(
							jen.If(
								cf.Builder.checkIfEmptyValue(jen.Id("o").Dot(idField.Name), idField.Type, false),
							).Block(
								jen.Add(cf.GetErr()).Op("=").Qual("errors", "New").Params(jen.Lit(fmt.Sprintf("validate: %s: empty", idField.Name))),
								jen.Return(),
							),
						)
					}
				}
				if cf.Entity.FB(FeaturesValidator, FVValidationRequired) {
					if action == MethodSet || action == MethodNew {
						fname := cf.Entity.FS(FeaturesValidator, FVValidateFunc)
						cf.Add(
							cf.GetErr().Op("=").Id(EngineVar).Dot(fname).Params(cf.GetParam(ParamContext), cf.GetParam(ParamObject)),
						)
						cf.AddCheckError()
						return true
					}
				}
			}
		}
	}
	return nil
}
