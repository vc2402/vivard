package vivard

import (
	"encoding/json"
	"fmt"
	dep "github.com/vc2402/vivard/dependencies"
	"net/http"

	"github.com/graphql-go/graphql"
	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type GQLEngine struct {
	schema     *gql.Schema
	descriptor *GQLDescriptor
}

type GQLTypeGenerator func() gql.Output
type GQLInputTypeGenerator func() gql.Input
type GQLQueryGenerator func() *gql.Field

type GQLDescriptor struct {
	// engine              *Engine
	types               map[string]gql.Output
	inputs              map[string]gql.Input
	typesGenerators     map[string]GQLTypeGenerator
	inputsGenerators    map[string]GQLInputTypeGenerator
	queriesGenerators   map[string]GQLQueryGenerator
	mutationsGenerators map[string]GQLQueryGenerator
}

const (
	//Special types

	//KVStringString - name for special type for Key and Value pair
	KVStringString = "__kv_string_string"
	//KVStringStringInput - name for special input type for Key and Value pair
	KVStringStringInput = "__kv_string_string"
	//KVStringInt - name for special type for Key and Value pair
	KVStringInt = "__kv_string_int"
	//KVStringIntInput - name for special input type for Key and Value pair
	KVStringIntInput        = "__kv_string_int"
	KVStringIntName         = "__kv_string_int"
	KVStringStringName      = "__kv_string_string"
	KVStringStringInputName = "__kv_string_string_input" //"StringStringKVInput"
	KVStringIntInputName    = "__kv_string_int_input"    //"StringIntKVInput"
)

// func NewGQLService() *GQLEngine {
//   return &GQLEngine{
//     descriptor: createGQLDescriptor(),
//   }
// }

func (gqe *GQLEngine) Descriptor() *GQLDescriptor {
	return gqe.descriptor
}

func createGQLDescriptor() *GQLDescriptor {
	return &GQLDescriptor{
		// engine:              eng,
		types:               map[string]gql.Output{},
		inputs:              map[string]gql.Input{},
		typesGenerators:     map[string]GQLTypeGenerator{},
		inputsGenerators:    map[string]GQLInputTypeGenerator{},
		queriesGenerators:   map[string]GQLQueryGenerator{},
		mutationsGenerators: map[string]GQLQueryGenerator{},
	}
}

func (gqe *GQLEngine) generate(_ *Engine) error {
	gqld := gqe.descriptor
	for tn, tg := range gqld.typesGenerators {
		if gqld.types[tn] == nil {
			gqld.types[tn] = tg()
		}
	}
	queries := gql.Fields{}
	for qn, qg := range gqld.queriesGenerators {
		queries[qn] = qg()
	}
	rootQuery := gql.ObjectConfig{Name: "Query", Fields: queries}
	mutations := gql.Fields{}
	for mn, mg := range gqld.mutationsGenerators {
		mutations[mn] = mg()
	}
	rootMutations := gql.ObjectConfig{Name: "Mutation", Fields: mutations}
	schemaConfig := gql.SchemaConfig{
		Query:    gql.NewObject(rootQuery),
		Mutation: gql.NewObject(rootMutations),
		// Subscription: gql.NewObject(subscription),
	}
	sch, err := gql.NewSchema(schemaConfig)
	if err != nil {
		return err
	}
	gqe.schema = &sch
	return nil
}

// func (gqld *GQLDescriptor) Engine() *Engine {
// 	return gqld.engine
// }

func (gqe *GQLEngine) Prepare(_ *Engine, _ dep.Provider) error {
	gqe.descriptor = createGQLDescriptor()
	return nil
}
func (gqe *GQLEngine) Start(eng *Engine, _ dep.Provider) error {
	return gqe.generate(eng)
}

func (gqe *GQLEngine) Provide() interface{} {
	return gqe
}

func (gqe *GQLEngine) HTTPHandler() http.HandlerFunc {
	pretty := true
	h := handler.New(&handler.Config{
		Schema:   gqe.schema,
		Pretty:   pretty,
		GraphiQL: true,
	})
	return func(w http.ResponseWriter, r *http.Request) {
		// 	h.ContextHandler(r.Context(), w, r)
		// for statistics implement it yourself
		opts := handler.NewRequestOptions(r)

		// execute graphql query
		params := graphql.Params{
			Schema:         *h.Schema,
			RequestString:  opts.Query,
			VariableValues: opts.Variables,
			OperationName:  opts.OperationName,
			Context:        r.Context(),
		}
		// st := app.startQueryStatistics(opts.OperationName, opts.Query)
		// log.Tracef("Handler: opname: %s, query: %s", opts.OperationName, opts.Query)
		result := graphql.Do(params)
		// st.finish(len(result.Errors) == 0)

		// use proper JSON Header
		w.Header().Add("Content-Type", "application/json; charset=utf-8")

		var buff []byte
		if pretty {
			w.WriteHeader(http.StatusOK)
			buff, _ = json.MarshalIndent(result, "", "\t")

			w.Write(buff)
		} else {
			w.WriteHeader(http.StatusOK)
			buff, _ = json.Marshal(result)

			w.Write(buff)
		}

	}

}
func (gqld *GQLDescriptor) GetType(name string) gql.Output {
	if t, ok := gqld.types[name]; ok {
		return t
	}
	if g, ok := gqld.typesGenerators[name]; ok {
		t := g()
		gqld.types[name] = t
		return t
	}
	switch name {
	case KVStringString:
		t := gqld.getKVStringStringType()
		gqld.types[name] = t
		return t
	case KVStringInt:
		t := gqld.getKVStringIntType()
		gqld.types[name] = t
		return t
	default:
		panic(fmt.Sprintf("undefined gql type '%s'", name))
	}
}
func (gqld *GQLDescriptor) GetInputType(name string) gql.Input {
	if t, ok := gqld.inputs[name]; ok {
		return t
	}
	if g, ok := gqld.inputsGenerators[name]; ok {
		t := g()
		gqld.inputs[name] = t
		return t
	}
	switch name {
	case KVStringStringInput:
		t := gqld.getKVStringStringInputType()
		gqld.inputs[name] = t
		return t
	case KVStringIntInput:
		t := gqld.getKVStringIntInputType()
		gqld.inputs[name] = t
		return t
	default:
		panic(fmt.Sprintf("undefined gql input type '%s'", name))
	}
}

func (gqld *GQLDescriptor) AddTypeGenerator(name string, g GQLTypeGenerator) {
	if gqld.typesGenerators[name] != nil {
		panic(fmt.Sprintf("duplicate gql type '%s'", name))
	}
	gqld.typesGenerators[name] = g
}

func (gqld *GQLDescriptor) AddInputGenerator(name string, g GQLInputTypeGenerator) {
	if gqld.inputsGenerators[name] != nil {
		panic(fmt.Sprintf("duplicate gql input type '%s'", name))
	}
	gqld.inputsGenerators[name] = g
}

func (gqld *GQLDescriptor) AddQueryGenerator(name string, g GQLQueryGenerator) {
	if gqld.queriesGenerators[name] != nil {
		panic(fmt.Sprintf("duplicate gql query '%s'", name))
	}
	gqld.queriesGenerators[name] = g
}

func (gqld *GQLDescriptor) AddMutationGenerator(name string, g GQLQueryGenerator) {
	if gqld.mutationsGenerators[name] != nil {
		panic(fmt.Sprintf("duplicate gql mutation '%s'", name))
	}
	gqld.mutationsGenerators[name] = g
}

func (gqld *GQLDescriptor) getKVStringStringType() *graphql.Object {
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name: KVStringStringName,
			Fields: graphql.Fields{
				"key": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						r := p.Source.(KVPair)
						return r.KeyString()
					},
				},
				"val": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						r := p.Source.(KVPair)
						return r.ValString()
					},
				},
			},
		},
	)
}

