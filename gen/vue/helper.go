package vue

import (
	"fmt"
	"github.com/alecthomas/participle/lexer"
	"os"
	"path/filepath"
	"text/template"

	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
)

type helper struct {
	templ      *template.Template
	e          *gen.Entity
	cg         *ClientGenerator
	outDir     string
	idField    *gen.Field
	err        error
	ctx        helperContext
	components map[string]vcComponentDescriptor
}

type annotationSet []*gen.Annotation
type fields []fieldDescriptor
type formTabs map[string]formTab

type helperContext struct {
	fields               fields
	width                string
	ann                  annotationSet
	kind                 string
	withRows             bool
	withTabs             bool
	useGrid              bool
	title                string
	tabs                 formTabs
	components           map[string]componentDescriptor
	needSecurity         bool
	needResourceSecurity bool
}

type formTab struct {
	ID       string
	label    string
	order    int
	roles    string
	resource string
}

type fieldDescriptor struct {
	fld      *gen.Field
	width    string
	w        int
	row      int
	ord      int
	title    string
	mask     string
	ann      annotationSet
	tab      string
	tabOrder int
	readonly bool
}
type componentDescriptor struct {
	name    string
	path    string
	relPath string
}

func (cg *ClientGenerator) getTypesPath(e any) (string, error) {
	var features gen.Features
	var name string
	var pos lexer.Position
	switch v := e.(type) {
	case *gen.Entity:
		features = v.Features
		name = v.Name
		pos = v.Pos
	case *gen.Enum:
		features = v.Features
		name = v.Name
		pos = v.Pos
	default:
		return "", fmt.Errorf("vue: invalid type for getTypePath: %T", e)
	}
	fp, ok := features.GetString(js.Features, js.FFilePath)
	if !ok {
		return "", fmt.Errorf("vue: at %v: file path not set for %s", pos, name)
	}
	tn := filepath.Base(fp)
	ext := filepath.Ext(tn)
	if ext != "" {
		tn = tn[:len(tn)-len(ext)]
	}
	return filepath.Join("..", "..", "..", "types", tn), nil
}

func (h *helper) parse(str string) *helper {
	if h.err != nil {
		return h
	}
	h.templ, h.err = h.templ.Parse(str)
	return h
}

func (h *helper) addComponent(cmp string, p string, file *gen.File, name string) {
	if cmp != "" && p != "" {
		if p[0] != '@' && p[0] != '.' && !filepath.IsAbs(p) {
			if file.Package == h.e.File.Package {
				if file.Name == h.e.File.Name {
					p = "." + string(os.PathSeparator) + p
				} else {
					p = filepath.Join("..", file.Name, p)
				}
			} else {
				p = filepath.Join("..", "..", file.Package, file.Name, p)
			}
		}
		h.components[cmp] = vcComponentDescriptor{
			Comp: cmp,
			Imp:  p,
		}
	} else {
		h.cg.desc.AddError(fmt.Errorf("internal error: addComponent was called with %s/%s for %s", cmp, p, name))
	}
}

func (cg *ClientGenerator) getTitleFieldName(e *gen.Entity) (name string, tip string) {
	name = ""
	tip = ""
	for {
		f := getTitleField(e)
		if f == nil {
			return "", ""
		}
		if name != "" {
			name += "."
		}
		name += f.Annotations.GetStringAnnotationDef(js.Annotation, js.AnnotationName, "")
		if !f.Type.Complex {
			tip = f.Type.Type
			break
		}

		dt, ok := cg.desc.FindType(f.Type.Type)
		if !ok {
			cg.b.AddError(fmt.Errorf("at %v: type not found: '%s'", f.Pos, f.Type.Type))
			return "", ""
		}
		e = dt.Entity()
		if e == nil {
			cg.b.AddError(fmt.Errorf("at %v: title: external types and Enums are not supported: '%s'", f.Pos, f.Type.Type))
			return "", ""
		}
	}
	return
}

func getTitleField(e *gen.Entity) *gen.Field {
	fld, ok := e.Annotations.GetStringAnnotation(js.Annotation, js.AnnotationTitle)
	if ok {
		f := e.GetField(fld)
		if f != nil {
			return f
		}
	}
	idf := e.GetIdField()
	return idf
}

func getTabs(e *gen.Entity) (ret formTabs, def string) {
	if ann := e.GetAnnotation(vueTabSet); ann != nil {
		ret = map[string]formTab{}
		for i, a := range ann.Values {
			tab := formTab{ID: a.Key, order: i, label: a.Key}
			if def == "" {
				def = a.Key
			}
			if a.Value != nil && a.Value.String != nil {
				tab.label = *a.Value.String
			} else {
				if ta := e.GetAnnotation(vueTab, a.Key); ta != nil {
					tab.order = ta.GetInt(vcaOrder, i)
					tab.label = ta.GetString(vcaLabel, tab.label)
					tab.roles = ta.GetString(vcaRoles, "")
					tab.resource = ta.GetString(vcaResource, "")
					if d := ta.GetBool(vcaDefault, false); d {
						def = tab.ID
					}
				}
			}
			ret[tab.ID] = tab
		}
	}
	return
}

func (ft formTabs) getOrder(id string) int {
	if ft != nil {
		return ft[id].order
	}
	return 0
}

func (ft formTabs) getLabel(id string) string {
	if ft != nil && ft[id].label != "" {
		return ft[id].label
	}
	return id
}

func (as annotationSet) getString(name string) (ret string, ok bool) {
	for _, a := range as {
		if ret, ok = a.GetStringTag(name); ok {
			return
		}
	}
	return "", false
}

func (as annotationSet) getInt(name string) (ret int, ok bool) {
	for _, a := range as {
		if ret, ok = a.GetIntTag(name); ok {
			return
		}
	}
	return 0, false
}

func (as annotationSet) getBool(name string) (ret bool, ok bool) {
	for _, a := range as {
		if ret, ok = a.GetBoolTag(name); ok {
			return
		}
	}
	return false, false
}

func (as annotationSet) getStringDef(name string, def string) string {
	if v, ok := as.getString(name); ok {
		return v
	}
	return def
}

func (as annotationSet) getIntDef(name string, def int) int {
	if v, ok := as.getInt(name); ok {
		return v
	}
	return def
}

func (as annotationSet) getBoolDef(name string, def bool) bool {
	if v, ok := as.getBool(name); ok {
		return v
	}
	return def
}
