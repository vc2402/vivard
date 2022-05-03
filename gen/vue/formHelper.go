package vue

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
)

func (cg *VueCLientGenerator) newFormHelper(name string, e *gen.Entity, annName string, annSpec string, outDir string) (*helper, error) {
	skipTabs := false

	idf := e.GetIdField()
	typesPath, err := cg.getTypesPath(e)
	if err != nil {
		return nil, err
	}
	ctx := helperContext{}
	a := e.GetAnnotation(annName, annSpec)
	if a != nil {
		ctx.ann = []*gen.Annotation{a}
	}
	a = e.GetAnnotation(annName)
	if a != nil {
		ctx.ann = append(ctx.ann, a)
	}
	a = e.GetAnnotation(vueAnnotation)
	if a != nil {
		ctx.ann = append(ctx.ann, a)
	}
	ctx.components = map[string]componentDescriptor{}
	if annSpec != "" {
		name := fmt.Sprintf("Form%s", annSpec)
		ctx.components["form"] = componentDescriptor{name: name, path: cg.getPathForComponent(e, name+".vue")}
		name = fmt.Sprintf("Dialog%s", annSpec)
		ctx.components["dialog"] = componentDescriptor{name: name, path: cg.getPathForComponent(e, name+".vue")}
		skipTabs = ctx.ann.getBoolDef(vueSkipTabs, false)
	}

	ctx.width = ctx.ann.getStringDef(vcaWidth, "")

	ctx.kind = ctx.ann.getStringDef(vcaLayout, vcalFlex)
	ctx.title = ctx.ann.getStringDef(vcaLabel, e.Name)

	var dt string
	if !skipTabs {
		ctx.tabs, dt = getTabs(e)
	}

	maxRow := -1
	hasStringWidth := false
	hasIntWidth := false
	for _, f := range e.GetFields(true, true) {
		as := annotationSet{}
		if f.FB(gen.FeaturesAPIKind, gen.FCIgnore) {
			continue
		}
		if _, ok := f.Features.Get(gen.FeatureHistKind, gen.FHHistoryOf); ok {
			continue
		}
		a = f.Annotations.Find(annName, annSpec)
		if a != nil {
			as = annotationSet{a}
		}
		a = f.Annotations[annName]
		if a != nil {
			as = append(as, a)
		}
		a = f.Annotations[vueAnnotation]
		if a != nil {
			as = append(as, a)
		}
		if as.getBoolDef(vueAnnotationIgnore, false) {
			continue
		}
		if annSpec != "" && !as.getBoolDef(vueAnnotationUse, false) && !as.getBoolDef(vueAnnotationReadonly, false) {
			//TODO make possibility to use one of ignore or use for specific forms
			continue
		}
		fd := fieldDescriptor{
			fld:      f,
			ann:      as,
			mask:     as.getStringDef(vcaMask, ""),
			ord:      as.getIntDef(vcaOrder, 1000),
			title:    as.getStringDef(vcaLabel, f.Name),
			width:    as.getStringDef(vcaWidth, ""),
			w:        as.getIntDef(vcaWidth, 0),
			readonly: as.getBoolDef(vueAnnotationReadonly, f.FB(gen.FeaturesCommonKind, gen.FCReadonly)),
		}

		if fd.width != "" {
			hasStringWidth = true
		}
		if fd.w != 0 {
			hasIntWidth = true
		}
		if r, ok := as.getInt(vcaRow); ok {
			ctx.withRows = true
			fd.row = r
			if maxRow < r {
				maxRow = r
			}
		} else {
			fd.row = -1
		}

		if !skipTabs {
			if t, ok := as.getString(vcaTab); ok {
				ctx.withTabs = true
				fd.tab = t
				fd.tabOrder = ctx.tabs.getOrder(t)
			} else {
				fd.tab = dt
				fd.tabOrder = ctx.tabs.getOrder(dt)
			}
		}
		ctx.fields = append(ctx.fields, fd)
	}
	if hasStringWidth && hasIntWidth {
		cg.desc.AddWarning(fmt.Sprintf("at %v: %s: both string and int 'width' values found in vue annotations", e.Pos, e.Name))
	}
	if hasIntWidth {
		ctx.useGrid = true
	}
	if ctx.withRows || hasIntWidth {
		for i := range ctx.fields {
			if ctx.withRows && ctx.fields[i].row == -1 {
				ctx.fields[i].row = maxRow + 1
			}
			if hasIntWidth && ctx.fields[i].w == 0 {
				if w, err := strconv.Atoi(ctx.width); err == nil && w <= 12 {
					ctx.fields[i].w = w
				} /*else {
					ctx.fields[i].w = 3
				}*/
			}
		}
	}
	sort.Sort(&ctx.fields)
	th := &helper{templ: template.New(name), cg: cg, e: e, idField: idf, outDir: outDir, ctx: ctx}

	components := map[string]string{}
	customComponents := map[string]vcCustomComponentDescriptor{}
	typesFromTS := []string{}
	funcs := template.FuncMap{
		"Label": func(f fieldDescriptor) string { return f.title },
		"Rows": func(it interface{}) [][]fieldDescriptor {
			switch v := it.(type) {
			case [][]fieldDescriptor:
				return v
			case helperContext, *helper:
				rows, _ := ctx.getRows(0)
				return rows
			}
			return [][]fieldDescriptor{}
		},
		"GetFields": func(h *helper) []fieldDescriptor {
			return h.ctx.fields
		},
		"Tabs": func() [][][]fieldDescriptor {
			tabs, _ := ctx.getTabs(0)
			return tabs
		},
		"TabLable": func(tab [][]fieldDescriptor) string {
			if len(tab) > 0 && len(tab[0]) > 0 {
				return ctx.tabs.getLabel(tab[0][0].tab)
			}
			return "-invalid-"
		},
		"ShowInDialog": func(f fieldDescriptor) bool {
			// in this kind of helper only needed fields are in range
			return true
		},
		"GridColAttrs": func(f fieldDescriptor) string {
			if f.w > 0 {
				return fmt.Sprintf("cols='%d'", f.w)
			} else {
				return ""
			}
		},
		"FieldName": func(f fieldDescriptor) string {
			return f.fld.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		},
		"AttrName": func(f fieldDescriptor) string { return cg.getJSAttrNameForDisplay(f.fld, false, false) },
		"TableAttrName": func(f fieldDescriptor) string {
			return cg.getJSAttrNameForDisplay(f.fld, true, false)
		},
		"TableIconName": func(f fieldDescriptor) string { return cg.getJSAttrNameForDisplay(f.fld, true, true) },
		"NeedIconForTable": func(f fieldDescriptor) bool {
			return f.fld.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false)
		},
		"Name": func() string {
			return e.Name
		},
		"TypeName": func(arg ...interface{}) string {
			return e.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		},
		"FieldType": func(f fieldDescriptor) string {
			return f.fld.FS(js.Features, js.FType)
			// return f.fld.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationType, "")
		},
		"GetQuery": func(arg ...interface{}) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationGet])
		},
		"SaveQuery": func(arg ...interface{}) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationSet])
		},
		"CreateQuery": func(arg ...interface{}) string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationCreate])
		},
		"ListQuery": func() string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationList])
		},
		// "ListQueryAttrs": func() string {
		// 	if e.FB(gen.FeatureDictKind, gen.FDQualified) {
		// 		return "this.qualifier"
		// 	}
		// 	return ""
		// },
		"DictWithQualifier": func(hlp *helper) bool {
			return e.FB(gen.FeatureDictKind, gen.FDQualified)
		},
		"DictQualifierTitle": func() string {
			if qf, ok := e.Features.GetField(gen.FeatureDictKind, gen.FDQualifier); ok {
				for _, fd := range th.ctx.fields {
					if fd.fld == qf {
						return fd.title
					}
				}
				return qf.Annotations.GetStringAnnotationDef(vueAnnotation, vcaLabel, qf.Name)
			}
			return ""
		},
		"LookupForQualifier": func(hlp *helper) string {
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
		"LookupQuery": func() string {
			return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationLookup])
		},
		"TypesFilePath": func(arg ...interface{}) string {
			return typesPath
		},
		"Title": func(arg ...interface{}) string { return ctx.title },
		"IDType": func(arg ...interface{}) string {
			t, _ := e.Features.GetString(js.Features, js.FIDType)
			return t
		},
		"IDField": func() (ret string) {
			if idf := e.GetIdField(); idf != nil {
				ret = idf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
			}
			return
		},
		"FormComponentType": func(f fieldDescriptor) string {
			if _, ok := f.fld.Parent().Annotations[gen.AnnotationFind]; ok {
				fld, _ := f.fld.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
				if fld != nil {
					return fld.Type.Type
				}
			}
			if f.fld.Type.Map != nil {
				return "map"
			}
			if f.fld.Type.Array != nil {
				return "array"
			}
			return f.fld.Type.Type
		},
		"FormComponent": func() string {
			cmp := e.FS(featureVueKind, fVKFormComponent)
			// path := e.FS(featureVueKind, fVKFormComponentPath)
			components[cmp] = cmp
			return cmp
		},
		"ViewComponent": func() string {
			cmp := e.FS(featureVueKind, fVKViewComponent)
			components[cmp] = cmp
			return cmp
		},
		"DialogComponent": func(hlp *helper) string {
			cmp := e.FS(featureVueKind, fVKDialogComponent)
			components[cmp] = cmp
			return cmp
		},
		"DictEditComponent": func(hlp *helper, addToRequired bool) string {
			cmp := e.FS(featureVueKind, fVKDictEditComponent)
			if addToRequired {
				components[cmp] = cmp
			}
			return cmp
		},
		"ArrayAsLookup": func(f fieldDescriptor) bool {
			if f.fld.Type.Array != nil {
				if gen.IsPrimitiveType(f.fld.Type.Array.Type) {
					return false
				}
				if f, ok := cg.desc.FindType(f.fld.Type.Array.Type); ok {
					return !f.Entity().HasModifier(gen.TypeModifierEmbeddable)
				}
			}
			return false
		},
		"ArrayAsList": func(f fieldDescriptor) bool {
			if f.fld.Type.Array != nil {
				if gen.IsPrimitiveType(f.fld.Type.Array.Type) {
					return false
				}
				if e, ok := cg.desc.FindType(f.fld.Type.Array.Type); ok {
					if e.Entity().HasModifier(gen.TypeModifierEmbeddable) {
						name := e.Entity().FS(js.Features, js.FInstanceGenerator)
						typesFromTS = append(typesFromTS, name)
						return true
					}
				}
			}
			return false
		},
		"ArrayAsChips": func(f fieldDescriptor) bool {
			if f.fld.Type.Array != nil {
				return f.fld.Type.Array.Type == gen.TipString
			}
			return false
		},
		"LookupComponent": func(f fieldDescriptor, addToRequired bool) string {
			if cus := f.fld.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATCustom, ""); cus != "" {
				return cus
			}

			tip := f.fld.Type
			if _, ok := f.fld.Parent().Annotations[gen.AnnotationFind]; ok {
				fld, _ := f.fld.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
				tip = fld.Type
			}
			typename := tip.Type
			if tip.Array != nil {
				typename = tip.Array.Type
			}
			var lc string
			if t, ok := e.Pckg.FindType(typename); ok {
				if t.Entity().HasModifier(gen.TypeModifierEmbeddable) {
					// if f.fld.FB(gen.FeaturesCommonKind, gen.FCReadonly) {
					// 	lc = t.Entity().FS(featureVueKind, fVKViewComponent)
					// } else {
					lc = t.Entity().FS(featureVueKind, fVKFormComponent)
					// }
				} else {
					typename = t.Entity().Name
				}
			}

			if lc == "" {
				ud := f.fld.FS(featureVueKind, fVKUseInDialog)

				switch ud {
				//TODO: get name and path from features
				case fVKUseInDialogLookup:
					lc = typename + "LookupComponent"
				case fVKUseInDialogForm:
					lc = typename + "Form"
				default:
					lc = typename + "LookupComponent"
				}
			}
			if addToRequired {
				components[lc] = lc
			}

			return lc
		},
		"LookupWithQualifier": func(f fieldDescriptor) bool {
			_, ok := f.fld.Features.GetField(gen.FeatureDictKind, gen.FDQualifiedBy)
			return ok
		},
		"AppendToField": func(f fieldDescriptor) []string {
			if he, ok := f.fld.Features.GetEntity(gen.FeatureHistKind, gen.FHHistoryEntity); ok {
				hc := he.FS(featureVueKind, fVKHistComponent)
				if hf, ok := f.fld.Features.GetField(gen.FeatureHistKind, gen.FHHistoryField); ok {
					fn := hf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
					components[hc] = hc
					return []string{fmt.Sprintf("<%s v-if=\"value && value.%s\" :items=\"value.%s\"/>", hc, fn, fn)}
				}
			}
			return nil
		},
		"FieldWithAppend": func(f fieldDescriptor) bool {
			return f.fld.FS(gen.FeatureHistKind, gen.FHHistoryEntityName) != ""
		},
		"Readonly": func(f ...fieldDescriptor) bool {
			if len(f) > 0 {
				return f[0].readonly
			}
			return e.FB(gen.FeaturesCommonKind, gen.FCReadonly)
		},
		"FieldAttrs": func(f fieldDescriptor) string {
			if iff, ok := f.fld.Annotations.GetStringAnnotation(vueAnnotation, vcaIf); ok {
				return fmt.Sprintf("v-if='%s'", iff)
			}
			return ""
		},
		"InputAttrs": func(f fieldDescriptor) string {
			if f.mask != "" {
				return fmt.Sprintf("v-mask=\"'%s'\"", f.mask)
			}
			return ""
		},
		"RequiredComponents": func() map[string]string {
			return components
		},
		"AdditionalComponents": func() map[string]vcCustomComponentDescriptor { return customComponents },
		"IsID": func(f fieldDescriptor, auto bool) bool {
			if auto {
				return f.fld.HasModifier(gen.AttrModifierIDAuto)
			}
			return f.fld.IsIdField()
		},
		"CustomComponent": func(param interface{}) string {
			switch v := param.(type) {
			case string:
				//registered well-known component
				if imp, ok := cg.options.Components[v]; ok {
					customComponents[imp.Name] = vcCustomComponentDescriptor{imp.Name, imp.Import}
					return imp.Name
					//}
				}
				return "v-text-field"
			case fieldDescriptor:
				return v.ann.getStringDef(vueATCustom, "")
			}
			return ""
		},
		"ConponentAddAttrs": func(f fieldDescriptor) string {
			switch f.fld.Type.Type {
			case gen.TipDate:
				t := f.fld.Annotations.GetStringAnnotationDef(vueAnnotation, vueAnnotationDisplayType, vueATISODate)
				return fmt.Sprintf(":type=\"'%s'\"", t)
			}
			return ""
		},
		"NotExported": func(t *gen.Entity) bool {
			level, ok := t.Features.GetString(gen.FeaturesAPIKind, gen.FAPILevel)
			return ok && level != gen.FAPILAll
		},
		"GUITableColor": func(f fieldDescriptor) string {
			c, _ := cg.getJSAttrColorForTable(f.fld)
			return c
		},
		"GUITableType": func(f fieldDescriptor) string {
			fromAnnotations := func(ann gen.Annotations, def string) string {
				return ann.GetStringAnnotationDef(vueTableAnnotation, vueAnnotationDisplayType,
					ann.GetStringAnnotationDef(vueAnnotation, vueAnnotationDisplayType, def),
				)
			}
			if cust := f.fld.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATCustom, ""); cust != "" {
				return "custom"
			}
			if f.fld.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false) {
				return "icon"
			}
			if !f.fld.Type.Complex {
				return fromAnnotations(f.fld.Annotations, f.fld.Type.Type)
			}
			t := f.fld.Type.Type
			if f.fld.Type.Array != nil {
				t = f.fld.Type.Array.Type
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
		"GUITableComponent": func(f fieldDescriptor) string {
			return f.fld.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATCustom, "")
		},
		"CanBeMultiple": func() bool {
			// if refsManyToMany, ok := e.Features.GetBool(gen.FeaturesCommonKind, gen.FCRefsAsManyToMany); ok && refsManyToMany {
			// 	return true
			// }
			// return false
			return e.IsDictionary()
		},
		"LookupAttrs": func(f fieldDescriptor) (ret string) {
			if _, itsManyToMany := f.fld.Features.GetEntity(gen.FeaturesCommonKind, gen.FCManyToManyType); itsManyToMany {
				ret = "multiple"
			} else if f.fld.Parent().Annotations[gen.AnnotationFind] != nil {
				ret = ":returnObject='false' hideAdd"
				if f.fld.Type.Array != nil {
					ret += " multiple"
				}
			}
			if qf, ok := f.fld.Features.GetField(gen.FeatureDictKind, gen.FDQualifiedBy); ok {
				ret += fmt.Sprintf(" :qualifier=\"value.%s\"", qf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, ""))
			}
			return
		},
		"ApolloClient": func() string {
			return cg.options.ApolloClientVar
		},
		"InstanceGenerator": func() string {
			return e.FS(js.Features, js.FInstanceGenerator) + "()"
		},
		"InstanceGeneratorName": func() string {
			return e.FS(js.Features, js.FInstanceGenerator)
		},
		"InstanceGeneratorForField": func(f fieldDescriptor) string {
			t := f.fld.Type.Type
			if f.fld.Type.Array != nil {
				t = f.fld.Type.Array.Type
			}
			if e, ok := e.Pckg.FindType(t); ok {
				name := e.Entity().FS(js.Features, js.FInstanceGenerator)
				// it is too late to do it here... will do earlier
				//typesFromTS = append(typesFromTS, name)
				return name + "()"
			}
			return ""
		},
		"TypesFromTS": func() string {
			return strings.Join(typesFromTS, ", ")
		},
		"SelfFormComponent": func() string {
			if cd, ok := th.ctx.getComponentDescriptor("form"); ok {
				return cd.name
			}
			return e.FS(featureVueKind, fVKFormComponent)
		},
		"SelfFormComponentPath": func() string {
			if cd, ok := th.ctx.getComponentDescriptor("form"); ok {
				return cd.path
			}
			return fmt.Sprintf("./%s.vue", e.FS(featureVueKind, fVKFormComponent))
		},
		"DialogWidth": func() string {
			if th.ctx.width != "" {
				return th.ctx.width
			}
			return "$vuetify.breakpoint.lgAndUp ? '60vw' : '80vw'"
		},
	}
	th.templ.Funcs(funcs)
	return th, nil
}

