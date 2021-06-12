package scripting

import (
	js "github.com/dop251/goja"
	// js "github.com/robertkrimen/otto"
)

type module struct {
	rt      *runtime
	exports map[string]js.Value
}

func (rt *runtime) require(call js.FunctionCall) js.Value {
	module := call.Argument(0)
	// if module.IsString() {
	mName := module.String()
	// val, err := rt.runtime.Get(mName)
	val, ok := rt.srv.modules[mName]
	// if err == nil && !val.IsUndefined() {
	if ok {
		rt.srv.log.Tracef("require: returning %+v", val)
		return rt.runtime.ToValue(val)
	} else {
		prg, err := rt.srv.getScript(mName, rt.runtime)
		if err != nil {
			rt.srv.log.Warnf("require: Problem while loading module %s: %v", mName, err)
		} else {
			exp := rt.runtime.Get("exports")
			rt.runtime.Set("exports", rt.runtime.NewObject())
			ret, err := rt.runtime.RunProgram(prg)
			if err != nil {
				rt.srv.log.Warnf("require: Problem while loading module %s: executing: %v", mName, err)
			} else {
				ret = rt.runtime.Get("exports")
				rt.runtime.Set("exports", exp)
				rt.srv.log.Tracef("require: returning %+v", ret)
				return ret
			}
		}
	}
	rt.srv.log.Warnf("require: Problem while loading module %s", mName)
	// return js.UndefinedValue()
	return js.Undefined()
}
