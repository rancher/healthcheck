package types

import (
	"github.com/rancher/go-rancher-metadata/metadata"
)

type Service struct {
	UUID        string
	HealthCheck metadata.HealthCheck
	Servers     []Server
}

type Server struct {
	UUID    string
	Address string
}

type MetadataPoller interface {
	GetHealthCheckServices() (services []Service, err error)
	OnChange(do func(string))
	Init() error
}

type ConfigUpdater interface {
	ApplyConfig(string)
	Start() error
	Run() error
}
