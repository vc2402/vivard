package vue

import (
	"fmt"
	"os"
	"path"
)

func (h *helper) generateForm(formName string) error {

	h.parse(htmlFormTemplate).
		parse(htmlFormListTemplate).
		parse(htmlFormCardTemplate).
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
		parse(htmlFormDisabledTemplate).
		parse(vueFormTSTemplate).
		parse(vueFormListTSTemplate).
		parse(vueFormCardTSTemplate).
		parse(cssFormTemplate)
	if h.err != nil {
		return fmt.Errorf("Error while parsing form template: %v", h.err)
	}

	if h.e.FB(featureVueKind, fVKFormRequired) {
		// p := path.Join(h.outDir, h.e.Name+"Form.vue")
		p := h.e.FS(featureVueKind, fVKFormComponentPath)
		p = path.Join(h.outDir, p)
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %v", p, err)
		}
		defer f.Close()

		h.parse("<template>{{template \"FORM\" .}}</template>\n{{template \"FORM.TS\" .}}\n{{template \"CSS\" .}}\n")
		if h.err != nil {
			return fmt.Errorf("Error while parsing form file template: %v", h.err)
		}
		err = h.templ.Execute(f, h.e)
		if err != nil {
			return fmt.Errorf("Error while executing form template: %v", err)
		}
	}
	if h.e.FB(featureVueKind, fVKFormListRequired) {
		p := h.e.FS(featureVueKind, fVKFormListComponentPath)
		p = path.Join(h.outDir, p)
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %v", p, err)
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

	if h.e.FB(featureVueKind, fVKCardRequired) {
		// p := path.Join(h.outDir, h.e.Name+"Form.vue")
		p := h.e.FS(featureVueKind, fVKCardComponentPath)
		p = path.Join(h.outDir, p)
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %v", p, err)
		}
		defer f.Close()

		h.parse("<template>{{template \"FORM-CARD\" .}}</template>\n{{template \"FORM-CARD.TS\" .}}\n{{template \"CSS\" .}}\n")
		if h.err != nil {
			return fmt.Errorf("Error while parsing card template: %v", h.err)
		}
		err = h.templ.Execute(f, h.e)
		if err != nil {
			return fmt.Errorf("Error while executing card template: %v", err)
		}
	}

	return nil
}

//Forms templates
var htmlFormTemplate = `
{{define "FORM"}}
  <div class="d-flex flex-row flex-wrap justify-space-around align-center">
    <slot name="pre-fields"></slot>
    {{range (GetFields .)}}{{if ShowInDialog .}}{{if IsID .}}<div v-if="!isNew">{{"{{"}}value.{{FieldName .}}{{"}}"}}</div>{{if NotAuto .}}<div class="mx-2" v-if="isNew">{{template "FORM_INPUT_FIELD" .}}</div>{{end}}
      {{else}}<div class="mx-2">
      {{template "FORM_INPUT_FIELD" .}}
    </div>{{end}}{{end}}
    {{end}}
    <v-btn v-if="!value" flat icon color="primary" @click="addValue">
      <v-icon>add</v-icon> {{Title .}}
    </v-btn>
    <slot name="post-fields"></slot>
  </div>
{{end}}
`

var htmlFormCardTemplate = `
{{define "FORM-CARD"}}
  <v-card >
    <v-card-title
      class="headline lighten-2"
      primary-title
    >
      <v-layout row justify-space-between>
        <slot name="pre-title"></slot>
        <slot name="title"><v-flex v-if="title">
          {{"{{title}}"}}
        </v-flex></slot>
        <slot name="post-title"></slot>
      </v-layout>
    </v-card-title>

    <v-card-text>
      <v-progress-linear v-if="loading" intermediate></v-progress-linear>
      <slot name="problem"><div v-if="problem">{{"{{problem}}"}}</div></slot>
      {{template "FORM" .}}
    </v-card-text>

    <v-divider></v-divider>

    <v-card-actions>
      <slot name="actions">
        <div v-if="withActions" class="d-flex flex-row justify-space-around">
          <v-btn
            color="primary"
            @click="close()"
          >
            {{"{{closeText}}"}}
          </v-btn>
          <v-btn
            color="primary"
            @click="close(true)"
          >
            {{"{{okText}}"}}
          </v-btn>
        </div>
      </slot>
    </v-card-actions>
  </v-card>
{{end}}
`

