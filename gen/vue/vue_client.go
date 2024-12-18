package vue

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
)

const (
	vueGeneratorName         = "vue"
	vueAnnotation            = "vue"
	vueAnnotationFilePath    = "path"
	vueAnnotationIgnore      = "ignore"
	vueAnnotationUse         = "use"
	vueAnnotationDisplayType = "type"
	vueAnnotationReadonly    = "readonly"

	// use iso date for date field
	vueATISODate = "isodate"
	// use datetime for date field
	vueATDate = "date"
	// use datetime for date field
	vueATDateTime = "datetime"
	// use datetime for date field
	vueATTime = "time"
	// custom may be used in tables and forms for custom component (should be registered in Vue App) : custom="CustomComponenName"
	vueATCustom = "custom"
	// redefine value for complex field (default - js:title annotatef field)
	vueATValue = "value"
	// tooltip descriptor
	vueATTooltip = "tooltip"
	// editable in table
	vueATEditable = "editable"
	// use for color
	vueATColorAttr = "color"

	//vueAnnotationComponent - string or bool for defining custom form component (in form: 'ComponentName from fileName') or request it's generation
	// vueAnnotationFormComponent = "formComponent"
	//vueAnnotationViewComponent - string or bool for defining custom view component (in form: 'ComponentName from fileName') or request it's generation
	// vueAnnotationViewComponent = "viewComponent"
	//vueAnnotationCardComponent - bool request card generation (by default false)
	// vueAnnotationCardComponent = "cardComponent"
	//vueAnnotationLookupComponent - string or bool for defining custom lookup component (in form: 'ComponentName from fileName') or request it's generation
	// vueAnnotationLookupComponent = "lookupComponent"
	//vueAnnotationDialogComponent - bool request dialog generation (by default true)
	// vueAnnotationDialogComponent = "dialogComponent"
	// view annotations:
	// vue-form can be in form of:
	//  $vue-form() or $vue-form:formName()
	//  the former creates form with default name and params
	//  the later one creates named form; this allows create few defferent forms or set params for fields
	// for all annotations may be used tags:
	//   ignore - not generate anything
	//   use - use existing component (string in form: 'ComponentName from fileName')
	vueFormAnnotation = "vue-form"
	//vfaCard - tag for vue-form (bool) - also generate card component
	vfaCard = "card"

	vueViewAnnotation = "vue-view"

	vueDialogAnnotation = "vue-dialog"

	vueLookupAnnotation = "vue-lookup"
	vlaMultiple         = "multiple"

	vueTableAnnotation = "vue-table"
	vtaUseIcon         = "useIcon"
	vueTabsAnnotation  = "vue-tabs"
	vueTreeAnnotation  = "vue-tree"

	// vueTabSet - set of tabs for forms/views; enumerates tabs id; may be in form tabid="tab-label"; or just tabid, then for each tab may be defined annotation vue-tab:tabid
	vueTabSet = "vue-tabs"
	//vueTab - description of tab (vue-tab:tabid); may content order, label
	vueTab = "vue-tab"
	// vueSkipTabs may be used in forms with specifier to ignore tabs definition
	vueSkipTabs = "skipTabs"

	//common annotations
	//vcaWidth = annotation for width;
	// for form should be string (in % or any dimension with dimension)
	// for field may string (with dimension) or number (1-12 in grid points or proportional value)
	vcaWidth = "width"
	//vcaLayout - layout. may be "grid", "flex" or "table" (not implemnted yet)
	vcaLayout = "layout"
	vcalGrid  = "grid"
	vcalFlex  = "flex"
	vcalTable = "table"

	//vcaLabel - label for field, title for dialog etc
	vcaLabel = "label"
	//vcaRoles - roles that has access to this element (tab currently)
	vcaRoles = "roles"
	// vcaResource - check access for resource
	vcaResource = "resource"
	// vcaComponent - use component for tab
	vcaComponent = "component"
	//vcaOrder - order for field (int) if not specified - 1000
	vcaOrder = "order"
	//vcaRow - row int (if not set - 1; if not set for all of the field - will be formed with wrap)
	vcaRow = "row"
	//vcaMask - mask for text input
	vcaMask = "mask"
	//vcaTab - id of tab (for tabbed forms/views)
	vcaTab = "tab"
	//vcaIf - condition (js) when field is shown in form (inside "vue" annotation)
	vcaIf = "if"
	//vcaSuffix - text suffix for v-text-field
	vcaSuffix = "suffix"
	//vcaPrefix - text suffix for v-text-field
	vcaPrefix = "prefix"
	//vcaPrependIcon - icon to put before component: name [color [<other v-icon attrs, like small...]]
	vcaPrependIcon = "prependIcon"
	//vcaAppendIcon - icon to put after component: name [color [<other v-icon attrs, like small...]]
	vcaAppendIcon = "appendIcon"
	//vcaReadonly - hide plus button in lookup
	vcaReadonly = "readonly"
	// may be used instead of text-field may be bool or int (number of lines)
	vcaTextArea = "textArea"
	vcaDefault  = "default"
	// vcaNoWrap may be used for flex forms (by default flex-wrap class used)
	vcaNoWrap = "noWrap"
	// vcaCompact generate form without paddings
	vcaCompact = "compact"
)

const (
	VCOptions              = "vue"
	VCOptionDateComponent  = "date"
	VCOptionMapComponent   = "map"
	VCOptionColorComponent = "color"
	VCOptionsApolloClient  = "apollo-client"

	vcoApolloClientDef = "this.$apollo.getClient()"
)