func (f fields) Len() int { return len(f) }
func (f fields) Less(i, j int) bool {
	return f[i].tabOrder < f[j].tabOrder ||
		f[i].tabOrder == f[j].tabOrder &&
			(f[i].row < f[j].row || f[i].row == f[j].row && f[i].ord < f[j].ord)
}
func (f fields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (ctx helperContext) getTabs(from int) (tabs [][][]fieldDescriptor, next int) {
	var rows [][]fieldDescriptor
	for {
		rows, next = ctx.getRows(from)
		if len(rows) > 0 {
			tabs = append(tabs, rows)
		}
		if next == len(ctx.fields) {
			break
		}
		from = next
	}
	return
}

func (ctx helperContext) getRows(from int) (rows [][]fieldDescriptor, next int) {
	current := ctx.fields[from].tab
	var row []fieldDescriptor
	for {
		row, next = ctx.getRow(from)
		if len(row) > 0 {
			rows = append(rows, row)
		}
		if next == len(ctx.fields) || current != ctx.fields[next].tab {
			break
		}
		from = next
	}
	return
}

func (ctx helperContext) getRow(from int) (row []fieldDescriptor, next int) {
	curRow := ctx.fields[from].row
	curTab := ctx.fields[from].tab
	for next = from; next < len(ctx.fields); next++ {
		if curRow != ctx.fields[next].row || curTab != ctx.fields[next].tab {
			return
		} else {
			row = append(row, ctx.fields[next])
		}
	}
	return
}

func (ctx helperContext) getComponentDescriptor(name string) (cd componentDescriptor, ok bool) {
	cd, ok = ctx.components[name]
	return
}

func (ctx helperContext) setComponentDescriptor(name string, componentName string, path string) {
	ctx.components[name] = componentDescriptor{name: componentName, path: path}
}
