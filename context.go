package vivard

import "context"

// Context is the interface for default request context object
type Context interface {
	UserID() int
	UserType() int
	UserName() string
	Source() string
	HasRole(role string) bool
	RolesMask() int
	GetExt(key string) interface{}
}

type DefaultContext struct {
	userID    int
	userType  int
	userName  string
	source    string
	roles     []string
	rolesMask int
	ext       map[string]interface{}
}

func (c DefaultContext) GetExt(key string) (interface{}, bool) {
	if c.ext == nil {
		return nil, false
	}
	ret, ok := c.ext[key]
	return ret, ok
}

var ContextID = &struct{}{}

func NewContext(ctx context.Context, userID int, userName string, source string, roles []string, rolesMask int, ext ...interface{}) context.Context {
	newCtx := DefaultContext{userID: userID, userName: userName, source: source, roles: roles, rolesMask: rolesMask}
	if len(ext) > 0 {
		newCtx.ext = map[string]interface{}{}
		for i := 0; i < len(ext)-1; i += 2 {
			if key, ok := ext[i].(string); ok {
				newCtx.ext[key] = ext[i+1]
			}
		}
	}
	return context.WithValue(ctx, ContextID, newCtx)
}

func RequestContext(ctx context.Context) DefaultContext {
	cv := ctx.Value(ContextID)
	if dc, ok := cv.(DefaultContext); ok {
		return dc
	}
	return DefaultContext{userID: -1}
}

func (c DefaultContext) UserID() int {
	return c.userID
}

func (c DefaultContext) UserType() int {
	return c.userType
}

func (c DefaultContext) UserName() string {
	return c.userName
}

func (c DefaultContext) Source() string {
	return c.source
}

func (c DefaultContext) HasRole(role string) bool {
	for _, r := range c.roles {
		if r == role {
			return true
		}
	}
	return false
}

func (c DefaultContext) Roles() []string {
	return c.roles
}

func (c DefaultContext) RolesMask() int {
	return c.rolesMask
}