const (
	// featureVueKind kind for vue features
	featureVueKind gen.FeatureKind = "vue-client"

	//fVKOutDir - string with path for component for this entity
	fVKOutDir = "out-dir"
	//fVKIgnore may be used for field for not show it in forms
	fVKIgnore = "ignore"
	//fVKFormRequired - bool, true if form instance required
	fVKFormRequired = "form-required"
	//fVKFormListRequired - bool, true if form instance required
	fVKFormListRequired = "form-list-required"
	//fVKCardRequired - bool, true if card instance required
	fVKCardRequired = "card-required"
	//fVKDialogRequired - bool, true if dialog instance required
	fVKDialogRequired = "dialog-required"
	//fVKViewRequired - bool, true if view instance required
	fVKViewRequired = "view-required"
	//fVKLookupRequired - bool, true if lookup instance required
	fVKLookupRequired = "lookup-required"

	//fVKFormComponent - string, name of form component
	fVKFormComponent = "form-component"
	//fVKFormListComponent - string, name of form component
	fVKFormListComponent = "form-list-component"
	//fVKCardComponent - string, name of card component
	fVKCardComponent = "card-component"
	//fVKDialogComponent - string, name of Dialog component
	fVKDialogComponent = "dialog-component"
	//fVKViewComponent - bool, true if view instance required
	fVKViewComponent = "view-component"
	//fVKHistComponent - string, name of history component
	fVKHistComponent     = "hist-component"
	fVKDictEditComponent = "dict-edit-component"
	fVKLookupComponent   = "lookup-component"

	// path values - path to components
	fVKFormComponentPath     = "form-component-path"
	fVKFormListComponentPath = "form-list-component-path"
	fVKCardComponentPath     = "card-component-path"
	fVKDialogComponentPath   = "dialog-component-path"
	fVKViewComponentPath     = "view-component-path"
	fVKConfComponentPath     = "conf-component-path"
	fVKLookupComponentPath   = "lookup-component-path"
	fVKTypeDescriptorPath    = "type-descriptor"
	fVKHistComponentPath     = "hist-component-path"
	fVKDictEditComponentPath = "dict-edit-component-path"

	//fVKUseInDialog - whether use lookup or form in dialog ('lookup'|'form'|'ignore')
	fVKUseInDialog       = "use-in-dialog"
	fVKUseInDialogIgnore = "ignore"
	fVKUseInDialogForm   = "form"
	fVKUseInDialogLookup = "lookup"
)

const (
	literalOKButton     = "okButton"
	literalDeleteButton = "deleteButton"
	literalCloseButton  = "closeButton"
	literalDeleteVerb   = "deleteVerb"
)

// var vcCustomComponents = map[string]string{
//   VCOptionDateComponent: "InputDateComponent",
// }

type vueComponentField struct {
	Name            string
	Type            string
	Required        bool
	Label           string
	LookupComponent string
	Field           *gen.Field
}
type vueComponent struct {
	Name          string
	ComponentName string
	TypeName      string
	Fields        []*vueComponentField
	Entity        *gen.Entity
}
type VCOptionComponentSpec struct {
	Name   string `json:"name"`
	Import string `json:"import"`
}
type ClientOptions struct {
	Components      map[string]VCOptionComponentSpec `json:"components"`
	ApolloClientVar string                           `json:"apollo_client_var"`
	OutputDir       string                           `json:"output_dir"`
}
type ClientGenerator struct {
	desc    *gen.Package
	options ClientOptions
	b       *gen.Builder
}

func init() {
	gen.RegisterPlugin(&ClientGenerator{})
}

func (cg *ClientGenerator) Name() string {
	return vueGeneratorName
}

func (cg *ClientGenerator) SetOptions(options any) error {
	return gen.OptionsAnyToStruct(options, &cg.options)
}

func (cg *ClientGenerator) CheckAnnotation(desc *gen.Package, ann *gen.Annotation, item interface{}) (bool, error) {
	// if t, ok := item.(*gen.Entity); ok && t.HasModifier(gen.TypeModifierConfig) {
	//   // at Prepare stage Required features should be set already
	//   cg.processForConfig(t)
	// }
	annname := strings.Split(ann.Name, ":")
	switch annname[0] {
	case vueAnnotation, vueDialogAnnotation, vueLookupAnnotation, vueTableAnnotation, vueFormAnnotation, vueViewAnnotation, vueTabSet, vueTab:
		return true, nil
	case gen.AnnotationConfig:
		if e, ok := item.(*gen.Entity); ok {
			// preventive setting list requred; really better find if we have array of this type...
			if ann.GetBool(gen.AnnCfgValue, false) {
				e.Features.Set(featureVueKind, fVKFormListRequired, true)
			}
			return true, nil
		}
	}
	return false, nil
}

