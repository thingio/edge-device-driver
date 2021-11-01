package di

import "sync"

// NewContainer is a factory method that returns an initialized Container receiver struct.
func NewContainer(serviceConstructors ServiceConstructorMap) *Container {
	c := &Container{
		serviceMap: map[string]service{},
		mutex:      sync.RWMutex{},
	}

	if serviceConstructors != nil {
		c.Update(serviceConstructors)
	}

	return c
}

type Get func(serviceName string) interface{}

// ServiceConstructor defines the contract for a function/closure to create a service.
type ServiceConstructor func(get Get) interface{}

// ServiceConstructorMap maps a service name to a function/closure to create that service.
type ServiceConstructorMap map[string]ServiceConstructor

// service is an internal structure used to track a specific service's constructor and constructed instance.
type service struct {
	constructor ServiceConstructor
	instance    interface{}
}

// Container is a receiver that maintains a list of services, their constructors, and their constructed instances
// in a thread-safe manner.
type Container struct {
	serviceMap map[string]service
	mutex      sync.RWMutex
}

// Update updates the container's internal services included in serviceConstructors.
func (c *Container) Update(serviceConstructors ServiceConstructorMap) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for serviceName, constructor := range serviceConstructors {
		c.serviceMap[serviceName] = service{
			constructor: constructor,
			instance:    nil,
		}
	}
}

// get looks up the requested serviceName and, if it exists, returns a constructed instance.
// If the requested service does not exist, it returns nil.
// Get wraps instance construction in a singleton; the implementation assumes an instance,
// once constructed, will be reused and returned for all subsequent get(serviceName) calls.
func (c *Container) get(serviceName string) interface{} {
	service, ok := c.serviceMap[serviceName]
	if !ok {
		return nil
	}

	if service.instance == nil {
		service.instance = service.constructor(c.get)
		c.serviceMap[serviceName] = service
	}

	return service.instance
}

// Get wraps get to make it thread-safe.
func (c *Container) Get(serviceName string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.get(serviceName)
}
