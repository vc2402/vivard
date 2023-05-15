package gen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
	"github.com/alecthomas/participle/lexer/regex"
	"github.com/dave/jennifer/jen"
)

type File struct {
	Pos      lexer.Position
	Name     string
	FileName string
	Package  string    `("package" @Ident ";")?`
	Meta     []*Meta   `( @@ `
	Entries  []*Entity `| @@ )*`
	Pckg     *Package
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

type Meta struct {
	Pos      lexer.Position
	TypeName string   ` "meta" ( "(" @Ident ")" )?`
	Lines    []string ` (@MetaLine)* `
	start    int
	end      int
	TypeRef  *Entity
	err      error
}

type Annotations map[string]*Annotation
type Features map[string]interface{}

type Entity struct {
	Pos          lexer.Position
	Modifiers    []*EntityModifier `( (@@)* )? `
	Name         string            `"type" (@Ident | @QualifiedName)  `
	BaseTypeName string            `( "extends" (@Ident | @QualifiedName) )? "{"`
	Entries      []*Entry          `( @@ )*`
	Incomplete   bool              `(@More)? "}"`
	Fields       []*Field
	Methods      []*Method
	FieldsIndex  map[string]*Field
	MethodsIndex map[string]*Method
	Annotations  Annotations
	// Features - generators created values based on Annotations, generators options...
	Features     Features
	TypeModifers map[TypeModifier]bool
	// BaseType      *Entity
	Pckg *Package
	// BaseFieldTags map[string]string
	BaseField       *Field
	File            *File
	FullAnnotations Annotations
}

type Entry struct {
	Pos       lexer.Position
	Field     *Field           `( @@  `
	Method    *Method          `|  @@ )`
	Modifiers []*EntryModifier `( "<" ( @@ )* ">")? ";"`
}

type Field struct {
	Pos         lexer.Position
	Modifiers   []*EntryModifier
	Name        string   `@Ident ":"`
	Type        *TypeRef `@@`
	Tags        map[string]string
	Annotations Annotations
	// Features - generators created values based on Annotations, generators options...
	Features Features
	parent   *Entity
}
type Method struct {
	Pos         lexer.Position
	Modifiers   []*EntryModifier
	Name        string         `@Ident "("`
	Params      []*MethodParam ` (  (@@) ("," @@)* )? `
	RetValue    *TypeRef       `")" (":" @@)?`
	Annotations Annotations
	// Features - generators created values based on Annotations, generators options...
	Features Features
	parent   *Entity
}
type MethodParam struct {
	Pos      lexer.Position
	Name     string   `@Ident ":"`
	Type     *TypeRef `@@`
	Features Features
}

type TypeRef struct {
	Array       *TypeRef `( "[" @@ "]"`
	Map         *MapType ` |  @@ `
	Ref         bool     ` | ( @"*" )?`
	Type        string   ` (@Ident | @QualifiedName | @"auto" ) )`
	NonNullable bool     `[ @"!" ]`
	Complex     bool
	Embedded    bool
}

type MapType struct {
	KeyType   string   `"map" "[" (@"int"|@"string") "]"`
	ValueType *TypeRef `@@`
}

type EntityModifier struct {
	Pos          lexer.Position
	Hook         *Hook             `( @@`
	Annotation   *Annotation       `| @@`
	TypeModifier *TypeModifierType `| @@ )`
}

type TypeModifierType struct {
	Modifier TypeModifier `@( "abstract" | "config" | "dictionary" | "transient" | "embeddable" | "singleton" | "extern" | "extendable" )`
}
type EntryModifier struct {
	Pos          lexer.Position
	Hook         *Hook       `( @@`
	AttrModifier string      `| @AttrModifier`
	Annotation   *Annotation `| @@ )`
}

type Hook struct {
	Key   string `@HookTag`
	Spec  string `( ":" @Ident )?`
	Value string `("=" @String)?`
}

type Annotation struct {
	Pos  lexer.Position
	Name string `@AnnotationTag `
	// Spec   string           //`( ":" @Ident )?`
	Values []*AnnotationTag `("(" (@@)* ")")?`
}
type AnnotationTag struct {
	Pos   lexer.Position
	Key   string           `(@Ident | @String)`
	Value *AnnotationValue `( "=" @@ )?`
}
type AnnotationValue struct {
	String    *string  ` ( @String `
	Bool      *Boolean `| @("true" | "false") `
	Number    *float64 `| @Number )`
	Interface interface{}
}

var (
	lex = lexer.Must(ebnf.New(`
Comment = ("//") { "\u0000"…"\uffff"-"\n" } .
MetaLine = ("#" | "\t") { "\u0000"…"\uffff"-"\n" } .
TypeModifier = "config" | "dictionary" | "embeddable" | "foreign" | "singleton" | "transient" .
AttrModifier = "id" | "auto" | "lookup" | "one-to-many" | "embedded" | "ref-embedded" | "calculated" .
More = "..." .
QualifiedName = Ident "." Ident .
Ident = (alpha | "_") { "_" | alpha | digit } .
AnnotationTag = "$" AnnotationName [ ":" AnnotationName ] .
AnnotationName = (alpha | "_") { "_" | alpha | digit | "-"} .
HookTag = "@" (alpha | "_") { "_" | alpha | digit | "-" } .
String = "\"" { "\u0000"…"\uffff"-"\""-"\\" | "\\" any } "\"" .
Number = ("." | digit) {"." | digit} .
Whitespace = " " | "\t" | "\n" | "\r" .
Punct = "!"…"/" | ":"…"@" | "["…` + "\"`\"" + ` | "{"…"~" .
BracketOpen = "[" .
BracketClose = "]" .
ModifierOpen = "<" .
ModifierClose = ">" .

any = "\u0000"…"\uffff" .																										
alpha = "a"…"z" | "A"…"Z" .
digit = "0"…"9" .
`))

	regexLex = lexer.Must(regex.New(`
Whitespace = [\s\t\n]+
Comment = \/\/.*^
BracketOpen = \[
BracketClose = \]
ModifierOpen = [<]
ModifierClose = [>]
TypeModifier = (foreign)|(embeddable)|(dictionary)|(transient)|(singleton) 
AttrModifier = (id)|(auto)|(lookup)|(one-to-many)|(embedded)
More = \.\.\.
QualifiedName = [[:ascii:]][\w\d]*\.[[:ascii:]][\w\d]*
AnnotationTag = [$][[:ascii:]][\w\d_-]*
HookTag = @[[:ascii:]][\w\d_-]*
String = "[^"]*"
Number = ([-+])?(\d+)|(\d*\.\d+)
Ident = [[:ascii:]][\w\d]*
`))

	// FieldTypeString = { "string" } .
	// FieldTypeInt = { "int" } .
	// FieldTypeDate = { "date" } .

	parser = participle.MustBuild(&File{},
		// participle.Lexer(regexLex),
		participle.Lexer(lex),
		participle.Elide("Comment", "Whitespace"),
		participle.Unquote("String"),
		participle.Map(annotationMapper, "AnnotationTag", "HookTag", "MetaLine"),
		participle.UseLookahead(1),
	)
)

func annotationMapper(token lexer.Token) (lexer.Token, error) {
	token.Value = token.Value[1:]
	return token, nil
}

func Parse(files []string) ([]*File, error) {
	ret := []*File{}
	for _, file := range files {
		ast := &File{}
		r, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("can't open file '%s': %w", file, err)
		}
		err = parser.Parse(r, ast)
		r.Close()
		if err != nil {
			return nil, fmt.Errorf("while parsing '%s': %w", file, err)
		}
		ast.FileName = path.Base(file)
		ast.Name = ast.FileName
		err = ast.postProcess()
		if err != nil {
			return nil, fmt.Errorf("while post processing '%s': %w", file, err)
		}
		if ext := strings.Index(ast.Name, "."); ext != -1 {
			ast.Name = ast.Name[:ext]
		}
		ret = append(ret, ast)
	}
	return ret, nil
}

