package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard"
	"github.com/vc2402/vivard/resource"
	"github.com/vc2402/vivard/utils"
	"regexp"
	"strings"
	"unicode"
)

type ServiceGenerator struct {
	proj     *Project
	desc     *Package
	b        *Builder
	services map[string]*serviceDescriptor
}

const (
	serviceGeneratorName    = "Service"
	serviceAnnotation       = "service"
	serviceInjectAnnotation = "inject-service"
	saName                  = "name"
	saVar                   = "var"
	saType                  = "type"
	saPointer               = "usePointer"
	saPackage               = "package"

	ServiceFeatureKind FeatureKind = "service-feature"
	// SFKProvide is used for Singleton's field if inject required
	SFKProvide = "provide"
	// SFKInject is used for Singleton if at least one field of it requires SFKProvide
	SFKInject = "inject"
	// SFKRegister is used for singleton if it is required to register it as service
	SFKRegister = "register"
	// SFKEngineService requests Engine ref to service; returns field's name
	SFKEngineService = "engine-service"
	// sFKServices map of Engine services (string -> serviceDescriptor)
	sFKServices = "services"
	// sFKServiceProvider should be set for package if it contains singleton-service
	sFKServiceProvider = "service-provider"

	optionService = "service"
	// map of serviceName -> {[package, type]} (if package and type are absent no conversion will  be made)
	osServices = "services"
	ossPackage = "package"
	ossType    = "type"
)

type serviceDescriptor struct {
	name    string
	pckg    string
	tip     string
	varName string
}

func init() {
	RegisterPlugin(&ServiceGenerator{})
}

func (cg *ServiceGenerator) Name() string {
	return sequenceGeneratorName
}

// SetDescriptor from DescriptorAware
func (cg *ServiceGenerator) SetDescriptor(proj *Project) {
	cg.proj = proj
}
func (cg *ServiceGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	if ann.Name == serviceAnnotation {
		if ent, ok := item.(*Entity); ok {
			if !ent.HasModifier(TypeModifierSingleton) {
				return false, fmt.Errorf("at %v: %s can be used with singleton only", ann.Pos, serviceAnnotation)
			}
			return true, nil
		}
	}
	if ann.Name == serviceInjectAnnotation {
		if fld, ok := item.(*Field); ok {
			if !fld.Parent().HasModifier(TypeModifierSingleton) {
				return false, fmt.Errorf("at %v: %s can be used for member of singleton only", ann.Pos, serviceInjectAnnotation)
			}
			return true, nil
		}
	}
	return false, nil
}

func (cg *ServiceGenerator) Prepare(desc *Package) error {
	cg.desc = desc
	if cg.services == nil {
		cg.services = map[string]*serviceDescriptor{}
		if opts, ok := desc.Options().Custom[optionService].(map[string]interface{}); ok {
			if srv, ok := opts[osServices].(map[string]interface{}); ok {
				for s, d := range srv {
					var desc *serviceDescriptor
					if dd, ok := d.(map[string]interface{}); ok {
						if p, ok := dd[ossPackage].(string); ok {
							if t, ok := dd[ossType].(string); ok {
								desc = &serviceDescriptor{name: s, pckg: p, tip: t}
							}
						}
					}
					cg.services[s] = desc
				}

			}
		}
	}
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if an, ok := t.Annotations[serviceAnnotation]; ok && t.HasModifier(TypeModifierSingleton) {
				name, _ := an.GetNameTag(saName)
				if name == "" {
					cg.desc.AddWarning(fmt.Sprintf("at %v: service name not found; ignoring", an.Pos))
					continue
				}
				t.Features.Set(ServiceFeatureKind, SFKRegister, name)
				t.Pckg.Features.Set(ServiceFeatureKind, sFKServiceProvider, true)
			}
			if t.HasModifier(TypeModifierSingleton) {
				for _, field := range t.Fields {
					if an, ok := field.Annotations[serviceInjectAnnotation]; ok {
						name, _ := an.GetNameTag(saName)
						if name == "" {
							cg.desc.AddWarning(fmt.Sprintf("at %v: service name not found; ignoring", an.Pos))
							continue
						}
						var pckg string
						var tip string
						if t, ok := cg.proj.FindType(field.Type.Type); ok {
							pckg = t.packagePath
							tip = t.name
						}
						if p, ok := an.GetStringTag(saPackage); ok {
							pckg = p
						}
						if t, ok := an.GetStringTag(saType); ok {
							tip = t
						} else if an.GetBool(saPointer, false) {
							if strings.Trim(tip, " \t")[0] != '*' {
								tip = "*" + tip
							}
						}

						field.Features.Set(ServiceFeatureKind, SFKProvide, name)
						t.Features.Set(ServiceFeatureKind, SFKInject, true)
						if pckg != "" && tip != "" {
							cg.proj.CallFeatureFunc(t, ServiceFeatureKind, SFKEngineService, name, pckg, tip)
						} else {
							cg.proj.CallFeatureFunc(t, ServiceFeatureKind, SFKEngineService, name)
						}
					}
				}
			}
		}
	}
	return nil
}

