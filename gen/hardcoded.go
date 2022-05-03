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
	Str  *string  `( @String `
	Num  *float64 `| @Number `
	Bool *bool    `| ("true" | "false") )`
}

var (
	hclex = lexer.Must(ebnf.New(`
Ident = (alpha | "_") { "_" | alpha | digit } .
String = "\"" { "\u0000"…"\uffff"-"\""-"\\" | "\\" any } "\"" .
Number = ("." | digit) {"." | digit} .
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
	r := regexp.MustCompile(`^\s*(hardcoded)|(init-values)\s*\{$`)
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

	flush := func() error {
		if obj != nil {
			defined := map[string]bool{}
			var name string
			if len(obj.Attrs) != len(m.TypeRef.Fields) {
				return fmt.Errorf("too many attrs for type %s", m.TypeName)
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
				var val jen.Code
				switch f.Type.Type {
				case TipInt:
					if a.Value.Num == nil {
						return fmt.Errorf("invalid value for attr %s of type %s: %v", name, TipInt, *a)
					}
					val = jen.Lit(int(*a.Value.Num))
				case TipFloat:
					if a.Value.Num == nil {
						return fmt.Errorf("invalid value for attr %s of type %s: %v", name, TipFloat, *a)
					}
					val = jen.Lit(*a.Value.Num)
				case TipString:
					if a.Value.Str == nil {
						return fmt.Errorf("invalid value for attr %s of type %s: %v", name, TipString, *a)
					}
					val = jen.Lit(*a.Value.Str)
				case TipBool:
					if a.Value.Bool == nil {
						return fmt.Errorf("invalid value for attr %s of type %s: %v", name, TipBool, *a)
					}
					val = jen.Lit(*a.Value.Bool)
				case TipDate:
					return fmt.Errorf("type datetime of field %s can not be used for hardcoded", name)
				default:
					return fmt.Errorf("type of field %s can not be used for hardcoded", name)
				}
				defined[name] = true
				values[jen.Id(name)] = val
			}
			items = append(items, jen.Op("&").Id(m.TypeRef.Name).Values(values))

		}
		obj = &hardcoded{}
		return nil
	}
	write := func() {
		flush()
		arr := jen.List(jen.Index().Op("*").Id(m.TypeRef.Name).Values(items...), jen.Nil())
		if readonly {
			m.TypeRef.Features.Set(FeatGoKind, FCDictGetter, arr)
			m.TypeRef.Features.Set(FeaturesCommonKind, FCReadonly, true)
		} else {
			m.TypeRef.Features.Set(FeatGoKind, FCDictIniter, arr)
		}

	}
	for pos := 1; pos < len(m.Current()); pos++ {
		l = strings.Trim(m.Current()[pos], " \t")
		if l == "}" {
			if phase == 0 {
				write()
				return true, nil
			}
			phase = 0
			continue
		}
		switch phase {
		case 0:
			err = flush()
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
