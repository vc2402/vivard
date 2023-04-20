package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/vc2402/vivard/resource"
	"github.com/vc2402/vivard/utils"
)

type ResourceGenerator struct {
	proj           *Project
	desc           *Package
	b              *Builder
	root           string
	prefix         string
	delimiter      string
	checkByDefault bool
}

type resourceDescriptor struct {
	key          string
	description  string
	parent       string
	checkAccess  bool
	idVarName    string
	keyConstName string
}

const (
	resourceGeneratorName = "Resource"
	resourceAnnotation    = "resource"
	raKey                 = "key"
	raDescription         = "description"
	raParent              = "parent"
	raCheckAccess         = "checkAccess"

	ResourceFeatureKind FeatureKind = "resource-feature"
	// RFKKey may be set tor Entity that is resource or Package to define its resource key
	RFKKey         = "key"
	RFKParent      = "parent"
	RFKCheckAccess = "check-access"
	RFConstName    = "const"
	RFVarName      = "var"
	// RFResources is map of required resources for package
	RFResources = "resources"
	// RFRequired set to true for package if there are resources
	RFRequired           = "required"
	RFManagerField       = "manager-field"
	RFAccessCheckerField = "access-checker"

	optionsResource = "resource"
	orRootResource  = "root"
	orKeyPrefix     = "prefix"
	orKeyDelimiter  = "delimiter"
)

const (
	funcInitResources = "initResources"
	ResourcePackage   = "github.com/vc2402/vivard/resource"
)

func init() {
	RegisterPlugin(&ResourceGenerator{})
}

func (cg *ResourceGenerator) Name() string {
	return resourceGeneratorName
}

func (cg *ResourceGenerator) SetOptions(options any) error {
	if opts, ok := options.(map[string]any); ok {
		if root, ok := opts[orRootResource].(string); ok {
			cg.root = root
		}
		if pr, ok := opts[orKeyPrefix].(string); ok {
			cg.prefix = pr
		} else {
			cg.prefix = cg.root
		}
		if d, ok := opts[orKeyDelimiter].(string); ok {
			cg.delimiter = d
		}
	}
	return nil
}

// SetDescriptor from DescriptorAware
func (cg *ResourceGenerator) SetDescriptor(proj *Project) {
	cg.proj = proj
}
func (cg *ResourceGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	if ann.Name == resourceAnnotation {
		return true, nil
	}
	return false, nil
}

func (cg *ResourceGenerator) Prepare(desc *Package) error {
	if cg.delimiter == "" {
		cg.delimiter = ":"
		if opts, ok := desc.Options().Custom[optionsResource].(map[string]interface{}); ok {
			if err := cg.SetOptions(opts); err != nil {
				return err
			}
		}
	}
	prefix := desc.Name
	if cg.prefix != "" {
		prefix = fmt.Sprintf("%s%s%s", cg.prefix, cg.delimiter, prefix)
	}
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			if a, ok := t.Annotations[resourceAnnotation]; ok {
				key := a.GetString(raKey, fmt.Sprintf("%s%s%s", prefix, cg.delimiter, t.Name))
				description := a.GetString(raDescription, t.Name)
				t.Features.Set(ResourceFeatureKind, RFKKey, key)
				var parent string
				if p, ok := a.GetStringTag(raParent); ok {
					parent = p
				} else {
					desc.Features.Set(ResourceFeatureKind, RFKKey, prefix)
					parent = prefix
				}
				checkAccess := a.GetBool(raCheckAccess, cg.checkByDefault)
				varName := fmt.Sprintf("Resource%sID", t.Name)
				constName := fmt.Sprintf("Resource%sKey", t.Name)
				t.Features.Set(ResourceFeatureKind, RFKParent, parent)
				t.Features.Set(ResourceFeatureKind, RFKCheckAccess, checkAccess)
				t.Features.Set(ResourceFeatureKind, RFConstName, constName)
				t.Features.Set(ResourceFeatureKind, RFVarName, varName)
				rd := resourceDescriptor{
					key:          key,
					description:  description,
					parent:       parent,
					checkAccess:  checkAccess,
					idVarName:    varName,
					keyConstName: constName,
				}
				if descriptors, ok := desc.Features.Get(ResourceFeatureKind, RFResources); ok {
					descriptors.(map[string]resourceDescriptor)[key] = rd
				} else {
					desc.Features.Set(ResourceFeatureKind, RFResources, map[string]resourceDescriptor{key: rd})
					desc.Features.Set(ResourceFeatureKind, RFRequired, true)
					desc.Features.Set(ResourceFeatureKind, RFManagerField, cg.proj.CallFeatureFunc(t, ServiceFeatureKind, SFKEngineService, resource.ServiceManager))
				}
				if checkAccess {
					desc.Features.Set(ResourceFeatureKind, RFAccessCheckerField, cg.proj.CallFeatureFunc(t, ServiceFeatureKind, SFKEngineService, resource.ServiceAccessChecker))
				}
			}
		}
	}
	return nil
}

