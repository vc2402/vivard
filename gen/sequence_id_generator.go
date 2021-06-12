package gen

import (
	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
)

const (
	engineSequenceProvider = "SeqProv"
)

const (
	sequenceFeatures FeatureKind = "seq-id"

	sfInited = "inited"
)

type SequnceIDGenerator struct {
	desc *Package
	b    *Builder
}

func (cg *SequnceIDGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	return false, nil
}
func (cg *SequnceIDGenerator) Prepare(desc *Package) error {
	cg.desc = desc

	desc.Engine.Fields.Add(jen.Id(engineSequenceProvider).Qual(vivardPackage, "SequenceProvider")).Line()

	return nil
}
func (cg *SequnceIDGenerator) Generate(bldr *Builder) (err error) {
	cg.desc = bldr.Descriptor
	cg.b = bldr
	if !cg.desc.Features.Bool(sequenceFeatures, sfInited) {
		bldr.Descriptor.Engine.Initializator.Add(jen.Id(EngineVar).Dot(engineSequenceProvider).Op("=").Id("v").Dot("GetService").Params(jen.Lit(vivard.ServiceSequenceProvider)).
			Assert(jen.Qual(vivardPackage, "SequenceProvider"))).Line()
		cg.desc.Features.Set(sequenceFeatures, sfInited, true)
	}
	for _, t := range bldr.File.Entries {
		idfld := t.GetIdField()
		if idfld != nil && idfld.HasModifier(AttrModifierIDAuto) && idfld.Type.Type == TipInt {
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
	f := jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(fname).Params(jen.Id("ctx").Qual("context", "Context")).Parens(
		jen.List(ret, jen.Id("err").Error()),
	).Block(
		jen.List(jen.Id("seq"), jen.Id("err")).Op(":=").Id(EngineVar).Dot(engineSequenceProvider).Dot("Sequence").Params(jen.Id("ctx"), jen.Lit(e.Name)),
		returnIfErr(),
		jen.Return(jen.Id("seq").Dot("Next").Params(jen.Id("ctx"))),
	).Line()
	cg.b.Functions.Add(f)
	return nil
}
