package resource

import "errors"

var (
	ErrForbidden       = errors.New("forbidden")
	ErrUnknownResource = errors.New("resource not found")
	ErrDuplicate       = errors.New("duplicate key")
)

const (
	ServiceAccessChecker  = "resource:access-checker"
	ServiceChangeNotifier = "resource:change-notifier"
	ServiceManager        = "resource:manager"
)

//Key may be used to uniquely identify resource by name; generally resource may have unique int id
type Key string

type ID int

type AccessKind int

const (
	AccessRead AccessKind = iota
	AccessWrite
	AccessCreate
	AccessDelete
	AccessList
)

//AccessChecker defines access check implementation interface
type AccessChecker interface {
	//CheckResourceAccess called when it is necessary to check access rights to Resource with given id;
	//  objectID may be nil or contain id of Resource instance
	//  accessKind defines action performing with resource
	//  return should be ErrForbidden if access is forbidden or nil if granted;
	//  in general may contain other value (e.g. ErrUnknownResource)
	CheckResourceAccess(id ID, objectID interface{}, accessKind AccessKind) (err error)
	// CheckResourceAccessByKey may be used for checking access by resource key; see CheckResourceAccess
	CheckResourceAccessByKey(key Key, objectID interface{}, accessKind AccessKind) (err error)
}

type ChangeKind int

const (
	ChangeModified ChangeKind = iota
	ChangeCreated
	ChangeDeleted
)

//ChangeNotifier may be used for notification of resource changes
type ChangeNotifier interface {
	NotifyResourceChanged(id ID, objectID interface{}, kind ChangeKind)
}

type Manager interface {
	// FindResource looks for resource dy key and returns its id or ErrUnknownResource
	FindResource(key Key) (ID, error)
	//CreateResource creates new resource; parent may be nil
	CreateResource(key Key, description string, parentKey *Key) (ID, error)
}
