package vue

import (
	"fmt"
	"os"
	"path/filepath"
)

func (h *helper) generateGridForm(formName string, compPath ...string) error {

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
		parse(newHtmlFormDisabledTemplate).
		// parse(htmlFormCardTemplate).
		parse(htmlFormInputTemplate).
		parse(htmlFormTextInputTemplate).
		parse(htmlFormTextAreaTemplate).
		parse(htmlFormDateInputTemplate).
		parse(htmlFormColorInputTemplate).
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
		return fmt.Errorf("error while parsing form template: %v", h.err)
	}

	if h.e.FB(featureVueKind, fVKFormRequired) || (len(compPath) > 0 && compPath[0] != "") {
		p := h.e.FS(featureVueKind, fVKFormComponentPath)
		if len(compPath) > 0 && compPath[0] != "" {
			p = compPath[0]
		}
		p = filepath.Join(h.outDir, p)
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("error opening file '%s': %v", p, err)
		}
		defer f.Close()

		h.parse("<template>{{template \"FORM\" .}}</template>\n{{template \"FORM.TS\" .}}\n{{template \"CSS\" .}}\n")
		if h.err != nil {
			return fmt.Errorf("error while parsing form file template: %v", h.err)
		}
		err = h.templ.Execute(f, h)
		if err != nil {
			return fmt.Errorf("error while executing form template: %v", err)
		}
	}

	if h.e.FB(featureVueKind, fVKFormListRequired) || (len(compPath) > 1 && compPath[1] != "") {
		p := h.e.FS(featureVueKind, fVKFormListComponentPath)
		if len(compPath) > 1 && compPath[1] != "" {
			p = compPath[0]
		}
		if p == "" {
			return fmt.Errorf("FormList: form list requested but path not generated for %s", h.e.Name)
		}
		p = filepath.Join(h.outDir, p)
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("FormList: Error opening file '%s': %v", p, err)
		}
		defer f.Close()

		h.parse("<template>{{template \"FORM-LIST\" .}}</template>\n{{template \"FORM-LIST.TS\" .}}\n")
		if h.err != nil {
			return fmt.Errorf("error while parsing form list template: %v", h.err)
		}
		err = h.templ.Execute(f, h.e)
		if err != nil {
			return fmt.Errorf("error while executing form template: %v", err)
		}
	}
	return nil
}

// Forms templates
var htmlTablessFormTemplate = `
{{define "FORM"}}
{{template "FORM_CONTENT" .}}
{{end}}
`
var htmlTabbedFormTemplate = `
{{define "FORM"}}
  <v-tabs>
    {{range Tabs}}<v-tab {{if NeedRolesSecurity}}v-if="hasAccess('{{TabID .}}')"{{else if NeedResourceSecurity}}{{if ne (ResourceForTab (TabID .)) ""}}v-if="{{TabID .}}Accessible"{{end}}{{end}}>{{TabLable .}}</v-tab>
    {{end}}
    {{range Tabs}}<v-tab-item {{if NeedRolesSecurity}}v-if="hasAccess('{{TabID .}}')"{{else if NeedResourceSecurity}}{{if ne (ResourceForTab (TabID .)) ""}}v-if="{{TabID .}}Accessible"{{end}}{{end}}>{{if ne (ComponentForTab (TabID .)) ""}}<{{(ComponentForTab (TabID .))}} v-model="value"/>{{else}}{{template "FORM_CONTENT" .}}{{end}}</v-tab-item>
    {{end}}
  </v-tabs>
{{end}}
`

const newHtmlFormDisabledTemplate = `{{define "DISABLED_IN_FORM"}}disabled===true || isFieldDisabled('{{FieldName .}}'){{end}}`