func (f *File) postProcess() error {
	for _, t := range f.Entries {
		if IsPrimitiveType(t.Name) {
			return fmt.Errorf("at %v: unexpected %s", t.Pos, t.Name)
		}
		if t.IsNameQualified() && !t.HasModifier(TypeModifierExternal) {
			return fmt.Errorf("at %v: unexpected qualified name for not external type %s", t.Pos, t.Name)
		}
		processTypeModifiers(t)
		if t.Incomplete && !t.HasModifier(TypeModifierExternal) {
			return fmt.Errorf("at %v: only external types may be incomplete", t.Pos)
		}
		t.Fields = []*Field{}
		t.Methods = []*Method{}
		t.FieldsIndex = map[string]*Field{}
		t.MethodsIndex = map[string]*Method{}
		t.Features = Features{}
		// stripAnnotations(t.Modifiers)
		for _, te := range t.Entries {
			if te.Field != nil {
				te.Field.Modifiers = te.Modifiers
				te.Field.Features = Features{}
				te.Field.parent = t
				t.Fields = append(t.Fields, te.Field)
				t.FieldsIndex[te.Field.Name] = te.Field
				te.Field.PostProcess()
			} else if te.Method != nil {
				te.Method.Modifiers = te.Modifiers
				te.Method.Features = Features{}
				te.Method.parent = t
				if te.Method.RetValue != nil {
					te.Method.RetValue.Complex = !IsPrimitiveType(te.Method.RetValue.Type)
				}
				t.Methods = append(t.Methods, te.Method)
				t.MethodsIndex[te.Method.Name] = te.Method
				for _, p := range te.Method.Params {
					p.Features = Features{}
				}
			} else {
				return fmt.Errorf("undefined entry at %v", te.Pos)
			}
		}
	}
	return nil
}

