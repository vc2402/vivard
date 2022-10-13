package vue

import (
	"fmt"
	"path"
	"text/template"

	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
)

type helper struct {
	templ   *template.Template
	e       *gen.Entity
	cg      *VueCLientGenerator
	outDir  string
	idField *gen.Field
	err     error
	ctx     helperContext
}

type annotationSet []*gen.Annotation
type fields []fieldDescriptor
type formTabs map[string]formTab

type helperContext struct {
	fields       fields
	width        string
	ann          annotationSet
	kind         string
	withRows     bool
	withTabs     bool
	useGrid      bool
	title        string
	tabs         formTabs
	components   map[string]componentDescriptor
	needSecurity bool
}

type formTab struct {
	ID    string
	label string
	order int
	roles string
}

type fieldDescriptor struct {
	fld      *gen.Field
	width    string
	w        int
	row      int
	ord      int
	title    string
	mask     string
	ann      annotationSet
	tab      string
	tabOrder int
	readonly bool
}
type componentDescriptor struct {
	name string
	path string
}

func (cg *VueCLientGenerator) getTypesPath(e *gen.Entity) (string, error) {
	fp, ok := e.Features.GetString(js.Features, js.FFilePath)
	if !ok {
		return "", fmt.Errorf("file path not set for %s", e.Name)
	}
	tn := path.Base(fp)
	ext := path.Ext(tn)
	if ext != "" {
		tn = tn[:len(tn)-len(ext)]
	}
	return path.Join("../types", tn), nil
}
func (cg *VueCLientGenerator) newHelper(name string, e *gen.Entity, outDir string) (*helper, error) {
	idf := e.GetIdField()
	typesPath, err := cg.getTypesPath(e)
	if err != nil {
		return nil, err
	}
	components := map[string]string{}
	customComponents := map[string]vcCustomComponentDescriptor{}
	funcs := template.FuncMap{
		"ShowInDialog": func(f *gen.Field) bool {
			if ignore, ok := f.Annotations.GetBoolAnnotation(vueDialogAnnotation, vueAnnotationIgnore); ok && ignore {
				return false
			}
			if ignore, ok := f.Annotations.GetBoolAnnotation(vueAnnotation, vueAnnotationIgnore); ok && ignore {
				return false
			}
			if ignore, ok := f.Features.GetBool(gen.FeaturesAPIKind, gen.FCIgnore); ok && ignore {
				return false
			}
			if f.Name == gen.ExtendableTypeDescriptorFieldName {
				return false
			}
			return true
		},
		"Label": func(f *gen.Field) string {
			name := f.Name
			if an, ok := f.Annotations[vueAnnotation]; ok {
				name = an.GetString(vcaLabel, name)
			}
			//TODO: make different label for tables?
			return name
		},
		"FieldName": func(f *gen.Field) string {
			return f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		},
		"AttrName": func(f *gen.Field) string { return cg.getJSAttrNameForDisplay(f, false, false) },
		"TableAttrName": func(f *gen.Field) string {
			if vfn := f.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATValue, ""); vfn != "" {
				return cg.getJSAttrForSubfield(f, vfn)
			}
			return cg.getJSAttrNameForDisplay(f, true, false)
		},
		"TableIconName": func(f *gen.Field) string { return cg.getJSAttrNameForDisplay(f, true, true) },
		"NeedIconForTable": func(f *gen.Field) bool {
			return f.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false)
		},
		"ShowInView": func(f *gen.Field) bool {
			return cg.getJSAttrNameForDisplay(f, false, false) != ""
		},
		"ShowInTable": func(f *gen.Field) bool {
			if _, ok := f.Features.GetField(gen.FeatureHistKind, gen.FHHistoryOf); ok || f.HasModifier(gen.AttrModifierAuxiliary) {
				return false
			}
			return !f.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vueAnnotationIgnore, false) &&
				f.Name != gen.ExtendableTypeDescriptorFieldName
		},
		"TypeName": func(e *gen.Entity) string {
			return e.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		},
		"FieldType": func(f *gen.Field) string {
			return f.FS(js.Features, js.FType)
			// return f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationType, "")
		},
		"GetQuery": func(e *gen.Entity) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationGet])
		},
		"SaveQuery": func(e *gen.Entity) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationSet])
		},
		"CreateQuery": func(e *gen.Entity) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationCreate])
		},
		"ListQuery": func(e *gen.Entity) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationList])
		},
		"ListQueryAttrs": func(e *gen.Entity) string {
			if e.FB(gen.FeatureDictKind, gen.FDQualified) {
				qt, _ := e.Features.GetEntity(gen.FeatureDictKind, gen.FDQualifierType)
				idfld := qt.GetIdField()
				return fmt.Sprintf("this.qualifier && [this.qualifier.%s]", idfld.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, ""))
			}
			return ""
		},
		"DictWithQualifier": func(e *gen.Entity) bool {
			return e.FB(gen.FeatureDictKind, gen.FDQualified)
		},
		"LookupForQualifier": func(e *gen.Entity) string {
			if e.FB(gen.FeatureDictKind, gen.FDQualified) {
				qt, _ := e.Features.GetEntity(gen.FeatureDictKind, gen.FDQualifierType)
				cmp := qt.FS(featureVueKind, fVKLookupComponent)
				if cmp != "" {
					components[cmp] = cmp
					return cmp
				}
			}
			return ""
		},
		"LookupQuery": func(e *gen.Entity) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationLookup])
		},
		"TypesFilePath": func(e *gen.Entity) string {
			return typesPath
		},
		//TODO: get title from annotations
		"Title": func(e *gen.Entity) string {
			if t, ok := e.Annotations.GetStringAnnotation(vueDialogAnnotation, vcaLabel); ok {
				return t
			}
			return fmt.Sprintf(`"%s"`, e.Name)
		},
		"IDType": func(e *gen.Entity) string {
			t, _ := e.Features.GetString(js.Features, js.FIDType)
			return t
		},
		"IDField": func(e *gen.Entity) (ret string) {
			if idf := e.GetIdField(); idf != nil {
				ret = idf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
			}
			return
		},
		"ItemText":  getTitleFieldName,
		"ItemValue": func() string { return idf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "") },
		"FormComponentType": func(f *gen.Field) string {
			if _, ok := f.Parent().Annotations[gen.AnnotationFind]; ok {
				fld, _ := f.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
				return fld.Type.Type
			}
			if f.Type.Map != nil {
				return "map"
			}
			if f.Type.Array != nil {
				return "array"
			}
			return f.Type.Type
		},
		"FormComponent": func(e *gen.Entity) string {
			cmp := e.FS(featureVueKind, fVKFormComponent)
			// path := e.FS(featureVueKind, fVKFormComponentPath)
			components[cmp] = cmp
			return cmp
		},
		"ViewComponent": func(e *gen.Entity) string {
			cmp := e.FS(featureVueKind, fVKViewComponent)
			components[cmp] = cmp
			return cmp
		},
		"DialogComponent": func(e *gen.Entity) string {
			cmp := e.FS(featureVueKind, fVKDialogComponent)
			components[cmp] = cmp
			return cmp
		},
		"DictEditComponent": func(e *gen.Entity, addToRequired bool) string {
			cmp := e.FS(featureVueKind, fVKDictEditComponent)
			if addToRequired {
				components[cmp] = cmp
			}
			return cmp
		},
		"ArrayAsLookup": func(f *gen.Field) bool {
			if f.Type.Array != nil {
				if gen.IsPrimitiveType(f.Type.Array.Type) {
					return false
				}
				if f, ok := cg.desc.FindType(f.Type.Array.Type); ok {
					return !f.Entity().HasModifier(gen.TypeModifierEmbeddable)
				}
			}
			return false
		},
		"ArrayAsList": func(f *gen.Field) bool {
			return false
		},
		"LookupComponent": func(f *gen.Field, addToRequired bool) string {
			tip := f.Type
			if _, ok := f.Parent().Annotations[gen.AnnotationFind]; ok {
				fld, _ := f.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
				tip = fld.Type
			}
			typename := tip.Type
			if tip.Array != nil {
				typename = tip.Array.Type
			}

			ud := f.FS(featureVueKind, fVKUseInDialog)
			var lc string
			switch ud {
			//TODO: get name and path from features
			case fVKUseInDialogLookup:
				lc = typename + "LookupComponent"
			case fVKUseInDialogForm:
				lc = typename + "Form"
			default:
				lc = typename + "LookupComponent"
			}
			if addToRequired {
				components[lc] = lc
			}

			return lc
		},
		"AppendToField": func(f *gen.Field) []string {
			if he, ok := f.Features.GetEntity(gen.FeatureHistKind, gen.FHHistoryEntity); ok {
				hc := he.FS(featureVueKind, fVKHistComponent)
				if hf, ok := f.Features.GetField(gen.FeatureHistKind, gen.FHHistoryField); ok {
					fn := hf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
					components[hc] = hc
					return []string{fmt.Sprintf("<%s v-if=\"value && value.%s\" :items=\"value.%s\"/>", hc, fn, fn)}
				}
			}
			return nil
		},
		"FieldWithAppend": func(f *gen.Field) bool {
			return f.FS(gen.FeatureHistKind, gen.FHHistoryEntityName) != ""
		},
		"Readonly": func(f ...*gen.Field) bool {
			if len(f) > 0 {
				return f[0].FB(gen.FeaturesCommonKind, gen.FCReadonly)
			}
			return e.FB(gen.FeaturesCommonKind, gen.FCReadonly)
		},
		"RequiredComponents":   func() map[string]string { return components },
		"AdditionalComponents": func() map[string]vcCustomComponentDescriptor { return customComponents },
		"IsID": func(f *gen.Field, auto bool) bool {
			if auto {
				return f.HasModifier(gen.AttrModifierIDAuto)
			}
			return f.IsIdField()
		},
		"CustomComponent": func(tip string) string {
			if imp, ok := cg.options.Components[tip]; ok {
				//if cmp, ok := vcCustomComponents[tip]; ok {
				customComponents[imp.Name] = vcCustomComponentDescriptor{imp.Name, imp.Import}
				return imp.Name
				//}
			}
			return "v-text-field"
		},
		"ConponentAddAttrs": func(f *gen.Field) string {
			switch f.Type.Type {
			case gen.TipDate:
				t := f.Annotations.GetStringAnnotationDef(vueAnnotation, vueAnnotationDisplayType, vueATDate)
				return fmt.Sprintf(":type=\"'%s'\"", t)
			}
			return ""
		},
		"SelfFormComponent": func() string {
			return e.FS(featureVueKind, fVKFormComponent)
		},
		"SelfFormComponentPath": func() string {
			return fmt.Sprintf("./%s.vue", e.FS(featureVueKind, fVKFormComponent))
		},
		"NotExported": func(t *gen.Entity) bool {
			level, ok := t.Features.GetString(gen.FeaturesAPIKind, gen.FAPILevel)
			return ok && level != gen.FAPILAll
		},
		"GUITableColor": func(f *gen.Field) string {
			c, _ := cg.getJSAttrColorForTable(f)
			return c
		},
		"GUITableType": func(f *gen.Field) string {
			fromAnnotations := func(ann gen.Annotations, def string) string {
				return ann.GetStringAnnotationDef(vueTableAnnotation, vueAnnotationDisplayType,
					ann.GetStringAnnotationDef(vueAnnotation, vueAnnotationDisplayType, def),
				)
			}
			if cust := f.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATCustom, ""); cust != "" {
				return "custom"
			}
			if f.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false) {
				return "icon"
			}
			if !f.Type.Complex {
				return fromAnnotations(f.Annotations, f.Type.Type)
			}
			if fromAnn := fromAnnotations(f.Annotations, ""); fromAnn != "" {
				// allow to override default
				return fromAnn
			}
			t := f.Type.Type
			if f.Type.Array != nil {
				t = f.Type.Array.Type
			}
			if e, ok := cg.desc.FindType(t); ok {
				if e.Entity().HasModifier(gen.TypeModifierEmbeddable) {
					if ret := fromAnnotations(e.Entity().Annotations, ""); ret != "" {
						return ret
					}
				}
				if ff := getTitleField(e.Entity()); ff != nil {
					return ff.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueAnnotationDisplayType,
						ff.Annotations.GetStringAnnotationDef(vueAnnotation, vueAnnotationDisplayType,
							ff.Type.Type,
						),
					)
				}
			}
			return "string"
		},
		"GUITableComponent": func(f *gen.Field) string {
			return f.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATCustom, "")
		},
		"GUITableTooltip": func(f *gen.Field) string {
			ret := ""
			ttfn := f.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATTooltip, "")
			if ttfn != "" {
				ret = cg.getJSAttrForSubfield(f, ttfn)
			}
			return ret
		},
		"EditableInTable": func(f *gen.Field) string {
			if f.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vueATEditable, false) {
				return "true"
			}
			return "false"
		},
		"CanBeMultiple": func(e *gen.Entity) bool {
			// if refsManyToMany, ok := e.Features.GetBool(gen.FeaturesCommonKind, gen.FCRefsAsManyToMany); ok && refsManyToMany {
			// 	return true
			// }
			// return false
			return e.IsDictionary()
		},
		"LookupAttrs": func(f *gen.Field) (ret string) {
			if _, itsManyToMany := f.Features.GetEntity(gen.FeaturesCommonKind, gen.FCManyToManyType); itsManyToMany {
				ret = "multiple"
			} else if f.Parent().Annotations[gen.AnnotationFind] != nil {
				ret = ":returnObject='false' hideAdd"
				if f.Type.Array != nil {
					ret += " multiple"
				}
			}
			if qf, ok := f.Features.GetField(gen.FeatureDictKind, gen.FDQualifiedBy); ok {
				ret += fmt.Sprintf(" :qualifier=\"value.%s\"", qf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, ""))
			}
			return
		},
		"LookupWithAdd": func(e *gen.Entity) bool {
			return !e.FB(gen.FeaturesCommonKind, gen.FCReadonly)
		},
		"GetFields": func(e *gen.Entity) []*gen.Field {
			return e.GetFields(true, true)
		},
		"FiltersImports": func() string {
			return `import { DateTimeFilter} from '@/filters/dateTimeFilter';
import {RoundNumber} from '@/filters/numberFilter';
`
		},
		"Filter": func(name string, withPipe ...bool) string {
			var ret string
			switch name {
			case "date":
				ret = "DateTimeFilter"
			case "number", "int", "float":
				ret = "RoundNumber"
			default:
				cg.desc.AddWarning(fmt.Sprintf("vue: undefined Filter requested: %s", name))
				return ""
			}
			prefix := ""
			if len(withPipe) > 0 && withPipe[0] {
				prefix = "|"
			}
			return fmt.Sprintf("%s%s", prefix, ret)
		},
		"ApolloClient": func() string {
			return cg.options.ApolloClientVar
		},
		"InstanceGenerator": func(e *gen.Entity) string {
			return e.FS(js.Features, js.FInstanceGenerator) + "()"
		},
		"InstanceGeneratorName": func(e *gen.Entity) string {
			return e.FS(js.Features, js.FInstanceGenerator)
		},
		"InputAttrs": func(f *gen.Field) string {
			return ""
		},
		"DialogWidth": func() string {
			return "$vuetify.breakpoint.lgAndUp ? '60vw' : '80vw'"
		},
		// "RequiresInputField": func(f *gen.Field) bool {
		// 	return !f.Annotations.GetBoolAnnotationDef(vueFormAnnotation, vueAnnotationIgnore, false)
		// },
	}

	th := &helper{templ: template.New(name), cg: cg, e: e, idField: idf, outDir: outDir}
	th.templ.Funcs(funcs)
	return th, nil
}

