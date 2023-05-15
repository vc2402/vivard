package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
)

const (
	objectRefPluginName     = "ObjectRef"
	objectRefAnnotation     = "object-ref"
	referenceableAnnotation = "referenceable"

	objectRefTypeName = "ObjectRef"
	derefFunctionName = "Deref"
	gqlUnionTypeName  = "ObjectRefUnion"
)

// ObjectRefGenerator generates Go code for $object-ref annotations
type ObjectRefGenerator struct {
	proj             *Project
	desc             *Package
	b                *Builder
	typeCreated      bool
	objectField      *Field
	referables       []*Entity
	requiresStringID bool
	warningCreated   bool
}

func init() {
	RegisterPlugin(&ObjectRefGenerator{})
}

func (cg *ObjectRefGenerator) Name() string {
	return objectRefPluginName
}

func (cg *ObjectRefGenerator) SetDescriptor(proj *Project) {
	cg.proj = proj
}

// CheckAnnotation checks that annotation may be utilized by CodeGeneration
func (cg *ObjectRefGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	cg.desc = desc
	if ann.Name == objectRefAnnotation {
		if fld, ok := item.(*Field); ok {
			if !cg.typeCreated {
				err := cg.createRefType()
				if err != nil {
					return true, err
				}
			}
			fld.Type.Type = fmt.Sprintf("%s.%s", InternalPackageName, objectRefTypeName)
			fld.Type.Complex = true
			fld.Type.Embedded = true
			if !fld.HasModifier(AttrModifierEmbedded) {
				fld.Modifiers = append(fld.Modifiers, &EntryModifier{AttrModifier: string(AttrModifierEmbedded)})
			}
			return true, nil
		}
		return true, fmt.Errorf("at %s: only fields can be object's referencies", ann.Pos)
	} else if ann.Name == referenceableAnnotation {
		if ent, ok := item.(*Entity); ok && !ent.HasModifier(TypeModifierSingleton) && !ent.HasModifier(TypeModifierTransient) {
			cg.referables = append(cg.referables, ent)
			return true, nil
		}
		return true, fmt.Errorf("at %s: only storeable types can be referenceable", ann.Pos)
	}
	return false, nil
}

// Prepare from Generator interface
func (cg *ObjectRefGenerator) Prepare(desc *Package) error {
	cg.desc = desc
	for _, file := range desc.Files {
		for _, ent := range file.Entries {
			if _, ok := ent.Annotations[referenceableAnnotation]; !ok {
				continue
			}
			idField := ent.GetIdField()
			if idField == nil {
				return fmt.Errorf("at %v: %s: there is no id field for type %s", ent.Pos, objectRefPluginName, ent.Name)
			}
			if idField.Type.Type != TipInt && idField.Type.Type != TipString {
				return fmt.Errorf("at %v: %s: type %s: only string and int may be used for referenceable types id", idField.Pos, objectRefPluginName, ent.Name)
			}
			if idField.Type.Type != TipInt {
				cg.requiresStringID = true
			}
		}
	}

	//cg.objectField.Features.Set(GQLFeatures, GQLFTypeTag, gqlUnionTypeName)
	if cg.typeCreated && len(cg.referables) == 0 && !cg.warningCreated {
		desc.Project.AddWarning("objectRef without referenceable types")
		cg.warningCreated = true
	}
	return nil
}

// Generate from generator interface
func (cg *ObjectRefGenerator) Generate(bldr *Builder) (err error) {
	cg.desc = bldr.Descriptor
	cg.b = bldr
	//cg.objectField.Features.Set(GQLFeatures, GQLFTypeTag, gqlUnionTypeName)

	if bldr.Descriptor.Name == InternalPackageName && bldr.File == cg.objectField.Parent().File {
		cg.proj.CallFeatureFunc(bldr, GQLFeatures, GQLGenerateUnionType, gqlUnionTypeName, cg.referables)
		cg.generateDeref(cg.objectField.Parent())
	}
	return nil
}

func (cg *ObjectRefGenerator) generateDeref(e *Entity) {
	f := jen.Func().Parens(jen.Id("o").Op("*").Id(e.FS(FeatGoKind, FCGName))).Id(derefFunctionName).Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id(EngineVar).Op("*").Id("Engine"),
	).Parens(jen.List(jen.Any(), jen.Error())).BlockFunc(func(g *jen.Group) {
		var ifstmt = &jen.Statement{}
		for i, e := range cg.referables {
			cond := jen.Id("o").Dot("Package").Op("==").Lit(e.Pckg.Name).Op("&&").
				Id("o").Dot("Type").Op("==").Lit(e.Name)
			stmt := jen.Return(jen.Id(EngineVar).Dot(cg.desc.GetExtEngineRef(e.Pckg.Name)).Dot(cg.desc.GetMethodName(MethodGet, e.Name)).
				Parens(jen.List(jen.Id("ctx"), jen.Id("o").Dot("ObjectID"))))

			if i > 0 {
				ifstmt.Add(jen.Else().If(cond).Block(stmt))
			} else {
				ifstmt.Add(jen.If(cond).Block(stmt))
			}

		}
		g.Add(ifstmt)
		g.Return(jen.List(jen.Nil(), jen.Qual("errors", "New").Parens(jen.Lit("object is undefined"))))
	})
	cg.b.Functions.Add(f)
}

func (cg *ObjectRefGenerator) createRefType() error {
	if cg.typeCreated {
		return nil
	}
	pckg := cg.proj.GetInternalPackage()
	ent, err := pckg.AddType(objectRefTypeName)
	if err != nil {
		return err
	}
	ent.TypeModifers = map[TypeModifier]bool{TypeModifierEmbeddable: true}
	ent.Features.Set(FeaturesDBKind, FCIgnore, true)
	f, _ := ent.AddField("Package", TipString)
	f.Type.NonNullable = true
	f, _ = ent.AddField("Type", TipString)
	f.Type.NonNullable = true
	idTip := TipInt
	if cg.requiresStringID {
		idTip = TipString
	}
	f, _ = ent.AddField("ObjectID", idTip)
	f.Type.NonNullable = true
	f, _ = ent.AddField("Object", TipAny)
	f.Type.NonNullable = true
	f.Features.Set(GQLFeatures, GQLFUseDefinedType, gqlUnionTypeName)
	f.Modifiers = append(f.Modifiers, &EntryModifier{AttrModifier: string(AttrModifierCalculated)})
	f.Modifiers = append(f.Modifiers, &EntryModifier{Hook: &Hook{Key: AttrHookCalculate, Value: derefFunctionName}})
	// may be not the most elegant way but the only one at the moment
	f.Modifiers = append(f.Modifiers, &EntryModifier{Annotation: &Annotation{Name: "vue", Values: []*AnnotationTag{{Key: "ignore"}}}})
	trueValue := Boolean(true)
	f.Annotations["vue"] = &Annotation{Name: "vue", Values: []*AnnotationTag{{Key: "ignore", Value: &AnnotationValue{Bool: &trueValue}}}}
	f.Annotations["gql"] = &Annotation{Name: "gql", Values: []*AnnotationTag{{Key: "skip", Value: &AnnotationValue{Bool: &trueValue}}}}
	cg.objectField = f
	cg.typeCreated = true
	return nil
}