func IsPrimitiveType(name string) bool {
	return name == TipBool || name == TipDate || name == TipFloat || name == TipInt || name == TipString || name == TipAny
}

// FS shorcut to Features.String()
func (e *Entity) FS(kind FeatureKind, name string) string {
	return e.Features.String(kind, name)
}

// FB shorcut to Features.String()
func (e *Entity) FB(kind FeatureKind, name string) bool {
	return e.Features.Bool(kind, name)
}

func (e *Entity) GetIdField() *Field {
	if e.BaseTypeName != "" {
		bc := e.GetBaseType()
		return bc.GetIdField()
	}
	for _, f := range e.Fields {
		if f.IsIdField() {
			return f
		}
	}
	return nil
}

func (e *Entity) GetBaseType() *Entity {
	if e.BaseTypeName != "" {
		bc, ok := e.Pckg.FindType(e.BaseTypeName)
		if !ok {
			panic(fmt.Sprintf("at %v: base type '%s' not found", e.Pos, e.BaseTypeName))
		}
		return bc.Entity()
	}
	panic(fmt.Sprintf("GetBaseType was called for non-derived type"))
}

func (e *Entity) IsNameQualified() bool {
	return strings.Index(e.Name, ".") != -1
}

func (e *Entity) Package() string {
	idx := strings.Index(e.Name, ".")
	if idx != -1 {
		return e.Name[:idx]
	}
	//TODO: find and return current package
	return ""
}
func (e *Entity) HasModifier(mod TypeModifier) bool {
	return e.TypeModifers[mod]
}

func (e *Entity) IsDictionary() bool {
	return e.HasModifier(TypeModifierDictionary)
}

func (e *Entity) GetField(name string) *Field {
	if fld, ok := e.FieldsIndex[name]; ok {
		return fld
	}
	if e.BaseTypeName != "" {
		bt := e.GetBaseType()
		return bt.GetField(name)
	}
	return nil
}

func (e *Entity) GetFields(includeBase bool, baseInline bool) []*Field {
	if e.BaseTypeName == "" || !includeBase {
		return e.Fields
	}
	bt := e.GetBaseType()
	var ret []*Field
	if baseInline {
		ret = bt.GetFields(includeBase, baseInline)
	} else {
		ret = []*Field{e.GetBaseField()}
	}
	return append(ret, e.Fields...)
}
func (e *Entity) GetBaseField() *Field {
	if e.BaseField == nil {
		if e.BaseTypeName == "" {
			panic(fmt.Sprintf("GetBaseField is called for non derived type %s", e.Name))
		}

		e.BaseField = &Field{
			parent:      e,
			Annotations: Annotations{},
			Features:    Features{},
			Name:        e.BaseTypeName,
			Pos:         e.Pos,
			Tags:        map[string]string{},
			Type:        &TypeRef{Complex: true, NonNullable: true, Type: e.BaseTypeName},
		}
	}
	return e.BaseField
}