func (th *helper) parse(str string) *helper {
	if th.err != nil {
		return th
	}
	th.templ, th.err = th.templ.Parse(str)
	return th
}

func getTitleFieldName(e *gen.Entity) string {
	f := getTitleField(e)
	if f != nil {
		return f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
	} else {
		return ""
	}
}

func getTitleField(e *gen.Entity) *gen.Field {
	fld, ok := e.Annotations.GetStringAnnotation(js.Annotation, js.AnnotationTitle)
	if ok {
		f := e.GetField(fld)
		if f != nil {
			return f
		}
	}
	idf := e.GetIdField()
	return idf
}

func getTabs(e *gen.Entity) (ret formTabs, def string) {
	if ann := e.GetAnnotation(vueTabSet); ann != nil {
		ret = map[string]formTab{}
		for i, a := range ann.Values {
			tab := formTab{ID: a.Key, order: i, label: a.Key}
			if def == "" {
				def = a.Key
			}
			if a.Value != nil && a.Value.String != nil {
				tab.label = *a.Value.String
			} else {
				if ta := e.GetAnnotation(vueTab, a.Key); ta != nil {
					tab.order = ta.GetInt(vcaOrder, i)
					tab.label = ta.GetString(vcaLabel, tab.label)
					tab.roles = ta.GetString(vcaRoles, "")
					if d := ta.GetBool(vcaDefault, false); d {
						def = tab.ID
					}
				}
			}
			ret[tab.ID] = tab
		}
	}
	return
}

func (ft formTabs) getOrder(id string) int {
	if ft != nil {
		return ft[id].order
	}
	return 0
}

func (ft formTabs) getLabel(id string) string {
	if ft != nil && ft[id].label != "" {
		return ft[id].label
	}
	return id
}

func (as annotationSet) getString(name string) (ret string, ok bool) {
	for _, a := range as {
		if ret, ok = a.GetStringTag(name); ok {
			return
		}
	}
	return "", false
}

func (as annotationSet) getInt(name string) (ret int, ok bool) {
	for _, a := range as {
		if ret, ok = a.GetIntTag(name); ok {
			return
		}
	}
	return 0, false
}

func (as annotationSet) getBool(name string) (ret bool, ok bool) {
	for _, a := range as {
		if ret, ok = a.GetBoolTag(name); ok {
			return
		}
	}
	return false, false
}

func (as annotationSet) getStringDef(name string, def string) string {
	if v, ok := as.getString(name); ok {
		return v
	}
	return def
}

func (as annotationSet) getIntDef(name string, def int) int {
	if v, ok := as.getInt(name); ok {
		return v
	}
	return def
}

func (as annotationSet) getBoolDef(name string, def bool) bool {
	if v, ok := as.getBool(name); ok {
		return v
	}
	return def
}
