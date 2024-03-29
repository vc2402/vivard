package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
)

const (
	engineSequenceProvider = "SeqProv"
)

const (
	sequenceGeneratorName             = "Sequence"
	SequenceFeatures      FeatureKind = "seq-id"

	sfInited          = "inited"
	SFSetCurrentValue = "set-current-value"
	// SFGenerateSequenceCall code feature returns function tah generates code for getting next value from sequence
	//  function params: sequenceName string, receiver jen.Code
	SFGenerateSequenceCall = "seq-call"
)

type SequnceIDGenerator struct {
	desc *Package
	b    *Builder
}

func init() {
	RegisterPlugin(&SequnceIDGenerator{})
}

func (cg *SequnceIDGenerator) Name() string {
	return sequenceGeneratorName
}

func (cg *SequnceIDGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	return false, nil
}
func (cg *SequnceIDGenerator) Prepare(desc *Package) error {
	cg.desc = desc

	desc.Engine.Fields.Add(jen.Id(engineSequenceProvider).Qual(VivardPackage, "SequenceProvider")).Line()

	return nil
}
func (cg *SequnceIDGenerator) Generate(bldr *Builder) (err error) {
	cg.desc = bldr.Descriptor
	cg.b = bldr
	if !cg.desc.Features.Bool(SequenceFeatures, sfInited) {
		bldr.Descriptor.Engine.Initializator.Add(
			jen.Id(EngineVar).Dot(engineSequenceProvider).Op("=").Id("v").Dot("GetService").Params(jen.Lit(vivard.ServiceSequenceProvider)).
				Assert(jen.Qual(VivardPackage, "SequenceProvider")),
		).Line()
		cg.desc.Features.Set(SequenceFeatures, sfInited, true)
	}
	for _, t := range bldr.File.Entries {
		idfld := t.GetIdField()
		if idfld != nil && idfld.HasModifier(AttrModifierIDAuto) && (idfld.Type.Type == TipInt || idfld.Type.Type == TipString) {
			err = cg.generateIDGeneratorFunc(t)
			if err != nil {
				return
			}
		}
	}
	return
}
func (cg *SequnceIDGenerator) generateIDGeneratorFunc(e *Entity) error {
	name := e.Name
	fname := cg.desc.GetMethodName(MethodGenerateID, name)
	ret, err := cg.b.addType(jen.Id("id"), e.GetIdField().Type)
	if err != nil {
		return err
	}
	idfld := e.GetIdField()

	seqName := e.Name
	tip := e
	// let's correct seqName for derived type
	//TODO: create feature for sequence name and define it earlier
	for tip.BaseTypeName != "" {
		tip = tip.GetBaseType()
		seqName = tip.Name
	}
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(
		jen.Id("ctx").Qual(
			"context",
			"Context",
		),
	).Parens(
		jen.List(ret, jen.Id("err").Error()),
	).BlockFunc(
		func(g *jen.Group) {
			if idfld.Type.Type == TipInt {
				g.List(
					jen.Id("seq"),
					jen.Id("err"),
				).Op(":=").Id(EngineVar).Dot(engineSequenceProvider).Dot("Sequence").Params(jen.Id("ctx"), jen.Lit(seqName))
				g.Add(returnIfErr())
				g.Return(jen.Id("seq").Dot("Next").Params(jen.Id("ctx")))
			} else {
				g.Return(jen.List(jen.Qual("github.com/google/uuid", "New").Params().Dot("String").Params(), jen.Nil()))
			}
		},
	).Line()

	cg.b.Functions.Add(f)
	return nil
}

// ProvideFeature from FeatureProvider interface
func (cg *SequnceIDGenerator) ProvideFeature(
	kind FeatureKind,
	name string,
	obj interface{},
) (feature interface{}, ok ProvideFeatureResult) {
	switch kind {
	case SequenceFeatures:
		switch name {
		case SFSetCurrentValue:
			if t, ok := obj.(*Entity); ok {
				idField := t.GetIdField()
				if idField.HasModifier(AttrModifierIDAuto) {
					return func(args ...interface{}) jen.Code {
						val := jen.Id("value")
						if len(args) > 0 {
							n, ok := args[0].(string)
							if ok {
								val = jen.Id(n)
							}
						}

						return jen.List(
							jen.Id("seq"),
							jen.Id("err"),
						).Op(":=").Id(EngineVar).Dot(engineSequenceProvider).Dot("Sequence").Params(
							jen.Id("ctx"),
							jen.Lit(t.Name),
						).Line().
							If(jen.Id("err").Op("==").Nil()).Block(
							jen.List(jen.Id("_"), jen.Id("err")).Op("=").Id("seq").Dot("SetCurrent").Params(jen.Id("ctx"), val),
						)
					}, FeatureProvided
				}
			}
		case SFGenerateSequenceCall:
			var fun CodeHelperFunc
			fun = func(args ...interface{}) jen.Code {
				if len(args) != 2 {
					panic(fmt.Sprintf("sequence: generate next: 2 params expected: string and jen.Code"))
				}
				name, ok := args[0].(string)
				if !ok {
					panic(fmt.Sprintf("sequence: generate next: first param should be string"))
				}
				rec, ok := args[1].(jen.Code)
				if !ok {
					panic(fmt.Sprintf("sequence: generate next: second param should be jen.Code"))
				}

				return cg.generateSequencesCall(name, rec)
			}
			return fun, FeatureProvided
		}
	}
	return nil, FeatureNotProvided
}

func (cg *SequnceIDGenerator) generateSequencesCall(seqName string, receiver jen.Code) *jen.Statement {
	return jen.List(
		jen.Id("seq"),
		jen.Id("err"),
	).Op(":=").Id(EngineVar).Dot(engineSequenceProvider).Dot("Sequence").Params(jen.Id("ctx"), jen.Lit(seqName)).Line().
		Add(returnIfErr()).Line().
		List(receiver, jen.Id("err")).Op("=").Id("seq").Dot("Next").Params(jen.Id("ctx"))
}