const htmlFormInputTemplate = `{{define "FORM_INPUT_FIELD"}}{{if ne (CustomComponent .) ""}}<{{CustomComponent .}} v-model="value.{{FieldName .}}" label="{{Label .}}" @change="changed('{{FieldName .}}')" :disabled="{{if Readonly .}}true{{else}}disabled{{end}}"/>
  {{else if eq (FormComponentType .) "string"}}{{template "TEXT_INPUT" .}}
  {{else if eq (FormComponentType .) "int"}}{{template "TEXT_INPUT" .}}
  {{else if eq (FormComponentType .) "float"}}{{template "TEXT_INPUT" .}}
  {{else if eq (FormComponentType .) "date"}}{{template "DATE_INPUT" .}}
  {{else if eq (FormComponentType .) "bool"}}{{template "BOOL_INPUT" .}}
  {{else if eq (FormComponentType .) "color"}}{{template "COLOR_INPUT" .}}
  {{else if eq (FormComponentType .) "map"}}{{template "MAP_INPUT" .}}
  {{else if eq (FormComponentType .) "array"}}{{template "ARRAY_INPUT" .}}
  {{else if eq (FormComponentType .) "text-area"}}{{template "TEXT_AREA_INPUT" .}}
  {{else}}{{template "LOOKUP_INPUT" .}}{{end}}{{end}}`

const htmlFormTextInputTemplate = `{{define "TEXT_INPUT"}}<v-text-field v-if="value"
    v-model="value.{{FieldName .}}"
    label="{{Label .}}" {{InputAttrs .}}
    @change="changed('{{FieldName .}}')"
    :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"{{if IsIcon .}}
    :append-icon="value.{{FieldName .}}"{{end}}
  >{{if FieldWithAppend .}}<template v-slot:append-outer>
     {{range AppendToField .}}{{.}}{{end}}
   </template>{{end}}{{if WithPrependIcon .}}<template v-slot:prepend>
     {{PrependIcon .}}</template>{{end}}{{if WithAppendIcon .}}<template v-slot:append>
     {{AppendIcon .}}</template>{{end}}</v-text-field>{{end}}`

const htmlFormTextAreaTemplate = `{{define "TEXT_AREA_INPUT"}}<v-textarea v-if="value"
    v-model="value.{{FieldName .}}"
    label="{{Label .}}" {{InputAttrs .}}
	auto-grow
	outlined
	rows="{{TextAreaRows .}}"
    @change="changed('{{FieldName .}}')"
    :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"{{if IsIcon .}}
    :append-icon="value.{{FieldName .}}"{{end}}
  >{{if FieldWithAppend .}}<template v-slot:append-outer>
     {{range AppendToField .}}{{.}}{{end}}
   </template>{{end}}{{if WithPrependIcon .}}<template v-slot:prepend>
     {{PrependIcon .}}</template>{{end}}{{if WithAppendIcon .}}<template v-slot:append>
     {{AppendIcon .}}</template>{{end}}</v-textarea>{{end}}`

const htmlFormDateInputTemplate = `{{define "DATE_INPUT"}}<{{CustomComponent "date"}}  v-if="value"
    v-model="value.{{FieldName .}}"
    label="{{Label .}}"
    @change="changed('{{FieldName .}}')"
    :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"
    {{ConponentAddAttrs .}}
  ></{{CustomComponent "date"}}>{{end}}`
const htmlFormColorInputTemplate = `{{define "COLOR_INPUT"}}<{{CustomComponent "color"}}  v-if="value"
    v-model="value.{{FieldName .}}"
    label="{{Label .}}"
    @change="changed('{{FieldName .}}')"
    :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"
  ></{{CustomComponent "color"}}>{{end}}`
const htmlFormMapInputTemplate = `{{define "MAP_INPUT"}}<{{CustomComponent "map"}}  v-if="value"
    v-model="value.{{FieldName .}}"
    label="{{Label .}}"
    @change="changed('{{FieldName .}}')"
    :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"
    {{ConponentAddAttrs .}}
  ></{{CustomComponent "map"}}>{{end}}`
