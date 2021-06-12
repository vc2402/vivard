package vue

import (
	"fmt"
	"os"
)

func (h *helper) generateHistoryComponent() error {
	htmlHistoryTemplate := htmlHistoryToolTipOnHoverTemplate // htmlHistoryTooltipOnClickTemplate
	h.parse(htmlHistoryTemplate).
		parse(vueHistoryTSTemplate).
		parse(cssFormTemplate)
	if h.err != nil {
		return fmt.Errorf("Error while parsing form template: %v", h.err)
	}
	p := h.e.FS(featureVueKind, fVKHistComponentPath)
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("Error opening file '%s': %v", p, err)
	}
	defer f.Close()

	h.parse("<template>{{template \"COMPONENT\" .}}</template>\n{{template \"TS\" .}}\n{{template \"CSS\" .}}\n")
	if h.err != nil {
		return fmt.Errorf("Error while parsing form file template: %v", h.err)
	}
	err = h.templ.Execute(f, h)
	if err != nil {
		return fmt.Errorf("Error while executing form template: %v", err)
	}
	return nil
}

var htmlHistoryToolTipOnHoverTemplate = `
{{define "COMPONENT"}}
  <div>
    <v-tooltip
      :open-on-hover="true"
      color="#eeeeee"
      left
    >
      <template v-slot:activator="{ on, attrs }">
        <v-icon
          :color="color"
          v-on="on"
          v-bind="attrs"
        >mdi-history</v-icon>
      </template>
      <div class="d-flex flex-column">
        <div v-if="items">
          <div class="d-flex flex-column" v-for="(it, idx) in items"  :key="idx">
            <{{ViewComponent}} :item="it"/>
          </div>
        </div>
      </div>
    </v-tooltip>
  </div>
{{end}}
`

var htmlHistoryTooltipOnClickTemplate = `
{{define "COMPONENT"}}
  <div>
    <v-tooltip
      :open-on-hover="false"
      :open-on-click="true"
      color="#eeeeee"
      left
    >
      <template v-slot:activator="{ on }">
        <v-icon
          :color="color"
          @click="on.click"
        >mdi-history</v-icon>
      </template>
      <div class="d-flex flex-column">
        <div v-if="items">
          <div class="d-flex flex-column" v-for="(it, idx) in items"  :key="idx">
            <{{ViewComponent}} :item="it"/>
          </div>
        </div>
      </div>
    </v-tooltip>
  </div>
{{end}}
`

const vueHistoryTSTemplate = `
{{define "TS"}}
<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator';
import { {{TypeName}}, } from '{{TypesFilePath}}';
{{range RequiredComponents}}
import {{.}} from './{{.}}.vue'{{end}}

@Component({
  components:{
    {{range RequiredComponents}}
      {{.}},{{end}}
  }
})
export default class {{Name}}HistoryComponent extends Vue {
  @Prop() items: {{TypeName}}[] | undefined;
  @Prop({default:"primary"}) color!: string;
  
}
</script>
{{end}}
`
