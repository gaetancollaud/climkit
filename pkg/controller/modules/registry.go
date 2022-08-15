package modules

import (
	"github.com/gaetancollaud/climkit/pkg/climkit"
	"github.com/gaetancollaud/climkit/pkg/config"
	"github.com/gaetancollaud/climkit/pkg/mqtt"
	"github.com/gaetancollaud/climkit/pkg/postgres"
)

// Interface for the different modules being
type Module interface {
	Eligible() bool
	Start() error
	Stop() error
}

type ModuleBuilder func(mqtt.Client, postgres.Client, climkit.Client, *config.Config) Module

// Register stores a builder function into the registy for external access.
// Register() can be called from init() on a module in this package and will
// automatically register a module.
func Register(name string, builder ModuleBuilder) {
	Modules[name] = builder
}

var Modules = map[string]ModuleBuilder{}
