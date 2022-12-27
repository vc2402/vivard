package resource

import "errors"

var (
	ErrForbidden       = errors.New("forbidden")
	ErrUnknownResource = errors.New("resource not found")
)

type ID string

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
	CreateResource()
}
