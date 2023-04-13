package gen

import (
	"errors"
	"fmt"

	"github.com/dave/jennifer/jen"
)

const (
	bscdGeneratorName          = "ChangeDetector"
	cdAnnotationChangeDetector = "record-changes"
)

const (
	bscdFeatureKind FeatureKind = "bs-change-det"

	bcdfFieldBitConst          = "field-bit-const"
	bcdfFieldBitOrder          = "field-bit-order"
	bcdfEntityBitMaskFieldName = "bit-mask-field"
	bcdfChangedRequired        = "changed-required"
	bcdfGenerateChecker        = "generate-checker"
)

func init() {
	RegisterPlugin(&BitSetChangeDetectorGenerator{})
}

// BitSetChangeDetectorGenerator generates code for historic fields or whole entities (recording changes)
type BitSetChangeDetectorGenerator struct {
	proj *Project
	desc *Package
	b    *Builder
}

func (cdg *BitSetChangeDetectorGenerator) Name() string {
	return bscdGeneratorName
}

// SetDescriptor from DescriptorAware
func (cdg *BitSetChangeDetectorGenerator) SetDescriptor(proj *Project) {
	cdg.proj = proj
}

// CheckAnnotation checks that annotation may be utilized by CodeGeneration
func (cdg *BitSetChangeDetectorGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	if ann.Name == cdAnnotationChangeDetector {
		switch v := item.(type) {
		case *Entity:
			requires := v.FS(FeaturesChangeDetectorKind, FCDRequired)
			if requires == FCDRField {
				return true, errors.New("change recording is not available at Entity and Field level at the same time")
			}
			v.Features.Set(FeaturesChangeDetectorKind, FCDRequired, FCDREntity)
			return true, nil
		case *Field:
			requires := v.parent.FS(FeaturesChangeDetectorKind, FCDRequired)
			if requires == FCDREntity {
				return true, errors.New("change recording is not available at Entity and Field level at the same time")
			}
			v.Features.Set(FeaturesChangeDetectorKind, FCDRequired, true)
			v.Features.Set(FeaturesChangeDetectorKind, bcdfGenerateChecker, true)
			v.parent.Features.Set(FeaturesChangeDetectorKind, FCDRequired, FCDRField)
			return true, nil
		}
	}
	return false, nil
}

// Prepare from Generator interface
func (cdg *BitSetChangeDetectorGenerator) Prepare(desc *Package) error {
	cdg.desc = desc

	for _, file := range desc.Files {
		for _, e := range file.Entries {
			requires := e.FS(FeaturesChangeDetectorKind, FCDRequired)
			if requires == FCDREntity || requires == FCDRField {
				bitsCount := 0
				for _, f := range e.Fields {
					if requires == FCDREntity || f.FB(FeaturesChangeDetectorKind, FCDRequired) {
						name := fmt.Sprintf("_%s%sBit", e.Name, f.Name)
						order := bitsCount
						f.Features.Set(bscdFeatureKind, bcdfFieldBitConst, name)
						f.Features.Set(bscdFeatureKind, bcdfFieldBitOrder, order)
						bitsCount++
					}
				}
				fieldName := fmt.Sprintf("_%sChanges", e.Name)
				bitsField := &Field{
					Name:        fieldName,
					parent:      e,
					Pos:         e.Pos,
					Annotations: Annotations{},
					Features:    Features{},
					Modifiers:   []*EntryModifier{{AttrModifier: string(AttrModifierAuxiliary)}},
					Tags:        map[string]string{},
					Type:        &TypeRef{Type: TipInt, NonNullable: true},
				}
				e.Fields = append(e.Fields, bitsField)
				e.FieldsIndex[bitsField.Name] = bitsField
				e.Features.Set(bscdFeatureKind, bcdfEntityBitMaskFieldName, fieldName)
				bitsField.Features.Set(FeatGoKind, FCGAttrType, jen.Int64())
				bitsField.Features.Set(FeatGoKind, FCGName, fieldName)
				bitsField.Features.Set(FeaturesAPIKind, FCIgnore, true)
			}
		}
	}
	return nil
}

// Generate from generator interface
func (cdg *BitSetChangeDetectorGenerator) Generate(b *Builder) (err error) {
	cdg.desc = b.Descriptor
	cdg.b = b
	for _, e := range b.File.Entries {
		requires := e.FS(FeaturesChangeDetectorKind, FCDRequired)
		if requires == FCDREntity || requires == FCDRField {
			constGroup := fmt.Sprintf("%s fields bits", e.Name)
			first := true
			for _, f := range e.Fields {
				name := f.FS(bscdFeatureKind, bcdfFieldBitConst)
				if name == "" {
					continue
				}
				stmt := jen.Id(name)
				if first {
					stmt = jen.Id(name).Op("=").Lit(0x01).Op("<<").Iota()
					first = false
				}
				b.AddConst(constGroup, stmt)
				if f.FB(FeaturesChangeDetectorKind, bcdfGenerateChecker) {
					b.Functions.Add(
						jen.Func().Parens(jen.Id("o").Op("*").Id(e.Name)).Id(cdg.b.GetMethodName(f, CGIsChangedMethod)).Params().Bool().BlockFunc(func(g *jen.Group) {
							g.Return(cdg.getFieldChangedStmt(f, jen.Id("o")))
						}),
						jen.Line(),
					)
				}
			}
		}
	}
	return nil
}

