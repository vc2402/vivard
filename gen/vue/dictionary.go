package vue

import (
	"fmt"
	"os"
)

func (h *helper) createDictEditor() error {
	p := h.e.FS(featureVueKind, fVKDictEditComponentPath)
	if p == "" {
		return nil
	}

	f, err := os.Create(p)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("while opening file '%s' for DictEditComponent for %s: %v", p, h.e.Name, err)
	}
	h.parse(vueDictHTMLTemplate).
		parse(vueDictTSTemplate).
		parse("{{template \"HTML\" .}}\n{{template \"TS\" .}}\n")
	if h.err != nil {
		return fmt.Errorf("error while parsing template: %v", h.err)
	}
	err = h.templ.Execute(f, h)
	if err != nil {
		return fmt.Errorf("while executing template for DictEditComponent for %s: %v", h.e.Name, err)
	}
	return nil
}

const vueDictHTMLTemplate = `{{define "HTML"}}
<template>
  <div>
    <div v-if="problem">{{"{{"}}problem{{"}}"}}</div>
    <div class="d-flex flex-row align-baseline">
      <v-text-field
        v-model="search"
        prepend-icon="mdi-magnify"
        label="Search"
        single-line
        hide-details
      ></v-text-field>
      {{if DictWithQualifier .}}<{{LookupForQualifier .}} class="mx-5" v-model="qualifier" :returnObject="false" :multiple="true" :hideAdd="true" label="{{DictQualifierTitle}}"/>
      {{end}}<v-btn v-if="!hideReload" color="success" icon text>
        <v-icon color="success" @click="load()">mdi-reload</v-icon>
      </v-btn>
      <v-btn v-if="showAddButton" color="success" icon text>
        <v-icon color="success" @click="add()">mdi-plus-box</v-icon>
      </v-btn>
    </div>
    <div class="table" >
      <v-data-table
        :headers="headers"
        :items="items"
        :search="search"
        dense
        fixed-header
        item-key="name"
        :loading="loading"
        height="calc( 100% - 48px )"
        class="ddb-table"
        :items-per-page="20"
        :footer-props="{itemsPerPageOptions: [10, 20, 30, 40, 50, -1]}"
        >
        <template v-slot:item="{item}">
          <tr>
            {{range GetFields .}}<td class="table-cell">
              {{if NeedIconForTable .}}<v-icon>{{TableIconName .}}</v-icon>{{else if eq (GUITableType .) "custom"}}<{{GUITableComponent .}} :item="item" :header="{value: '{{TableAttrName .}}'}" :value="item.{{TableAttrName .}}" />{{else}}{{"{{"}}item.{{TableAttrName .}}{{"}}"}}{{end}}
            </td>{{end}}
            <td>
              <div class="d-flex flex-row">
              <v-btn v-if="!readonly" color="success" icon text @click="edit(item)">
                <v-icon>mdi-pencil</v-icon>
              </v-btn>
              <v-btn v-if="!readonly" color="warn" icon text @click="del(item)">
                <v-icon>mdi-delete</v-icon>
              </v-btn>
              </div>
            </td>
          </tr> 
        </template>
      </v-data-table>
    </div>
    <{{DialogComponent .}} ref="dialog"/>
  </div>
</template>
{{end}}`

const vueDictTSTemplate = `{{define "TS"}}
<script lang="ts">
import { Component, Prop, Vue, Watch, Inject, Emit } from 'vue-property-decorator';
import { {{TypeName}}, {{InstanceGeneratorName}}, {{ListQuery}} } from '{{TypesFilePath}}';
import VueRx from 'vue-rx';
{{range RequiredComponents}}import {{.}} from './{{.}}.vue'
{{end}}

@Component({
  components: {
    {{range RequiredComponents}}{{.}},
    {{end}}
  },
})
export default class {{DictEditComponent . false}} extends Vue {
  @Prop({default:false}) readonly!: boolean;
  @Prop({default:true}) showAddButton!: boolean;
  @Prop({default:false}) hideReload!: boolean;
  private problem = "";
  private search: string = "";
  private items: {{TypeName}}[] = [];
  private headers = [{{range (GetFields .)}}
  {text: "{{Label .}}", value: "{{TableAttrName .}}", {{if NeedIconForTable .}}icon: "{{TableIconName .}}", {{end}}type: "{{GUITableType .}}", color: "{{GUITableColor .}}"{{if ne (GUITableComponent .) ""}}, component: "{{GUITableComponent .}}"{{end}} }, {{end}}
  {text:"", value: ""}];
  
  private loading = false;{{if DictWithQualifier .}}
	qualifier: any = null;{{end}}

  mounted() {
    this.load();
  }
  
  public async load() {
    this.loading = true;
    this.problem = "";
    try {
      this.items = await {{ListQuery}}({{ApolloClient}}{{ if DictWithQualifier .}}, this.qualifier{{end}});  
    } catch(exc) {
      console.log("exception: ", exc);
      this.problem = "Problema: " + exc.toString();
    }
    this.loading = false;
  }
  async edit(item: any) {
    try {
      let res = await (this.$refs.dialog as {{DialogComponent .}}).show(item);
      if(res) {
        this.load();
      }
    } catch(exc) {
      this.problem = exc.toString();
    }
  }
  async del(item: any) {
    // if(this.delFn)
    //   this.delFn(this.getApollo(), item.id);
  }
  async add() {
    try {
      let res = await (this.$refs.dialog as {{DialogComponent .}}).show(null);
      if(res) {
        this.load();
      }
    } catch(exc) {
      this.problem = exc.toString();
    }
  }
}
</script>{{end}}
`