func (cg *ServiceGenerator) Generate(b *Builder) (err error) {
	cg.desc = b.Descriptor
	cg.b = b
	//for _, t := range b.File.Entries {
	//	if t.FB(ServiceFeatureKind, SFKInject) {
	//		for _, field := range t.Fields {
	//			name := field.FS(ServiceFeatureKind, SFKProvide)
	//			if name != "" {
	//				cg.proj.CallFeatureFunc(t, ServiceFeatureKind, SFKEngineService, name)
	//			}
	//		}
	//
	//	}
	//}
	return nil
}

func (cg *ServiceGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	if kind == ServiceFeatureKind {
		if name == SFKEngineService {
			var p *Package
			switch v := obj.(type) {
			case *Entity:
				p = v.Pckg
			case *Package:
				p = v
			default:
				return
			}
			var sm map[string]serviceDescriptor
			smf, ok := p.Features.Get(ServiceFeatureKind, sFKServices)
			if !ok {
				sm = map[string]serviceDescriptor{}
				p.Features.Set(ServiceFeatureKind, sFKServices, sm)
			} else {
				sm = smf.(map[string]serviceDescriptor)
			}
			ret := func(args ...interface{}) jen.Code {
				if len(args) > 0 {
					if srvName, ok := args[0].(string); ok {
						var pckg string
						var tip string
						if wkSrv, ok := cg.services[srvName]; ok {
							pckg = wkSrv.pckg
							tip = wkSrv.tip
						} else {
							if len(args) > 2 {
								if p, ok := args[1].(string); ok {
									if t, ok := args[2].(string); ok {
										pckg = p
										tip = t
									}
								}
							}
							if pckg == "" || tip == "" {
								switch srvName {
								case resource.ServiceManager:
									pckg = ResourcePackage
									tip = "Manager"
								case resource.ServiceAccessChecker:
									pckg = ResourcePackage
									tip = "AccessChecker"
								case vivard.ServiceCRON:
									pckg = vivardPackage
									tip = "*CRONService"
								default:
									panic("package and type should be provided for service for 'engine-service' feature")
								}

							}
						}
						varName := cg.normalizeName(fmt.Sprintf("srv%s%s", strings.ToUpper(srvName[:1]), srvName[1:]))
						sm[srvName] = serviceDescriptor{
							name:    srvName,
							pckg:    pckg,
							tip:     tip,
							varName: varName,
						}
						return jen.Id(EngineVar).Dot(varName)
					}
				}
				panic("there is no service name for 'engine-service' feature")
			}
			return CodeHelperFunc(ret), FeatureProvided
		}
	}
	return
}

