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
	loggerGeneratorName             = "Logger"
	loggerAttr                      = "log"
	LogFeatureKind      FeatureKind = "logger-feature"
	lfInited                        = "inited"
	// LFWarn returns code for warn (params: message and pairs key/value)
	LFWarn = "warn"
	// LFDebug returns code for debug (params: message and pairs key/value)
	LFDebug = "debug"
	// LFError returns code for error (params: message and pairs key/value)
	LFError = "error"

	OptionLoggerGenerator = "logger-generator"
	OptionsVariant        = "variant"
)

func init() {
	RegisterPlugin(&LoggerGenerator{variant: "zap"})
}

func (cg *LoggerGenerator) Name() string {
	return loggerGeneratorName
}

func (cg *LoggerGenerator) SetOptions(options any) error {
	if o, ok := options.(map[string]interface{}); ok {
		if variant, ok := o[OptionsVariant].(string); ok {
			cg.variant = variant
		}
	} else {
		cg.variant = "zap"
	}
	return nil
}

// SetDescriptor from DescriptorAware
func (cg *LoggerGenerator) SetDescriptor(proj *Project) {
	cg.proj = proj
	if opt, ok := proj.Options.Custom[OptionLoggerGenerator]; ok {
		_ = cg.SetOptions(opt)
	} else if cg.variant == "" {
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
	if !cg.desc.Features.Bool(LogFeatureKind, lfInited) {
		cg.desc.Engine.Initializator.Add(jen.Id(EngineVar).Dot(loggerAttr).Op("=").Id("v").Dot("GetService").Params(jen.Lit(ld.service)).Assert(jen.Op("*").Qual(VivardPackage, ld.tip)).Dot("Log").Params().Line())
		cg.desc.Features.Set(LogFeatureKind, lfInited, true)
	}
	return nil
}

type loggerField struct {
	name string
	val  jen.Code
}

func (cg *LoggerGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	switch kind {
	case LogFeatureKind:
		switch name {
		case LFDebug, LFWarn, LFError:
			if cg.variant == "zap" {
				var fun CodeHelperFunc
				fun = func(args ...interface{}) jen.Code {
					if len(args) < 1 {
						panic(fmt.Sprintf("logger feature: %s: at least one param expected", name))
					}
					msg, ok := args[0].(jen.Code)
					if !ok {
						if m, ok := args[0].(string); ok {
							msg = jen.Lit(m)
						} else {
							panic(fmt.Sprintf("logger feature: %s: first param should be jen.Code or string", name))
						}
					}
					var fields []loggerField
					for i := 1; i+1 < len(args); i += 2 {
						fn, ok := args[i].(string)
						if !ok {
							panic(fmt.Sprintf("logger feature: %s: first param of field should be string", name))
						}
						fv, ok := args[i+1].(jen.Code)
						if !ok {
							panic(fmt.Sprintf("logger feature: %s: second param of field should be jen.Code", name))
						}
						fields = append(fields, loggerField{fn, fv})
					}

					return cg.generateLoggerCall(name, msg, fields)
				}
				return fun, FeatureProvided
			}
			return nil, FeatureNotProvided
		}
	}
	return
}

func (cg *LoggerGenerator) generateLoggerCall(kind string, msg jen.Code, fields []loggerField) *jen.Statement {
	var funcName string
	switch kind {
	case LFDebug:
		funcName = "Debug"
	case LFWarn:
		funcName = "Warn"
	case LFError:
		funcName = "Error"
	}
	//zap only so far
	return jen.Id(EngineVar).Dot(loggerAttr).Dot(funcName).ParamsFunc(func(g *jen.Group) {
		g.Add(msg)
		for _, field := range fields {
			g.Qual(variants[cg.variant].pkg, "Any").Params(
				jen.Lit(field.name),
				field.val,
			)
		}
	})
}
