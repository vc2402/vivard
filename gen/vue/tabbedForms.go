package vue

import (
	"fmt"
	"os"
)

func (h *helper) generateGridForm(formName string, path ...string) error {

	baseTempl := htmlTablessFormTemplate
	if h.ctx.withTabs {
		baseTempl = htmlTabbedFormTemplate
	}
	formTempl := htmlFlexFormTemplate
	if h.ctx.useGrid {
		formTempl = htmlGridFormTemplate
	}
	h.parse(baseTempl).
		parse(formTempl).
		parse(newHtmlFormListTemplate).
		// parse(htmlFormCardTemplate).
		parse(htmlFormInputTemplate).
		parse(htmlFormTextInputTemplate).
		parse(htmlFormDateInputTemplate).
		parse(htmlFormMapInputTemplate).
		parse(htmlFormArrayInputTemplate).
		parse(htmlFormArrayAsListTemplate).
		parse(htmlFormArrayAsChipsTemplate).
		parse(htmlFormLookupInputTemplate).
		parse(htmlFormBoolInputTemplate).
		parse(vueTabbedFormTSTemplate).
		parse(newFormListTSTemplate).
		// parse(vueFormCardTSTemplate).
		parse(cssFormTemplate)
	if h.err != nil {
		return fmt.Errorf("Error while parsing form template: %v", h.err)
	}

	if h.e.FB(featureVueKind, fVKFormRequired) || (len(path) > 0 && path[0] != "") {
		// p := path.Join(h.outDir, h.e.Name+"GridForm.vue")
		p := h.e.FS(featureVueKind, fVKFormComponentPath)
		if len(path) > 0 && path[0] != "" {
			p = path[0]
		}
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %v", p, err)
		}
		defer f.Close()

		h.parse("<template>{{template \"FORM\" .}}</template>\n{{template \"FORM.TS\" .}}\n{{template \"CSS\" .}}\n")
		if h.err != nil {
			return fmt.Errorf("Error while parsing form file template: %v", h.err)
		}
		err = h.templ.Execute(f, h)
		if err != nil {
			return fmt.Errorf("Error while executing form template: %v", err)
		}
	}

	if h.e.FB(featureVueKind, fVKFormListRequired) || (len(path) > 1 && path[1] != "") {
		p := h.e.FS(featureVueKind, fVKFormListComponentPath)
		if len(path) > 1 && path[1] != "" {
			p = path[0]
		}
		if p == "" {
			return fmt.Errorf("FormList: form list requested but path not generated for %s", h.e.Name)
		}
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("FormList: Error opening file '%s': %v", p, err)
		}
		defer f.Close()

		h.parse("<template>{{template \"FORM-LIST\" .}}</template>\n{{template \"FORM-LIST.TS\" .}}\n")
		if h.err != nil {
			return fmt.Errorf("Error while parsing form list template: %v", h.err)
		}
		err = h.templ.Execute(f, h.e)
		if err != nil {
			return fmt.Errorf("Error while executing form template: %v", err)
		}
	}
	return nil
}

//Forms templates
var htmlTablessFormTemplate = `
{{define "FORM"}}
{{template "FORM_CONTENT" .}}
{{end}}
`
var htmlTabbedFormTemplate = `
{{define "FORM"}}
  <v-tabs>
    {{range Tabs}}<v-tab>{{TabLable .}}</v-tab>
    {{end}}
    {{range Tabs}}<v-tab-item>{{template "FORM_CONTENT" .}}</v-tab-item>
    {{end}}
  </v-tabs>
{{end}}
`

