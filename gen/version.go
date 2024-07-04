package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
)

type VersionGenerator struct {
	proj *Project
	desc *Package
	b    *Builder
	o    VersionOptions
}

type VersionOptions struct {
	// DefaultBehaviour behaviour on save with different version; default is warn (log file)
	// vaBehaviourWarning (warn), vaBehaviourError (err) or vaBehaviourNothing (ignore)
	DefaultBehaviour string
	// TryMerge try check whether changes are intersecting or no (requires changes recording)
	TryMerge         bool
	DefaultFieldName string
	// Scope whether object or type
	Scope string
}

const (
	versionGeneratorName = "Version"
	versionAnnotation    = "version"
	vaField              = "field"
	vaBehaviourWarning   = "warn"
	vaBehaviourError     = "err"
	vaBehaviourNothing   = "ignore"
	vaTryMerge           = "tryMerge"
	vaObjectWise         = "object"
	vaTypeWise           = "type"
	// to set name of sequence; bu default it is <typeName>Version
	vaSequenceName = "sequence"

	VersionFeatureKind FeatureKind = "f-ver"
	// VFField if set for Entity version should be tracked in field with name as feature value
	VFField = "field"
	// VFBehaviour is string (one of  vaBehaviourWarning, vaBehaviourError or vaBehaviourNothing)
	//  means what to do if version of saving object is not the same as current
	VFBehaviour    = "beh"
	VFTryMerge     = "mer"
	VFScope        = "scope"
	VFSequenceName = "seq-name"

	VersionDefaultFieldName = "Version"
)

func init() {
	RegisterPlugin(
		&VersionGenerator{
			o: VersionOptions{
				DefaultBehaviour: vaBehaviourWarning,
				DefaultFieldName: VersionDefaultFieldName,
				Scope:            vaTypeWise,
			},
		},
	)
}

func (vg *VersionGenerator) Name() string {
	return versionGeneratorName
}

func (vg *VersionGenerator) SetOptions(options any) error {
	if options != nil {
		return OptionsAnyToStruct(options, &vg.o)
	}
	return nil
}

// SetDescriptor from DescriptorAware
func (vg *VersionGenerator) SetDescriptor(proj *Project) {
	vg.proj = proj
}
func (vg *VersionGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	if ann.Name == versionAnnotation {
		if ent, ok := item.(*Entity); ok {
			beh := vg.o.DefaultBehaviour
			tryMerge := vg.o.TryMerge
			fieldName := vg.o.DefaultFieldName
			scope := vg.o.Scope
			sequenceName := fmt.Sprintf("%sVersion", ent.Name)
			for _, value := range ann.Values {
				switch value.Key {
				case vaField:
					if value.Value == nil || value.Value.String == nil {
						return true, fmt.Errorf("at %v: annotation '%s:%s' should be a string", ann.Pos, versionAnnotation, vaField)
					}
					fieldName = *value.Value.String
				case vaBehaviourWarning, vaBehaviourError, vaBehaviourNothing:
					if b, ok := value.GetBool(); ok {
						if b {
							beh = value.Key
						}
					} else {
						return true, fmt.Errorf("at %v: annotation '%s:%s' should be bool", ann.Pos, versionAnnotation, value.Key)
					}
				case vaTryMerge:
					if b, ok := value.GetBool(); ok {
						tryMerge = b
					} else {
						return true, fmt.Errorf("at %v: annotation '%s:%s' should be bool", ann.Pos, versionAnnotation, value.Key)
					}
				case vaTypeWise, vaObjectWise:
					if b, ok := value.GetBool(); ok {
						if b {
							scope = value.Key
						}
					} else {
						return true, fmt.Errorf("at %v: annotation '%s:%s' should be bool", ann.Pos, versionAnnotation, value.Key)
					}
				case vaSequenceName:
					if value.Value == nil || value.Value.String == nil {
						return true, fmt.Errorf(
							"at %v: annotation '%s:%s' should be a string",
							ann.Pos,
							versionAnnotation,
							vaSequenceName,
						)
					}
					sequenceName = *value.Value.String
				}
			}
			if ent.GetField(fieldName) != nil {
				return true, fmt.Errorf("at %v: version: type '%s' already has field '%s'", ent.Pos, ent.Name, fieldName)
			}
			ent.Features.Set(VersionFeatureKind, VFField, fieldName)
			ent.Features.Set(VersionFeatureKind, VFBehaviour, beh)
			ent.Features.Set(VersionFeatureKind, VFTryMerge, tryMerge)
			ent.Features.Set(VersionFeatureKind, VFScope, scope)
			ent.Features.Set(VersionFeatureKind, VFSequenceName, sequenceName)
			versionField := &Field{
				Name:        fieldName,
				parent:      ent,
				Pos:         ent.Pos,
				Annotations: Annotations{},
				Features:    Features{},
				Modifiers:   []*EntryModifier{},
				Tags:        map[string]string{},
				Type:        &TypeRef{Type: TipInt, NonNullable: true},
			}
			ent.Fields = append(ent.Fields, versionField)
			ent.FieldsIndex[versionField.Name] = versionField
			versionField.Features.Set(FeatGoKind, FCGAttrType, jen.Int())
			versionField.Features.Set(FeatGoKind, FCGName, fieldName)
		} else {
			return true, fmt.Errorf("at %v: annotation '%s' can be used for entity only", ann.Pos, versionAnnotation)
		}
		return true, nil
	}
	return false, nil
}

