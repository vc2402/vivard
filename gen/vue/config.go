package vue

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
)

const TabAsSpace = "  "

func (ch *configHelper) generate() error {
	tabs := ch.e.Annotations.GetBoolAnnotationDef(gen.AnnotationConfig, vueTabsAnnotation, false)
	if tabs {
		ch.parse(configTabsHTMLTemplate).
			parse(vueConfigTabsTSTemplate)

	} else {
		ch.parse(configTreeHTMLTemplate).
			parse(vueConfigTreeTSTemplate)
	}
	ch.parse(cssConfigTemplate).
		parse(configListItemHTMLTemplate).
		parse(configValueItemHTMLTemplate).
		parse(configDictEditHTMLTemplate).
		parse(configValueHTMLTemplate).
		parse(vueConfigTSCommonTemplate)
	if ch.err != nil {
		return fmt.Errorf("Error while parsing config template: %v", ch.err)
	}
	p := ch.e.FS(featureVueKind, fVKConfComponentPath)
	p = path.Join(ch.outDir, p)

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("Error opening file '%s': %v", p, err)
	}
	defer f.Close()

	ch.parse("<template>{{template \"TEMPL_CONFIG\" .}}</template>\n{{template \"TS_CONFIG\" .}}\n{{template \"CSS\" .}}\n")
	if ch.err != nil {
		return fmt.Errorf("Error while parsing config file template: %v", ch.err)
	}
	err = ch.templ.Execute(f, ch.e)
	if err != nil {
		return fmt.Errorf("Error while executing form template: %v", err)
	}
	return nil
}

const configTabsHTMLTemplate = `
{{define "TEMPL_CONFIG"}}
  <v-tabs>
    {{range .Fields}} 
    <v-tab> {{Label .}}</v-tab>{{end}}
    <v-btn v-if="!value" flat icon color="primary" @click="addValue">
      <v-icon>apply</v-icon> {{Title .}}
    </v-btn>
  </v-tabs>
  <v-tabs-items>
  {{range .Fields}} 
    <v-tab-item> {{Label .}}</v-tab-item>{{end}}
  </v-tabs-items>
{{end}}
`
const configTreeHTMLTemplate = `
{{define "TEMPL_CONFIG"}}
	<div class="d-flex flex-row" style="height:100%;">
      <div class="d-flex flex-column px-8 pt-2" style="height:100%; overflow-y: auto">
        <v-treeview
          :items="items"
					rounded
					hoverable
          dense
					activatable
					return-object
					@update:active="activeChanged"
        ></v-treeview>
      </div>
      <v-divider vertical></v-divider>
      <div 
        class="d-flex flex-column px-8 py-4"
      >
			<div  class="d-flex flex-row py-4">
				<v-btn text @click="save" color="primary">  <v-icon>mdi-content-save-all</v-icon> Save</v-btn>
				<v-btn text @click="load" color="primary">  <v-icon>mdi-refresh</v-icon> Reload</v-btn>
			</div>
      <v-divider></v-divider>
			
        <template v-if="!active">
          Select the item
        </template>
        <template v-else>
					{{range Leafs}}<div v-if="active.id == '{{.}}'">
						{{template "CONF_VALUE" (Leaf .)}}
					</div>{{end}}
        </template>
      </div>
  </div>
{{end}}
`

const vueConfigTreeTSTemplate = `
{{define "TS_CONFIG"}}
{{template "TS_CONFIG_COMMON" .}}
export default class {{.Name}}TreeComponent extends Vue {
  items: TreeItem[] = [];
	active: TreeItem|null = null;
	value: {{TypeName .}}|null = null;
	loading = false;
  
	created() {
		this.items = {{GetTreeItems .}};
		this.load();
	}
	activeChanged(act: TreeItem[]) {
		if(!act.length || !act[0].leaf) {
			this.active = null;
		} else {
			this.active = act[0];
		}
	}
	async load() {
    this.loading = true;
    try {
      this.value = await {{GetQuery .}}({{ApolloClient}});
    } catch(exc) {
      console.log("problem: ", exc);
    }
    this.loading = false;
  }
	async save() {
		if(this.value) {
			this.loading = true;
			try {
				await {{SaveQuery .}}({{ApolloClient}}, this.value);
			} catch(exc) {
				console.log("problem: ", exc);
			}
			this.loading = false;
		}
  }
}
</script>
{{end}}
`