func (cg *ServiceGenerator) ProvideCodeFragment(module interface{}, action interface{}, point interface{}, ctx interface{}) interface{} {
	if module == CodeFragmentModuleGeneral {
		if cf, ok := ctx.(*CodeFragmentContext); ok && cf.Package != nil {
			if smf, ok := cf.Package.Features.Get(ServiceFeatureKind, sFKServices); ok {
				services := smf.(map[string]serviceDescriptor)
				switch action {
				case EngineNotAMethod:
					if point == CFGEngineMembers {
						first := true
						utils.WalkMap(
							services,
							func(desc serviceDescriptor, _ string) error {
								if first {
									first = false
								} else {
									cf.Add(jen.Line())
								}
								cf.Add(jen.Id(desc.varName).Add(desc.getType()))
								return nil
							},
						)
						//for _, desc := range services {
						//	if first {
						//		first = false
						//	} else {
						//		cf.Add(jen.Line())
						//	}
						//	cf.Add(jen.Id(desc.varName).Add(desc.getType()))
						//}
						return true
					}
				case MethodEnginePrepare:
					if point == CFGEngineEnter {
						utils.WalkMap(
							services,
							func(desc serviceDescriptor, srv string) error {
								cf.Add(jen.Id(EngineVar).Dot(desc.varName).Op("=").Id("v").Dot("GetService").Params(jen.Lit(srv)).Dot("Provide").Params().Assert(desc.getType()))
								return nil
							},
						)
						//for srv, desc := range services {
						//	cf.Add(jen.Id(EngineVar).Dot(desc.varName).Op("=").Id("v").Dot("GetService").Params(jen.Lit(srv)).Dot("Provide").Params().Assert(desc.getType()))
						//}
						return true
					}
					if point == CFGEngineExit {
						for _, file := range cf.Package.Files {
							for _, t := range file.Entries {
								if t.HasModifier(TypeModifierSingleton) && t.FB(ServiceFeatureKind, SFKInject) {
									for _, field := range t.Fields {
										if srv := field.FS(ServiceFeatureKind, SFKProvide); srv != "" {
											if sd, ok := services[srv]; ok {
												attrName := t.FS(FeatGoKind, FCGSingletonAttrName)
												cf.Add(jen.Id(EngineVar).Dot(attrName).Dot(field.Name).Op("=").Id(EngineVar).Dot(sd.varName))
											}
										}
									}
								}
							}
						}
						return true
					}
				}
			}
			if sp, ok := cf.Package.Features.GetBool(ServiceFeatureKind, sFKServiceProvider); ok && sp && action == MethodEngineRegisterService {
				ret := false
				for _, file := range cf.Package.Files {
					for _, t := range file.Entries {
						if t.HasModifier(TypeModifierSingleton) {
							if name := t.FS(ServiceFeatureKind, SFKRegister); name != "" {
								if st, ok := cf.Package.Engine.SingletonInits[t.Name]; ok {
									cf.Add(st)
									delete(cf.Package.Engine.SingletonInits, t.Name)
									for i, n := range strings.Fields(name) {
										if i > 0 {
											cf.Add(jen.Line())
										}
										cf.Add(jen.Id("v").Dot("WithService").Params(jen.Lit(n), jen.Id(EngineVar).Dot(t.FS(FeatGoKind, FCGSingletonAttrName))))
									}
									ret = true
								}
							}
						}
					}
				}
				return ret
			}
		}
	}
	return nil
}

var serviceFieldNameRegExp = regexp.MustCompile(`[a-zA-Z_0-9]+`)

func (cg *ServiceGenerator) normalizeName(name string) string {
	matches := serviceFieldNameRegExp.FindAllStringIndex(name, -1)
	if matches == nil {
		return name
	}
	ret := strings.Builder{}

	for i, match := range matches {
		add := []rune(name[match[0]:match[1]])
		if i > 0 {
			add[0] = unicode.ToUpper(add[0])
		}
		ret.WriteString(string(add))
	}
	return ret.String()
}

func (sd *serviceDescriptor) getType() jen.Code {
	if sd.tip != "" && sd.tip[0] == '*' {
		return jen.Op("*").Qual(sd.pckg, sd.tip[1:])
	}
	return jen.Qual(sd.pckg, sd.tip)
}