// ProvideFeature from FeatureProvider interface
func (cdg *BitSetChangeDetectorGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
	if kind == FeaturesChangeDetectorKind {
		switch name {
		case FCDChangedHook:
			switch o := obj.(type) {
			case *Field:
				o.Features.Set(FeaturesChangeDetectorKind, bcdfChangedRequired, true)
				o.Parent().Features.Set(FeaturesChangeDetectorKind, bcdfChangedRequired, true)
				return true, FeatureProvided
			case *Entity:
				o.Features.Set(FeaturesChangeDetectorKind, bcdfChangedRequired, true)
				return true, FeatureProvided
			default:
			}
		case FCDChangedCode:
			if f, isField := obj.(*Field); isField && f.FS(bscdFeatureKind, bcdfFieldBitConst) != "" {
				return func(args ...interface{}) jen.Code {
					a := &FeatureArguments{desc: cdg.desc}
					a.init("obj").parse(args)
					return cdg.getFieldChangedStmt(f, a.get("obj"))
				}, FeatureProvided
			}
		}
	}
	return
}

// OnEntityHook implements GeneratorHookHolder
func (cdg *BitSetChangeDetectorGenerator) OnEntityHook(name HookType, mod HookModifier, e *Entity, vars *GeneratorHookVars) (code *jen.Statement, order int) {
	if name == HookSave || name == HookUpdate || name == HookCreate {
		if mod == HMStart && e.FB(FeaturesChangeDetectorKind, bcdfChangedRequired) {
			requires := e.FS(FeaturesChangeDetectorKind, FCDRequired)
			for _, f := range e.Fields {
				if requires == FCDREntity || f.FB(FeaturesChangeDetectorKind, bcdfChangedRequired) {
					block := cdg.proj.OnHook(name, HMModified, f, vars)
					if block != nil {
						stmt := jen.If(cdg.getFieldChangedStmt(f, vars.GetObject())).Block(block).Line()
						if code == nil {
							code = stmt
						} else {
							code.Add(stmt)
						}
					}
				}
			}
		} else if mod == HMExit && e.FB(FeaturesChangeDetectorKind, bcdfChangedRequired) {
			bitsField := e.FS(bscdFeatureKind, bcdfEntityBitMaskFieldName)
			code = vars.GetObject().Dot(bitsField).Op("=").Lit(0)
		}
	}
	return
}

// OnFieldHook implements GeneratorHookHolder
func (cdg *BitSetChangeDetectorGenerator) OnFieldHook(name HookType, mod HookModifier, f *Field, vars *GeneratorHookVars) (code *jen.Statement, order int) {
	if (name == HookSet || name == HookSetNull) && mod == HMStart &&
		(f.Parent().FS(FeaturesChangeDetectorKind, FCDRequired) == FCDREntity || f.FB(FeaturesChangeDetectorKind, FCDRequired)) {
		code = cdg.getSetFieldChangedStmt(f, vars.GetObject())
		fieldName := f.FS(FeatGoKind, FCGName)
		switch name {
		case HookSet:
			newVal := vars.GetVar("newValue")
			if newVal == nil {
				newVal = jen.Id("val")
			}
			if newVal != nil {
				cond := &jen.Statement{}
				if f.Type.Array != nil {
					code = jen.Id("changed").Op(":=").Add(vars.GetObject()).Dot(fieldName).Op("==").Nil().Op("||").Len(jen.Add(vars.GetObject()).Dot(fieldName)).Op("!=").Len(newVal).Line().
						If(jen.Op("!").Id("changed")).Block(
						jen.For(jen.Id("i").Op(":=").Lit(0), jen.Id("i").Op("<").Len(newVal), jen.Id("i").Op("++")).Block(
							jen.If(jen.Add(vars.GetObject()).Dot(fieldName).Index(jen.Id("i")).Op("!=").Add(newVal).Index(jen.Id("i"))).Block(
								jen.Id("changed").Op("=").True(),
								jen.Break(),
							),
						),
					).Line().
						If(jen.Id("changed")).Block(code)
				} else {
					if f.FB(FeatGoKind, FCGPointer) {
						cond = jen.Add(vars.GetObject()).Dot(fieldName).Op("==").Nil().Op("||").Op("*")
					}
					code = jen.If(cond.Add(vars.GetObject()).Dot(fieldName).Op("!=").Add(newVal)).Block(code)
				}
			}
		case HookSetNull:
			code = jen.If(jen.Add(vars.GetObject()).Dot(fieldName).Op("!=").Nil()).Block(code)
		}
	}
	return
}

// OnMethodHook implements GeneratorHookHolder
func (cdg *BitSetChangeDetectorGenerator) OnMethodHook(name HookType, mod HookModifier, m *Method, vars *GeneratorHookVars) (code *jen.Statement, order int) {
	return nil, 0

}

func (cdg *BitSetChangeDetectorGenerator) getFieldChangedStmt(f *Field, obj *jen.Statement) *jen.Statement {
	if f.FS(bscdFeatureKind, bcdfFieldBitConst) == "" {
		return nil
	}
	bitsField := f.Parent().FS(bscdFeatureKind, bcdfEntityBitMaskFieldName)
	return jen.Add(obj).Dot(bitsField).Op("&").Id(f.FS(bscdFeatureKind, bcdfFieldBitConst)).Op("!=").Lit(0)
}

func (cdg *BitSetChangeDetectorGenerator) getSetFieldChangedStmt(f *Field, obj *jen.Statement) *jen.Statement {
	if f.FS(bscdFeatureKind, bcdfFieldBitConst) == "" {
		return nil
	}
	bitsField := f.Parent().FS(bscdFeatureKind, bcdfEntityBitMaskFieldName)
	return jen.Add(obj).Dot(bitsField).Op("=").Add(obj).Dot(bitsField).Op("|").Id(f.FS(bscdFeatureKind, bcdfFieldBitConst))
}
