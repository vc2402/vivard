package gen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

//HistoryGenerator generates code for historic fields or whole entities (recording changes)
type HistoryGenerator struct {
	proj *Project
	desc *Package
	b    *Builder
}

const (
	historyAnn               = "historic"
	historyAnnIncapsulate    = "incapsulate"
	historyAnnFields         = "fields"
	historyAnnFieldTimestamp = "timestamp"
	historyAnnFieldUserID    = "uid"
	historyAnnFieldUserName  = "uname"
	historyAnnFieldSource    = "source"
)

const (
	FeatureHistKind FeatureKind = "historic"

	// FHCollect - boolean for field: history collection is required
	FHCollect = "collect"

	//FHHistoryFieldName - string: name of historic attr in object
	FHHistoryFieldName = "hist-field-name"

	//FHHistoryField - string: name of historic attr in object
	FHHistoryField = "hist-field"

	//FHHistoryEntityName - string: name of entity for history
	FHHistoryEntityName = "hist-entity-name"

	//FHHistoryEntity - *Entity: entity for history
	FHHistoryEntity = "hist-entity"

	//FHHistoryOf - *Field - for field refs field it holds history of; for entity - entity for which field this entity is created
	FHHistoryOf = "history-of"

	FHKind       = "kind"
	FHkField     = "field"
	FHkTimestamp = "timestamp"
	FHkUserID    = "uid"
	FHkUserName  = "username"
	FHkSource    = "source"
	FHkCustom    = "custom"

	FHAttr = "attr"

	FHType      = "type"
	FHtSetter   = "setter"
	FHtFunction = "function"
)

//SetDescriptor from DescriptorAware
func (hg *HistoryGenerator) SetDescriptor(proj *Project) {
	hg.proj = proj
}

//CheckAnnotation checks that annotation may be utilized by CodeGeneration
func (hg *HistoryGenerator) CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error) {
	if ann.Name == historyAnn {
		if field, ok := item.(*Field); ok {
			cdf := desc.GetFeature(field, FeaturesChangeDetectorKind, FCDChangedHook)
			if cdf != true {
				return true, errors.New("history generator requires change detector; try to add BitSetChangeDetectorGenerator")
			}
			histEntName := fmt.Sprintf("%s%sHistory", field.Parent().Name, field.Name)
			he := &Entity{
				Pos:          field.Pos,
				Annotations:  Annotations{},
				Features:     Features{},
				File:         field.Parent().File,
				Name:         histEntName,
				TypeModifers: map[TypeModifier]bool{TypeModifierEmbeddable: true},
				Pckg:         field.Parent().Pckg,
				FieldsIndex:  map[string]*Field{},
			}
			field.Features.Set(FeatureHistKind, FHHistoryEntityName, histEntName)
			field.Features.Set(FeatureHistKind, FHHistoryEntity, he)
			he.Features.Set(FeaturesCommonKind, FCReadonly, true)
			he.Features.Set(FeatureHistKind, FHHistoryOf, field)

			fa := strings.Split(ann.GetString(historyAnnFields, historyAnnFieldTimestamp), ",")
			fields := make([]*Field, 1, 5)
			fields[0] = &Field{
				Name:        field.Name,
				parent:      he,
				Annotations: Annotations{},
				Features:    Features{},
				Pos:         field.Pos,
				Tags:        map[string]string{},
				Type:        field.Type,
			}
			fields[0].Features.Set(FeatureHistKind, FHKind, FHkField)
			he.FieldsIndex[fields[0].Name] = fields[0]
			for _, f := range fa {
				fld := &Field{
					parent:      he,
					Annotations: Annotations{},
					Features:    Features{},
					Pos:         field.Pos,
					Tags:        map[string]string{},
				}
				switch strings.Trim(f, " \t") {
				case historyAnnFieldTimestamp:
					fld.Name = "Timestamp"
					fld.Type = &TypeRef{Type: TipDate, NonNullable: true}
					fld.Features.Set(FeatureHistKind, FHKind, FHkTimestamp)
				case historyAnnFieldUserID:
					fld.Name = "UserID"
					fld.Type = &TypeRef{Type: TipInt, NonNullable: true}
					fld.Features.Set(FeatureHistKind, FHKind, FHkUserID)
				case historyAnnFieldUserName:
					fld.Name = "UserName"
					fld.Type = &TypeRef{Type: TipString, NonNullable: true}
					fld.Features.Set(FeatureHistKind, FHKind, FHkUserName)
				case historyAnnFieldSource:
					fld.Name = "Source"
					fld.Type = &TypeRef{Type: TipString, NonNullable: true}
					fld.Features.Set(FeatureHistKind, FHKind, FHkSource)
				default:
					//TODO add possibility set custom fields
					return true, fmt.Errorf("at %v: unknown attr for historic field: %s", field.Pos, f)
				}
				fields = append(fields, fld)
				he.FieldsIndex[fld.Name] = fld
			}

			he.Fields = fields

			field.Parent().File.AddEntity(he)
			hisFldName := fmt.Sprintf("%sHistory", field.Name)
			hisField := &Field{
				Name:        hisFldName,
				parent:      field.Parent(),
				Pos:         field.Pos,
				Annotations: Annotations{},
				Features:    Features{},
				Modifiers:   []*EntryModifier{{AttrModifier: string(AttrModifierEmbedded)}},
				Tags:        map[string]string{},
				Type:        &TypeRef{Array: &TypeRef{Type: he.Name, NonNullable: true}, NonNullable: true, Complex: true, Embedded: true},
			}
			hisField.PostProcess()
			field.Parent().Fields = append(field.Parent().Fields, hisField)
			field.Parent().FieldsIndex[hisField.Name] = hisField
			if field.Parent().FS(FeaturesChangeDetectorKind, FCDRequired) == "" {
				field.Parent().Features.Set(FeaturesChangeDetectorKind, FCDRequired, FCDRField)
			}
			field.Features.Set(FeaturesChangeDetectorKind, FCDRequired, true)
			field.Features.Set(FeatureHistKind, FHCollect, true)
			field.Features.Set(FeatureHistKind, FHHistoryFieldName, hisFldName)
			field.Features.Set(FeatureHistKind, FHHistoryField, hisField)

			hisField.Features.Set(FeaturesCommonKind, FCReadonly, true)
			hisField.Features.Set(FeatureHistKind, FHHistoryOf, field)

			return true, nil
		}
		return true, errors.New("historic may be used with fields only at the moment")
	}
	return false, nil
}

