package gen

import "github.com/dave/jennifer/jen"

// Generator interface can be used for creating custom generators
type Generator interface {
	// Name returns the unique name of generator (plugin)
	Name() string
	// CheckAnnotation calls for every annotation to check that it is understandable
	// item contains whether *Entity, *Field or *Method
	// return (true, nil) if Generator does understand this annotation
	//        (false, nil) if Generator does not understand this annotation
	// error should be returned if annotation is understandable but contains errors
	CheckAnnotation(desc *Package, ann *Annotation, item interface{}) (bool, error)
	// Prepare - prepare to generation
	Prepare(desc *Package) error
	// Generate - do generate
	Generate(bldr *Builder) error
}

// MetaProcessor meta processor interface
type MetaProcessor interface {
	//ProcessMeta should try to process current slice in given meta and return true on success;
	//  error shows that there are problems in known format
	ProcessMeta(desc *Package, m *Meta) (bool, error)
}

// OptionsSetter may be implemented by Generator if it depends on options
type OptionsSetter interface {
	// SetOptions sets plugin specific options
	SetOptions(options any) error
}

// ProvideFeatureResult special type for ProvideFeature return value
type ProvideFeatureResult int

const (
	// FeatureNotProvided - provider can not provide this feature
	FeatureNotProvided ProvideFeatureResult = iota
	// FeatureProvided - feature result present feature that can be cached
	FeatureProvided
	// FeatureProvidedNonCacheable - feature result present feature that can not be cached
	FeatureProvidedNonCacheable
)

// FeatureProvider provides features for code generation
type FeatureProvider interface {
	//ProvideFeature - returns feature result and ok if can provided requested feature; nil and false otherwise
	// params: kind and name of feature,
	// obj - whether *Entity, *Field or *Method
	ProvideFeature(kind FeatureKind, name string, obj interface{}) (feature interface{}, ok ProvideFeatureResult)
}

// DescriptorAware may be used in case if generator requires reference to descriptor object
type DescriptorAware interface {
	//SetDescriptor will be called for all objects before first call of Generator methods
	SetDescriptor(proj *Project)
}

// CodeFragmentProvider provides fragment of code
type CodeFragmentProvider interface {
	//ProvideCodeFragment should return code fragment for given point
	// module is high-level generation abstraction (e.g., go skeleton or db)
	// action is more low-level (like setter) and depends on module
	// point is place inside action (e.g. function-enter)
	// ctx contains context like params, variables names etc.
	// for go it should be *CodeFragmentContext instance and code should be added via this object;
	//  return value of nil in this case shows that nothing was added;
	ProvideCodeFragment(module interface{}, action interface{}, point interface{}, ctx interface{}) interface{}
}

// CodeHelperFunc Feature type for helping generate code (params depends on feature)
type CodeHelperFunc func(args ...interface{}) jen.Code

// FeatureFunc may be returned by feature provider
type FeatureFunc func(args ...interface{}) (any, error)

type HookArgParam struct {
	Name  string
	Param interface{}
}

// HookArgsDescriptor holds args for feature hook function
type HookArgsDescriptor struct {
	// Str  - string arg (defaultName for Go hooks)
	Str string
	// Ctx - variable name or context *jen.Statement if it is not "ctx"
	Ctx interface{}
	// Eng - variable name or Engine *jen.Statement if it is not "eng"; false to force not add engine arg
	Eng interface{}
	// Obj - variable name or object *jen.Statement if it is not "obj"
	Obj interface{}
	//Params - additional params with their names
	Params []HookArgParam
	//ErrVar - assign error return value to variable
	ErrVar interface{}
}

// HookFeatureFunc func can be returned as feature that can create code for hook
type HookFeatureFunc func(args HookArgsDescriptor) jen.Code

// MethodKind type of generated method
type MethodKind int