func (vg *VersionGenerator) Prepare(desc *Package) error {
	//for _, file := range desc.Files {
	//	for _, e := range file.Entries {
	//		if fn := e.FS(VersionFeatureKind, VFField); fn != "" {
	//
	//		}
	//	}
	//}
	return nil
}

func (vg *VersionGenerator) Generate(b *Builder) (err error) {
	return nil
}

func (vg *VersionGenerator) ProvideCodeFragment(
	module interface{},
	action interface{},
	point interface{},
	ctx interface{},
) interface{} {
	if module == CodeFragmentModuleGeneral {
		if cf, ok := ctx.(*CodeFragmentContext); ok {
			if cf.Entity != nil {
				if fn := cf.Entity.FS(VersionFeatureKind, VFField); fn != "" {
					if (action == MethodSet || action == MethodNew) &&
						point == CFGPointEnterBeforeHooks {
						stmt := &jen.Statement{}
						if action == MethodSet {
							stmt = jen.If(
								cf.GetParam(ParamObject).Dot(fn).Op("!=").Id(cf.ObjVar).Dot(fn).BlockFunc(
									func(g *jen.Group) {
										//TODO check behaviour
										idField := cf.Entity.GetIdField()
										g.Add(
											vg.proj.CallCodeFeatureFunc(
												cf.Entity, LogFeatureKind, LFWarn,
												"version mismatch",
												"f", jen.Lit(cf.MethodName),
												"id", cf.GetParam(ParamObject).Dot(idField.Name),
												"ver", jen.Id("o").Dot(fn),
												"curr", jen.Id(cf.ObjVar).Dot(fn),
											),
										)
									},
								),
							).Line()
						}
						if cf.Entity.FS(VersionFeatureKind, VFScope) == vaTypeWise {
							seqName := cf.Entity.FS(VersionFeatureKind, VFSequenceName)
							stmt.Add(
								vg.proj.CallCodeFeatureFunc(
									cf.Entity,
									SequenceFeatures,
									SFGenerateSequenceCall,
									seqName,
									jen.Id("o").Dot(fn),
								),
							)
						} else {
							stmt.Add(jen.Id("o").Dot(fn).Op("++"))
						}
						stmt.Line()
						cf.Add(stmt)
						return true
					}
				}
			}
		}
	}
	return nil
}