func (cg *ClientGenerator) Prepare(desc *gen.Package) error {
	cg.desc = desc
	if _, err := desc.Options().CustomToStruct(VCOptions, &cg.options); err != nil {
		desc.AddWarning(fmt.Sprintf("problem while setting custom options for vue: %v", err))
	}
	if cg.options.ApolloClientVar == "" {
		cg.options.ApolloClientVar = vcoApolloClientDef
	}
	//outDir := cg.getOutputDir()
	for _, file := range desc.Files {
		for _, t := range file.Entries {
			outDir := cg.getOutputDirForEntity(t)
			t.Features.Set(featureVueKind, fVKOutDir, outDir)
			if t.HasModifier(gen.TypeModifierConfig) {
				t.Features.Set(featureVueKind, fVKConfComponentPath, t.Name+".vue")
			} else {
				generateCommon := true
				if t.Annotations.GetBoolAnnotationDef(gen.AnnotationConfig, gen.AnnCfgGroup, false) {
					generateCommon = false
				}
				if generateCommon {
					// lookup
					n := t.Name + "LookupComponent"
					t.Features.Set(featureVueKind, fVKLookupComponent, n)
					t.Features.Set(featureVueKind, fVKLookupComponentPath, n+".vue")

					// type descriptor
					t.Features.Set(featureVueKind, fVKTypeDescriptorPath, t.Name+"TypeDescriptor.ts")
					// form
					if !t.FB(gen.FeaturesCommonKind, gen.FCReadonly) {
						// form-list
						if t.FB(featureVueKind, fVKFormListRequired) {
							t.Features.Set(featureVueKind, fVKFormRequired, true)
							c := t.Name + "FormListComponent"
							t.Features.Set(featureVueKind, fVKFormListComponent, c)
							t.Features.Set(featureVueKind, fVKFormListComponentPath, c+".vue")
						}
						if a, ok := t.Annotations.GetStringAnnotation(vueFormAnnotation, vueAnnotationUse); ok {
							c, p, ok := parseComponentAnnotation(a)
							if !ok {
								return fmt.Errorf(
									"at %v: invalid component annotation format: '%s'",
									t.Annotations[vueFormAnnotation].Pos,
									a,
								)
							}
							t.Features.Set(featureVueKind, fVKFormRequired, false)
							t.Features.Set(featureVueKind, fVKFormComponent, c)
							t.Features.Set(featureVueKind, fVKFormComponentPath, p)
						} else {
							if ignore, ok := t.Annotations.GetBoolAnnotation(vueFormAnnotation, vueAnnotationIgnore); !ok || !ignore {
								c := t.Name + "FormComponent"
								t.Features.Set(featureVueKind, fVKFormRequired, true)
								t.Features.Set(featureVueKind, fVKFormComponent, c)
								t.Features.Set(featureVueKind, fVKFormComponentPath, c+".vue")
							}
						}
					}
					// do not generate by default
					if create, ok := t.Annotations.GetBoolAnnotation(vueFormAnnotation, vfaCard); ok && create {
						c := t.Name + "Card"
						t.Features.Set(featureVueKind, fVKCardRequired, true)
						t.Features.Set(featureVueKind, fVKCardComponent, c)
						t.Features.Set(featureVueKind, fVKCardComponentPath, c+".vue")
					}
				}

				if t.HasModifier(gen.TypeModifierTransient) {
					continue
				}

				// view
				if a, ok := t.Annotations.GetStringAnnotation(vueViewAnnotation, vueAnnotationUse); ok {
					c, p, ok := parseComponentAnnotation(a)
					if !ok {
						return fmt.Errorf(
							"at %v: invalid component annotation format: '%s'",
							t.Annotations[vueViewAnnotation].Pos,
							a,
						)
					}
					t.Features.Set(featureVueKind, fVKViewRequired, false)
					t.Features.Set(featureVueKind, fVKViewComponent, c)
					t.Features.Set(featureVueKind, fVKViewComponentPath, p)
				} else {
					if ignore, ok := t.Annotations.GetBoolAnnotation(vueViewAnnotation, vueAnnotationIgnore); !ok || !ignore {
						c := t.Name + "View"
						t.Features.Set(featureVueKind, fVKViewRequired, true)
						t.Features.Set(featureVueKind, fVKViewComponent, c)
						t.Features.Set(featureVueKind, fVKViewComponentPath, c+".vue")
					}
				}
				//history
				if _, ok := t.Features.Get(gen.FeatureHistKind, gen.FHHistoryOf); ok {
					c := t.Name + "HistComponent"
					t.Features.Set(featureVueKind, fVKHistComponent, c)
					t.Features.Set(featureVueKind, fVKHistComponentPath, c+".vue")
				}
				if t.FB(featureVueKind, fVKFormRequired) {
					//let's check for lookups requirements
					for _, f := range t.GetFields(true, true) {
						if ignore, ok := f.Annotations.GetBoolAnnotation(vueAnnotation, vueAnnotationIgnore); ok && ignore {
							f.Annotations.AddTag(vueFormAnnotation, vueAnnotationIgnore, true)
							continue
						}
						if ignore, ok := f.Annotations.GetBoolAnnotation(
							gen.GQLAnnotation,
							gen.GQLAnnotationSkipTag,
						); ok && ignore {
							f.Annotations.AddTag(vueFormAnnotation, vueAnnotationIgnore, true)
							continue
						}
						if _, ok := f.Annotations.GetStringAnnotation(vueAnnotation, vcaIf); ok {
							err := cg.checkFieldIfStatment(t, f)
							if err != nil {
								return err
							}
						}
						//TODO check specific annotations
						if f.Type.Complex {
							if f.Type.Map != nil {
								//TODO: out map values
								continue
							}
							tn := f.Type.Type
							if f.Type.Array != nil {
								tn = f.Type.Array.Type
							}
							if f.Type.Array == nil || f.Type.Array.Complex {
								t, ok := desc.FindType(tn)
								if !ok || t.Entity() == nil {
									f.Annotations.AddTag(vueFormAnnotation, vueAnnotationIgnore, true)
									desc.AddWarning(fmt.Sprintf("at %v: can not find type '%s' for form generation; ignoring", f.Pos, tn))
									continue
								}
								if t.Entity().HasModifier(gen.TypeModifierTransient) {
									f.Annotations.AddTag(vueFormAnnotation, vueAnnotationIgnore, true)
									continue
								}
								if t.Entity().HasModifier(gen.TypeModifierEmbeddable) {
									f.Features.Set(featureVueKind, fVKUseInDialog, fVKUseInDialogForm)
									continue
								}
								f.Features.Set(featureVueKind, fVKUseInDialog, fVKUseInDialogLookup)
							}
						}
					}
				}

				if t.HasModifier(gen.TypeModifierEmbeddable) {
					continue
				}

				if !t.FB(gen.FeaturesCommonKind, gen.FCReadonly) {
					if ignore, ok := t.Annotations.GetBoolAnnotation(vueDialogAnnotation, vueAnnotationIgnore); !ok || !ignore {
						c := t.Name + "DialogComponent"
						t.Features.Set(featureVueKind, fVKDialogRequired, true)
						t.Features.Set(featureVueKind, fVKDialogComponent, c)
						t.Features.Set(featureVueKind, fVKDialogComponentPath, c+".vue")
					}
				}

				if t.HasModifier(gen.TypeModifierDictionary) && !t.FB(gen.FeaturesCommonKind, gen.FCReadonly) {
					if ignore, ok := t.Annotations.GetBoolAnnotation(vueAnnotation, vueAnnotationIgnore); !ok || !ignore {
						c := t.Name + "DictEditComponent"
						t.Features.Set(featureVueKind, fVKDialogRequired, true)
						t.Features.Set(featureVueKind, fVKDictEditComponent, c)
						t.Features.Set(featureVueKind, fVKDictEditComponentPath, c+".vue")
					}
				}
			}
		}
		for _, enum := range file.Enums {
			n := enum.Name + "LookupComponent"
			enum.Features.Set(featureVueKind, fVKLookupComponent, n)
			enum.Features.Set(featureVueKind, fVKLookupComponentPath, n+".vue")
		}
	}
	return nil
}