var htmlGridFormTemplate = `
{{define "FORM_CONTENT"}}
  <div class="d-flex flex-column" {{FormStyles}}>
    <slot name="pre-fields"></slot>
    {{range Rows .}}
      <v-row justify="space-between" align="baseline"{{if Compact}} no-gutters{{end}}>
        {{range .}}
        <v-col {{GridColAttrs .}}>
          {{if IsID .}}<div v-if="!isNew">{{"{{"}}value && value.{{FieldName .}}{{"}}"}}</div>
          {{if NotAuto .}}<div class="mx-2" v-if="isNew"> {{template "FORM_INPUT_FIELD" .}}</div>{{end}}
          {{else}}<div class="mx-2" {{FieldAttrs .}} >
            {{template "FORM_INPUT_FIELD" .}}
          </div>{{end}}
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
  <div class="d-flex flex-row {{FlexWrap}} {{FlexJustify}} align-baseline" {{FormStyles}}>
    <slot name="pre-fields"></slot>
    {{range (GetFields .)}}{{if IsID .}}<div v-if="!isNew" {{FlexFieldStyles .}}>{{"{{"}}value && value.{{FieldName .}}{{"}}"}}</div>{{if NotAuto .}}<div v-else class="mx-2" {{FieldAttrs .}} {{FlexFieldStyles .}}>{{template "FORM_INPUT_FIELD" .}}</div>{{end}}
    {{else}}<div class="mx-2" {{FieldAttrs .}} {{FlexFieldStyles .}}>
      {{template "FORM_INPUT_FIELD" .}}
    </div>{{end}}
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
import { {{TypeName}}, {{InstanceGeneratorName}} } from '{{TypesFilePath}}';
{{TypesFromTS}}
{{range RequiredComponents}}
import {{.Comp}} from '{{.Imp}}'{{end}}
{{range AdditionalComponents}}
import {{.Comp}} from '{{.Imp}}';{{end}}
{{if NeedSecurity}}{{SecurityImport}}{{end}}

@Component({
  name: "{{Name}}FormComponent",
  components:{
    {{range RequiredComponents}}
      {{.Comp}},{{end}}
    {{range AdditionalComponents}}
      {{.Comp}},{{end}}
  }
})
export default class {{Name}}FormComponent extends Vue {
  @Prop() value!: {{TypeName}} | undefined;
  @Prop({default:false}) isNew!: boolean;
  @Prop({default:false}) disabled!: boolean|{[key:string]:boolean};
  @Prop() validator: any;
{{if NeedSecurity}}{{SecurityInject}}{{end}}
  {{range Tabs}}{{if ne (ResourceForTab (TabID .)) ""}}{{TabID .}}Accessible = false;{{end}}
    {{end}}
{{if LateInitRequired}}
  beforeCreate() {
    {{range LateInitRequiredComponents}}this.$options.components!.{{.Comp}} = require('{{.Imp}}').default 
{{end}}
  }
{{end}}
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
  add{{FieldName . "U"}}() {
    if(!this.value)
      this.addValue()
    if(!this.value!.{{FieldName .}})
      this.$set(this.value!, "{{FieldName .}}",  []);
    this.value!.{{FieldName .}}!.push({{InstanceGeneratorForField .}})
  }

  remove{{FieldName . "U"}}(idx: number) {
    if(this.value && this.value.{{FieldName .}} && this.value.{{FieldName .}}[idx])
      this.value.{{FieldName .}}.splice(idx, 1);
  }{{end}}{{end}}

  {{if NeedRolesSecurity}}hasAccess(tab: string): boolean {
    let roles: string[] = [];
    switch(tab) {
      {{range Tabs}}{{if ne (RolesForTab (TabID .)) ""}}case '{{TabID .}}': roles = [{{RolesForTab (TabID .)}}]; break;{{end}}{{end}}
      default: return true;
    }
    for(let r of roles) {
      if(this.loginManager.hasRole(r))
        return true;
    }
    return false;
  }{{end}}
  {{if NeedResourceSecurity}}async requestAccessRights() {
    let resources: string[] = [{{range Tabs}}{{if ne (ResourceForTab (TabID .)) ""}}
      "{{ResourceForTab (TabID .)}}", {{end}}
    {{end}}];
    const result = await this.loginManager.getResources(resources);

    {{range Tabs}}{{if ne (ResourceForTab (TabID .)) ""}}this.{{TabID .}}Accessible = result.resource("{{ResourceForTab (TabID .)}}") && result.resource("{{ResourceForTab (TabID .)}}")!.checkAccessRight("r") || false;{{end}}
    {{end}}
  }
  mounted() {
    this.requestAccessRights();
  }{{end}}
  isFieldDisabled(field: string): boolean {
    return typeof this.disabled == "object" && this.disabled[field];
  }
}
</script>
{{end}}
`
const newHtmlFormListTemplate = `
{{define "FORM-LIST"}}
<div>
  <div v-if="value">
    <div v-for="(d, idx) in value" :key="idx" class="d-flex flex-row align-baseline justify-space-between">
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
import {{.Comp}} from '{{.Imp}}'{{end}}

@Component({
  name: "{{.Name}}FormListComponent",
  components:{
    {{range RequiredComponents}}
      {{.Comp}},{{end}}
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
