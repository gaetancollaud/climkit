package modules

import (
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/mqtt"
)

// Interface for the different modules being
type Module interface {
	Start() error
	Stop() error
}

type ModuleBuilder func(mqtt.Client, climkit.Climkit, *config.Config) Module

// Register stores a builder function into the registy for external access.
// Register() can be called from init() on a module in this package and will
// automatically register a module.
func Register(name string, builder ModuleBuilder) {
	Modules[name] = builder
}

var Modules = map[string]ModuleBuilder{}
