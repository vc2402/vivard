package vue

import (
	"fmt"
	"os"
)

//TODO not to include create, set functions for readonly dictionaries

func (h *helper) createDialog(path ...string) error {
	p := h.e.FS(featureVueKind, fVKDialogComponentPath)
	if len(path) > 0 {
		p = path[0]
	}

	f, err := os.Create(p)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("while opening file for DialogComponent for %s: %v", h.e.Name, err)
	}
	h.parse(htmlGridDialogTemplate).
		parse(vueGridDialogTSTemplate).
		parse("{{template \"HTML\" .}}\n{{template \"TS\" .}}\n")
	if h.err != nil {
		return fmt.Errorf("error while parsing template: %v", h.err)
	}
	err = h.templ.Execute(f, h.e)
	if err != nil {
		return fmt.Errorf("while executing template for DialogComponent for %s: %v", h.e.Name, err)
	}
	return nil
}

//Dialog
var htmlGridDialogTemplate = `
{{define "HTML"}}
<template>
  <v-dialog
    v-model="showDialog"
    persistent
    scrollable
    :fullscreen="$vuetify.breakpoint.sm"
    :max-width="{{DialogWidth}}"
  >
    <v-card >
      <v-card-title
        class="headline lighten-2"
        primary-title
      >
        <v-layout row justify-space-between>
          <v-flex>
            {{"{{title}}"}}
          </v-flex>
          
          <v-spacer></v-spacer>
          <v-btn
            text
            icon
            color="primary"
            @click="close()"
          >
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-layout>
      </v-card-title>

      <v-card-text>
        <v-progress-linear v-if="loading" intermediate></v-progress-linear>
        <div v-if="problem">{{"{{problem}}"}}</div>
        <{{SelfFormComponent}} v-model="value" />
      </v-card-text>
      <v-divider></v-divider>
      <v-card-actions>
        <div class="d-flex flex-row justify-center flex-grow-1">
          <v-btn
            color="primary"
            @click="close(true)"
          >
            {{"{{okText}}"}}
          </v-btn>
          </div><div class="d-flex flex-row justify-end">
          <v-btn
            color="primary"
            text
            @click="close()"
          >
            Close
          </v-btn>
        </div>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
{{end}}
`

const vueGridDialogTSTemplate = `
{{define "TS"}}
<script lang="ts">
import { Component, Prop, Vue, Emit, Inject } from 'vue-property-decorator';
import VueApollo from 'vue-apollo';
import { {{TypeName .}}, New{{TypeName .}}Instance, {{GetQuery .}}, {{if not Readonly}}{{SaveQuery .}}, {{CreateQuery .}}{{end}} } from '{{TypesFilePath .}}';
import {{SelfFormComponent}} from '{{SelfFormComponentPath}}';

@Component({
  components:{
    {{SelfFormComponent}}
  }
})
export default class {{.Name}}DialogComponent extends Vue {
  private value: {{TypeName .}} | null = null;
  private isNew = false;
  private title: string = "{{Title .}}";
  private okText: string = "OK";
  private showDialog: boolean = false;
  private resolve: (res: {{TypeName .}}|null)=>void = () => {};
  private reject: (err: any)=>void = () => {};
  private loading = false;
  private problem = "";
  doNotGQL = {{NotExported .}}
  
  show(v: {{TypeName .}}|{{IDType .}}|null): Promise<{{TypeName .}}|null> {
    this.showDialog = true;
    this.problem = "";
    if(!v) {
      this.value = New{{TypeName .}}Instance();
      this.isNew = true;
    } else if(typeof v === "object") {
      this.value = v;
      this.isNew = false;
    } else {
      this.loadAndShow(v);
      this.isNew = false;
    }
    return new Promise((resolve, reject) => {
      this.resolve = resolve;
      this.reject = reject;    
    });
  }
  async loadAndShow(id: {{IDType .}}) {
    this.loading = true;
    this.problem = "";
    try {
      this.value = await {{GetQuery .}}({{ApolloClient}}, id);
    } catch(exc) {
      this.problem = exc.toString();
    }
    this.loading = false;
  }
  async saveAndClose() {
    //TODO: check that all necessary fields are filled
    try {
      let res: {{TypeName .}}
      {{if Readonly}}this.resolve(res);{{else}}
      if(this.doNotGQL) {
        this.resolve(this.value)
      } else {
        if(this.isNew) {
          res = await {{CreateQuery .}}({{ApolloClient}}, this.value!);
        } else {
          res = await {{SaveQuery .}}({{ApolloClient}}, this.value!);
        }
        this.resolve(res);
      }{{end}}
      this.showDialog = false;
    } catch(exc) {
      this.reject(exc);
    }
  }
  async close(save: boolean) {
    if(!save) {
      this.resolve(null);
      this.showDialog = false;
    } else {
      this.saveAndClose();
    }
  }
}
</script>
{{end}}
`
