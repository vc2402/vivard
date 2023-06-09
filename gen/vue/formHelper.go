package vue

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
)

func (cg *VueCLientGenerator) newFormHelper(name string, e *gen.Entity, annName string, annSpec string, outDir string) (*helper, error) {
	skipTabs := false
	skipRows := false
	if annName == vueTableAnnotation {
		skipTabs = true
		skipRows = true
	}
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
		name := fmt.Sprintf("%sForm%s", e.Name, annSpec)
		p := cg.getPathForComponent(e, name+".vue")
		rp := cg.pathToRelative(cg.getOutputDirForEntity(e), p)
		ctx.components["form"] = componentDescriptor{name: name, path: name + ".vue", relPath: rp}
		name = fmt.Sprintf("%sDialog%s", e.Name, annSpec)
		p = cg.getPathForComponent(e, name+".vue")
		rp = cg.pathToRelative(cg.getOutputDirForEntity(e), p)
		ctx.components["dialog"] = componentDescriptor{name: name, path: name + ".vue", relPath: rp}
		skipTabs = ctx.ann.getBoolDef(vueSkipTabs, false)
	}

	ctx.width = ctx.ann.getStringDef(vcaWidth, "")

	ctx.kind = ctx.ann.getStringDef(vcaLayout, vcalFlex)
	ctx.title = ctx.ann.getStringDef(vcaLabel, e.Name)

	var dt string
	if !skipTabs {
		ctx.tabs, dt = getTabs(e)
		for _, t := range ctx.tabs {
			if t.resource != "" {
				ctx.needResourceSecurity = true
				break
			} else if t.roles != "" {
				ctx.needSecurity = true
				break
			}
		}
	}

	maxRow := -1
	hasStringWidth := false
	hasIntWidth := false
	for idx, f := range e.GetFields(true, true) {
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
			ord:      as.getIntDef(vcaOrder, 1000*idx),
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
		if r, ok := as.getInt(vcaRow); ok && !skipRows {
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
	th := &helper{templ: template.New(name), cg: cg, e: e, idField: idf, outDir: outDir, ctx: ctx, components: map[string]vcComponentDescriptor{}}

	//components := map[string]vcComponentDescriptor{}
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
		"HideAddForLookup": func(f fieldDescriptor) bool {
			return f.fld.Annotations.GetBoolAnnotationDef(vueLookupAnnotation, vcaReadonly, false) ||
				f.fld.HasModifier(gen.AttrModifierEmbeddedRef) ||
				f.fld.FB(gen.GQLFeatures, gen.GQLFIDOnly)
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
		"TabID": func(tab [][]fieldDescriptor) string {
			if len(tab) > 0 && len(tab[0]) > 0 {
				return tab[0][0].tab
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
			if vfn := f.fld.Annotations.GetStringAnnotationDef(vueTableAnnotation, vueATValue, ""); vfn != "" {
				return cg.getJSAttrForSubfield(f.fld, vfn)
			}
			return cg.getJSAttrNameForDisplay(f.fld, true, false)
		},
		"TableIconName": func(f fieldDescriptor) string { return cg.getJSAttrNameForDisplay(f.fld, true, true) },
		"NeedIconForTable": func(f fieldDescriptor) bool {
			return f.fld.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false) ||
				f.fld.Annotations.GetBoolAnnotationDef(js.Annotation, js.AnnotationIcon, false)
		},
		"IsIcon": func(f fieldDescriptor) bool {
			return f.fld.Annotations.GetBoolAnnotationDef(js.Annotation, js.AnnotationIcon, false)
		},
		"Name": func() string {
			return e.Name
		},
		"ShowInTable": func(f fieldDescriptor) bool {
			if _, ok := f.fld.Features.GetField(gen.FeatureHistKind, gen.FHHistoryOf); ok || f.fld.HasModifier(gen.AttrModifierAuxiliary) {
				return false
			}
			return !f.fld.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vueAnnotationIgnore, false) &&
				f.fld.Name != gen.ExtendableTypeDescriptorFieldName
		},
		"TypeName": func(arg ...interface{}) string {
			return e.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		},
		"FieldType": func(f fieldDescriptor) string {
			return f.fld.FS(js.Features, js.FType)
			// return f.fld.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationType, "")
		},
		"ShowInView": func(f fieldDescriptor) bool {
			return cg.getJSAttrNameForDisplay(f.fld, false, false) != ""
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
		"LookupWithAdd": func() bool {
			return !e.FB(gen.FeaturesCommonKind, gen.FCReadonly) &&
				!e.Annotations.GetBoolAnnotationDef(vueLookupAnnotation, vcaReadonly, false)
		},
		"ListQueryAttrs": func() string {
			if e.FB(gen.FeatureDictKind, gen.FDQualified) {
				qt, _ := e.Features.GetEntity(gen.FeatureDictKind, gen.FDQualifierType)
				idfld := qt.GetIdField()
				return fmt.Sprintf("this.qualifier && [this.qualifier.%s]", idfld.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, ""))
			}
			return ""
		},
		"IsQualifierFilled": func() string {
			if e.FB(gen.FeatureDictKind, gen.FDQualified) {
				qt, _ := e.Features.GetEntity(gen.FeatureDictKind, gen.FDQualifierType)
				idfld := qt.GetIdField()
				return fmt.Sprintf("this.qualifier && this.qualifier.%s", idfld.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, ""))
			}
			return ""

		},
		"GUITableTooltip": func(f fieldDescriptor) string {
			ret := ""
			ttfn := f.ann.getStringDef(vueATTooltip, "")
			if ttfn != "" {
				ret = cg.getJSAttrForSubfield(f.fld, ttfn)
			}
			return ret
		},
		"GetQuery": func(arg ...interface{}) string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationGet)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationGet])
		},
		"SaveQuery": func(arg ...interface{}) string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationSet)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationSet])
		},
		"CreateQuery": func(arg ...interface{}) string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationCreate)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationCreate])
		},
		"DeleteQuery": func(arg ...interface{}) string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationDelete)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationDelete])
		},
		"ListQuery": func() string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationList)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationList])
		},
		"LookupQuery": func() string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationLookup)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationLookup])
		},
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
				p := qt.FS(featureVueKind, fVKLookupComponentPath)
				if cmp != "" && p != "" {
					hlp.addComponent(cmp, p, qt)
					return cmp
				}
			}
			return ""
		},
		"TypesFilePath": func(arg ...interface{}) string {
			return typesPath
		},
		"Title": func(arg ...interface{}) string {
			return fmt.Sprintf("`%s`", ctx.title)
		},
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
				var fld *gen.Field
				if f, ok := f.fld.Features.Get(gen.FeaturesAPIKind, gen.FAPIFindForEmbedded); ok {
					fields := f.([]*gen.Field)
					fld = fields[len(fields)-1]
				}
				if fld == nil {
					fld, _ = f.fld.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
				}
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
			if col, ok := f.ann.getBool(vueATColorAttr); ok && col {
				return "color"
			}
			if col, ok := f.fld.Annotations.GetBoolAnnotation(js.Annotation, js.AnnotationColor); ok && col {
				return "color"
			}
			if f.ann.getBoolDef(vcaTextArea, false) {
				return "text-area"
			}
			if _, ok := f.ann.getInt(vcaTextArea); ok {
				return "text-area"
			}
			return f.fld.Type.Type
		},
		"TextAreaRows": func(f fieldDescriptor) int {
			return f.ann.getIntDef(vcaTextArea, 2)
		},
		"FormComponent": func() string {
			cmp := e.FS(featureVueKind, fVKFormComponent)
			p := e.FS(featureVueKind, fVKFormComponentPath)
			th.addComponent(cmp, p, e)
			return cmp
		},
		"ViewComponent": func() string {
			cmp := e.FS(featureVueKind, fVKViewComponent)
			p := e.FS(featureVueKind, fVKViewComponentPath)
			th.addComponent(cmp, p, e)
			return cmp
		},
		"DialogComponent": func(hlp *helper) string {
			cmp := e.FS(featureVueKind, fVKDialogComponent)
			p := e.FS(featureVueKind, fVKDialogComponentPath)
			th.addComponent(cmp, p, e)
			return cmp
		},
		"DictEditComponent": func(hlp *helper, addToRequired bool) string {
			cmp := e.FS(featureVueKind, fVKDictEditComponent)
			if addToRequired {
				p := e.FS(featureVueKind, fVKDictEditComponentPath)
				th.addComponent(cmp, p, e)
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
			if cus := f.fld.Annotations.GetStringAnnotationDef(vueLookupAnnotation, vueATCustom, ""); cus != "" {
				return cus
			}

			tip := f.fld.Type
			if _, ok := f.fld.Parent().Annotations[gen.AnnotationFind]; ok {
				var fld *gen.Field
				if f, ok := f.fld.Features.Get(gen.FeaturesAPIKind, gen.FAPIFindForEmbedded); ok {
					fields := f.([]*gen.Field)
					fld = fields[len(fields)-1]
				}
				if fld == nil {
					fld, _ = f.fld.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
				}
				tip = fld.Type
			}
			for tip.Array != nil {
				tip = tip.Array
			}
			typename := tip.Type
			var lc, lcp string
			if t, ok := e.Pckg.FindType(typename); ok {
				if t.Entity().HasModifier(gen.TypeModifierEmbeddable) {
					lc = t.Entity().FS(featureVueKind, fVKFormComponent)
					lcp = t.Entity().FS(featureVueKind, fVKFormComponentPath)
				}
				if lc == "" {
					ud := f.fld.FS(featureVueKind, fVKUseInDialog)

					switch ud {
					case fVKUseInDialogLookup:
						lc = t.Entity().FS(featureVueKind, fVKLookupComponent)
						lcp = t.Entity().FS(featureVueKind, fVKLookupComponentPath)
					case fVKUseInDialogForm:
						lc = t.Entity().FS(featureVueKind, fVKFormComponent)
						lcp = t.Entity().FS(featureVueKind, fVKFormComponentPath)
					default:
						lc = t.Entity().FS(featureVueKind, fVKLookupComponent)
						lcp = t.Entity().FS(featureVueKind, fVKLookupComponentPath)
					}
				}
				if lc != "" && lcp != "" {
					if addToRequired {
						th.addComponent(lc, lcp, t.Entity())
					}
				}
			}

			if lc == "" || lcp == "" {
				cg.desc.AddWarning(fmt.Sprintf("at %v: lookupComponent not found for field %s", f.fld.Pos, f.fld.Name))
			}

			return lc
		},
		"LookupWithQualifier": func(f fieldDescriptor) bool {
			_, ok := f.fld.Features.GetField(gen.FeatureDictKind, gen.FDQualifiedBy)
			return ok
		},
		"ByRefField": func(f fieldDescriptor) bool {
			return f.fld.HasModifier(gen.AttrModifierEmbeddedRef) || f.fld.FB(gen.GQLFeatures, gen.GQLFIDOnly)
		},
		"AppendToField": func(f fieldDescriptor) []string {
			if he, ok := f.fld.Features.GetEntity(gen.FeatureHistKind, gen.FHHistoryEntity); ok {
				hc := he.FS(featureVueKind, fVKHistComponent)
				hcp := he.FS(featureVueKind, fVKHistComponentPath)
				if hf, ok := f.fld.Features.GetField(gen.FeatureHistKind, gen.FHHistoryField); ok {
					fn := hf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
					th.addComponent(hc, hcp, he)
					return []string{fmt.Sprintf("<%s v-if=\"value && value.%s\" :items=\"value.%s\"/>", hc, fn, fn)}
				}
			}
			return nil
		},
		"ItemText": func() string {
			return getTitleFieldName(e)
		},
		"ItemValue": func() string { return idf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "") },
		"FieldWithAppend": func(f fieldDescriptor) bool {
			return f.fld.FS(gen.FeatureHistKind, gen.FHHistoryEntityName) != ""
		},
		"WithPrependIcon": func(f fieldDescriptor) bool {
			return f.fld.Annotations.GetStringAnnotationDefTrimmed(vueAnnotation, vcaPrependIcon, "") != ""
		},
		"WithAppendIcon": func(f fieldDescriptor) bool {
			return f.fld.Annotations.GetStringAnnotationDefTrimmed(vueAnnotation, vcaAppendIcon, "") != ""
		},
		"PrependIcon": func(f fieldDescriptor) string {
			if pi := f.fld.Annotations.GetStringAnnotationDefTrimmed(vueAnnotation, vcaPrependIcon, ""); pi != "" {
				fields := strings.Fields(pi)
				attrs := ""
				if len(fields) > 1 {
					if fields[1] != "unset" {
						attrs = fmt.Sprintf("color=\"%s\" ", fields[1])
					}
					for i := 2; i < len(fields); i++ {
						attrs += fields[i] + " "
					}
				}
				return fmt.Sprintf("<v-icon %s>%s</v-icon>", attrs, fields[0])
			}
			return ""
		},
		"AppendIcon": func(f fieldDescriptor) string {
			if pi := f.fld.Annotations.GetStringAnnotationDefTrimmed(vueAnnotation, vcaAppendIcon, ""); pi != "" {
				fields := strings.Fields(pi)
				attrs := ""
				if len(fields) > 1 {
					if fields[1] != "unset" {
						attrs = fmt.Sprintf("color=\"%s\" ", fields[1])
					}
					for i := 2; i < len(fields); i++ {
						attrs += fields[i] + " "
					}
				}
				return fmt.Sprintf("<v-icon %s>%s</v-icon>", attrs, fields[0])
			}
			return ""
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
			ret := ""
			if f.mask != "" {
				ret = fmt.Sprintf("v-mask=\"'%s'\" ", f.mask)
			}
			if pref, ok := f.fld.Annotations.GetStringAnnotation(vueAnnotation, vcaPrefix); ok {
				ret += fmt.Sprintf("prefix=\"%s\" ", pref)
			}
			if suff, ok := f.fld.Annotations.GetStringAnnotation(vueAnnotation, vcaSuffix); ok {
				ret += fmt.Sprintf("suffix=\"%s\" ", suff)
			}
			return ret
		},
		"RequiredComponents": func() map[string]vcComponentDescriptor {
			return th.components
		},
		"AdditionalComponents": func() map[string]vcCustomComponentDescriptor { return customComponents },
		"IsID": func(f fieldDescriptor) bool {
			return f.fld.IsIdField()
		},
		"NotAuto": func(f fieldDescriptor) bool {
			return f.fld.IsIdField() && !f.fld.HasModifier(gen.AttrModifierIDAuto)
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
			if f.fld.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false) ||
				f.fld.Annotations.GetBoolAnnotationDef(js.Annotation, js.AnnotationIcon, false) {
				return "icon"
			}
			if f.fld.Annotations.GetBoolAnnotationDef(js.Annotation, js.AnnotationColor, false) {
				return "color"
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
		"EditableInTable": func(f fieldDescriptor) string {
			if f.fld.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vueATEditable, false) {
				return "true"
			}
			return "false"
		},
		"TypeForView": func(f fieldDescriptor) string {
			return f.fld.Type.Type
		},
		"CanBeMultiple": func() bool {
			// if refsManyToMany, ok := e.Features.GetBool(gen.FeaturesCommonKind, gen.FCRefsAsManyToMany); ok && refsManyToMany {
			// 	return true
			// }
			// return false
			return e.IsDictionary() || e.Annotations.GetBoolAnnotationDef(vueLookupAnnotation, vlaMultiple, false)
		},
		"LookupAttrs": func(f fieldDescriptor) (ret string) {
			if _, itsManyToMany := f.fld.Features.GetEntity(gen.FeaturesCommonKind, gen.FCManyToManyType); itsManyToMany {
				ret = "multiple"
			} else if f.fld.Parent().Annotations[gen.AnnotationFind] != nil {
				ret = ":returnObject='false' hideAdd"
				if f.fld.Type.Array != nil {
					ret += " multiple"
				}
			} else if f.fld.Type.Array != nil {
				ret = "multiple"
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
		"IsNullable": func(f fieldDescriptor) bool {
			return !f.fld.Type.NonNullable
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
				return cd.relPath
			}
			p := e.FS(featureVueKind, fVKFormComponentPath)
			if p[0] != '@' && p[0] != '.' && !path.IsAbs(p) {
				p = "." + string(os.PathSeparator) + p
			}
			return p
		},
		"DialogWidth": func() string {
			if th.ctx.width != "" {
				return th.ctx.width
			}
			//TODO: check fields and form widths; if set - set to fit-content
			if a, ok := e.Annotations[vueFormAnnotation]; ok {
				if a.GetTag(vcaWidth) != nil {
					return "fit-content"
				}
			}
			wideWidth := "60vw"
			if len(th.ctx.fields) < 4 {
				wideWidth = "40vw"
			} else if len(th.ctx.fields) < 6 {
				wideWidth = "50vw"
			}
			return fmt.Sprintf("$vuetify.breakpoint.lgAndUp ? '%s' : '80vw'", wideWidth)
		},
		"NeedSecurity": func() bool {
			return ctx.needSecurity || ctx.needResourceSecurity
		},
		"NeedRolesSecurity": func() bool {
			return ctx.needSecurity
		},
		"NeedResourceSecurity": func() bool {
			return ctx.needResourceSecurity
		},
		"SecurityImport": func() string {
			if ctx.needSecurity || ctx.needResourceSecurity {
				//TODO: get from options
				return "import { LoginManager } from '@/plugins/loginManager';"
			}
			return ""
		},
		"SecurityInject": func() string {
			if ctx.needSecurity || ctx.needResourceSecurity {
				//TODO: get from options
				return "  @Inject(\"loginManager\") loginManager!: LoginManager;"
			}
			return ""
		},
		"RolesForTab": func(tab string) string {
			for _, t := range ctx.tabs {
				if t.ID == tab {
					if t.roles != "" {
						roles := strings.Fields(t.roles)
						ret := ""
						for i, r := range roles {
							if r != "" {
								if i > 0 {
									ret += ", "
								}
								ret += fmt.Sprintf("'%s'", r)
							}
						}
						return ret
					}
					break
				}
			}
			return ""
		},
		"ResourceForTab": func(tab string) string {
			for _, t := range ctx.tabs {
				if t.ID == tab {
					return t.resource
				}
			}
			return ""
		},
		"FieldRoles": func(f fieldDescriptor) string {
			roles := f.ann.getStringDef(vcaRoles, "")
			if roles == "" {
				return ""
			}
			rs := strings.Fields(roles)
			bldr := strings.Builder{}
			for i, role := range rs {
				if i > 0 {
					bldr.WriteRune(',')
					bldr.WriteRune(' ')
				}
				bldr.WriteRune('"')
				bldr.WriteString(role)
				bldr.WriteRune('"')
			}
			return bldr.String()
		},
		"FlexWrap": func() string {
			if !ctx.ann.getBoolDef(vcaNoWrap, false) {
				return "flex-wrap"
			}
			return ""
		},
		"FlexJustify": func() string {
			if len(ctx.fields) == 2 {
				return "justify-space-around"
			}
			return "justify-space-between"
		},
		"FormStyles": func() string {
			var width string
			if w, ok := ctx.ann.getString(vcaWidth); ok {
				width = w
			} else if w, ok := ctx.ann.getInt(vcaWidth); ok {
				width = fmt.Sprintf("%dpx", w)
			}
			if width != "" {
				return fmt.Sprintf(`style="width: %s"`, width)
			}
			return ""
		},
		"FlexFieldStyles": func(f fieldDescriptor) string {
			var width string
			if w, ok := f.ann.getString(vcaWidth); ok {
				width = w
			} else if w, ok := f.ann.getInt(vcaWidth); ok {
				width = fmt.Sprintf("%dpx", w)
			}
			if width != "" {
				return fmt.Sprintf(`style="width: %s"`, width)
			}
			return ""
		},
		"Literal": func(key string) string {
			//TODO get from options
			switch key {
			case literalOKButton:
				return "OK"
			case literalDeleteButton:
				return "Delete"
			case literalCloseButton:
				return "Close"
			case literalDeleteVerb:
				return "Delete"
			default:
				cg.desc.AddError(fmt.Errorf("undefined key for Literal: %s", key))
			}
			return ""
		},
		"WithValidator": func() bool {
			return e.FB(js.FeaturesValidator, js.FVGenerate) ||
				e.FB(gen.FeaturesValidator, gen.FVValidationRequired)
		},
		"ValidatorClass": func() string {
			return e.FS(js.FeaturesValidator, js.FVValidatorClass)
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