var htmlGridFormTemplate = `
{{define "FORM_CONTENT"}}
  <div class="d-flex flex-column">
    <slot name="pre-fields"></slot>
    {{range Rows .}}
      <v-row justify="space-between">
        {{range .}}
        <v-col {{GridColAttrs .}}>
          {{if IsID . false}}<div v-if="!isNew">{{"{{"}}value && value.{{FieldName .}}{{"}}"}}</div>{{end}}
          <div class="mx-5" {{if IsID . true}}v-if="isNew" {{else}} {{FieldAttrs .}} {{end}}>
            {{template "FORM_INPUT_FIELD" .}}
          </div>
        </v-col>
        {{end}}
      </v-row>
    {{end}}
    <v-btn v-if="!value && !disabled" text icon color="primary" @click="addValue">
      <v-icon>add</v-icon> {{Title}}
    </v-btn>
    <slot name="post-fields"></slot>
  </div>
{{end}}
`
var htmlFlexFormTemplate = `
{{define "FORM_CONTENT"}}
  <div class="d-flex flex-row flex-wrap justify-space-between align-baseline">
    <slot name="pre-fields"></slot>
    {{range (GetFields .)}}{{if IsID . false}}<div v-if="!isNew">{{"{{"}}value && value.{{FieldName .}}{{"}}"}}</div>{{end}}<div class="mx-5" {{if IsID . true}}v-if="isNew" {{else}} {{FieldAttrs .}} {{end}}>
      {{template "FORM_INPUT_FIELD" .}}
    </div>
    {{end}}
    <v-btn v-if="!value && !disabled" text icon color="primary" @click="addValue">
      <v-icon>add</v-icon> {{Title}}
    </v-btn>
    <slot name="post-fields"></slot>
  </div>
{{end}}
`

const vueTabbedFormTSTemplate = `
{{define "FORM.TS"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName}}, {{InstanceGeneratorName}}, {{TypesFromTS}} } from '{{TypesFilePath}}';
{{range RequiredComponents}}
import {{.}} from './{{.}}.vue'{{end}}
{{range AdditionalComponents}}
import {{.Comp}} from '{{.Imp}}';{{end}}

@Component({
  components:{
    {{range RequiredComponents}}
      {{.}},{{end}}
    {{range AdditionalComponents}}
      {{.Comp}},{{end}}
  }
})
export default class {{Name}}DialogComponent extends Vue {
  @Prop() value!: {{TypeName}} | undefined;
  @Prop({default:false}) isNew!: boolean;
  @Prop({default:false}) disabled!: boolean;
  
  @Emit("input")
  emitValue() {
    return this.value;
  }
  @Emit("change")
  emitChanged(fld: string) {
    return fld;
  }
  changed(fld: keyof {{TypeName}}) {
    //if(this.value[fld] == "")
    //  this.value[fld] = null;
    this.emitChanged(fld);
    this.emitValue();
  }
  addValue() {
    this.value = {{InstanceGenerator}}
  }
  {{range (GetFields .)}}{{if and (eq (FormComponentType .) "array") (ArrayAsList .)}}
  add{{FieldName .}}() {
    if(!this.value)
      this.addValue()
    if(!this.value!.{{FieldName .}})
      this.value!.{{FieldName .}} = [];
    this.value!.{{FieldName .}}.push({{InstanceGeneratorForField .}})
  }{{end}}
  {{end}}
}
</script>
{{end}}
`
const newHtmlFormListTemplate = `
{{define "FORM-LIST"}}
<div>
  <div v-if="value">
    <div v-for="(d, idx) in value" :key="idx" class="d-flex flex-row align-center justify-space-between">
      <{{FormComponent}}  :value="d" />
      <v-btn icon color="warning" @click="onDelItem(idx)"><v-icon>delete</v-icon></v-btn>
    </div>
  </div>
  <v-btn icon color="primary" @click="onAddItem"><v-icon>add</v-icon></v-btn>
</div>
{{end}}  
`
const newFormListTSTemplate = `
{{define "FORM-LIST.TS"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName}}, {{InstanceGeneratorName}} } from '{{TypesFilePath}}';
{{range RequiredComponents}}
import {{.}} from './{{.}}.vue'{{end}}

@Component({
  components:{
    {{range RequiredComponents}}
      {{.}},{{end}}
  }
})
export default class {{.Name}}FormListComponent extends Vue {
  @Prop({default:()=>[]}) value!: {{TypeName}}[];
  
  @Emit("input")
  emitValue() {
    return this.value;
  }
  @Emit("change")
  emitChanged() {
    return this.value;
  }
  onDelItem(idx: number) {
    this.value.splice(idx, 1);
    this.emitValue();
    this.emitChanged();
  }
  onAddItem() {
    if(!this.value) {
      this.value = [];
    }
    this.value.push({{InstanceGenerator}})
    this.emitValue();
    this.emitChanged();
  }
}
</script>
{{end}}
`