func (gqld *GQLDescriptor) getKVStringStringInputType() *graphql.InputObject {
	return graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name: KVStringStringInputName,
			Fields: graphql.InputObjectConfigFieldMap{
				"key": &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				"val": &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
			},
		},
	)
}

func GQLArgToKVStringString(arg map[string]interface{}) (ret KVPair, err error) {
	var ok bool
	ret.Key, ok = arg["key"].(string)
	if !ok {
		err = fmt.Errorf("can not get <string> key from %v", arg)
		return
	}
	ret.Val, ok = arg["val"].(string)
	if !ok {
		err = fmt.Errorf("can not get <string> val from %v", arg)
		return
	}
	return
}

func GQLArgToArrKVStringString(arg []interface{}) (ret ArrKV, err error) {
	ret = make(ArrKV, len(arg))
	for i, a := range arg {
		if am, ok := a.(map[string]interface{}); ok {
			ret[i], err = GQLArgToKVStringString(am)
			if err != nil {
				return
			}
		} else {
			err = fmt.Errorf("map[string]interface{} is expected; but found: %v", a)
			return
		}
	}
	return
}

func GQLArgToMapStringString(arg []interface{}) (ret map[string]string, err error) {
	ret = make(map[string]string, len(arg))
	for _, a := range arg {
		if am, ok := a.(map[string]interface{}); ok {
			kv, e := GQLArgToKVStringString(am)
			if e != nil {
				err = e
				return
			}
			ret[kv.KeyStr()] = kv.ValStr()
		} else {

			err = fmt.Errorf("map[string]interface{} is expected; but found: %v", a)
			return
		}
	}
	return
}