const (
	// methods for retrieving items from cache for example (entry points)

	//MethodGet for Getter for type
	MethodGet MethodKind = iota
	//MethodSet is setter
	MethodSet
	//MethodNew is creator of new instance (gets already filled struct as argument)
	MethodNew
	//MethodDelete deletes instance
	MethodDelete
	//MethodGenerateID returnes id for new entity
	MethodGenerateID
	//MethodGetAll may be used for types with small amount of instances (dictionaries, etc)
	MethodGetAll
	//MethodInit inits new struct (if necessary fills id with auto value)
	MethodInit
	// MethodNewBulk creates set of new instances
	MethodNewBulk

	// methods for store-based actions (db)

	//MethodLoad loads entity from a store
	MethodLoad
	//MethodSave saves existing entity into the store (replace entity with modified one)
	MethodSave
	//MethodUpdate updates existing entity into the store (only modified fields if available)
	MethodUpdate
	//MethodCreate creates (inserts) new entity
	MethodCreate
	//MethodRemove removes (or marks as deleted) entity
	MethodRemove
	//MethodRemoveFK removes entities for given parent key
	MethodRemoveFK
	//MethodReplaceFK removes entities for given parent key and inserts new (params: id and array of entities)
	MethodReplaceFK
	//MethodLookup returns items for given query
	MethodLookup
	//MethodList returns all items (for dictionaries)
	MethodList
	//MethodListFK returns all items (for one-to-many fields)
	MethodListFK
	//MethodFind can be used for looking for objects by some parameters
	MethodFind
	//MethodChanged returns true if attr was changed
	MethodChanged

	methodMax
	EngineNotAMethod
	MethodEnginePrepare
	MethodEngineStart
	MethodEngineRegisterService
	TypeFieldNotAMethod
)

// MethodsNamesTemplates contains templates for standart methods names
var MethodsNamesTemplates = [methodMax]string{
	"Get%s",
	"Set%s",
	"New%s",
	"Delete%s",
	"Generate%sID",
	"GetAll%s",
	"Init%s",
	"New%ss",
	"Load%s",
	"Save%s",
	"Update%s",
	"Create%s",
	"Remove%s",
	"RemoveFK%s",
	"ReplaceFK%s",
	"Lookup%s",
	"List%s",
	"ListFK%s",
	"Find%s",
	"Is%sChanged",
}

const (
	TipString = "string"
	TipInt    = "int"
	TipFloat  = "float"
	TipBool   = "bool"
	TipDate   = "date"
	TipAny    = "any"
	TipAuto   = "auto"

	EngineVar = "eng"

	ExtendableTypeDescriptorFieldName = "V_Type"
)

type AttrModifier string

const (
	AttrModifierEmbedded    AttrModifier = "embedded"
	AttrModifierEmbeddedRef AttrModifier = "ref-embedded"
	AttrModifierID          AttrModifier = "id"
	AttrModifierIDAuto      AttrModifier = "auto"
	AttrModifierForeignKey  AttrModifier = "foreign-key"
	AttrModifierOneToMany   AttrModifier = "one-to-many"
	AttrModifierCalculated  AttrModifier = "calculated"
	AttrModifierAuxiliary   AttrModifier = "auxiliary"
)

type TypeModifier string

const (
	TypeModifierAbstract   TypeModifier = "abstract"
	TypeModifierConfig     TypeModifier = "config"
	TypeModifierDictionary TypeModifier = "dictionary"
	TypeModifierEmbeddable TypeModifier = "embeddable"
	TypeModifierExtendable TypeModifier = "extendable"
	TypeModifierExternal   TypeModifier = "extern"
	TypeModifierSingleton  TypeModifier = "singleton"
	TypeModifierTransient  TypeModifier = "transient"
)
const (
	autoGeneratedIDFieldName = "ID"

	VivardPackage       = "github.com/vc2402/vivard"
	VivardPackageAlias  = "vivard"
	dependenciesPackage = "github.com/vc2402/vivard/dependencies"
)