// GetFullAnnotations returns annotation including base type annotations
func (e *Entity) GetFullAnnotations() Annotations {
	if e.BaseTypeName != "" {
		if e.FullAnnotations == nil {
			bt := e.GetBaseType()
			e.FullAnnotations = Annotations{}
			e.FullAnnotations.Append(bt.GetFullAnnotations())
			e.FullAnnotations.Append(e.Annotations)
		}
		return e.FullAnnotations
	}
	return e.Annotations
}

// GetAnnotation looks for annotation in FullAnnotations (including BaseType annotations)
func (e *Entity) GetAnnotation(name string, spec ...string) *Annotation {
	ann := e.GetFullAnnotations()
	if len(spec) == 0 || spec[0] == "" {
		return ann[name]
	} else {
		return ann.Find(name, spec[0])
	}
}

func (e *Entity) HaveHook(key string) (val *Hook, ok bool) {
	for _, m := range e.Modifiers {
		if m.Hook != nil && m.Hook.Key == key {
			return m.Hook, true
		}
	}
	return
}

func (e *Entity) AddDescendant(desc *Entity) {
	var descendants []*Entity
	if d, ok := e.Features.Get(FeaturesCommonKind, FCDescendants); ok {
		descendants = d.([]*Entity)
	}
	descendants = append(descendants, desc)
	e.Features.Set(FeaturesCommonKind, FCDescendants, descendants)
	if e.BaseTypeName != "" {
		if bt := e.GetBaseType(); bt != nil {
			bt.AddDescendant(desc)
		}
	}
}

func (e *Entity) AddField(name string, tip string) (*Field, error) {
	if _, ok := e.FieldsIndex[name]; ok {
		return nil, fmt.Errorf("field '%s' is already exists in type '%s", name, e.Name)
	}
	fld := &Field{
		Modifiers:   []*EntryModifier{},
		Name:        name,
		Type:        &TypeRef{Type: tip},
		Tags:        map[string]string{},
		Annotations: Annotations{},
		Features:    Features{},
		parent:      e,
	}
	e.Fields = append(e.Fields, fld)
	e.FieldsIndex[name] = fld
	return fld, nil
}

func (e *Entity) TypeRef() *TypeRef {
	return &TypeRef{
		Type: fmt.Sprintf("%s.%s", e.Pckg.Name, e.Name),
	}
}

func (f *Field) IsIdField() bool {
	for _, a := range f.Modifiers {
		if a.AttrModifier == "id" || a.AttrModifier == "auto" {
			return true
		}
	}
	return false
}

func (f *Field) HasModifier(mod AttrModifier) bool {
	for _, a := range f.Modifiers {
		if a.AttrModifier == string(mod) {
			return true
		}
	}
	return false
}

func (f *Field) PostProcess() {
	f.Type.Complex = !IsPrimitiveType(f.Type.Type)
	f.Type.Embedded = f.HasModifier(AttrModifierEmbedded)
	//trying to fill complex for ref types of arrays...
	arrType := f.Type.Array
	for arrType != nil {
		arrType.Complex = !IsPrimitiveType(arrType.Type)
		arrType = arrType.Array
	}
}

// Parent returns enclosing entity
func (f *Field) Parent() *Entity {
	return f.parent
}

// FS shorcut to Features.String()
func (f *Field) FS(kind FeatureKind, name string) string {
	return f.Features.String(kind, name)
}

// FB shorcut to Features.String()
func (f *Field) FB(kind FeatureKind, name string) bool {
	return f.Features.Bool(kind, name)
}

// FS shorcut to Features.String()
func (m *Method) FS(kind FeatureKind, name string) string {
	return m.Features.String(kind, name)
}

// FB shorcut to Features.String()
func (m *Method) FB(kind FeatureKind, name string) bool {
	return m.Features.Bool(kind, name)
}

// Parent returns enclosing entity
func (m *Method) Parent() *Entity {
	return m.parent
}

// Name returns string representing given kind and name
func (f Features) Name(kind FeatureKind, name string) string {
	return fmt.Sprintf("%s:%s", kind, name)
}

// Get looks for feature and returns it if any
func (f Features) Get(kind FeatureKind, name string) (val interface{}, ok bool) {
	val, ok = f[f.Name(kind, name)]
	return
}

// Set adds feature to a feature set
func (f Features) Set(kind FeatureKind, name string, feat interface{}) {
	f[f.Name(kind, name)] = feat
}

