package gen

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
	"github.com/dave/jennifer/jen"
)

type hardcoded struct {
	Attrs []*hardcodedAttr `(@@)*`
}

type hardcodedAttr struct {
	Name  string   `(@Ident ":")?`
	Value *hcValue `@@ ( "," )?`
}

type hcValue struct {
	Null  bool       `( @Nil `
	Str   *string    `| @String `
	Num   *float64   `| @Number `
	Bool  *Boolean   `| @("true" | "false")`
	Map   []*hcMap   `| "{" (@@ ( "," )? )* "}" `
	Arr   []*hcValue `| "[" (@@)* "]"`
	Ident *string    `| @Ident )`
}

type hcMap struct {
	Key string   `(@Ident | @String) ":"`
	Val *hcValue `@@ ( "," )?`
}

var (
	hclex = lexer.Must(ebnf.New(`
Nil = "null" .
Ident = (alpha | "_") { "_" | alpha | digit } .
True = "true" .
False = "false" .
String = "\"" { "\u0000"…"\uffff"-"\""-"\\" | "\\" any } "\"" .
Number = "-" ("." | digit) {"." | digit} | ("." | digit) {"." | digit} .
Whitespace = " " | "\t" | "\n" | "\r" .
Punct = "!"…"/" | ":"…"@" | "["…` + "\"`\"" + ` | "{"…"~" .
                                                    
any = "\u0000"…"\uffff" .																										
alpha = "a"…"z" | "A"…"Z" .
digit = "0"…"9" .
`))

	hcparser = participle.MustBuild(&hardcoded{},
		participle.Lexer(hclex),
		participle.Elide("Whitespace"),
		participle.Unquote("String"),
		participle.UseLookahead(1),
	)
)

