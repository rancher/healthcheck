package metadata

import (
	"fmt"
	"strings"

	"github.com/leodotcloud/log"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/healthcheck/types"
)

type RMetaFetcher struct {
	metadataClient metadata.Client
}

type Poller struct {
	MetaFetcher Fetcher
}

type Fetcher interface {
	OnChange(intervalSeconds int, do func(string))
	GetServices() ([]metadata.Service, error)
	GetSelfHost() (metadata.Host, error)
	GetContainers() ([]metadata.Container, error)
}

func (mf RMetaFetcher) OnChange(intervalSeconds int, do func(string)) {
	mf.metadataClient.OnChange(intervalSeconds, do)
}

func (mf RMetaFetcher) GetServices() ([]metadata.Service, error) {
	return mf.metadataClient.GetServices()
}

func (mf RMetaFetcher) GetContainers() ([]metadata.Container, error) {
	return mf.metadataClient.GetContainers()
}

func (mf RMetaFetcher) GetSelfHost() (metadata.Host, error) {
	return mf.metadataClient.GetSelfHost()
}

func (p *Poller) OnChange(do func(string)) {
	p.MetaFetcher.OnChange(5, do)
}

func (p *Poller) Init(metadataURL string) error {
	metadataClient, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		return err
	}

	p.MetaFetcher = RMetaFetcher{
		metadataClient: metadataClient,
	}
	return nil
}

func (p *Poller) GetHealthCheckServices() (services []types.Service, err error) {
	svcs, err := p.MetaFetcher.GetServices()
	if err != nil {
		return nil, err
	}
	cs, err := p.MetaFetcher.GetContainers()
	if err != nil {
		return nil, err
	}
	selfHost, err := p.MetaFetcher.GetSelfHost()
	if err != nil {
		return nil, err
	}
	var ses []types.Service
	// process services
	uuidToPrimaryIP := make(map[string]string)
	for _, svc := range svcs {
		for _, c := range svc.Containers {
			if c.PrimaryIp == "" {
				continue
			}
			uuidToPrimaryIP[c.UUID] = c.PrimaryIp
		}
	}
	for _, svc := range svcs {
		if svc.HealthCheck.Port == 0 {
			continue
		}

		for _, c := range svc.Containers {
			var servers []types.Server
			if c.PrimaryIp == "" && c.NetworkFromContainerUUID != "" {
				c.PrimaryIp = uuidToPrimaryIP[c.NetworkFromContainerUUID]
			}
			addServer(&c, &servers, &selfHost)
			if len(servers) == 0 {
				continue
			}
			var hc metadata.HealthCheck
			if c.HealthCheck.Port != 0 {
				hc = c.HealthCheck
			} else {
				hc = svc.HealthCheck
			}
			s := types.Service{
				UUID:        c.UUID,
				HealthCheck: hc,
				Servers:     servers,
			}
			ses = append(ses, s)
		}
	}
	// process standalone containers
	for _, c := range cs {
		if c.ServiceName != "" || c.HealthCheck.Port == 0 {
			continue
		}
		var servers []types.Server
		addServer(&c, &servers, &selfHost)
		if len(servers) == 0 {
			continue
		}
		s := types.Service{
			UUID:        c.UUID,
			HealthCheck: c.HealthCheck,
			Servers:     servers,
		}
		ses = append(ses, s)
	}
	return ses, nil
}

func addServer(c *metadata.Container, servers *[]types.Server, selfHost *metadata.Host) {
	if c.PrimaryIp == "" {
		return
	}
	if !(strings.EqualFold(c.State, "running") || strings.EqualFold(c.State, "starting")) {
		return
	}
	//only configure on the host
	// when container health checkers contain host id
	skipCheck := true
	for _, hostUUID := range c.HealthCheckHosts {
		if strings.EqualFold(selfHost.UUID, hostUUID) {
			skipCheck = false
			break
		}
	}
	if skipCheck {
		log.Debugf("Health check for container [%s] is not configured on this host, skipping", c.Name)
		return
	}

	srvr := types.Server{
		UUID:    fmt.Sprintf("cattle-%s_%s_%v", selfHost.UUID, c.UUID, c.StartCount),
		Address: c.PrimaryIp,
	}
	*servers = append(*servers, srvr)
}
