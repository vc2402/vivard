package vue

import (
	"fmt"
	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
	"os"
	"path"
	"text/template"
)

func (cg *VueCLientGenerator) generateEnum(outDir string, e *gen.Enum) error {
	fileName := e.Features.String(featureVueKind, fVKLookupComponentPath)
	p := path.Join(outDir, fileName)
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("while opening file for ViewComponent for %s: %v", e.Name, err)
	}
	defer f.Close()
	templ := template.New("ENUM").
		Funcs(cg.getEnumFuncMap(e))

	templ, err = templ.Parse(htmlEnumLookupTemplate)
	if err != nil {
		return err
	}
	templ, err = templ.Parse(vueEnumLookupTSTemplate)
	if err != nil {
		return err
	}
	templ, err = templ.Parse("{{template \"HTML\" .}}\n{{template \"TS\" .}}\n")
	if err != nil {
		return err
	}
	err = templ.Execute(f, e)
	if err != nil {
		return fmt.Errorf("while executing template for Enum '%s': %v", e.Name, err)
	}
	return nil
}

func (cg *VueCLientGenerator) getEnumFuncMap(e *gen.Enum) template.FuncMap {
	return template.FuncMap{
		"TypeName": func(e *gen.Enum) string {
			return e.Features.String(js.Features, js.FName)
		},
		"ComponentName": func(e *gen.Enum) string {
			return e.Features.String(featureVueKind, fVKLookupComponent)
		},
		"TypesFilePath": func(arg ...interface{}) string {
			fp, _ := cg.getTypesPath(e)
			return fp
		},
		"FieldLabel": func(ef *gen.EnumField) string {
			return ef.Name
		},
		"DefaultValue": func(ef *gen.Enum) string {
			if len(ef.Fields) > 0 {
				return ef.Fields[0].Name
			}
			return "null"
		},
	}
}

var htmlEnumLookupTemplate = `
{{define "HTML"}}
<template>
  <div class="flex-row">
    <v-select
      v-model="selected"
      :hint="hint"
      :items="items"
      :readonly="readonly"
      :disabled="disabled"
      :label="label"
      :return-object="false"
      hide-no-data
    >
    </v-select>
  </div>
</template>
{{end}}
`

const vueEnumLookupTSTemplate = `
{{define "TS"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Watch } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, {{range .Fields}}{{.Name}},{{end}} } from '{{TypesFilePath .}}';

@Component({
  name: "{{ComponentName .}}",
})
export default class {{ComponentName .}} extends Vue {
  @Prop() value!: {{TypeName .}};
  @Prop() hint!: string;
  @Prop() label!: string;
  @Prop() readonly!: boolean;
  @Prop({default:false}) disabled!: boolean; 

  private selected: {{TypeName .}}|null = {{DefaultValue .}};
  private items = [
  {{range .Fields}}{text: "{{FieldLabel .}}", value: {{.Name}}},{{end}}
  ];
  
  created() {
    this.onValueChange();
  }
  
  @Watch('value') onValueChange() {
    this.selected = this.value;
  }
  @Emit('input') selectedChanged(): {{TypeName .}}|null {
    this.emitChanged();
    return this.selected;
  }

  @Emit('change') emitChanged() {
    
  }
  @Watch('selected') onSelectedChanged() {
    this.selectedChanged();
  }
}
</script>
{{end}}
`
