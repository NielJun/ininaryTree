package registry

import "context"

// interface to override the registry func
type Registry interface {
	// 获得服务的名字
	Name() string

	// 初始化
	Init(ctx context.Context, opts ...Option) (err error)

	// 服务注册
	Register(ctx *context.Context, service *Service) (err error)

	// 服务反注册
	UnRegister(ctx *context.Context, service *Service) (err error)

	// 服务发现 通过服务的名字获取服务的信息（ip和port 列表）
	GetService(ctx *context.Context, serviceName string) (service *Service, err error)
}