func (cg *ClientGenerator) Generate(b *gen.Builder) (err error) {
	cg.desc = b.Descriptor
	cg.b = b
	for _, t := range b.File.Entries {
		outDir := cg.getOutputDirForEntity(t)
		err := cg.generateFor(outDir, t)
		if err != nil {
			return err
		}
	}
	for _, e := range b.File.Enums {
		outDir := cg.getOutputDirForEnum(e)
		err := cg.generateEnum(outDir, e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cg *ClientGenerator) generateFor(outDir string, e *gen.Entity) (err error) {
	if e.HasModifier(gen.TypeModifierTransient) || e.HasModifier(gen.TypeModifierSingleton) || e.HasModifier(gen.TypeModifierExternal) {
		return
	}
	if ignore, ok := e.Annotations.GetBoolAnnotation(vueAnnotation, vueAnnotationIgnore); ok && ignore {
		return
	}

	if e.HasModifier(gen.TypeModifierConfig) {
		ch, err := cg.newConfigHelper("VUE-CONFIG", e, outDir)
		if err != nil {
			return err
		}
		err = ch.generate()
		if err != nil {
			return err
		}
	} else {
		//th, err := cg.newHelper("VUE", e, outDir)
		th, err := cg.newFormHelper("LOOKUP", e, vueLookupAnnotation, "", outDir)
		if err != nil {
			return err
		}

		//if e.FB(featureVueKind, fVKDialogRequired) {
		//	err := th.createDialog()
		//	if err != nil {
		//		return err
		//	}
		//}

		if th.idField != nil {
			p := e.FS(featureVueKind, fVKLookupComponentPath)
			p = filepath.Join(outDir, p)
			f, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("while opening file for LookupComponent for %s: %v", e.Name, err)

			}
			defer f.Close()
			if e.IsDictionary() {
				th.parse(vueDictionaryLookupTSTemplateBody)
			} else {
				th.parse(vueEntityLookupTSTemplateBody)
			}
			th.parse(htmlLookupTemplate).parse("{{template \"TS\" .}}\n{{template \"HTML\" .}}\n")
			if th.err != nil {
				return fmt.Errorf("Error while parsing template for LookupComponent: %v\n", th.err)
			}
			buffer := bytes.Buffer{}
			err = th.templ.Execute(&buffer, th)
			if err != nil {
				return fmt.Errorf("while executing template for LookupComponent for %s: %v", th.e.Name, err)
			}

			if e.IsDictionary() {
				th.parse(vueDictionaryLookupTSTemplateHeader)
			} else {
				th.parse(vueEntityLookupTSTemplateHeader)
			}
			th.parse("{{template \"TS-HEADER\" .}}\n")
			if th.err != nil {
				return fmt.Errorf("error while parsing template for LookupComponent header: %v", th.err)
			}
			err = th.templ.Execute(f, th)
			if err != nil {
				return fmt.Errorf("while executing template for LookupComponent header for %s: %v", th.e.Name, err)
			}
			_, err = buffer.WriteTo(f)
			if err != nil {
				return fmt.Errorf("while executing template for LookupComponent header for %s: %v", th.e.Name, err)
			}
		}

		// type descriptor
		p := e.FS(featureVueKind, fVKTypeDescriptorPath)

		if p != "" {
			th, err := cg.newFormHelper("TABLE", e, vueTableAnnotation, "", outDir)
			if err != nil {
				return err
			}
			p = filepath.Join(outDir, p)
			f, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("while opening file for TypeDescriptor for %s: %v", e.Name, err)
			}
			defer f.Close()
			th.parse(typeDescriptorTSTemplate)
			if th.err != nil {
				return fmt.Errorf("while parsing template: %v", th.err)
			}
			err = th.templ.Execute(f, th)
			if err != nil {
				return fmt.Errorf("while executing template for TypeDescriptor for %s: %v", e.Name, err)
			}
		}
		//card view
		p = e.FS(featureVueKind, fVKViewComponentPath)

		if p != "" {
			th, err = cg.newFormHelper("VIEW", e, vueViewAnnotation, "", outDir)
			p = filepath.Join(outDir, p)
			f, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("while opening file for ViewComponent for %s: %v", e.Name, err)
			}
			defer f.Close()
			th.parse(htmlCardViewTemplate).
				parse(htmlFieldViewTemplate).
				parse(htmlFieldViewBoolTemplate).
				parse(htmlFieldViewComplexTemplate).
				parse(htmlFieldViewDateTemplate).
				parse(htmlFieldViewFloatTemplate).
				parse(htmlFieldViewIntTemplate).
				parse(htmlFieldViewTextTemplate).
				parse(vueViewTSTemplateBody).
				parse(htmlViewCSS).
				parse("{{template \"TS\" .}}\n{{template \"HTML\" .}}\n{{template \"CSS\" .}}\n")
			if th.err != nil {
				return fmt.Errorf("error while parsing view template: %v", th.err)
			}
			buffer := bytes.Buffer{}
			err = th.templ.Execute(&buffer, th)
			if err != nil {
				return fmt.Errorf("while executing template for ViewComponent for %s: %v", e.Name, err)
			}
			th.parse(vueViewTSTemplateHeader).
				parse("{{template \"TS-HEADER\" .}}\n")
			if th.err != nil {
				return fmt.Errorf("error while parsing template for View: %v", th.err)
			}
			err = th.templ.Execute(f, th)
			if err != nil {
				return fmt.Errorf("while executing template for View for %s: %v", th.e.Name, err)
			}
			_, err = buffer.WriteTo(f)
			if err != nil {
				return fmt.Errorf("while writing body for View for %s: %v", th.e.Name, err)
			}
		}

		if e.FB(featureVueKind, fVKDialogRequired) {
			th, err = cg.newFormHelper("DIALOG", e, vueDialogAnnotation, "", outDir)
			if err != nil {
				return err
			}
			err := th.createDialog()
			if err != nil {
				return err
			}
		}

		th, err = cg.newFormHelper("GRID", e, vueFormAnnotation, "", outDir)
		if err != nil {
			return err
		}
		err = th.generateGridForm("main")
		if err != nil {
			return err
		}
		for spec := range e.Annotations.ByPrefix(vueFormAnnotation, false) {
			th, err = cg.newFormHelper("GRID:"+spec, e, vueFormAnnotation, spec, outDir)
			if err != nil {
				return err
			}
			if cd, ok := th.ctx.getComponentDescriptor("form"); ok {
				err = th.generateGridForm("form"+spec, cd.path)
				if err != nil {
					return err
				}
			}
			th, err = cg.newFormHelper("GRID:"+spec, e, vueDialogAnnotation, spec, outDir)
			if err != nil {
				return err
			}
			if cd, ok := th.ctx.getComponentDescriptor("dialog"); ok {
				err = th.createDialog(cd.path)
				if err != nil {
					return err
				}
			}
		}
		if e.IsDictionary() {
			err = th.createDictEditor()
			if err != nil {
				return err
			}
		}
		if e.FS(featureVueKind, fVKHistComponent) != "" {
			err = th.generateHistoryComponent()
			if err != nil {
				return err
			}
		}

		if it, ok := e.Features.GetEntity(gen.FeaturesAPIKind, gen.FAPIFindParamType); ok {
			h, err := cg.newFormHelper("Find", it, vueFormAnnotation, "", outDir)
			if err != nil {
				return err
			}
			err = h.generateGridForm("find")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cg *ClientGenerator) getOutputDir() (ret string) {
	ret = filepath.Join(cg.getClientOutputDir(), "components")
	os.MkdirAll(ret, os.ModeDir|os.ModePerm)
	return
}

func (cg *ClientGenerator) getClientOutputDir() (ret string) {
	ret = "./gql-ts"
	if opt := cg.desc.Options().ClientOutputDir; opt != "" {
		ret = opt
	}
	if opt := cg.options.OutputDir; opt != "" {
		ret = opt
	}
	ret = filepath.FromSlash(ret)
	return
}

func (cg *ClientGenerator) getOutputDirForEntity(e *gen.Entity) (ret string) {
	return cg.getOutputDirForFile(e.File.Package, e.File.Name)
}

func (cg *ClientGenerator) getOutputDirForEnum(e *gen.Enum) (ret string) {
	return cg.getOutputDirForFile(e.File.Package, e.File.Name)
}

func (cg *ClientGenerator) getOutputDirForFile(packageName, fileName string) (ret string) {
	dir := filepath.Join(cg.getOutputDir(), packageName, fileName)
	err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		cg.desc.AddError(err)
	}
	return dir
}