const configValueHTMLTemplate = `
{{define "CONF_VALUE"}}
	{{if eq .Tip 0}}{{template "CONFIG_LIST_ITEM" .}}
	{{else if eq .Tip 1}}{{template "CONFIG_VALUE_ITEM" .}}
	{{else if eq .Tip 2}}{{template "CONFIG_DICT_EDIT" .}}{{end}}
{{end}}
`
const configListItemHTMLTemplate = `
{{define "CONFIG_LIST_ITEM"}}
	<{{FormListComponent .}} v-model="value{{.Path}}"/>
{{end}}
`
const configValueItemHTMLTemplate = `
{{define "CONFIG_VALUE_ITEM"}}
  <{{FormComponent .}} :value="value{{.Path}}"/>
{{end}}
`
const configDictEditHTMLTemplate = `
{{define "CONFIG_DICT_EDIT"}}
  <{{DictComponent .}}/>
{{end}}
`

const vueConfigTSCommonTemplate = `
{{define "TS_CONFIG_COMMON"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, {{GetQuery .}}, {{SaveQuery .}}, {{InstanceGeneratorName .}} } from '{{TypesFilePath .}}';
{{range RequiredComponents}}
import {{.Comp}} from '{{.Imp}}';{{end}}
{{range AdditionalComponents}}
import {{.Comp}} from '{{.Imp}}';{{end}}

type TreeItem = {id:string, name:string, leaf: boolean, children?: TreeItem[]};

@Component({
  components:{
    {{range RequiredComponents}}
      {{.Comp}},{{end}}
    {{range AdditionalComponents}}
      {{.Comp}},{{end}}
  }
})
{{end}}
`

const vueConfigTabsTSTemplate = `
{{define "TS_CONFIG"}}
{{template "TS_CONFIG_COMMON" .}}
export default class {{.Name}}TabsComponent extends Vue {
  @Prop() value!: {{TypeName .}} | undefined;
  @Prop({default:false}) isNew!: boolean;
  

}
</script>
{{end}}
`

const cssConfigTemplate = `
{{define "CSS"}}
{{end}}}
`

type configHelper struct {
	templ      *template.Template
	e          *gen.Entity
	cg         *VueCLientGenerator
	components map[string]vcComponentDescriptor
	outDir     string
	err        error
}

type leafType int

const (
	ltList leafType = iota
	ltForm
	ltDictionary
)

type leafDescriptor struct {
	ID   string
	Tip  leafType
	Ent  *gen.Entity
	Path string
}