const (
	//AnnotationFind may be used for defining Find params definition
	AnnotationFind = "find"

	//AnnFndFieldTag tag for find Field annotation - links input field to Entity field
	AnnFndFieldTag = "field"
	//AnnFndTypeTag tag for comparision type (see AFT*values)
	AnnFndTypeTag = "type"
	//AFFDeleted special value for field deleted
	AFFDeleted = "_deleted_"
	//AFTEqual - value for type find tag - equal to (default value)
	AFTEqual = "eq"
	//AFTNotEqual - value for type find tag - not equal to
	AFTNotEqual = "ne"
	// AFTGreaterThan - value for type find tag - greater than
	AFTGreaterThan = "gt"
	// AFTGreaterThanOrEqual - value for type find tag - greater than or equal to
	AFTGreaterThanOrEqual = "gte"
	// AFTLessThan - value for type find tag - less than
	AFTLessThan = "lt"
	// AFTLessThanOrEqual - value for type find tag - less than or equal to
	AFTLessThanOrEqual = "lte"
	// AFTStartsWith - like or regexp for start
	AFTStartsWith = "starts-with"
	// AFTContains - like or regexp for start
	AFTContains = "contains"
	// AFTStartsWithIgnoreCase - like or regexp for start
	AFTStartsWithIgnoreCase = "starts-with-ignore-case"
	// AFTContainsIgnoreCase - like or regexp for start
	AFTContainsIgnoreCase = "contains-ignore-case"
	// AFTIgnore - for deleted; if true - include deleted items
	AFTIgnore = "ignore"
	// AFTIsNull - null nullable fields (value should be bool)
	AFTIsNull = "is-null"

	// AnnotationLookup for lookup function generation
	AnnotationLookup       = "lookup"
	ALEqual                = "eq"
	ALNotEqual             = "ne"
	ALStartsWith           = "startsWith"
	ALStartsWithIgnoreCase = "startsWithIgnoreCase"
	ALContains             = "contains"
	ALContainsIgnoreCase   = "containsIgnoreCase"

	// AnnotationCall - annotation for hook method
	AnnotationCall = "call"
	//AnnCallName - name of function to call
	AnnCallName = "name"
	//AnnCallJS - bool - name is JS script name
	AnnCallJS = "js"

	//AnnotationGo for common go-specific annotation
	AnnotationGo = "go"
	//AnnGoPackage defines package name for external type
	AnnGoPackage = "package"

	//AnnotationConfig for config members
	AnnotationConfig = "config"
	//AnnCfgValue shows that type describes cfg value
	AnnCfgValue = "value"
	//AnnCfgGroup shows that type describes cfg group
	AnnCfgGroup = "group"
	//AnnCfgMutable shows that this config field can be changed (generates save func for it)
	AnnCfgMutable = "mutable"

	//AnnotationDefault set default value for entity's field
	AnnotationDefault = "default"

	// AnnotationSort indicate to use this field for sorting (usually at db access moment)
	AnnotationSort = "sort"
	// AnnSortAscending sort in ascending order
	AnnSortAscending = "asc"
	// AnnSortDescending sort in descending order
	AnnSortDescending = "desc"
)

type FeatureKind string

