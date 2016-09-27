package metadata

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/healthcheck/types"
	"strings"
)

type RMetaFetcher struct {
	metadataClient metadata.Client
}

type Poller struct {
	MetaFetcher Fetcher
}

const (
	metadataURL = "http://rancher-metadata/2015-12-19"
)

type Fetcher interface {
	OnChange(intervalSeconds int, do func(string))
	GetServices() ([]metadata.Service, error)
	GetSelfHost() (metadata.Host, error)
}

func (mf RMetaFetcher) OnChange(intervalSeconds int, do func(string)) {
	mf.metadataClient.OnChange(intervalSeconds, do)
}

func (mf RMetaFetcher) GetServices() ([]metadata.Service, error) {
	return mf.metadataClient.GetServices()
}

func (mf RMetaFetcher) GetSelfHost() (metadata.Host, error) {
	return mf.metadataClient.GetSelfHost()
}

func (p *Poller) OnChange(do func(string)) {
	p.MetaFetcher.OnChange(5, do)
}

func (p *Poller) Init() error {
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

	host, err := p.MetaFetcher.GetSelfHost()
	if err != nil {
		return nil, err
	}
	var ses []types.Service
	for _, svc := range svcs {
		if svc.HealthCheck.Port == 0 {
			continue
		}

		var servers []types.Server
		for _, c := range svc.Containers {
			if c.PrimaryIp == "" {
				continue
			}
			if !(strings.EqualFold(c.State, "running") || strings.EqualFold(c.State, "starting")) {
				continue
			}
			//only configure on the host
			// when container health checkers contain host id
			skipCheck := true
			for _, hostUUID := range c.HealthCheckHosts {
				if strings.EqualFold(host.UUID, hostUUID) {
					skipCheck = false
					break
				}
			}
			if skipCheck {
				logrus.Debugf("Health check for container [%s] is not configured on this host, skipping", c.Name)
				continue
			}

			srvr := types.Server{
				UUID:    fmt.Sprintf("cattle-%s_%s_%v", host.UUID, c.UUID, c.StartCount),
				Address: c.PrimaryIp,
			}
			servers = append(servers, srvr)
		}
		s := types.Service{
			UUID:        svc.UUID,
			HealthCheck: svc.HealthCheck,
			Servers:     servers,
		}
		ses = append(ses, s)
	}
	return ses, nil
}
