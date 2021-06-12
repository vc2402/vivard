package gen

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

const (
	cronGeneratorAnnotation = "cron"

	cronSingletonDefaultFunctionName = "AtTime"
	cronSingletonFunctionPrefix      = "->"
)

//CroneGenerator generates Go code for @time hook
type CroneGenerator struct {
	proj *Project
	desc *Package
	b    *Builder
}

//CheckAnnotation checks that annotation may be utilized by CodeGeneration
func (cg *CroneGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	cg.desc = desc
	if ann.Name == cronGeneratorAnnotation {
		return true, nil
	}
	return false, nil
}

//Prepare from Generator interface
func (cg *CroneGenerator) Prepare(desc *Package) error {
	cg.desc = desc
	return nil
}

//Generate from generator interface
func (cg *CroneGenerator) Generate(bldr *Builder) (err error) {
	cg.desc = bldr.Descriptor
	cg.b = bldr
	for _, t := range bldr.File.Entries {
		if !t.HasModifier(TypeModifierSingleton) {
			continue
		}
		name := t.Name
		if n, ok := t.Annotations.GetStringAnnotation(codeGeneratorAnnotation, codeGenAnnoSingletonEngineAttr); ok {
			name = n
		}
		for _, m := range t.Modifiers {
			if h := m.Hook; h != nil && h.Key == TypeHookTime {
				cg.desc.Features.Set(FeatGoKind, FCGCronRequired, true)
				fname := cronSingletonDefaultFunctionName
				spec := h.Value
				if pref := strings.LastIndex(h.Value, cronSingletonFunctionPrefix); pref != -1 {
					fname = strings.Trim(h.Value[(pref+len(cronSingletonFunctionPrefix)):], " \t")
					if fname == "" {
						return fmt.Errorf("at %v: empty cron function name found", m.Pos)
					}
					spec = strings.Trim(h.Value[:pref], " \t")
				}
				if spec == "" {
					return fmt.Errorf("at %v: empty cron specification found", m.Pos)
				}
				cg.desc.Engine.Initializator.Add(
					jen.List(jen.Id("_"), jen.Id("err")).Op("=").Id(cronEngineVar).Dot("AddFunc").Params(
						jen.Lit(spec),
						jen.Func().Params().BlockFunc(func(g *jen.Group) {
							if h.Spec == HookJSPrefix {
								cg.desc.Features.Set(FeatGoKind, FCGScriptingRequired, true)
								g.Id(EngineVar).Dot(scriptingEngineField).Dot("ProcessSingleRet").Params(
									jen.Qual("context", "TODO"),
									jen.Lit(fname),
									jen.Map(jen.String()).Interface().Values(
										jen.Dict{jen.Lit(name): jen.Id(EngineVar).Dot(name)},
									),
								)
							} else {
								g.Id(EngineVar).Dot(name).Dot(fname).ParamsFunc(func(g *jen.Group) {
									g.Qual("context", "TODO").Params()
									g.Id(EngineVar)
								})
							}
						}),
					).Line(),
				)
				// if spec == "" || spec == HookGoPrefix {
				// 	return cg.getGoHookFuncFeature(value), FeatureProvidedNonCacheable
				// }
				// if spec == HookJSPrefix {
				// 	return cg.getJSHookFuncFeature(value), FeatureProvidedNonCacheable
				// }
			}
		}
		for _, m := range t.Methods {
			if h, ok := m.HaveHook(MethodHookTime); ok {
				cg.desc.Features.Set(FeatGoKind, FCGCronRequired, true)
				cg.desc.Engine.Initializator.Add(
					jen.List(jen.Id("_"), jen.Id("err")).Op("=").Id(cronEngineVar).Dot("AddFunc").Params(
						jen.Lit(h.Value),
						jen.Func().Params().Block(
							jen.Id(EngineVar).Dot(name).Dot(m.Name).ParamsFunc(func(g *jen.Group) {
								g.Qual("context", "TODO").Params()
								g.Id(EngineVar)
								for _, p := range m.Params {
									g.Add(bldr.goEmptyValue(p.Type))
								}
							}),
						),
					).Line(),
				)
			}
		}
	}
	return nil
}
