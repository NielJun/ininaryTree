package registry

import (
	"context"
	"fmt"
	"sync"
)

var (
	pluginMgr = &PluginMgr{
		pluginsMap: make(map[string]Registry),
		lock:       sync.Mutex{},
	}
)

type PluginMgr struct {
	// the manager map of all the plugins.
	pluginsMap map[string]Registry
	// single process use lock
	lock sync.Mutex
}

// Register the plugin
func (p *PluginMgr) registerPlugin(plugin Registry) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// if the map contain current plugin
	_, ok := p.pluginsMap[plugin.Name()]
	if ok {
		err = fmt.Errorf("the %s plugin have been registied")
		return
	}

	p.pluginsMap[plugin.Name()] = plugin
	return

}

// UnRegister the plugin
func (p *PluginMgr) unRegisterPlugin(plugin Registry) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// if the map contain current plugin
	_, ok := p.pluginsMap[plugin.Name()]
	if !ok {
		err = fmt.Errorf("the %s plugin have not been registied")
		return
	}

	delete(p.pluginsMap, plugin.Name())
	return

}

// init the plugins
func (p *PluginMgr) initRegistry(ctx context.Context, name string, opts ...Option) (registry Registry, err error) {

	p.lock.Lock()
	defer p.lock.Unlock()

	// if the plugin have been registered in map
	plugin, ok := p.pluginsMap[name]

	if !ok {
		err = fmt.Errorf("the %s plugin have not been registied")
		return
	}

	registry = plugin
	registry.Init(ctx, opts...)
	return
}

// func to been exposed to others
func RegisterPlugin(plugin Registry) (err error) {
	return pluginMgr.registerPlugin(plugin)
}

// func to been exposed to others
func InitRegistry(ctx context.Context, name string, opts ...Option) (registry Registry, err error) {
	return pluginMgr.initRegistry(ctx, name, opts...)
}
