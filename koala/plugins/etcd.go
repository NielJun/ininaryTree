package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/who7708/etcd/clientv3"
	"github.com/daniel/ininaryTree/koala/registry"
	"path"
	"time"
)

// the unique value in global
var (
	etcdRegister *EtcdRegistry = &EtcdRegistry{
		ServiceCh:          make(chan *registry.Service, Max_Service_Value),
		RegistryServiceMap: make(map[string]*RegisterService, Max_Service_Value),
	}
)

// const values
const (
	Max_Service_Value = 8
)

// the plugin struct of the  etcd
type EtcdRegistry struct {
	options   *registry.Options
	client    *clientv3.Client
	ServiceCh chan *registry.Service

	// store the services in the map
	RegistryServiceMap map[string]*RegisterService
}

//
type RegisterService struct {
	id      clientv3.LeaseID
	service *registry.Service
	// if the service had been resisted
	beenRegistied bool
}

func init() {
	registry.RegisterPlugin(etcdRegister)
	go etcdRegister.run()
}

func (e *EtcdRegistry) Name() string {
	return "etcd"
}

// the init func
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

// this is the func witch run in background . it works to register in register center
func (e *EtcdRegistry) run() {

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

// keep alive for the services from the map
func (e *EtcdRegistry) registerOrKeepAlive() {

	for _, registryService := range e.RegistryServiceMap {
		if registryService.beenRegistied {
			e.keepAlive(registryService)
			continue
		}

		e.registerService(registryService)
	}

}

// keep alive for Specific service
func (e *EtcdRegistry) keepAlive(registService *RegisterService) {
	return
}

// register service  to the map
func (e *EtcdRegistry) registerService(registService *RegisterService) (err error) {
	// get the grant from client
	resp, err := e.client.Grant(context.TODO(), e.options.HeartBeat)
	if err != nil {
		return
	}
	registService.id = resp.ID

	for _, node := range registService.service {

		temp := registry.Service{
			Name:  registry.Service{}.Name,
			Nodes: []*registry.Node{node,},
		}

		data, err := json.Marshal(temp)
		if err != nil {
			continue
		}

		key := e.serviceNodePath(&temp)

		_, err = e.client.Put(context.TODO(), key, string(data), clientv3.WithLease(resp.ID))
		if err != nil {
			continue
		}

	}

	return
}

func (e *EtcdRegistry) serviceNodePath(service *registry.Service) string {
	return path.Join(e.options.RegisterPath, service.Name)
}