func (gqld *GQLDescriptor) getKVStringIntType() *graphql.Object {
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name: KVStringIntName,
			Fields: graphql.Fields{
				"key": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						r := p.Source.(KVPair)
						return r.KeyString()
					},
				},
				"val": &graphql.Field{
					Type: graphql.NewNonNull(graphql.Int),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						r := p.Source.(KVPair)
						return r.ValInteger()
					},
				},
			},
		},
	)
}

func (gqld *GQLDescriptor) getKVStringIntInputType() *graphql.InputObject {
	return graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name: KVStringIntInputName,
			Fields: graphql.InputObjectConfigFieldMap{
				"key": &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				"val": &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
			},
		},
	)
}

func GQLArgToKVStringInt(arg map[string]interface{}) (ret KVPair, err error) {
	var ok bool
	ret.Key, ok = arg["key"].(string)
	if !ok {
		err = fmt.Errorf("can not get <string> key from %v", arg)
		return
	}
	ret.Val, ok = arg["val"].(int)
	if !ok {
		err = fmt.Errorf("can not get <string> val from %v", arg)
		return
	}
	return
}

func GQLArgToArrKVStringInt(arg []interface{}) (ret ArrKV, err error) {
	ret = make(ArrKV, len(arg))
	for i, a := range arg {
		if am, ok := a.(map[string]interface{}); ok {
			ret[i], err = GQLArgToKVStringInt(am)
			if err != nil {
				return
			}
		} else {
			err = fmt.Errorf("map[string]interface{} is expected; but found: %v", a)
			return
		}
	}
	return
}

func GQLArgToMapStringInt(arg []interface{}) (ret map[string]int, err error) {
	ret = make(map[string]int, len(arg))
	for _, a := range arg {
		if am, ok := a.(map[string]interface{}); ok {
			kv, e := GQLArgToKVStringInt(am)
			if e != nil {
				err = e
				return
			}
			ret[kv.KeyStr()] = kv.ValInt()
		} else {
			err = fmt.Errorf("map[string]interface{} is expected; but found: %v", a)
			return
		}
	}
	return
}