func (cg *ClientGenerator) pathToRelative(from, to string) (ret string) {
	var err error
	ret, err = filepath.Rel(from, to)
	if err != nil {
		cg.b.AddWarning(fmt.Sprintf("problem while getting relative path for '%s': %v", to, err))
		ret = to
	} else {
		if ret[0] != '.' {
			ret = fmt.Sprintf(".%c%s", filepath.Separator, ret)
		}
	}
	return
}

func (cg *ClientGenerator) getJSAttrNameForDisplay(f *gen.Field, forTable bool, forIcon bool) string {
	annName := js.AnnotationTitle
	if forIcon && f.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false) {
		annName = js.AnnotationIcon
	}
	ret := f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
	_, custom := f.Annotations.GetStringAnnotation(vueAnnotation, vueATCustom)
	if forTable {
		if _, ok := f.Annotations.GetStringAnnotation(vueTableAnnotation, vueATCustom); ok {
			custom = true
		}
	}
	if f.Type.Complex && !custom {
		if f.Type.Array != nil || f.Type.Map != nil {
			return ""
		} else if t, ok := cg.desc.FindType(f.Type.Type); ok && t.Entity() != nil {
			found := false
			for _, ff := range t.Entity().GetFields(true, true) {
				if _, ok := ff.Annotations.GetBoolAnnotation(js.Annotation, annName); ok {
					if ff.Type.Complex || !forTable {
						ret = ret + "." + cg.getJSAttrNameForDisplay(ff, forTable, forIcon)
					} else {
						ret = ret + "." + ff.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
					}
					found = true
					break
				}
			}
			if !found && !f.Parent().HasModifier(gen.TypeModifierEmbeddable) {
				if forTable {
					cg.desc.AddWarning(
						fmt.Sprintf(
							"vue: at %v: can not find attr for table in type %s for %s",
							f.Pos,
							f.Type.Type,
							f.Name,
						),
					)
				} else {
					if _, ok := t.Entity().Annotations[gen.AnnotationConfig]; !ok {
						cg.desc.AddWarning(
							fmt.Sprintf(
								"vue: at %v: can not find title field in type %s for %s",
								f.Pos,
								f.Type.Type,
								f.Name,
							),
						)
					}
				}
			}
		} else {
			cg.desc.AddWarning(fmt.Sprintf("vue: at %v: type %s not found for %s", f.Pos, f.Type.Type, f.Name))
		}
	}
	return ret
}

func (cg *ClientGenerator) processForConfig(t *gen.Entity) {
	for _, f := range t.Fields {
		if f.Type.Array != nil {
			if at, ok := cg.b.FindType(f.Type.Array.Type); ok && at.Entity() != nil {
				e := at.Entity()
				if e.Annotations.GetBoolAnnotationDef(gen.AnnotationConfig, gen.AnnCfgValue, false) {
					e.Features.Set(featureVueKind, fVKFormListRequired, true)
				} else if e.Annotations.GetBoolAnnotationDef(gen.AnnotationConfig, gen.AnnCfgGroup, false) {
					cg.processForConfig(e)
				}
			}
		}

	}
}

func (cg *ClientGenerator) checkFieldIfStatment(t *gen.Entity, f *gen.Field) error {
	// if iff, ok := f.Annotations.GetStringAnnotation(vueAnnotation, vcaIf); ok {
	//   // so far only boolean values
	//   parts := strings.Split(iff, ".")
	// }
	return nil
}

