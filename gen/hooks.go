package gen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type HookType string
type HookModifier string

const (
	//well known var names
	//GHVObject default name of object (this) var
	GHVObject = "obj"
	//GHVContext default name of context var
	GHVContext = "ctx"
	//GHVEngine default name of engine var
	GHVEngine = "eng"
)

//GeneratorHookVars holds available variables for generator hooks
type GeneratorHookVars struct {
	// Ctx - variable name or context *jen.Statement if it is not "ctx"; if ctx is not available should be false
	Ctx interface{}
	// Eng - variable name or Engine *jen.Statement if it is not "eng"; if there is no Engine in context should be false
	Eng interface{}
	// Obj - variable name or object *jen.Statement if it is not "obj"
	Obj interface{}
	//Others - additional well known for some hooks vars with their names (there are no defaults for others - all should be set as var name or statement)
	Others map[string]interface{}
}

//GeneratorHookHolder - interface for hooks code generation
type GeneratorHookHolder interface {
	//OnEntityHook may be called for Entity for given hook name
	//  if returned value is not nil, it will be added to code in consideration to order
	//  vars contains var names for common objects (ctx, eng, obj); may be nil, in this case defaults will be used
	OnEntityHook(name HookType, mod HookModifier, e *Entity, vars *GeneratorHookVars) (code *jen.Statement, order int)

	//OnFieldHook may be called for Field for given hook name
	//  if returned value is not nil, it will be added to code in consideration to order
	//  vars contains var names for common objects (ctx, eng, obj); may be nil, in this case defaults will be used
	OnFieldHook(name HookType, mod HookModifier, f *Field, vars *GeneratorHookVars) (code *jen.Statement, order int)

	//OnMethodHook may be called for Method for given hook name
	//  if returned value is not nil, it will be added to code in consideration to order
	//  vars contains var names for common objects (ctx, eng, obj); may be nil, in this case defaults will be used
	OnMethodHook(name HookType, mod HookModifier, m *Method, vars *GeneratorHookVars) (code *jen.Statement, order int)
}

//standart hooks
const (
	//HookSet will be called on Set for field (start, exit, error)
	HookSet HookType = "gen:Set"
	//HookSetNull will be called on Set for field (start, exit, error)
	HookSetNull HookType = "gen:SetNull"
	//HookGet will be called on Get for field (start, exit, error)
	HookGet HookType = "gen:Get"
	//HookNew will be called on New for entity (start, exit, error)
	HookNew HookType = "gen:New"
	//HookDelete will be called on Delete for entity (start, exit, error)
	HookDelete HookType = "gen:Del"
	//HookLoad will be called on Load for entity (start, exit, error)
	HookLoad HookType = "gen:Load"
	//HookSave will be called on Save for entity (start, exit, error) and for field (modified)
	HookSave HookType = "gen:Save"
	//HookUpdate will be called on Update for entity (start, exit, error) and for field (modified)
	HookUpdate HookType = "gen:Update"
	//HookSave will be called on Create for entity (start, exit, error) and for field (modified)
	HookCreate HookType = "gen:Create"
)

const (
	//HMStart - start of function
	HMStart HookModifier = "start"
	//HMExit - last statements of function
	HMExit HookModifier = "exit"
	//HMError - error in function
	HMError HookModifier = "error"
	//HMModified - field was modified (just before save)
	HMModified HookModifier = "modified"
)

//OnHook may be called during generation to add hooks code for item (may be *Field, *Method, *Entity)
//  vars contains var names for common objects (ctx, eng, obj); may be nil, in this case defaults will be used
func (p *Project) OnHook(name HookType, mod HookModifier, item interface{}, vars *GeneratorHookVars) (st *jen.Statement) {
	for _, hh := range p.hooks {
		var hs *jen.Statement
		switch t := item.(type) {
		case *Entity:
			hs, _ = hh.OnEntityHook(name, mod, t, vars)
		case *Field:
			hs, _ = hh.OnFieldHook(name, mod, t, vars)
		case *Method:
			hs, _ = hh.OnMethodHook(name, mod, t, vars)
		default:
			p.AddError(fmt.Errorf("OnHook was called for undefined type %T", item))
			return st
		}
		if hs != nil {
			if st == nil {
				st = hs
			} else {
				st = st.Add(hs).Line()
			}
		}
	}
	return
}

//GetCtx returns statement for context var; returns nil it is unavailabe
func (ha *GeneratorHookVars) GetCtx() *jen.Statement {
	if ha != nil {
		switch v := ha.Ctx.(type) {
		case bool:
			if v == false {
				return nil
			}
		case string:
			return jen.Id(v)
		case *jen.Statement:
			return v
		}
	}
	return jen.Id("ctx")
}

//MustCtx returns statement for context var; panics if it is unavailable
func (ha *GeneratorHookVars) MustCtx() *jen.Statement {
	if stmt := ha.GetCtx(); stmt != nil {
		return stmt
	}
	panic("generator hook: Ctx var is not accessible")
}

//GetEngine returns statement for engine var; returns nil it is unavailabe
func (ha *GeneratorHookVars) GetEngine() *jen.Statement {
	if ha != nil {
		switch v := ha.Eng.(type) {
		case bool:
			if v == false {
				return nil
			}
		case string:
			return jen.Id(v)
		case *jen.Statement:
			return v
		}
	}
	return jen.Id("eng")
}

//MustEngine returns statement for engine var; panics if it is unavailable
func (ha *GeneratorHookVars) MustEngine() *jen.Statement {
	if stmt := ha.GetCtx(); stmt != nil {
		return stmt
	}
	panic("generator hook: Engine var is not accessible")
}

//GetObject returns statement for object (this) var; returns nil it is unavailabe
func (ha *GeneratorHookVars) GetObject() *jen.Statement {
	if ha != nil {
		switch v := ha.Obj.(type) {
		case bool:
			if v == false {
				return nil
			}
		case string:
			return jen.Id(v)
		case *jen.Statement:
			return v
		}
	}
	return jen.Id("obj")
}

//MustEngine returns statement for engine var; panics if it is unavailable
func (ha *GeneratorHookVars) MustObject() *jen.Statement {
	if stmt := ha.GetCtx(); stmt != nil {
		return stmt
	}
	panic("generator hook: Object var is not accessible")
}

//GetObject returns statement for object (this) var; returns nil it is unavailabe
func (ha *GeneratorHookVars) GetVar(name string) *jen.Statement {
	if ha != nil && ha.Others != nil {
		switch v := ha.Others[name].(type) {
		case string:
			return jen.Id(v)
		case *jen.Statement:
			return v
		}
	}
	return nil
}

// NewHookVars creates new instance of *GeneratorHookVars with vars given as args
//  each var is a pair of name and value(string or *jen.Statement)
//  names may be well known names
func NewHookVars(vars ...interface{}) *GeneratorHookVars {
	ret := &GeneratorHookVars{}
	for i := 0; i < len(vars); i++ {
		name, ok := vars[i].(string)
		if !ok {
			break
		}
		i++
		if i >= len(vars) {
			break
		}
		v := vars[i]
		switch name {
		case GHVObject:
			ret.Obj = v
		case GHVContext:
			ret.Ctx = v
		case GHVEngine:
			ret.Eng = v
		default:
			if ret.Others == nil {
				ret.Others = map[string]interface{}{name: v}
			} else {
				ret.Others[name] = v
			}
		}
	}
	return ret
}
