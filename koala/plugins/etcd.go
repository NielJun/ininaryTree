package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/who7708/etcd/clientv3"
	"github.com/daniel/ininaryTree/koala/registry"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

// const values
const (
	MAX_SERVICE_NUM = 8
)

// etcd的插件的数据结构
type EtcdRegistry struct {
	options            *registry.Options           // 选项模式初始化的数据
	client             *clientv3.Client            // etcd v3版本的客户端
	ServiceCh          chan *registry.Service      // 服务对象
	value              atomic.Value                // 原子操作的对象 用来存储从etcd拉取的所有数据 因为读取的时候需要枷锁进行保护
	lock               sync.Mutex                  // 已经需要从etcd上面拉取数据 这个锁是为了给etcd链接进行保护
	RegistryServiceMap map[string]*RegisterService // 存储服务和服务名字的map
}

// 注册服务时用到的诗句结构
type RegisterService struct {
	id             clientv3.LeaseID                        //续约用到的ID标实
	service        *registry.Service                       //服务对象
	beenRegistered bool                                    //当前的服务是否被注册过 主要是在注册完成后down掉时候作为一个标实
	keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse // 用来存储 续约所用到的管道
}

// 存储服务信息的数据结构 作为从etcd连拉取的缓存对象
type AllServiceInfo struct {
	serviceMap map[string]*registry.Service
}

// the unique value in global
var (
	etcdRegister *EtcdRegistry = &EtcdRegistry{
		ServiceCh:          make(chan *registry.Service, MAX_SERVICE_NUM),
		RegistryServiceMap: make(map[string]*RegisterService, MAX_SERVICE_NUM),
	}
)

func init() {

	// 在初始化的时候 定义出来服务信息的数据结构对象 并且存储在Etcd的注册器里面
	allServiceInfo := &AllServiceInfo{serviceMap: make(map[string]*registry.Service, MAX_SERVICE_NUM)}
	//存储在源自操作对象里面
	etcdRegister.value.Store(allServiceInfo)

	registry.RegisterPlugin(etcdRegister)
	// 启动一个携程在后台不断的处理注册请求
	go etcdRegister.run()
}

func (e *EtcdRegistry) Name() string {
	return "etcd"
}

// 初始化函数
func (e *EtcdRegistry) Init(ctx context.Context, opts ...registry.Option) (err error) {

	e.options = &registry.Options{}
	for _, opt := range opts {
		opt(e.options)
	}
	// etcd client v3 is using grpg license instead of http+json
	e.client, err = clientv3.New(clientv3.Config{
		Endpoints:   e.options.Addrs,
		DialTimeout: e.options.Timeout,
	})
	if err != nil {
		fmt.Println("Init etcd failed")
	}
	return
}

// 服务注册
// 当发起服务注册的请求的时候，我们会把它放在一个服务的注册管道内
// 然后通过诗歌后台的gorouting (run())来完成注册的操作
func (e *EtcdRegistry) Register(ctx *context.Context, service *registry.Service) (err error) {

	select {
	case e.ServiceCh <- service:
	default:
		err = fmt.Errorf("register chan is full")
		return
	}
	return
}

func (e *EtcdRegistry) UnRegister(ctx *context.Context, service *registry.Service) (err error) {
	panic("implement me")
}

// 通过名字获取服务
// 1.直接从缓存中间拿到
// 2.如果缓存中间没有，则从etcd中间拉取
// 3.返回
func (e *EtcdRegistry) GetService(ctx *context.Context, serviceName string) (service *registry.Service, err error) {

	// 从全局存储的位于etcdRegister里面的源自操作对象里面取出map
	allServiceInfo := etcdRegister.value.Load().(*AllServiceInfo)

	// 判断当前缓存是否包含该服务信息
	service, ok := allServiceInfo.serviceMap[serviceName]
	if !ok {
		return
	}
	//缓存中没有则从etcd上面拉取 （仅仅允许一个对象进行etcd的数据拉取）
	e.lock.Lock()

	// 拉取之前加锁以后在做一次校验
	service, ok = allServiceInfo.serviceMap[serviceName]
	if !ok {
		return
	}

	// 如果还是咩有 则去拉取


	defer e.lock.Unlock()

	return
}

// 运行在后台的注册协程
func (e *EtcdRegistry) run() {
	for {
		select {
		case service := <-e.ServiceCh:
			// if service have been store in the map
			_, ok := e.RegistryServiceMap[service.Name]
			if ok {
				break
			}

			// map have `not stored the service
			registryService := &RegisterService{
				service: service,
			}
			e.RegistryServiceMap[service.Name] = registryService
		default:
			// keep alive
			e.registerOrKeepAlive()
			time.Sleep(500 * time.Millisecond)
		}
	}

}

// select the service status if keep alive or register for the services from the map
func (e *EtcdRegistry) registerOrKeepAlive() {

	for _, registryService := range e.RegistryServiceMap {

		if registryService.beenRegistered {
			e.keepAlive(registryService)
			continue
		}

		e.registerService(registryService)
	}

}

// keep alive for Specific service
func (e *EtcdRegistry) keepAlive(registService *RegisterService) {
	select {
	// 判断需要进行续约的管道中有没有对象
	case resp := <-registService.keepAliveChan:
		if resp == nil {
			registService.beenRegistered = false
			return
		}

		fmt.Sprintf("service: %s,node : %s,ttl:%v", registService.service.Name, registService.service.Nodes[0].IP, registService.service.Nodes[0].Port)
	}

	return
}

// register service  to the map
func (e *EtcdRegistry) registerService(registService *RegisterService) (err error) {

	// 从客户端获得租约的信息
	resp, err := e.client.Grant(context.TODO(), e.options.HeartBeat)
	if err != nil {
		return
	}
	//设置租约ID
	registService.id = resp.ID

	for _, node := range registService.service.Nodes {

		service := registry.Service{
			Name:  registry.Service{}.Name,
			Nodes: []*registry.Node{node,},
		}

		data, err := json.Marshal(service)
		if err != nil {
			continue
		}

		key := e.serviceNodePath(&service)

		// 把key放入etcd中间
		_, err = e.client.Put(context.TODO(), key, string(data), clientv3.WithLease(resp.ID))
		if err != nil {
			continue
		}

		// 进行续约
		channel, err := e.client.KeepAlive(context.TODO(), resp.ID)
		if err != nil {
			continue
		}

		// 如果服务注册完毕 则是指该服务的状态
		registService.keepAliveChan = channel
		registService.beenRegistered = true

	}

	return
}

// 通过节点的信息 构造一个节点的路径
func (e *EtcdRegistry) serviceNodePath(service *registry.Service) string {
	nodeIp := fmt.Sprintf("%s:%d", service.Nodes[0].IP, service.Nodes[0].Port)
	return path.Join(e.options.RegisterPath, service.Name, nodeIp)
}


// 通过节点的名字 构造一个节点的前缀路径
func (e *EtcdRegistry) serviceNodePrePath(serviceName string) string {
	return path.Join(e.options.RegisterPath, serviceName)
}