func (cg *ClientGenerator) getJSAttrColorForTable(f *gen.Field) (string, bool) {
	if f.Type.Complex {
		ret := f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		found := false
		if f.Type.Array != nil || f.Type.Map != nil {
			return "", false
		} else if t, ok := cg.desc.FindType(f.Type.Type); ok && t.Entity() != nil {
			for _, ff := range t.Entity().GetFields(true, true) {
				if _, ok := ff.Annotations.GetBoolAnnotation(js.Annotation, js.AnnotationColor); ok {
					if ff.Type.Complex {
						ca, ok := cg.getJSAttrColorForTable(ff)
						if !ok || ca == "" {
							return "", false
						}
						ret = ret + "." + ca
					} else {
						ret = ret + "." + ff.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
					}

					found = true
				}
			}
		}
		if !found {
			ret = ""
		}
		return ret, found
	}
	return "", false
}

func (cg *ClientGenerator) getJSAttrForSubfield(f *gen.Field, fieldName string) string {
	fn := f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
	if f.Type.Complex && fieldName != "" && fn != "" {
		if f.Type.Array == nil && f.Type.Map == nil {
			if t, ok := cg.desc.FindType(f.Type.Type); ok && t.Entity() != nil {
				if ttf := t.Entity().GetField(fieldName); ttf != nil {
					return fmt.Sprintf("%s.%s", fn, ttf.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, ""))
				}
			}
		}
	}
	return ""
}

func (cg *ClientGenerator) getPathForComponent(e *gen.Entity, name string) string {
	outDir := cg.getOutputDirForEntity(e)
	return filepath.Join(outDir, name)
}

func parseComponentAnnotation(ann string) (component string, path string, ok bool) {
	re := regexp.MustCompile(`^[ \t]*([a-zA-Z_][a-zA-Z_0-9]*)[ \t]+from[ \t]+(([.@/][./a-zA-Z0-9_-]*)|('([^']+)'))[ \t]*$`)
	parts := re.FindStringSubmatch(ann)
	if len(parts) == 6 {
		component = parts[1]
		if parts[5] != "" {
			path = parts[5]
		} else {
			path = parts[3]
		}
	}
	ok = component != "" && path != ""
	return
}

type vcCustomComponentDescriptor struct {
	Comp string
	Imp  string
}

type vcComponentDescriptor struct {
	Comp string
	Imp  string
}

const typeDescriptorTSTemplate = `
export const {{TypeName .}}Descriptor = {
  id: "{{IDField}}",
  headers: [{{range (GetFields .)}} {{if ShowInTable .}}
    {text: "{{Label .}}", value: "{{TableAttrName .}}", {{if NeedIconForTable .}}icon: "{{TableIconName .}}", {{end}}type: "{{GUITableType .}}"{{if ne (GUITableColor .) ""}}, color: "{{GUITableColor .}}"{{end}}{{if ne (GUITableComponent .) ""}}, component: "{{GUITableComponent .}}"{{end}}{{if ne (GUITableTooltip .) ""}}, tooltip: "{{GUITableTooltip .}}"{{end}}, editable: {{EditableInTable .}}{{if ne (FieldRoles .) ""}}, roles: [{{FieldRoles .}}]{{end}}},{{end}}{{end}}
  ]
};
`

// CardView
var htmlCardViewTemplate = `
{{define "HTML"}}
<template>
  <v-card >
    <v-card-text>
      <v-progress-linear v-if="loading" intermediate></v-progress-linear>
      <div class="d-flex flex-row flex-wrap justify-space-around align-baseline" v-if="value">
        {{range (GetFields .)}}{{if ShowInDialog .}}<div class="mx-2">
          {{template "VIEW_FIELD" .}}
        </div>{{end}}
        {{end}}
      </div>
    </v-card-text>
  </v-card>
</template>
{{end}}
`

const htmlFieldViewTemplate = `{{define "VIEW_FIELD"}}{{if ShowInView .}}<div v-if="value.{{AttrName .}} != undefined || showEmpty" class="d-flex flex-column"><div class="field-value">{{if eq (TypeForView .) "string"}}{{template "TEXT_VIEW" .}}
  {{else if eq (TypeForView .) "int"}}{{template "TEXT_VIEW" .}}
  {{else if eq (TypeForView .) "float"}}{{template "TEXT_VIEW" .}}
  {{else if eq (TypeForView .) "date"}}{{template "DATE_VIEW" .}}
  {{else if eq (TypeForView .) "bool"}}{{template "BOOL_VIEW" .}}
  {{else}}{{template "COMPLEX_VIEW" .}}{{end}}</div>
  <div class="field-title">{{Label .}}</div></div>{{end}}{{end}}`

const htmlFieldViewTextTemplate = `{{define "TEXT_VIEW"}}{{if IsIcon .}}<v-icon>{{end}}{{"{{"}}value.{{FieldName .}}{{"}}"}}{{if IsIcon .}}</v-icon>{{end}}{{end}}`
const htmlFieldViewIntTemplate = `{{define "INT_VIEW"}}{{"{{"}}value.{{FieldName .}}{{Filter "int" true}}{{"}}"}}{{end}}`
const htmlFieldViewFloatTemplate = `{{define "FLOAT_VIEW"}}{{"{{"}}value.{{FieldName .}}{{Filter "float" true}}{{"}}"}}{{end}}`
const htmlFieldViewDateTemplate = `{{define "DATE_VIEW"}}{{"{{"}}value.{{FieldName .}}{{Filter "date" true}}{{"}}"}}{{end}}`

const htmlFieldViewComplexTemplate = `{{define "COMPLEX_VIEW"}}{{"{{"}}value.{{AttrName .}}{{"}}"}}{{end}}`

const htmlFieldViewBoolTemplate = `{{define "BOOL_VIEW"}}<v-icon>{{"{{"}}value.{{FieldName .}}?'mdi-checkbox-marked':'mdi-checkbox-blank-outline'{{"}}"}}</v-icon>{{end}}`