// GetString looks for feature and asserts it to string if any
func (f Features) GetString(kind FeatureKind, name string) (val string, ok bool) {
	val, ok = f[f.Name(kind, name)].(string)
	return
}

// GetBool looks for feature and asserts it to bool if any
func (f Features) GetBool(kind FeatureKind, name string) (val bool, ok bool) {
	val, ok = f[f.Name(kind, name)].(bool)
	return
}

// String looks for feature and asserts it to string if any; otherwise returns empty string
func (f Features) String(kind FeatureKind, name string) (val string) {
	val, _ = f[f.Name(kind, name)].(string)
	return
}

// Stmt looks for feature and asserts it to *jen.Statement if any; otherwise returns empty statement
func (f Features) Stmt(kind FeatureKind, name string) (val *jen.Statement) {
	val = &jen.Statement{}
	val, _ = f[f.Name(kind, name)].(*jen.Statement)
	return
}

// Bool looks for feature and asserts it to bool if any; otherwise return false
func (f Features) Bool(kind FeatureKind, name string) (val bool) {
	val, ok := f[f.Name(kind, name)].(bool)
	return ok && val
}

// GetEntity looks for feature and asserts it to *Entity if any
func (f Features) GetEntity(kind FeatureKind, name string) (val *Entity, ok bool) {
	val, ok = f[f.Name(kind, name)].(*Entity)
	return
}

// GetField looks for feature and asserts it to *Field if any
func (f Features) GetField(kind FeatureKind, name string) (val *Field, ok bool) {
	val, ok = f[f.Name(kind, name)].(*Field)
	return
}

// Find returns Annotation of kind prefix:name if any; nil otherwise
func (a Annotations) Find(name, spec string) *Annotation {
	an := fmt.Sprintf("%s:%s", name, spec)
	return a[an]
}

// ByPrefix returns Annotations of kind prefix:name with given prefix; keys are names only
// params: prefix - first part of name
// includeUnspec - include annotation without specifier (with name == kind)
func (a Annotations) ByPrefix(prefix string, includeUnspec bool) map[string]*Annotation {
	ret := map[string]*Annotation{}
	for n, an := range a {
		if includeUnspec && n == prefix {
			ret[""] = an
		} else {
			parts := strings.SplitN(n, ":", 2)
			if len(parts) == 2 && parts[0] == prefix {
				ret[parts[1]] = an
			}
		}
	}
	return ret
}

// GetStringAnnotation returns string annotation with given name if any; ok is false otherwise
func (a Annotations) GetStringAnnotation(name string, key string) (val string, ok bool) {
	if a == nil {
		return
	}
	if an, ok := a[name]; ok {
		return an.GetStringTag(key)
	}
	return
}
func (a Annotations) GetStringAnnotationDef(name string, key string, def string) (val string) {
	val, ok := a.GetStringAnnotation(name, key)
	if !ok {
		val = def
	}
	return
}

// GetNameAnnotation returns string tag annotation with given name if any or first tag key if it is bool and true
//
//	e.g. in case $ann(someValue) it returns 'someValue' and for $ann(someValue name="somName") returns someName
func (a Annotations) GetNameAnnotation(name string, key string) (val string, ok bool) {
	if a == nil {
		return
	}
	if an, ok := a[name]; ok {
		return an.GetNameTag(key)
	}
	return
}
func (a Annotations) GetStringAnnotationDefTrimmed(name string, key string, def string) (val string) {
	val, ok := a.GetStringAnnotation(name, key)
	if !ok {
		val = def
	}
	val = strings.Trim(val, " \t\n")
	return
}
func (a Annotations) GetBoolAnnotation(name string, key string) (val bool, ok bool) {
	if a == nil {
		return
	}
	if an, ok := a[name]; ok {
		return an.GetBoolTag(key)
	}
	return
}
func (a Annotations) GetBoolAnnotationDef(name string, key string, def bool) bool {
	if a == nil {
		return def
	}
	if an, ok := a[name]; ok {
		if v, ok := an.GetBoolTag(key); ok {
			return v
		}
	}
	return def
}
func (a Annotations) GetInterfaceAnnotation(name string, key string) (val interface{}) {
	if a == nil {
		return
	}
	if an, ok := a[name]; ok {
		if ret, ok := an.GetInterfaceTag(key); ok {
			return ret
		}
	}
	return
}
func (a Annotations) AddTag(name string, key string, val interface{}, spec ...string) error {
	if a == nil {
		return errors.New("Annotations.AddTag was called for nil object")
	}
	av := &AnnotationValue{}
	switch v := val.(type) {
	case string:
		av.String = &v
	case int:
		n := float64(v)
		av.Number = &n
	case float64:
		av.Number = &v
	case bool:
		av.Bool = new(Boolean)
		*av.Bool = Boolean(v)
	default:
		av.Interface = val
	}
	at := &AnnotationTag{Key: key, Value: av}
	an, ok := a[name]
	if !ok {
		a[name] = &Annotation{Name: name, Values: []*AnnotationTag{at}}
	} else {
		an.Values = append(an.Values, at)
	}
	return nil
}

