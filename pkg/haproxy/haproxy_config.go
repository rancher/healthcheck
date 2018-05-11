package haproxy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"

	"github.com/leodotcloud/log"
	"github.com/rancher/healthcheck/types"
)

type Provider struct {
	builder *haproxyConfigBuilder
	Poller  types.MetadataPoller
}

type Backend struct {
	UUID            string
	Protocol        string
	ResponseTimeout int
	RequestLine     string
	Server          *Server
}

type Server struct {
	UUID    string
	Address string
	Port    int
	Config  string
}

type haproxyConfigBuilder struct {
	Name      string
	ReloadCmd string
	StartCmd  string
	Config    string
	Template  string
}

type Backends []*Backend

func (p *Provider) Start() error {
	log.Info("Starting haproxy listener")

	p.builder = &haproxyConfigBuilder{
		ReloadCmd: "haproxy_reload.sh /etc/haproxy/haproxy.cfg reload",
		StartCmd:  "haproxy_reload.sh /etc/haproxy/haproxy.cfg start",
		Config:    "/etc/haproxy/haproxy_new.cfg",
		Template:  "/etc/haproxy/haproxy_template.cfg",
		Name:      "healthCheck",
	}
	// start haproxy
	return p.applyHaproxyConfig(nil, false)
}

func (p *Provider) ApplyConfig(string) {
	log.Info("Scheduling apply config")
	bes, err := p.buildConfig()
	if err != nil {
		log.Errorf("Failed to build haproxy config: %v", err)
	}
	err = p.applyHaproxyConfig(bes, true)
	if err != nil {
		log.Errorf("Error applying config: %v", err)
	}
}

func (p *Provider) Run() {
	p.Poller.OnChange(p.ApplyConfig)
}

func (p *Provider) buildConfig() (backends Backends, err error) {
	svcs, err := p.Poller.GetHealthCheckServices()
	if err != nil {
		return nil, err
	}

	var bes Backends
	for _, svc := range svcs {
		hc := svc.HealthCheck
		port := hc.Port
		config := ""
		if hc.Interval != 0 {
			config = fmt.Sprintf("inter %v", hc.Interval)
		}
		if hc.HealthyThreshold != 0 {
			config = fmt.Sprintf("%s rise %v", config, hc.HealthyThreshold)
		}
		if hc.UnhealthyThreshold != 0 {
			config = fmt.Sprintf("%s fall %v", config, hc.UnhealthyThreshold)
		}
		proto := "tcp"
		if hc.RequestLine != "" {
			proto = "http"
		}
		for _, c := range svc.Servers {
			srvr := &Server{
				UUID:    c.UUID,
				Address: c.Address,
				Port:    port,
				Config:  config,
			}
			be := &Backend{
				UUID:            c.UUID,
				Protocol:        proto,
				ResponseTimeout: hc.ResponseTimeout,
				RequestLine:     hc.RequestLine,
				Server:          srvr,
			}
			bes = append(bes, be)
		}

	}
	sort.Sort(bes)

	return bes, nil
}

func (b *haproxyConfigBuilder) write(backends Backends) (err error) {
	var w io.Writer
	w, err = os.Create(b.Config)
	if err != nil {
		return err
	}
	var t *template.Template
	t, err = template.ParseFiles(b.Template)
	if err != nil {
		return err
	}
	conf := make(map[string]interface{})
	conf["backends"] = backends
	err = t.Execute(w, conf)
	return err
}

func (p *Provider) applyHaproxyConfig(backends Backends, reload bool) error {
	if reload {
		// apply config
		if err := p.builder.write(backends); err != nil {
			return err
		}
		return p.builder.exec(p.builder.ReloadCmd)
	}
	return p.builder.exec(p.builder.StartCmd)
}

func (b *haproxyConfigBuilder) exec(cmd string) error {
	output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	msg := fmt.Sprintf("%v -- %v", b.Name, string(output))
	log.Info(msg)
	if err != nil {
		return fmt.Errorf("error reloading %v: %v", msg, err)
	}
	return nil
}

func (s Backends) Len() int {
	return len(s)
}
func (s Backends) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Backends) Less(i, j int) bool {
	return strings.Compare(s[i].UUID, s[j].UUID) > 0
}
