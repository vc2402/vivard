package js

import (
	"github.com/vc2402/vivard/gen"
	"io"
	"text/template"
)

func (cg *GQLCLientGenerator) generateEnum(wr io.Writer, e *gen.Enum) (err error) {
	tip := template.New("ENUM").
		Funcs(cg.getFuncsMap())
	tip, err = tip.Parse(enumTemplate)
	if err != nil {
		return err
	}
	err = tip.Execute(wr, e)
	if err != nil {
		return err
	}
	return nil
}

const enumTemplate = `
export type {{EnumName .}} = {{EnumType .}};
export type {{EnumInputName .}} = {{EnumType .}};
{{range .Fields}}
export const {{EnumFieldName .}} = {{EnumFieldValue .}}{{end}}
`
