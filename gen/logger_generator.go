package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
)

type LoggerGenerator struct {
	proj    *Project
	desc    *Package
	b       *Builder
	variant string
}

type loggerDescriptor struct {
	pkg     string
	logger  string
	service string
	tip     string
}

var variants = map[string]loggerDescriptor{
	"zap": {
		pkg:     "go.uber.org/zap",
		logger:  "Logger",
		service: vivard.ServiceLoggingZap,
		tip:     "LoggerService",
	},
	"logrus": {
		pkg:     "github.com/sirupsen/logrus",
		logger:  "Entry",
		service: vivard.ServiceLoggingLogrus,
		tip:     "LogrusService",
	},
}

const (
	loggerAttr                 = "log"
	logFeatureKind FeatureKind = "logger-feature"
	lfInited                   = "inited"

	OptionLoggerGenerator = "logger-generator"
	OptionsVariant        = "variant"
)

// SetDescriptor from DescriptorAware
func (cg *LoggerGenerator) SetDescriptor(proj *Project) {
	cg.proj = proj
	if opt, ok := proj.Options.Custom[OptionLoggerGenerator].(map[string]interface{}); ok {
		if variant, ok := opt[OptionsVariant].(string); ok {
			cg.variant = variant
		}
	} else {
		cg.variant = "zap"
	}
}
func (cg *LoggerGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	return false, nil
}

func (cg *LoggerGenerator) Prepare(desc *Package) error {
	ld, ok := variants[cg.variant]
	if !ok {
		return fmt.Errorf("undefined logger variant: %s", cg.variant)
	}
	desc.Engine.Fields.Add(jen.Id(loggerAttr).Op("*").Qual(ld.pkg, ld.logger)).Line()
	return nil
}

func (cg *LoggerGenerator) Generate(b *Builder) (err error) {
	cg.desc = b.Descriptor
	ld, _ := variants[cg.variant]
	if !cg.desc.Features.Bool(logFeatureKind, lfInited) {
		cg.desc.Engine.Initializator.Add(jen.Id(EngineVar).Dot(loggerAttr).Op("=").Id("v").Dot("GetService").Params(jen.Lit(ld.service)).Assert(jen.Op("*").Qual(vivardPackage, ld.tip)).Dot("Log").Params().Line())
		cg.desc.Features.Set(logFeatureKind, lfInited, true)
	}
	return nil
}

func (cg *LoggerGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	return
}