func (cg *VueCLientGenerator) newConfigHelper(name string, e *gen.Entity, outDir string) (*configHelper, error) {
	fp, ok := e.Features.GetString(js.Features, js.FFilePath)
	if !ok {
		return nil, fmt.Errorf("file path not set for %s", e.Name)
	}
	tn := path.Base(fp)
	ext := path.Ext(tn)
	if ext != "" {
		tn = tn[:len(tn)-len(ext)]
	}
	typesPath := path.Join("..", "..", "..", "/types", tn)
	//components := map[string]string{}
	customComponents := map[string]vcCustomComponentDescriptor{}
	leafs := map[string]leafDescriptor{}
	tree := cg.buildLeafs(e, leafs)
	th := &configHelper{templ: template.New(name), cg: cg, e: e, outDir: outDir, components: map[string]vcComponentDescriptor{}}
	funcs := template.FuncMap{
		"GetTreeItems": func(e *gen.Entity) string {
			return tree
		},
		"Leafs": func() []string {
			ret := make([]string, len(leafs))
			i := 0
			for l := range leafs {
				ret[i] = l
				i++
			}
			return ret
		},
		"Leaf": func(id string) leafDescriptor {
			return leafs[id]
		},
		"Label": func(f *gen.Field) string { return f.Name },
		"FieldName": func(f *gen.Field) string {
			return f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		},
		"AttrName":      func(f *gen.Field) string { return cg.getJSAttrNameForDisplay(f, false, false) },
		"TableAttrName": func(f *gen.Field) string { return cg.getJSAttrNameForDisplay(f, true, false) },
		"TableIconName": func(f *gen.Field) string { return cg.getJSAttrNameForDisplay(f, true, true) },
		"NeedIconForTable": func(f *gen.Field) bool {
			return f.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false)
		},
		"TypeName": func(e *gen.Entity) string {
			return e.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		},
		"FieldType": func(f *gen.Field) string {
			//return f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationType, "")
			return f.FS(js.Features, js.FType)
		},
		"FormComponent": func(leaf leafDescriptor) string {
			cmp := leaf.Ent.FS(featureVueKind, fVKFormComponent)
			p := leaf.Ent.FS(featureVueKind, fVKFormComponentPath)
			th.addComponent(cmp, p, leaf.Ent)
			return cmp
		},
		"FormListComponent": func(leaf leafDescriptor) string {
			cmp := leaf.Ent.FS(featureVueKind, fVKFormListComponent)
			p := leaf.Ent.FS(featureVueKind, fVKFormListComponentPath)
			th.addComponent(cmp, p, leaf.Ent)
			return cmp
		},
		"DictComponent": func(leaf leafDescriptor) string {
			cmp := leaf.Ent.FS(featureVueKind, fVKDictEditComponent)
			p := leaf.Ent.FS(featureVueKind, fVKDictEditComponentPath)
			th.addComponent(cmp, p, leaf.Ent)
			return cmp
		},
		"GetQuery": func(e *gen.Entity) string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationGet)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationGet])
		},
		"SaveQuery": func(e *gen.Entity) string {
			name, err := cg.desc.Project.CallFeatureFunc(e, js.Features, js.FFunctionName, gen.GQLOperationSet)
			if err != nil {
				cg.b.AddError(err)
			}
			return name.(string)
			//return e.Features.String(gen.GQLFeatures, gen.GQLOperationsAnnotationsTags[gen.GQLOperationSet])
		},
		"TypesFilePath": func(e *gen.Entity) string {
			return typesPath
		},
		//TODO: get title from annotations
		"Title": func(e *gen.Entity) string { return e.Name },
		"FormComponentType": func(f *gen.Field) string {
			if _, ok := f.Parent().Annotations[gen.AnnotationFind]; ok {
				fld, _ := f.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
				return fld.Type.Type
			}
			return f.Type.Type
		},
		"LookupComponent": func(f *gen.Field, addToRequired bool) string {
			tip := f.Type
			if _, ok := f.Parent().Annotations[gen.AnnotationFind]; ok {
				fld, _ := f.Features.GetField(gen.FeaturesAPIKind, gen.FAPIFindFor)
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
					ud := f.FS(featureVueKind, fVKUseInDialog)
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
					if lc != "" && lcp != "" {
						if addToRequired {
							th.addComponent(lc, lcp, t.Entity())
						}
					}
				}
			}

			if lc == "" || lcp == "" {
				cg.desc.AddWarning(fmt.Sprintf("at %v: lookupComponent not found for field %s", f.Pos, f.Name))
			}
			return lc
		},
		"RequiredComponents":   func() map[string]vcComponentDescriptor { return th.components },
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
			if f.Annotations.GetBoolAnnotationDef(vueTableAnnotation, vtaUseIcon, false) {
				return "icon"
			}
			if !f.Type.Complex {
				return fromAnnotations(f.Annotations, f.Type.Type)
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
		"CanBeMultiple": func(e *gen.Entity) bool {
			// if refsManyToMany, ok := e.Features.GetBool(gen.FeaturesCommonKind, gen.FCRefsAsManyToMany); ok && refsManyToMany {
			// 	return true
			// }
			// return false
			return true
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
			return
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
		// "RequiresInputField": func(f *gen.Field) bool {
		// 	return !f.Annotations.GetBoolAnnotationDef(vueFormAnnotation, vueAnnotationIgnore, false)
		// },
	}

	th.templ.Funcs(funcs)
	return th, nil
}