// Append appends all annotation from another to a
func (a Annotations) Append(another Annotations) {
	for k, v := range another {
		if an, ok := a[k]; ok {
			an.Append(v)
		} else {
			a[k] = v
		}
	}
}

// Add adds values from another to a rewrites values with the same key
func (a *Annotation) Append(another *Annotation) {
	for _, av := range another.Values {
		if v := a.GetTag(av.Key); v != nil {
			v.Value = av.Value
		} else {
			a.Values = append(a.Values, av)
		}
	}
}

func (a *Annotation) GetTag(key string) *AnnotationTag {
	for _, t := range a.Values {
		if t.Key == key {
			return t
		}
	}
	return nil
}

func (a *Annotation) GetStringTag(key string) (ret string, ok bool) {
	if t := a.GetTag(key); t != nil {
		return t.GetString()
	}
	return
}

func (a *Annotation) GetNameTag(key string) (ret string, ok bool) {
	if t := a.GetTag(key); t != nil {
		return t.GetString()
	}
	if len(a.Values) > 0 && a.Values[0].Value == nil {
		ret = a.Values[0].Key
		ok = true
	}
	return
}
func (a *Annotation) GetBoolTag(key string) (ret bool, ok bool) {
	if t := a.GetTag(key); t != nil {
		return t.GetBool()
	}
	return
}

func (a *Annotation) GetIntTag(key string) (ret int, ok bool) {
	if t := a.GetTag(key); t != nil {
		return t.GetInt()
	}
	return
}

func (a *Annotation) GetFloatTag(key string) (ret float64, ok bool) {
	if t := a.GetTag(key); t != nil {
		return t.GetFloat()
	}
	return
}

func (a *Annotation) GetInterfaceTag(key string) (ret interface{}, ok bool) {
	if t := a.GetTag(key); t != nil && t.Value != nil {
		return t.Value.Interface, true
	}
	return
}
func (a *Annotation) GetString(key string, def string) (ret string) {
	ret, ok := a.GetStringTag(key)
	if !ok {
		ret = def
	}
	return
}

func (a *Annotation) GetBool(key string, def bool) (ret bool) {
	ret, ok := a.GetBoolTag(key)
	if !ok {
		ret = def
	}
	return
}

func (a *Annotation) GetInt(key string, def int) (ret int) {
	ret, ok := a.GetIntTag(key)
	if !ok {
		ret = def
	}
	return
}

func (a *Annotation) GetFloat(key string, def float64) (ret float64) {
	ret, ok := a.GetFloatTag(key)
	if !ok {
		ret = def
	}
	return
}

func (a *Annotation) SetTag(key string, val interface{}) {
	t := a.GetTag(key)
	if t == nil {
		t = &AnnotationTag{Key: key, Value: &AnnotationValue{}}
		a.Values = append(a.Values, t)
	} else {
		t.Value.Bool = nil
		t.Value.String = nil
		t.Value.Number = nil
	}
	switch v := val.(type) {
	case string:
		t.Value.String = &v
	case int:
		n := float64(v)
		t.Value.Number = &n
	case float64:
		t.Value.Number = &v
	case bool:
		t.Value.Bool = new(Boolean)
		*t.Value.Bool = Boolean(v)
	default:
		t.Value.Interface = val
	}
	return
}
func (at *AnnotationTag) GetString() (ret string, ok bool) {
	if at.Value != nil && at.Value.String != nil {
		return *at.Value.String, true
	}
	return
}