//Prepare from Generator interface
func (hg *HistoryGenerator) Prepare(desc *Package) error {
	return nil
}

//Generate from generator interface
func (hg *HistoryGenerator) Generate(b *Builder) (err error) {
	return nil
}

// func (hg *HistoryGenerator) ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult) {
// 	return
// }

//OnEntityHook implements GeneratorHookHolder
func (hg *HistoryGenerator) OnEntityHook(name HookType, mod HookModifier, e *Entity, vars *GeneratorHookVars) (code *jen.Statement, order int) {

	return nil, 0
}

//OnFieldHook implements GeneratorHookHolder
func (hg *HistoryGenerator) OnFieldHook(name HookType, mod HookModifier, f *Field, vars *GeneratorHookVars) (code *jen.Statement, order int) {
	if (name == HookSave || name == HookUpdate || name == HookCreate) && mod == HMModified && f.FB(FeatureHistKind, FHCollect) {
		hfName := f.FS(FeatureHistKind, FHHistoryFieldName)
		entName := f.FS(FeatureHistKind, FHHistoryEntityName)
		histEnt, _ := f.Features.GetEntity(FeatureHistKind, FHHistoryEntity)
		fieldName := f.FS(FeatGoKind, FCGName)
		cond := &jen.Statement{}
		compareWith := jen.Add(vars.GetObject().Dot(hfName).Index(jen.Len(vars.GetObject().Dot(hfName)).Op("-").Lit(1)).Dot(f.Name))
		if f.FB(FeatGoKind, FCGPointer) {
			cond = jen.Add(vars.GetObject()).Dot(fieldName).Op("==").Nil().Op("&&").Add(compareWith).Op("!=").Nil().Op("||").Line().
				Add(vars.GetObject()).Dot(fieldName).Op("!=").Nil().Op("&&").Add(compareWith).Op("==").Nil().Op("||").Line().
				Add(vars.GetObject()).Dot(fieldName).Op("!=").Nil().Op("&&").Add(compareWith).Op("!=").Nil().Op("&&").Op("*")
			compareWith = jen.Op("*").Add(compareWith)
		}
		cond = cond.Add(vars.GetObject().Dot(f.Name).Op("!=").Add(compareWith))
		code =
			jen.If(
				jen.Len(vars.GetObject().Dot(hfName)).Op("==").Lit(0).Op("||").Line().
					Add(cond),
			).BlockFunc(
				func(g *jen.Group) {
					g.Add(vars.GetObject().Dot(hfName).Op("=").Append(
						vars.GetObject().Dot(hfName),
						jen.Op("&").Id(entName).Values(jen.DictFunc(
							func(d jen.Dict) {

								for _, hef := range histEnt.Fields {
									switch hef.FS(FeatureHistKind, FHKind) {
									case FHkField:
										d[jen.Id(hef.Name)] = vars.GetObject().Dot(f.Name)
									case FHkTimestamp:
										d[jen.Id(hef.Name)] = jen.Qual("time", "Now").Params()
									case FHkUserID:
									case FHkUserName:
									case FHkSource:
									case FHkCustom:
									}
								}
							},
						)),
					))
				},
			)

	}
	return

}

//OnMethodHook implements GeneratorHookHolder
func (hg *HistoryGenerator) OnMethodHook(name HookType, mod HookModifier, m *Method, vars *GeneratorHookVars) (code *jen.Statement, order int) {
	return nil, 0

}
