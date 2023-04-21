package scripting

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"runtime/debug"
	"sync"
	"time"

	js "github.com/dop251/goja"
	// js "github.com/robertkrimen/otto"
)

type ErrorCode int

const (
	FileNotFound ErrorCode = iota
	CompilationError
	RuntimeError
)

var errorsDesc = [...]string{
	"Script file locating problem",
	"Compilation problem",
	"Execution problem",
}

type Error struct {
	Description string
	Code        ErrorCode
}
type script struct {
	name         string
	lastModified time.Time
	// byteCode     *js.Script
	byteCode *js.Program
	locker   sync.Mutex
}

type runtime struct {
	srv     *Service
	runtime *js.Runtime
}

// SetContext sets context for all subsequantial calls
func (s *Service) SetContext(ctx map[string]interface{}) {
	s.context = ctx
}

// AddModule adds internal module (may be loaded by require())
func (s *Service) AddModule(name string, module interface{}) {
	s.modules[name] = module
}

// Process looks for script to process given operation and runs it
// context objects will be put in the global scope of script
// return value contains the result of execution (with key "_")
// and variables's listed in results argument values
func (s *Service) Process(ctx context.Context,
	operation string,
	context map[string]interface{},
	results []string) (map[string]interface{}, error) {
	exitedChan := make(chan struct{})

	defer func() {
		close(exitedChan)
		if caught := recover(); caught != nil {
			s.log.Warn("Process: recovered from %v", zap.Any("recover", caught))
			s.log.Warn(string(debug.Stack()))
		}
	}()
	vm := js.New()

	// for otto
	// interruptChan := make(chan func(), 1)
	// vm.Interrupt = interruptChan
	go func() {
		select {
		case <-exitedChan:
		case <-ctx.Done():
			// for otto
			// interruptChan <- func() { panic(errors.New("interrupted")) }
			// for goya
			vm.Interrupt("interrupted")
		}
	}()
	if s.context != nil {
		for name, value := range s.context {
			vm.Set(name, value)
		}
	}
	if context != nil {
		for name, value := range context {
			vm.Set(name, value)
		}
	}
	rt := &runtime{srv: s, runtime: vm}
	vm.Set("ctx", ctx)
	vm.Set("trace", rt.scriptTrace)
	vm.Set("tracef", rt.scriptTracef)
	vm.Set("require", rt.require)
	// for ottot
	// exports, _ :=
	// vm.Object("aa = {}")
	exports := vm.NewObject()
	vm.Set("exports", exports)
	scr, er := s.getScript(operation, vm)
	if er != nil {
		return nil, er
	}
	s.log.Debug("Process: going to execute script", zap.String("operation", operation))
	// for otto
	// val, err := vm.Run(scr)
	val, err := vm.RunProgram(scr)

	if err != nil {
		s.log.Warn("Process: error executing script: %+v", zap.Error(err))
		// vm.
		return nil, newError(RuntimeError, err.Error())
	}
	returns := map[string]interface{}{}
	// for otto
	// returns["_"], _ = val.Export()
	returns["_"] = val.Export()
	s.log.Debug("script finished", zap.Any("result", returns["_"]), zap.Error(err))
	if results != nil {
		for _, variable := range results {
			// for otto
			// ottores, err := vm.Get(variable)
			// if err == nil {
			ottores := vm.Get(variable)
			if ottores != nil {
				// res, _ := ottores.Export()
				res := ottores.Export()
				returns[variable] = res
			} else {
				s.log.Debug("error from js", zap.Error(err))
				returns[variable] = err
			}
		}
	}

	return returns, nil
}

// ProcessSingleRet looks for script to process given operation and runs it
// context objects will be put in the global scope of script
// return value is the result of execution
func (s *Service) ProcessSingleRet(ctx context.Context,
	operation string,
	context map[string]interface{}) (interface{}, error) {
	ret, err := s.Process(ctx, operation, context, nil)
	if err != nil {
		return nil, err
	}
	return ret["_"], nil
}

func (s *Service) getScript(name string,
	// for otto
	// vm *js.Otto) (*js.Script, *Error) {
	vm *js.Runtime) (*js.Program, *Error) {

	fileName := s.prefix + name + s.suffix
	stat, err := os.Stat(fileName)
	if err != nil {
		s.log.Warn("getScript", zap.Error(err))
		return nil, newError(FileNotFound, err.Error())
	}
	mt := stat.ModTime()
	s.locker.Lock()
	scr, ok := s.scripts[name]
	if !ok {
		scr = &script{name: name}
		s.scripts[name] = scr
	}
	s.locker.Unlock()
	s.locker.Lock()
	defer s.locker.Unlock()
	if scr.lastModified.Before(mt) {
		scr.lastModified = mt
		// scr, err1 := vm.Compile(fileName, nil)
		if file, err1 := ioutil.ReadFile(fileName); err1 == nil {
			bc, err1 := js.Compile(fileName, string(file), false)
			if err1 == nil {
				scr.byteCode = bc
			}
			err = err1
		}
	}
	if err != nil {
		s.log.Warn("getScript: error while compiling", zap.Error(err))
		return nil, newError(CompilationError, err.Error())
	}
	return scr.byteCode, nil
}

func newError(code ErrorCode, problem string) *Error {
	return &Error{Code: code, Description: problem}
}

func (rt *runtime) scriptTrace(call js.FunctionCall) js.Value {
	message := "script trace:"
	for _, arg := range call.Arguments {
		message += fmt.Sprintf(" %v", arg.Export())
	}
	rt.srv.log.Debug(message)
	return rt.runtime.ToValue(true)
}

func (rt *runtime) scriptTracef(call js.FunctionCall) js.Value {
	args := make([]interface{}, len(call.Arguments)-1, len(call.Arguments)-1)
	for i := 1; i < len(call.Arguments); i++ {
		args[i-1] = call.Argument(i).Export()
	}
	format := call.Arguments[0].String()
	rt.srv.log.Debug(fmt.Sprintf("script tracef: "+format, args...))
	return rt.runtime.ToValue(true)
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", errorsDesc[e.Code], e.Description)
}