const (
	// FeaturesCommonKind kind for common features
	FeaturesCommonKind FeatureKind = "gen"

	//FeaturesHookCodeKind - code for *Entity; creates hook code for hook's string (empty string or name with prefix);
	//    args: defaultNameOfFunction - string for function name if not set by modifier
	//          obj - object to call hook on
	//          newObj - new (old) value for object where applicable (nil if not)
	//          ctx - context object (string or jen.Code) (ctx by default)
	//	        eng - engine (eng by default)
	FeaturesHookCodeKind FeatureKind = "hook"

	//FCIgnore - bool; ignore common actions for this field/type
	FCIgnore = "ignore"
	// FCOneToManyField - *Field; common feature for description one to many field
	FCOneToManyField = "one-to-many-field"
	// FCOneToManyType - *Entity; common feature, reference to on to many type
	FCOneToManyType = "one-to-many-type"
	// FCForeignKey - *Entity; common feature for type showing that this type is many-to-one relation entity (value - type it references to)
	FCForeignKey = "foreign-key"
	// FCForeignKeyField - *Field; common feature for type, reference to its field that holds FK id
	FCForeignKeyField = "foreign-key-field"
	//FCModifiedFieldName - string; name for generated field storing value that field value was modified; should be requested via Descriptor
	FCModifiedFieldName = "modified-field-name"
	//FCViaParentAccessors - bool; generate accessors only via parent's object (for foreign-key types)
	FCViaParentAccessors = "access-via-parent"
	//FCComplexAccessor - bool; for field; if true - getter and setter for this field are complex (via eng)
	FCComplexAccessor = "complex-accessor"
	//FCSkipAccessors - bool, *Entity; do not generate setter and getter methods for Type
	FCSkipAccessors = "skip-accessors"
	//FCAttrIsPointer - bool; *Field; true if field for attr is pointer in the Go struct
	FCAttrIsPointer = "attr-is-pointer"
	//FCManyToManyType - type references to
	FCManyToManyType = "many-to-many-type"
	//FCManyToManyIDField - reference to id Field
	FCManyToManyIDField = "many-to-many-id"
	//FCRefsAsManyToMany - boolean for *Entity; true if there are refs to this entity as many to many field
	FCRefsAsManyToMany = "refs-many-to-many"
	//FCReadonly - boolean for *Entity(hardcodded) or *Field - the field is unmutable from api
	FCReadonly = "readonly"

	//Code features

	//FCObjIDCode returns code for *Entity: obj.Id(param if any may be string (var name) or jen.Code (var); by default - obj)
	FCObjIDCode = "id-accessor-code"
	//FCSetterCode returns code For *Field for setter(params if any are obj, value, context, defaults: obj, val, ctx)
	FCSetterCode = "setter-code"
	//FCGetterCode returns code for:
	//   *Field for getter(param if any are obj and context, defaults: obj, ctx);
	// or *Entity get object(params if any are objectID, ctx and engine, defaults: id, ctx, eng)
	//  first bool param is 'append ret with error' (by default true fo Entity)
	FCGetterCode = "getter-code"
	//FCIsNullCode returns code for IsAttrNull
	FCIsNullCode = "is-attr-null"
	//FCSetNullCode returns code for SetAttrNull
	FCSetNullCode = "set-attr-null"
	//FCAttrValueCode - code for access dereferenced attr value (returns code for e.g. *obj.attr); param (if any): obj
	FCAttrValueCode = "get-attr-code"
	//FCAttrSetCode - code to set dereferenced attr value  (returns code for e.g. *obj.attr = val); params (if any): obj, val
	FCAttrSetCode = "set-attr-code"
	//FCAttrIsEmptyCode - code for create bool value to check whether attr is empty  (returns code for e.g. obj.attr = nil); params (if any): obj
	//  if bool param found and it is true, returns nil if attr can't be empty
	FCAttrIsEmptyCode = "attr-is-empty-code"
	//FCListDictByIDCode - code for *Entity; list dictionary items by their ids; params: ids - array of ids, ctx
	FCListDictByIDCode = "fc-list-dict-by-id"
	//FCListByIDCode - code for *Entity; list items by their ids; params: ids - array of ids, ctx
	FCListByIDCode = "fc-list-by-id"
	//FCDictGetter - code for readonly dict getter (if any)
	FCDictGetter = "fc-dict-getter"
	//FCDictIniter - code for init dict values (if empty)
	FCDictIniter = "fc-dict-initer"
	//FCEngineVar - code feature for Field - returns code for engine to access this field (usualy *Engine but checks package)
	FCEngineVar = "engine-var"
	//FCDescendants - array of *Entity that are descendants of this type
	FCDescendants = "descendants"
)

const (
	//FeaturesDBKind - common kind for db features
	FeaturesDBKind FeatureKind = "db"

	//FDBIncapsulate - bool; (for one-to-many field) store as array in document db
	FDBIncapsulate = "incapsulate"
	//FDBFlushDict - code for flushing whole dictionary to storage
	FDBFlushDict = "flush_dict"
)