const htmlFormArrayInputTemplate = `{{define "ARRAY_INPUT"}}{{if ArrayAsLookup .}}<{{LookupComponent . true}}  v-if="value" v-model="value.{{FieldName .}}" label="{{Label .}}" @change="changed('{{FieldName .}}')" {{LookupAttrs .}} :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"{{if HideAddForLookup .}} :hideAdd="true"{{end}}/>{{else if ArrayAsList .}}{{template "ARRAY_AS_LIST" .}}{{else if ArrayAsChips .}}{{template "ARRAY_AS_CHIPS" .}}{{end}}{{end}}`

const htmlFormLookupInputTemplate = `{{define "LOOKUP_INPUT"}} 
<{{LookupComponent . true}}  
  v-if="value" 
  v-model="value.{{FieldName .}}" 
  label="{{Label .}}" 
  @change="changed('{{FieldName .}}')" 
  {{LookupAttrs .}} 
  :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"{{if ByRefField .}}
  :returnId="true"{{end}}{{if HideAddForLookup .}}
  :hideAdd="true"{{end}}>{{if FieldWithAppend .}}
    <template v-slot:append>{{range AppendToField .}}
      {{.}}{{end}}
    </template>{{end}}
</{{LookupComponent . true}}>{{end}}`

const htmlFormBoolInputTemplate = `{{define "BOOL_INPUT"}}<v-checkbox  v-if="value" v-model="value.{{FieldName .}}" label="{{Label .}}" @change="changed('{{FieldName .}}')" :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}"/>{{end}}`

const htmlFormArrayAsListTemplate = `{{define "ARRAY_AS_LIST"}}<div class="d-flex flex-column mt-2">
  <div class="d-flex flex-row justify-space-around"><h3>{{Label .}}</h3>{{if not (Readonly .)}}<v-btn if="!disabled" text color="primary" @click="add{{FieldName .}}"><v-icon>add</v-icon> {{Label .}}</v-btn>{{end}}</div>
  <div v-if="value && value.{{FieldName .}}">
    <div class="d-flex flex-column" v-for="(it, idx) in value.{{FieldName .}}"  :key="idx">
      <div class="d-flex flex-row align-center mt-3">
        <div class="flex-grow-1 flex-shrink-0">
          <{{LookupComponent . true}} v-model="value.{{FieldName .}}[idx]" :disabled="{{if Readonly .}}true{{else}}{{template "DISABLED_IN_FORM" .}}{{end}}" @change="changed('{{FieldName .}}')"/>
        </div>
        <div class="flex-grow-0 flex-shrink-1 ml-2">
          <v-icon if="!disabled" color="error" @click="remove{{FieldName .}}(idx)" title="delete">mdi-close-circle</v-icon>
        </div>
      </div>
      <v-divider v-if="idx < value.{{FieldName .}}.length"></v-divider>
    </div>
  </div>
  </div>
 {{end}}`

const htmlFormArrayAsChipsTemplate = `{{define "ARRAY_AS_CHIPS"}}<div class="d-flex flex-row align-center" v-if="value">
  <span class="mr-3">{{Label .}}:</span>
  <v-chip-group column>
    <v-chip v-for="(key, idx) in value.{{FieldName .}}" :key="key" close @click:close="value.{{FieldName .}}.splice(idx, 1)" color="primary">
    {{"{{"}}key{{"}}"}}
    </v-chip>
  </v-chip-group>
  <v-text-field
    class="mx-3"
    ref="new{{FieldName .}}Input"
    label="Add" 
    @keydown.enter="value.{{FieldName .}}?value.{{FieldName .}}.push($refs.new{{FieldName .}}Input.internalValue):value.{{FieldName .}}=[$refs.new{{FieldName .}}Input.internalValue]; $refs.new{{FieldName .}}Input.internalValue = ''"
    :disabled="{{template "DISABLED_IN_FORM" .}}"
  >
    <template v-slot:append-outer>
      <v-icon
        :disabled="!$refs.new{{FieldName .}}Input||!$refs.new{{FieldName .}}Input.internalValue"
        color="success"
        @click="value.{{FieldName .}}?value.{{FieldName .}}.push($refs.new{{FieldName .}}Input.internalValue):value.{{FieldName .}}=[$refs.new{{FieldName .}}Input.internalValue]; $refs.new{{FieldName .}}Input.internalValue = ''"
      >mdi-plus-box</v-icon>
    </template>
  </v-text-field>
</div>
{{end}}`