func (th *configHelper) parse(str string) *configHelper {
	if th.err != nil {
		return th
	}
	th.templ, th.err = th.templ.Parse(str)
	return th
}

func (th *configHelper) addComponent(cmp string, p string, entity *gen.Entity) {
	if cmp != "" && p != "" {
		if p[0] != '@' && p[0] != '.' && !path.IsAbs(p) {
			if entity.File.Package == th.e.File.Package {
				if entity.File.Name == th.e.File.Name {
					p = "." + string(os.PathSeparator) + p
				} else {
					p = path.Join("..", entity.File.Name, p)
				}
			} else {
				p = path.Join("..", "..", entity.File.Package, entity.File.Name, p)
			}
		}
		th.components[cmp] = vcComponentDescriptor{
			Comp: cmp,
			Imp:  p,
		}
	}
}

func (cg *VueCLientGenerator) getTreeItem(prefix string, e *gen.Entity, tabs string, leafs map[string]leafDescriptor, tip leafType, path string, fromField *gen.Field) string {
	grp := e.Annotations.GetBoolAnnotationDef(gen.AnnotationConfig, gen.AnnCfgGroup, false) || e.HasModifier(gen.TypeModifierConfig)
	val := e.Annotations.GetBoolAnnotationDef(gen.AnnotationConfig, gen.AnnCfgValue, false) || e.HasModifier(gen.TypeModifierDictionary)
	name := e.Annotations.GetStringAnnotationDef(vueAnnotation, vcaLabel, e.Name)
	if fromField != nil {
		name = fromField.Annotations.GetStringAnnotationDef(vueAnnotation, vcaLabel, name)
	}
	if grp {
		children := strings.Builder{}
		for _, f := range e.Fields {
			tn := f.Type.Type
			if f.Type.Array != nil {
				tn = f.Type.Array.Type
			}
			t, ok := e.Pckg.FindType(tn)
			if !ok {
				cg.b.AddWarning(fmt.Sprintf("reference type not found: %s", tn))
				continue
			}
			var tip leafType
			if t.Entity().HasModifier(gen.TypeModifierDictionary) {
				tip = ltDictionary
			} else {
				tip = ltForm
				if f.Type.Array != nil {
					tip = ltList
				}
			}
			p := fmt.Sprintf("%s.%s", path, f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, ""))
			children.WriteString(tabs)
			children.WriteString(cg.getTreeItem(prefix+"_"+f.Name, t.Entity(), tabs+TabAsSpace, leafs, tip, p, f))
		}
		return fmt.Sprintf("%s{id:'%s', name: '%s', leaf: false, children: \n%s[\n%s%s]\n%s},", tabs, prefix, name, tabs+TabAsSpace, children.String(), tabs+TabAsSpace, tabs)
	} else if val {
		leafs[prefix] = leafDescriptor{ID: prefix, Ent: e, Tip: tip, Path: path}
		return fmt.Sprintf("%s{id:'%s', name: '%s', leaf: true},\n", tabs, prefix, name)
	} else {
		cg.b.AddWarning(fmt.Sprintf("vue-config: at %v: there is no 'config' annotation for non-dictionary type %s", e.Pos, e.Name))
		return ""
	}
}

func (cg *VueCLientGenerator) buildLeafs(e *gen.Entity, leafs map[string]leafDescriptor) string {
	return fmt.Sprintf("[\n%s\n%s]", cg.getTreeItem(e.Name, e, TabAsSpace, leafs, ltForm, "", nil), TabAsSpace)
}
