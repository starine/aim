package naming

import (
	"errors"

	"github.com/starine/aim"
)

// errors
var (
	ErrNotFound = errors.New("service no found")
)

// Naming defined methods of the naming service
type Naming interface {
	Find(serviceName string, tags ...string) ([]aim.ServiceRegistration, error)
	Subscribe(serviceName string, callback func(services []aim.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(service aim.ServiceRegistration) error
	Deregister(serviceID string) error
}
