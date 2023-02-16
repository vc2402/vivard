package js

import (
	"fmt"
	"github.com/vc2402/vivard/gen"
	"text/template"
)

const (
	FeaturesValidator = "ts-validator"
	FVGenerate        = "generate"
	FVValidatorClass  = "validator"
)

type TSValidatorGenerator struct {
	desc *gen.Package
}

func (cg *TSValidatorGenerator) CheckAnnotation(desc *gen.Package, ann *gen.Annotation, item interface{}) (bool, error) {
	if ann.Name == Annotation {
		return true, nil
	}
	return false, nil
}

func (cg *TSValidatorGenerator) Prepare(desc *gen.Package) error {
	cg.desc = desc
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if t.FB(gen.FeaturesValidator, gen.FVValidationRequired) {
				t.Features.Set(FeaturesValidator, FVGenerate, true)
			}
			className := fmt.Sprintf("%sValidator", t.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, ""))
			t.Features.Set(FeaturesValidator, FVValidatorClass, className)
		}
	}
	return nil
}

func (cg *TSValidatorGenerator) Generate(b *gen.Builder) (err error) {
	return nil
}

func (cg *TSValidatorGenerator) ProvideCodeFragment(module interface{}, action interface{}, point interface{}, ctx interface{}) interface{} {
	if module == CodeFragmentModule {
		if cfc, ok := ctx.(CodeFragmentContext); ok {
			provided := false
			switch action {
			case CodeFragmentActionFile:
				for _, e := range cfc.File.Entries {
					if e.FB(FeaturesValidator, FVGenerate) ||
						e.FB(gen.FeaturesValidator, gen.FVValidationRequired) {
						cfc.Error = cg.generateValidator(e, cfc)
						provided = true
					}
				}
			case CodeFragmentActionImport:
				for _, e := range cfc.File.Entries {
					if e.FB(FeaturesValidator, FVGenerate) ||
						e.FB(gen.FeaturesValidator, gen.FVValidationRequired) {
						cfc.addImport("./vivard", "ValidatorBase")
						provided = true
					}
				}
			}
			if provided {
				return true
			}
		}
	}
	return nil
}

func (cg *TSValidatorGenerator) generateValidator(e *gen.Entity, cfc CodeFragmentContext) (err error) {
	funcs := template.FuncMap{
		"ClassName": func() string {
			return e.FS(FeaturesValidator, FVValidatorClass)
		},
		"GetFields": func(e *gen.Entity) []*gen.Field {
			allFields := e.GetFields(true, true)
			fields := make([]*gen.Field, len(allFields))
			i := 0
			for _, f := range allFields {
				if _, ok := f.Annotations.GetStringAnnotation(Annotation, AnnotationName); ok &&
					f.FB(gen.FeaturesValidator, gen.FVValidationRequired) {
					fields[i] = f
					i++
				}
			}
			return fields[:i]
		},
		"FieldName": func(f *gen.Field) string {
			return f.Annotations.GetStringAnnotationDef(Annotation, AnnotationName, "")
		},
	}
	tip := template.New("VALIDATOR").
		Funcs(funcs)
	tip, err = tip.Parse(validatorClassTemplate)
	if err != nil {
		return err
	}
	err = tip.Execute(cfc.Output, e)

	return err
}

var validatorClassTemplate = `
export class {{ClassName}} extends ValidatorBase {
  constructor() {
    super({ {{range GetFields .}}
      {{FieldName .}}:[],{{end}}
    });
  }

  {{range GetFields .}}{{FieldName .}}Rules(v: any): string[] {
    return [];
  }
  {{end}}

}
`