const vueViewTSTemplateHeader = `
{{define "TS-HEADER"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject, Watch } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}{{if ne (IDType .) "" }}, {{GetQuery .}}{{end}} } from '{{TypesFilePath .}}';
{{FiltersImports}}
{{TypesFromTS}}
{{end}}
`
const vueViewTSTemplateBody = `
{{define "TS"}}

@Component({
  name: "{{TypeName}}CardViewComponent",
  filters: { 
    {{Filter "date"}},
    {{Filter "number"}}
  }
})
export default class {{TypeName}}CardViewComponent extends Vue {
  @Prop() item: {{TypeName .}} {{if ne (IDType .) "" }} | {{IDType .}} {{end}}| undefined;
  @Prop({default:false}) showEmpty!: boolean;
  private value: {{TypeName .}} {{if ne (IDType .) "" }} | {{IDType .}} {{end}}| null = null;
  private loading = false;

  @Watch('item') onValueChanged() {
    if(!this.item) {
      this.value = null;
    } else {
      if(typeof this.item == "object") {
        this.value = this.item;
      } {{if ne (IDType .) "" }}else if(typeof this.item == "{{IDTypeJS .}}") {
        this.load(this.item as {{IDType .}});
      }{{end}}
    }
  }

  created() {
    this.onValueChanged();
  }
  {{if ne (IDType .) "" }}async load(id: {{IDType .}}) {
    this.loading = true;
    try {
      this.value = await {{GetQuery .}}({{ApolloClient}}, id);
    } catch(exc) {
      console.log("problem: ", exc);
    }
    this.loading = false;
  }{{end}}
}
</script>
{{end}}
`
const htmlViewCSS = `
{{define "CSS" }}
<style scoped lang="scss">
  .field-value {
    font-size: 120%;
    font-weight: bolder;
  }
  .field-title {
    font-size: 70%;
    text-align: center;
  }
</style>
{{end}}
`

// Lookups

var htmlLookupTemplate = `
{{define "HTML"}}
<template>
  <div class="flex-row">
    <v-autocomplete
      v-model="selected"
      :hint="hint"
      :items="items"
      :readonly="readonly"
      :disabled="disabled"
      :label="label"
      :item-text="'{{ItemText}}'"
      :item-value="'{{ItemValue}}'"
      :item-disabled="disabledProperty"
      :return-object="returnObject"
      :loading="loading"
      :error-messages="problem || errorMessages"
      hide-no-data
      hide-details="auto"
      {{if CanBeMultiple}}:multiple="multiple"
      :chips="multiple"
      :rules="rules"
      small-chips{{end}}
      @update:search-input="onChange($event)"
    >
    <template v-slot:append-outer v-if="hideAdd == undefined && !disabled">
        {{if LookupWithAdd}}<v-icon
          color="success"
          @click="onAdd()"
        >mdi-plus-box</v-icon>{{end}}
        <slot name="append"></slot>
    </template>
    </v-autocomplete>
    {{if LookupWithAdd}}<{{DialogComponent .}}  v-if="hideAdd == undefined" ref="dialog"/>{{end}}
  </div>
</template>
{{end}}
`

const vueDictionaryLookupTSTemplateHeader = `
{{define "TS-HEADER"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Watch } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, {{ListQuery}} } from '{{TypesFilePath .}}';
{{if LookupWithAdd}}import {{DialogComponent .}} from './{{DialogComponent .}}.vue';{{end}}
{{TypesFromTS}}
{{end}}
`

// TODO emit changed after input
const vueDictionaryLookupTSTemplateBody = `
{{define "TS"}}

@Component({
  name: "{{TypeName}}LookupComponent",
  components: {
    {{if LookupWithAdd}}{{DialogComponent .}}{{end}}
  }
})
export default class {{TypeName}}LookupComponent extends Vue {
  @Prop() value!: {{TypeName .}}|{{IDType .}}{{if CanBeMultiple}}|{{TypeName .}}[]|{{IDType .}}[]{{end}};
  @Prop() hint!: string;
  @Prop() label!: string;
  @Prop() readonly!: boolean;{{if CanBeMultiple}}
  @Prop({default:false}) multiple!: boolean;{{end}}
  @Prop({default:true}) returnObject!: boolean
  @Prop({default:undefined}) hideAdd: boolean|undefined
  @Prop({default:false}) disabled!: boolean;{{if DictWithQualifier .}}
  @Prop() qualifier: any;
  @Prop({default: true}) qualifiedByObject!: boolean; 
  @Prop() allowEmptyQualifier!: boolean;{{end}} 
  @Prop({default:()=>[]}) rules!: string[] | ((v:any)=>string|boolean)[];
  @Prop({default:()=>[]}) errorMessages!: string|string[];
  @Prop() filter!: (value: {{TypeName .}}) => boolean;
  @Prop() autoSelect!: ((value: {{TypeName .}}) => boolean)|string|number|boolean;
  @Prop({default: "disabled"}) disabledProperty!: string; 

  private selected: {{TypeName .}}|{{IDType .}}{{if CanBeMultiple}}|{{TypeName .}}[]|{{IDType .}}[]{{end}}|null = null;
  private items: {{TypeName .}}[] = [];
  private loading = false;
  private problem = "";
  
  @Watch('value') onValueChange() {
    this.selected = this.value;
  }
  @Emit('input') selectedChanged(): {{TypeName .}}|{{IDType .}}{{if CanBeMultiple}}|{{TypeName .}}[]|{{IDType .}}[]{{end}}|null {
    //delete (this.selected as any).__typename;
    this.emitChanged();
    return this.selected;
  }

  @Emit('change') emitChanged() {
    
  }
  @Watch('selected') onSelectedChanged() {
    this.selectedChanged();
  } {{if DictWithQualifier .}}
  @Watch('qualifier') onQualifierChanged() {
    this.load();
  }{{end}}
  created() {
    this.onCreated();
  }
  async onCreated() {
    await this.load();
    if(this.value != undefined)
      this.onValueChange();
  }
  async load() {
    {{if DictWithQualifier .}}if(!({{IsQualifierFilled}}) && !this.allowEmptyQualifier)
      return;{{end}}
    this.loading = true;
    this.items = [];
    this.problem = "";
    try {
      let res = await {{ListQuery}}({{ApolloClient}},{{ListQueryAttrs}});
      if(res) {
        this.items = res;
        if(this.filter) {
          this.items = this.items.filter(this.filter);
        }
        if(this.autoSelect) {
          if(typeof this.autoSelect == "function") {
            this.selected = this.items.find(this.autoSelect) || null;
          } else if(typeof this.autoSelect == "string") {
            this.selected = this.items.find(item => item[this.autoSelect as keyof {{TypeName .}}]) || null;
          } else if(typeof this.autoSelect == "number") {
            if(this.items && this.items.length > this.autoSelect)
              this.selected = this.items[this.autoSelect];
          } else if(typeof this.autoSelect == "boolean") {
            if(this.items && this.items.length)
              this.selected = this.items[0];
          } 
        }
      }
    } catch(exc) {
      this.problem = exc.toString();
    }
    this.loading = false;
  }
  
   {{if LookupWithAdd}}async onAdd() {
    try {
      let res = await (this.$refs.dialog as {{DialogComponent .}}).show(null);
      if(res) {
        this.load();
        this.selected = res;
        this.selectedChanged();
      }
    } catch(exc) {
      this.problem = exc.toString();
    }
  }{{end}}
  onChange(event: any) {
  }
}
</script>
{{end}}
`
const vueEntityLookupTSTemplateHeader = `
{{define "TS-HEADER"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Watch } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, {{LookupQuery}}, {{GetQuery .}}{{if HasFindType}}, {{FindQuery}}, {{FindTypeName}}{{end}}} from '{{TypesFilePath .}}';
import {{DialogComponent .}} from './{{DialogComponent .}}.vue';
{{end}}
`