func (cg *ResourceGenerator) Generate(b *Builder) (err error) {
	cg.desc = b.Descriptor
	cg.b = b
	//for _, t := range b.File.Entries {
	//	cg.generateVarsAndConstants(t)
	//}
	return nil
}

func (cg *ResourceGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	return
}

func (cg *ResourceGenerator) ProvideCodeFragment(module interface{}, action interface{}, point interface{}, ctx interface{}) interface{} {
	if module == CodeFragmentModuleGeneral {
		if cf, ok := ctx.(*CodeFragmentContext); ok {
			if cf.Package != nil && cf.Package.Features.Bool(ResourceFeatureKind, RFRequired) {
				switch action {
				case MethodEngineStart:
					if point == CFGEngineExit {
						cf.Add(jen.Id("err").Op("=").Id(EngineVar).Dot(funcInitResources).Params())
						cf.AddCheckError()
					}
				case EngineNotAMethod:
					switch point {
					case CFGEngineFileFunctions:
						cg.generateResourcesInitializer(cf)
						return true
					case CFGEngineFileGlobals:
						cg.generateVarsAndConstants(cf)
						return true
					}
				}
			} else if cf.Entity != nil && cf.Entity.FB(ResourceFeatureKind, RFKCheckAccess) && point == CFGPointEnterBeforeHooks {
				var accessKind jen.Code
				var obj jen.Code
				switch cf.MethodKind {
				case MethodSet:
					accessKind = jen.Qual(ResourcePackage, "AccessWrite")
					obj = cf.GetParam(ParamObject)
				case MethodGet:
					accessKind = jen.Qual(ResourcePackage, "AccessRead")
					obj = cf.GetParam(ParamID)
				case MethodNew:
					accessKind = jen.Qual(ResourcePackage, "AccessCreate")
					obj = cf.GetParam(ParamObject)
				case MethodDelete:
					accessKind = jen.Qual(ResourcePackage, "AccessDelete")
					obj = cf.GetParam(ParamID)
				case MethodList, MethodLookup, MethodFind:
					accessKind = jen.Qual(ResourcePackage, "AccessList")
					obj = jen.Nil()
				default:
					return nil
				}
				checker := cf.Entity.Pckg.Features.Stmt(ResourceFeatureKind, RFAccessCheckerField)
				cf.Add(
					cf.GetErr().Op("=").Add(checker).Dot("CheckResourceAccess").Params(
						cf.GetParam(ParamContext),
						jen.Id(cf.Entity.FS(ResourceFeatureKind, RFVarName)),
						obj,
						accessKind,
					),
				)
				cf.AddCheckError()
			}
		}
	}
	return nil
}

