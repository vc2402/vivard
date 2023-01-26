package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
)

var (
	//CodeFragmentModuleGeneral module of general functionality
	CodeFragmentModuleGeneral = struct{}{}
)

const (
	ParamContext      = "ctx"
	ParamID           = "id"
	ParamObject       = "obj"
	ParamVivardEngine = "vivard"
)

//CodeFragmentContext holds context of function
type CodeFragmentContext struct {
	MethodName string
	TypeName   string
	MethodKind
	//EngineAvailable true if eng var does exist in scope
	EngineAvailable bool
	//ErrVar holds err var name or empty if err not defined yet
	ErrVar string
	//ObjVar holds name of object var
	ObjVar string
	//Builder reference to Builder
	*Builder
	//Package refers to Package object if it is Package context (e.g. Engine fragments)
	*Package
	//Entity holds reference to Entity definition if it is entity-level function
	Entity *Entity
	//Field holds reference to Field if it is Field-level function
	Field *Field
	//Params holds function params by well-known names
	Params map[string]string
	//ErrorRet should be set to correct code generated by AddCheckError method; if empty return statement will be also empty
	ErrorRet []jen.Code
	//BeforeReturnError can be set to add process before error return statement
	BeforeReturnError func()
	//body holds current code
	body             *jen.Statement
	currentStatement *jen.Statement
	currentGroup     *jen.Group
	groupsStack      []*jen.Group
}

func (cf *CodeFragmentContext) GetErrName() string {
	if cf.ErrVar == "" {
		cf.ErrVar = "err"
		cf.Add(jen.Var().Id(cf.ErrVar).Error())
	}
	return cf.ErrVar
}

func (cf *CodeFragmentContext) GetErr() *jen.Statement {
	return jen.Id(cf.GetErrName())
}

func (cf *CodeFragmentContext) GetObjVarName() string {
	if cf.ObjVar == "" {
		if cf.Entity != nil {
			cf.ObjVar = "obj"
			cf.Add(jen.Var().Id(cf.ObjVar).Op("*").Id(cf.Entity.Name))
		} else {
			panic("can not determine obj type")
		}
	}
	return cf.ObjVar
}

func (cf *CodeFragmentContext) GetObjVar() *jen.Statement {
	return jen.Id(cf.GetObjVarName())
}

func (cf *CodeFragmentContext) Add(code ...jen.Code) *jen.Statement {
	if cf.currentGroup != nil {
		return cf.currentGroup.Add(code...)
	}
	if cf.currentStatement != nil {
		return cf.currentStatement.Add(code...)
	}
	if cf.body == nil {
		cf.body = &jen.Statement{}
	}
	return cf.body.Add(code...)
}

func (cf *CodeFragmentContext) GetParamName(name string) string {
	if param, ok := cf.Params[name]; ok {
		return param
	}
	panic(fmt.Sprintf("param '%s' not defined for Method %s", name, cf.MethodName))

}
func (cf *CodeFragmentContext) GetParam(name string) *jen.Statement {
	return jen.Id(cf.GetParamName(name))
}

func (cf *CodeFragmentContext) AddCheckError() {
	cf.Add(jen.If(cf.GetErr()).Op("!=").Nil().BlockFunc(func(g *jen.Group) {
		cf.Push(g)
		cf.AddOnErrorReturnStatement()
		cf.Pop()
	}))
}

func (cf *CodeFragmentContext) AddOnErrorReturnStatement() {
	if cf.BeforeReturnError != nil {
		cf.BeforeReturnError()
	} else if p := cf.Project(); p != nil {
		p.ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGPointExitError, cf, false)
	}
	switch len(cf.ErrorRet) {
	case 0:
		cf.Add(jen.Return())
	case 1:
		cf.Add(jen.Return(cf.ErrorRet[0]))
	default:
		cf.Add(jen.Return(jen.List(cf.ErrorRet...)))
	}
}
func (cf *CodeFragmentContext) Enter(beforeHooks bool) {
	if beforeHooks {
		cf.MustProject().ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGPointEnterBeforeHooks, cf, false)
	} else {
		cf.MustProject().ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGPointEnterAfterHooks, cf, false)
	}
}

func (cf *CodeFragmentContext) Exit(beforeHooks bool) {
	if beforeHooks {
		cf.MustProject().ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGPointExitBeforeHooks, cf, false)
	} else {
		cf.MustProject().ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGPointExitAfterHooks, cf, false)
	}
}

func (cf *CodeFragmentContext) MainAction() bool {
	return cf.MustProject().ProvideCodeFragment(CodeFragmentModuleGeneral, cf.MethodKind, CFGPointMainAction, cf, true) != nil
}

func (cf *CodeFragmentContext) Push(g *jen.Group) {
	if cf.currentGroup != nil {
		cf.groupsStack = append(cf.groupsStack, cf.currentGroup)
	}
	cf.currentGroup = g
}

func (cf *CodeFragmentContext) Pop() {
	stackLen := len(cf.groupsStack)
	if stackLen > 0 {
		cf.currentGroup = cf.groupsStack[stackLen-1]
		cf.groupsStack = cf.groupsStack[:stackLen-1]
	} else {
		cf.currentGroup = nil
	}
}

func (cf *CodeFragmentContext) Project() *Project {
	if cf.Package != nil {
		return cf.Package.Project
	} else if cf.Builder != nil {
		return cf.Builder.Project
	} else {
		return nil
	}
}

func (cf *CodeFragmentContext) MustProject() *Project {
	p := cf.Project()
	if p != nil {
		return p
	} else {
		panic("no way to get Project object")
	}
}