// TODO onChange: use toString() if necessary
const vueEntityLookupTSTemplateBody = `
{{define "TS"}}

@Component({
  name: "{{TypeName}}LookupComponent",
  components: {
    {{DialogComponent .}}
  }
})
export default class {{TypeName}}LookupComponent extends Vue {
  @Prop() value!: {{TypeName .}}|{{IDType .}}{{if CanBeMultiple}}|{{TypeName .}}[]|{{IDType .}}[]{{end}};
  @Prop() hint!: string;
  @Prop() label!: string;
  @Prop() readonly!: boolean;
  @Prop({default:true}) returnObject!: boolean
  @Prop({default:undefined}) hideAdd: boolean|undefined
  @Prop({default:false}) disabled!: boolean;{{if CanBeMultiple}}
  @Prop({default:false}) multiple!: boolean;{{end}}
  @Prop({default:()=>[]}) rules!: string[] | ((v:any)=>string|boolean)[];
  @Prop({default:()=>[]}) errorMessages!: string|string[];{{if HasFindType}}
  @Prop({default:null}) query!:{{FindTypeName}}|null;{{end}}
  @Prop() filter!: (value: {{TypeName .}}) => boolean;
  @Prop({default: "disabled"}) disabledProperty!: string; 

  private selected: {{TypeName .}}|{{IDType .}}{{if CanBeMultiple}}|{{TypeName .}}[]|{{IDType .}}[]{{end}}|null = null;
  private items: {{TypeName .}}[] = [];
  private loading = false;
  private problem = "";
  private lastSearch: string|null = null;
  private searchString: string = "";
  private timer: any = null;
  
  @Watch('value') onValueChange() {
    this.selected = this.value;
    if(this.selected && !this.items.length) {
      if(this.returnObject)
      this.items = {{if CanBeMultiple}}this.multiple ? this.selected as {{TypeName .}}[] : {{end}}[this.selected as {{TypeName .}}];
    }
    
  }{{if HasFindType}}
  @Watch('query') onQueryChange() {
    this.search();
  }{{end}}
  @Emit('input') selectedChanged(): {{TypeName .}}|{{IDType .}}{{if CanBeMultiple}}|{{TypeName .}}[]|{{IDType .}}[]{{end}}|{{IDType .}}|null {
    return this.selected;
  }
  @Watch('selected') onSelectedChanged() {
    this.selectedChanged();
  }
  created() {
    if(this.value)
      this.onValueChange();
  }
  async search() {
    this.loading = true;
    this.items = [];
    this.problem = "";
    try { {{if HasFindType}}
      let useQuery = false;
      if(this.query) {
        let prop: keyof {{FindTypeName}};
        for(prop in this.query) {
          if(this.query[prop]) {
            useQuery = true;
          }
        }
      }
      if(useQuery) {
        this.items = await {{FindQuery}}(this.$apollo.getClient(), this.query!);
      } else if(this.searchString) { {{end}}
        this.lastSearch = this.searchString;
        let res = await {{LookupQuery}}({{ApolloClient}}, this.lastSearch);
        if(res) {
          this.items = res;
        }{{if HasFindType}}
      } {{end}}
			if(this.filter) {
          this.items = this.items.filter(this.filter);
      }
    } catch(exc) {
      this.problem = exc.toString();
    }
    this.loading = false;
  }
  
  async onAdd() {
    try {
      let res = await (this.$refs.dialog as {{DialogComponent .}}).show(null);
      if(res) {
        this.items = [res];
        this.selected = res;
        this.selectedChanged();
      }
    } catch(exc) {
      this.problem = exc.toString();
    }
  }
  doSearch() {
    this.timer = null;
    if(this.searchString && (!this.lastSearch || !this.searchString.startsWith(this.lastSearch))) {
      if(this.loading)
        this.onChange(this.searchString)
      else
        this.search();
    }
  }
  onChange(event: string) { {{if HasFindType}}
    if(this.query)
      return;{{end}}
    if(this.searchString == event {{if ItemTypeIsString}}|| this.selected && (this.selected as {{TypeName .}}).{{ItemText}} && (this.selected as {{TypeName .}}).{{ItemText}} == event{{end}})
      return
    if(this.timer)
      clearTimeout(this.timer);
    this.timer = setTimeout(()=> this.doSearch(), 500);
    this.searchString = event;
  }
  async fillSelectedFromId(id: {{IDType .}}) {
    try {
      this.selected = await {{GetQuery .}}({{ApolloClient}}, id);
      if(this.selected)
        this.items = [this.selected];
    } catch(exc) {
      this.problem = exc.toString();
    }
  }
}
</script>
{{end}}
`