func (cg *ResourceGenerator) generateVarsAndConstants(cf *CodeFragmentContext) error {
	if ds, ok := cf.Package.Features.Get(ResourceFeatureKind, RFResources); ok {
		descriptors := ds.(map[string]resourceDescriptor)
		utils.WalkMap(
			descriptors,
			func(descriptor resourceDescriptor, key string) error {
				cf.Add(jen.Const().Id(descriptor.keyConstName).Op("=").Qual(ResourcePackage, "Key").Parens(jen.Lit(key)).Line())
				cf.Add(jen.Var().Id(descriptor.idVarName).Qual(ResourcePackage, "ID").Line())
				return nil
			},
		)
		//for key, descriptor := range descriptors {
		//	cf.Add(jen.Const().Id(descriptor.keyConstName).Op("=").Qual(ResourcePackage, "Key").Parens(jen.Lit(key)).Line())
		//	cf.Add(jen.Var().Id(descriptor.idVarName).Qual(ResourcePackage, "ID").Line())
		//}
	}
	//constName := t.FS(ResourceFeatureKind, RFConstName)
	//if constName != "" {
	//	key := t.FS(ResourceFeatureKind, RFKKey)
	//	cg.b.consts["resource"] = append(cg.b.consts["resource"], jen.Id(constName).Op("=").Lit(key))
	//	varName := t.FS(ResourceFeatureKind, RFVarName)
	//	cg.b.vars["resource"] = append(cg.b.vars["resource"], jen.Id(varName).Int())
	//}
	return nil
}

func (cg *ResourceGenerator) generateResourcesInitializer(cf *CodeFragmentContext) error {
	if ds, ok := cf.Package.Features.Get(ResourceFeatureKind, RFResources); ok {
		descriptors := ds.(map[string]resourceDescriptor)
		fld, _ := cf.Package.Features.Get(ResourceFeatureKind, RFManagerField)
		fldCode := fld.(jen.Code)
		cf.Add(jen.Func().Parens(jen.Id(EngineVar).Op("*").Id("Engine")).Id(funcInitResources).Params().Parens(jen.Err().Error()).BlockFunc(func(g *jen.Group) {
			cf.Push(g)
			if rk, ok := cf.Package.Features.GetString(ResourceFeatureKind, RFKKey); ok {
				cf.Add(jen.List(jen.Id("_"), jen.Id("err")).Op("=").Add(fldCode).Dot("FindResource").Params(jen.Lit(rk)))
				cf.Add(jen.If(jen.Id("err").Op("!=").Nil()).Block(
					jen.List(jen.Id("_"), jen.Id("err")).Op("=").Add(fldCode).Dot("CreateResource").Params(
						jen.Lit(rk),
						jen.Lit(fmt.Sprintf("%s Package", cf.Package.Name)),
						jen.Qual(ResourcePackage, "Key").Parens(jen.Lit(cg.root)),
					),
				))
				cf.Add(jen.If(jen.Id("err").Op("!=").Nil()).Block(
					jen.Return(jen.Id("err"))),
				)
			}
			utils.WalkMap(
				descriptors,
				func(desc resourceDescriptor, _ string) error {
					cf.Add(jen.List(jen.Id(desc.idVarName), jen.Id("err")).Op("=").Add(fldCode).Dot("FindResource").Params(jen.Id(desc.keyConstName)))
					cf.Add(jen.If(jen.Id("err").Op("!=").Nil()).Block(
						jen.List(jen.Id(desc.idVarName), jen.Id("err")).Op("=").Add(fldCode).Dot("CreateResource").Params(
							jen.Id(desc.keyConstName),
							jen.Lit(desc.description),
							jen.Qual(ResourcePackage, "Key").Parens(jen.Lit(desc.parent)),
						),
					))
					cf.Add(jen.If(jen.Id("err").Op("!=").Nil()).Block(
						jen.Return(jen.Id("err"))),
					)
					return nil
				},
			)
			cf.Add(jen.Return(jen.Nil()))

			cf.Pop()
		}))
	}
	return nil
}