func (at *AnnotationTag) GetBool() (ret bool, ok bool) {
	if at.Value != nil && at.Value.Bool != nil {
		return bool(*at.Value.Bool), true
	}
	if at.Value == nil {
		return true, true
	}
	return
}

func (at *AnnotationTag) GetInt() (ret int, ok bool) {
	if at.Value != nil && at.Value.Number != nil {
		return int(*at.Value.Number), true
	}
	return
}

func (at *AnnotationTag) GetFloat() (ret float64, ok bool) {
	if at.Value != nil && at.Value.Number != nil {
		return *at.Value.Number, true
	}
	return
}

func (at *AnnotationValue) strip() {
	if at.String != nil && (*at.String)[0:1] == "\"" {
		*at.String = (*at.String)[1 : len(*at.String)-1]
		*at.String = strings.ReplaceAll(*at.String, "\\\"", "\"")
	}
}

func (a *AnnotationTag) strip() {
	if a.Value != nil {
		a.Value.strip()
	}
}

// func (a *EntityModifier) strip() {
// 	if a.Annotation != nil {
// 		for _, at := range a.Annotation.Values {
// 			if at.Value != nil {
// 				at.Value.strip()
// 			}
// 		}
// 	} else if a.Hook != nil {
// 		if a.Hook.Value != "" && a.Hook.Value[0:1] == "\"" {
// 			a.Hook.Value = (a.Hook.Value)[1 : len(a.Hook.Value)-1]
// 			a.Hook.Value = strings.ReplaceAll(a.Hook.Value, "\\\"", "\"")
// 		}
// 	}
// }

func (a *EntryModifier) strip() {
	if a.Annotation != nil {
		for _, at := range a.Annotation.Values {
			if at.Value != nil {
				at.Value.strip()
			}
		}
	} else if a.Hook != nil {
		if a.Hook.Value != "" && a.Hook.Value[0:1] == "\"" {
			a.Hook.Value = (a.Hook.Value)[1 : len(a.Hook.Value)-1]
			a.Hook.Value = strings.ReplaceAll(a.Hook.Value, "\\\"", "\"")
		}
	}
}

func (f *Field) HaveHook(key string) (val *Hook, ok bool) {
	for _, m := range f.Modifiers {
		if m.Hook != nil && m.Hook.Key == key {
			return m.Hook, true
		}
	}
	return
}

func (m *Method) HaveHook(key string) (val *Hook, ok bool) {
	for _, mod := range m.Modifiers {
		if mod.Hook != nil && mod.Hook.Key == key {
			return mod.Hook, true
		}
	}
	return
}

func stripAnnotations(an []*EntryModifier) {
	for _, a := range an {
		a.strip()
	}
}

// Next looks for next slice; returns true if any false if no more slices found
func (m *Meta) Next() bool {
	if m.start > 0 && m.end == m.start {
		m.err = errors.New("at end")
		return false
	}
	m.start = m.end
	for m.start < len(m.Lines) && strings.Trim(m.Lines[m.start], " \t") == "" {
		m.start++
	}
	if m.start >= len(m.Lines) {
		return false
	}
	m.end = m.start
	for m.end < len(m.Lines) && strings.Trim(m.Lines[m.end], " \t") != "" {
		m.end++
	}
	if m.end == m.start && m.start == 0 {
		m.err = errors.New("empty")
	}
	return m.end > m.start
}

// Current returns current slice of meta lines (delimited by empty lines)
func (m *Meta) Current() []string {
	return m.Lines[m.start:m.end]
}

// Err returns current error state
func (m *Meta) Err() error { return m.err }

// Position returns position with lines offset
func (m *Meta) Position() lexer.Position {
	pos := m.Pos
	pos.Line += m.start
	return pos
}
func processTypeModifiers(t *Entity) {
	// var mod []*EntityModifier
	// tmc := 0
	t.TypeModifers = map[TypeModifier]bool{}
	for _, m := range t.Modifiers {
		if m.TypeModifier != nil {
			t.TypeModifers[m.TypeModifier.Modifier] = true
			// if len(mod) == 0 && tmc
			// } else if len(mod) > 0 {
			// 	mod = append(mod, m)
			// }
		}
		// if len(mod) > 0 {
		// 	t.Modifiers = mod
	}
}