const htmlFormDisabledTemplate = `{{define "DISABLED_IN_FORM"}}disabled{{end}}`

const vueFormTSTemplate = `
{{define "FORM.TS"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, {{InstanceGeneratorName .}} } from '{{TypesFilePath .}}';
{{range RequiredComponents}}
import {{.Comp}} from '{{.Imp}}'{{end}}
{{range AdditionalComponents}}
import {{.Comp}} from '{{.Imp}}';{{end}}

@Component({
  name: "{{.Name}}DialogComponent",
  components:{
    {{range RequiredComponents}}
      {{.Comp}},{{end}}
    {{range AdditionalComponents}}
      {{.Comp}},{{end}}
  }
})
export default class {{.Name}}DialogComponent extends Vue {
  @Prop() value!: {{TypeName .}} | undefined;
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
  changed(fld: keyof {{TypeName .}}) {
    //if(this.value[fld] == "")
    //  this.value[fld] = null;
    this.emitChanged(fld);
    this.emitValue();
  }
  addValue() {
    this.value = {{InstanceGenerator .}}
  }
}
</script>
{{end}}
`

const htmlFormListTemplate = `
{{define "FORM-LIST"}}
<div>
  <div v-if="value">
    <div v-for="(d, idx) in value" :key="idx" class="d-flex flex-row align-center justify-space-between">
      <{{FormComponent .}}  :value="d" />
      <v-btn icon color="warning" @click="onDelItem(idx)"><v-icon>delete</v-icon></v-btn>
    </div>
  </div>
  <v-btn icon color="primary" @click="onAddItem"><v-icon>add</v-icon></v-btn>
</div>
{{end}}  
`
const vueFormListTSTemplate = `
{{define "FORM-LIST.TS"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, {{InstanceGeneratorName .}} } from '{{TypesFilePath .}}';
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
  @Prop({default:()=>[]}) value!: {{TypeName .}}[];
  
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
    this.value.push({{InstanceGenerator .}})
    this.emitValue();
    this.emitChanged();
  }
}
</script>
{{end}}
`
const vueFormCardTSTemplate = `
{{define "FORM-CARD.TS"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, New{{TypeName .}}Instance } from '{{TypesFilePath .}}';
{{range RequiredComponents}}
import {{.Comp}} from '{{.Imp}}'{{end}}
{{range AdditionalComponents}}
import {{.Comp}} from '{{.Imp}}';{{end}}

@Component({
  name: "{{.Name}}DialogComponent",
  components:{
    {{range RequiredComponents}}
      {{.Comp}},{{end}}
    {{range AdditionalComponents}}
      {{.Comp}},{{end}}
  }
})
export default class {{.Name}}DialogComponent extends Vue {
  @Prop() value!: {{TypeName .}} | undefined;
  @Prop({default:false}) isNew!: boolean;
  @Prop({default:""}) title!: string;
  @Prop({default:"OK"}) okText!: string;
  @Prop({default:"Cancel"}) cancelText!: string;
  @Prop({default:false}) withActions!: boolean;
  @Prop({default:false}) loading!: boolean;
  @Prop() problem!: string | undefined;
  
  @Emit("action")
  close(ok: boolean) {
    return ok
  }
  @Emit("input")
  emitValue() {
    return this.value;
  }
  @Emit("change")
  emitChanged(fld: string) {
    return fld;
  }
  changed(fld: keyof {{TypeName .}}) {
    if(this.value[fld] == "")
      this.value[fld] = null;
    this.emitChanged(fld);
    this.emitValue();
  }
}
</script>
{{end}}
`
const cssFormTemplate = `
{{define "CSS"}}
{{end}}}
`
