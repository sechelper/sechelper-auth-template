// Package application contains the only business-module composition boundary.
// Framework packages must never import this package's business implementations.
package application

import "fmt"

// BusinessModule is the stable extension protocol exposed to second-party code.
// Implementations must depend only on framework interfaces supplied by the host.
type BusinessModule interface {
	Name() string
}

type Registry struct{ modules []BusinessModule }

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Add(module BusinessModule) error {
	if module == nil || module.Name() == "" {
		return fmt.Errorf("business module name is required")
	}
	for _, existing := range r.modules {
		if existing.Name() == module.Name() {
			return fmt.Errorf("duplicate business module: %s", module.Name())
		}
	}
	r.modules = append(r.modules, module)
	return nil
}

func (r *Registry) Modules() []BusinessModule { return append([]BusinessModule(nil), r.modules...) }
