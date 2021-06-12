package gen

import (
	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
)

type LoggerGenerator struct {
	proj *Project
	desc *Package
	b    *Builder
}

const (
	loggerAttr    = "log"
	logrusPackage = "github.com/sirupsen/logrus"

	logFeatureKind FeatureKind = "logger-feature"
	lfInited                   = "inited"
)

//SetDescriptor from DescriptorAware
func (cg *LoggerGenerator) SetDescriptor(proj *Project) {
	cg.proj = proj
}
func (cg *LoggerGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	return false, nil
}

func (lg *LoggerGenerator) Prepare(desc *Package) error {
	desc.Engine.Fields.Add(jen.Id(loggerAttr).Op("*").Qual(logrusPackage, "Entry")).Line()
	return nil
}

func (cg *LoggerGenerator) Generate(b *Builder) (err error) {
	cg.desc = b.Descriptor
	if !cg.desc.Features.Bool(logFeatureKind, lfInited) {
		cg.desc.Engine.Initializator.Add(jen.Id(EngineVar).Dot(loggerAttr).Op("=").Id("v").Dot("GetService").Params(jen.Lit(vivard.ServiceLoggingLogrus)).Assert(jen.Op("*").Qual(vivardPackage, "LogrusService")).Dot("Log").Params().Line())
		cg.desc.Features.Set(logFeatureKind, lfInited, true)
	}
	return nil
}

func (cg *LoggerGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	return
}