func (cg *CodeGenerator) parseHardcoded(m *Meta) (ok bool, err error) {
	// hardcoded meta format:
	// option1:
	// hardcoded {
	//   [attr:] value [,[attr:]] value  <- object1
	//   [attr:] value [,[attr:]] value  <- object2
	// }
	//
	//option 2:
	// hardcoded {
	//   {
	//     attr: value
	//     attr: value
	//   }
	// }

	l := strings.Trim(m.Current()[0], " \t")
	r := regexp.MustCompile(`^\s*(hardcoded)|(init-values)|(values)\s*\{$`)
	match := r.FindStringSubmatch(l)
	if match == nil {
		return
	}
	if m.TypeRef == nil {
		return false, fmt.Errorf("hardcoded without context")
	}

	var obj *hardcoded
	phase := 0
	kind := 0
	items := []jen.Code{}
	readonly := match[1] != ""
	initValues := match[2] != ""
	var name string

	var valueFor func(tr *TypeRef, a *hcValue, pos int, notNull ...bool) (val jen.Code, err error)
	valueFor = func(tr *TypeRef, a *hcValue, pos int, notNull ...bool) (val jen.Code, err error) {
		notNullable := tr.NonNullable
		if len(notNull) != 0 {
			notNullable = notNull[0]
		}
		if a.Null {
			if notNullable {
				return nil, fmt.Errorf("at %s:%d: found null value for non nullable attr %s of type %s", m.Pos.Filename, m.Pos.Line+pos, name, TipInt)
			}
			return jen.Nil(), nil
		}
		switch tr.Type {
		case TipInt:
			if notNullable {
				if a.Num == nil {
					return nil, fmt.Errorf("at %s:%d: invalid value for attr %s of type %s: %v", m.Pos.Filename, m.Pos.Line+pos, name, TipInt, *a)
				}
				val = jen.Lit(int(*a.Num))
			} else {
				if a.Null {
					val = jen.Qual(VivardPackage, "Ptr").Params(jen.Nil())
				} else if a.Num == nil {
					return nil, fmt.Errorf("at %s:%d: invalid value for attr %s of type %s: %v", m.Pos.Filename, m.Pos.Line+pos, name, TipInt, *a)
				} else {
					val = jen.Qual(VivardPackage, "Ptr").Params(jen.Lit(int(*a.Num)))
				}
			}
		case TipFloat:
			if a.Num == nil {
				return nil, fmt.Errorf("at %s:%d: invalid value for attr %s of type %s: %v", m.Pos.Filename, m.Pos.Line+pos, name, TipFloat, *a)
			}
			if notNullable {
				val = jen.Lit(*a.Num)
			} else {
				val = jen.Qual(VivardPackage, "Ptr").Params(jen.Lit(*a.Num))
			}
		case TipString:
			if a.Str == nil {
				return nil, fmt.Errorf("at %s:%d: invalid value for attr %s of type %s: %v", m.Pos.Filename, m.Pos.Line+pos, name, TipString, *a)
			}
			if notNullable {
				val = jen.Lit(*a.Str)
			} else {
				val = jen.Qual(VivardPackage, "Ptr").Params(jen.Lit(*a.Str))
			}
		case TipBool:
			if a.Bool == nil {
				return nil, fmt.Errorf("at %s:%d: invalid value for attr %s of type %s: %v", m.Pos.Filename, m.Pos.Line+pos, name, TipBool, *a)
			}
			if notNullable {
				val = jen.Lit(bool(*a.Bool))
			} else {
				val = jen.Qual(VivardPackage, "Ptr").Params(jen.Lit(bool(*a.Bool)))
			}
		case TipDate:
			return nil, fmt.Errorf("at %s:%d: type datetime of field %s can not be used for hardcoded", m.Pos.Filename, m.Pos.Line+pos, name)
		default:
			if tr.Array != nil {
				val = cg.goType(tr).ValuesFunc(func(g *jen.Group) {
					for _, v := range a.Arr {
						val, e := valueFor(tr.Array, v, pos, tr.Array.NonNullable)
						if e != nil {
							err = e
							return
						}
						g.Add(val)
					}
				})
				if err != nil {
					return
				}
				break
			} else if tr.Map != nil {
				values := jen.DictFunc(func(d jen.Dict) {
					for _, mv := range a.Map {
						v, e := valueFor(tr.Map.ValueType, mv.Val, pos, tr.Map.ValueType.NonNullable)
						if e != nil {
							err = e
							return
						}
						d[jen.Lit(mv.Key)] = v
					}
				})
				if err != nil {
					return
				}
				val = cg.goType(tr).Values(values)
				break
			} else if t, ok := m.TypeRef.Pckg.FindType(tr.Type); ok {
				if t.entry != nil {
					if idfld := t.entry.GetIdField(); idfld != nil {
						return valueFor(idfld.Type, a, pos, notNullable)
					}
				} else if t.enum != nil && a.Ident != nil {
					for _, fld := range t.enum.Fields {
						if *a.Ident == fld.Name {
							if cg.desc == t.enum.Pckg {
								return jen.Id(fld.Name), nil
							} else {
								return jen.Qual(t.enum.Pckg.fullPackage, fld.Name), nil
							}
						}
					}
					return nil, fmt.Errorf("at: %s:%d: value '%s' not found in enum %s", m.Pos.Filename, m.Pos.Line+pos, *a.Ident, t.enum.Name)
				}
			}
			return nil, fmt.Errorf("at: %s:%d: type of field %s can not be used for hardcoded", m.Pos.Filename, m.Pos.Line+pos, name)
		}
		return
	}
	flush := func(pos int) error {
		if obj != nil {
			defined := map[string]bool{}
			if len(obj.Attrs) > len(m.TypeRef.Fields) {
				return fmt.Errorf("at %s:%d: too many attrs for type %s", m.Pos.Filename, m.Pos.Line+pos, m.TypeName)
			} else if len(obj.Attrs) < len(m.TypeRef.Fields) {
				return fmt.Errorf("at %s:%d: too few attrs for type %s", m.Pos.Filename, m.Pos.Line+pos, m.TypeName)
			}
			values := jen.Dict{}

			for i, a := range obj.Attrs {
				if a.Name != "" {
					name = a.Name
				} else {
					name = m.TypeRef.Fields[i].Name
					a.Name = name
				}
				if defined[name] {
					return fmt.Errorf("duplicate value for %s ", name)
				}
				f := m.TypeRef.GetField(name)
				val, err := valueFor(f.Type, a.Value, pos)
				if err != nil {
					return err
				}

				defined[name] = true
				values[jen.Id(name)] = val
			}
			items = append(items, jen.Op("&").Id(m.TypeRef.Name).Values(values))

		}
		obj = &hardcoded{}
		return nil
	}
	write := func(pos int) error {
		err := flush(pos)
		if err != nil {
			return err
		}
		arr := jen.List(jen.Index().Op("*").Id(m.TypeRef.Name).Values(items...), jen.Nil())
		if readonly {
			m.TypeRef.Features.Set(FeatGoKind, FCDictGetter, arr)
			m.TypeRef.Features.Set(FeaturesCommonKind, FCReadonly, true)
		} else {
			if initValues {
				m.TypeRef.Features.Set(FeatGoKind, FCDictIniter, arr)
			} else {
				m.TypeRef.Features.Set(FeatGoKind, FCDictEnsurer, arr)
			}
		}
		return nil
	}
	for pos := 1; pos < len(m.Current()); pos++ {
		l = strings.Trim(m.Current()[pos], " \t\r")
		if l == "}" {
			if phase == 0 {
				err = write(pos)
				return true, err
			}
			phase = 0
			continue
		}
		switch phase {
		case 0:
			err = flush(pos)
			if err != nil {
				return false, err
			}
			phase = 1
			if l == "{" {
				kind = 2
				continue
			} else {
				kind = 1
			}
			fallthrough
		case 1:
			buf := obj
			if kind == 2 {
				buf = &hardcoded{}
			}
			err := hcparser.ParseString(l, buf)
			if err != nil {
				return false, fmt.Errorf("while parsing hardcoded for '%s' (%s): %v", m.TypeRef.Name, l, err)
			}
			switch kind {
			case 1:
				phase = 0
			case 2:
				obj.Attrs = append(obj.Attrs, buf.Attrs...)
			}
		}
	}
	err = errors.New("unclosed hardcoded section")
	return
}

func (v hcValue) String() string {
	switch {
	case v.Str != nil:
		return *v.Str
	case v.Num != nil:
		return fmt.Sprintf("%v", *v.Num)
	case v.Bool != nil:
		return fmt.Sprintf("%v", *v.Bool)
	default:
		return "-empty-"
	}
}
