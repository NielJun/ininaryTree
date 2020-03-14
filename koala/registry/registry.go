package registry

import "context"

// interface to override the registry func
type Registry interface {

	// the name of plugin
	Name() string

	// init interface func
	Init(ctx context.Context,opts ...Option)(err error)

	// register interface func
	Register(ctx *context.Context,service * Service)(err error)


	// unregister interface func
	UnRegister(ctx * context.Context,service * Service)(err error)
}