const (
	//hooks for type
	// creates for *Type; params: ctx Context.context, eng *Engine, old(new)Value *Type

	//TypeHookCreate creates hook to call before new object creation; default name: "OnCreate"; can be set only for singleton
	TypeHookCreate = "create"
	//TypeHookChange creates hook to call before object change; default name: "OnChange";
	// params: ctx Context.context, eng *Engine, oldValue *Type, newValue *Type
	//  oldValue is nil for Create operation; newValue is nil for Delete operation
	//  should return error, if return not nil changes will not be saved
	TypeHookChange = "change"
	//TypeHookChanged creates hook to call after object was changed; default name: "OnSaved";
	// params: ctx Context.context, eng *Engine, oldValue *Type, newValue *Type
	//  oldValue is nil for Create operation; newValue is nil for Delete operation
	//  should return error, if return not nil changes will not be saved
	TypeHookChanged = "changed"
	//TypeHookStart creates hook to call on singleton after all the objects created; default name: "OnStart";
	TypeHookStart = "start"
	//TypeHookMethod is used for Methods calls
	TypeHookMethod = "method"
	//TypeHookTime allows to call method of singleton at specific time or periodically
	TypeHookTime = "time"
	//TypeHookDelete sets Entity's hook should be called before deleting an instance; default name OnDelete.
	//  Function should return error; if not nil object will not be deleted
	TypeHookDelete = "delete"

	//hooks for fields

	//AttrHookSet - set may be used for complex fields (will be called before save)
	AttrHookSet = "set"
	//AttrHookCalculate - will be called when calculated field should be resolved (method of *Engine with params ctx and *object)
	AttrHookCalculate = "resolve"

	//hooks for methods

	//MethodHookTime allows to call method of singleton at specific time or periodically
	MethodHookTime = "time"

	//HookJSPrefix - prefix for js script
	HookJSPrefix = "js"
	//HookGoPrefix - prefix for go function
	HookGoPrefix = "go"
	//WithoutEngSuffix - suffix of name string showing not to include engine param
	WithoutEngSuffix = "%eng"
)

var hookFuncsTemmplates = map[string]string{
	TypeHookCreate:    "OnCreate",
	TypeHookChange:    "OnChange",
	TypeHookChanged:   "OnSaved",
	TypeHookStart:     "OnStart",
	AttrHookSet:       "On%s%sSet",
	AttrHookCalculate: "%sResolve%s",
	TypeHookDelete:    "OnDelete",
}

const (
	//FeaturesAPIKind - common kind of api specific fetures and flags
	FeaturesAPIKind FeatureKind = "api"

	//FAPILevel - string; for *Entity; level of API generation
	FAPILevel = "level"
	//FAPILIgnore - value for FAPILevel: generate nothing
	FAPILIgnore = "ignore"
	//FAPILTypes - value for FAPILevel: generate types only
	FAPILTypes = "types"
	//FAPILAll - value for FAPILevel: generate everything
	FAPILAll = "all"

	//FAPIFindParam - *Entity; for *Entity; type for find method param
	FAPIFindParamType = "find-param-type"
	//FAPIFindFor - for *Entity or *Field - ref to object this is find for
	FAPIFindFor = "find-for"
	//FAPIFindForEmbedded - for *Field - array of fields (if find for attr of embedded type)
	FAPIFindForEmbedded = "find-for-emb"
	//FAPIFindForName - for *Field - name of type this is find for
	FAPIFindForName = "find-for-name"
	//FAPIFindParam - string; for *Field; find param descriptor (values as for @AnnFndTypeTag)
	FAPIFindParam = "find-param"
)

const (
	//FeatureChangeDetectorKind - kind of features related to detecting of fields changes
	FeaturesChangeDetectorKind FeatureKind = "changes"

	//FCDRequired for Field should be bool; for Entity - one of FCDREntity or FCDRField
	FCDRequired = "required"
	//FCDREntity - value for FCDRequired (for Entity) - detect changes for all the fields of Entity
	FCDREntity = "entity"
	// FCDRField - value for FCDRequired (for Entity) - detect changes only for selected fields (with FCDRequired = true)
	FCDRField = "field"
	//FCDChangedHook bool feature for Entity or Field - returns true if exists change hook generator and hook will be called (use via Project.GetFeature)
	FCDChangedHook = "changed-hook"
	//FCDChangedCode - code for *Field - generates bool expression, that returns true if field was changed (better approach is to use hook)
	FCDChangedCode = "changed"
)
