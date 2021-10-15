package vivard

import "context"

type Context interface {
	UserID() int
	UserName() string
	Source() string
	HasRole(role string) bool
}

type DefaultContext struct {
	userID   int
	userName string
	source   string
	roles    []string
}

var ContextID = struct{}{}

func NewContext(ctx context.Context, userID int, userName string, source string, roles []string) context.Context {
	return context.WithValue(ctx, ContextID, DefaultContext{userID: userID, userName: userName, source: source, roles: roles})
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
