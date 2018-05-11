package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/leodotcloud/log"
	"github.com/patrickmn/go-cache"
	"github.com/rancher/go-rancher/client"
	"github.com/rancher/healthcheck/metadata"
	"github.com/rancher/healthcheck/pkg/haproxy"
	"github.com/rancher/healthcheck/util"
)

var Prefix = "cattle-"
var ServerName = "svname"
var Status = "status"

func Poll(metadataURL string) error {
	client, err := util.GetRancherClient()
	if err != nil {
		return err
	}
	if client == nil {
		return fmt.Errorf("Can not create RancherClient, No credentials found")
	}

	metadataPoller := &metadata.Poller{}
	if err = metadataPoller.Init(metadataURL); err != nil {
		return err
	}

	configUpdater := &haproxy.Provider{
		Poller: metadataPoller,
	}
	if err = configUpdater.Start(); err != nil {
		return err
	}

	go configUpdater.Run()

	c := cache.New(1*time.Hour, 1*time.Minute)
	m := &Monitor{
		client:         client,
		reportedStatus: c,
	}

	for stat := range m.getStats() {
		m.processStat(stat)
	}

	return nil
}

type Monitor struct {
	client         *client.RancherClient
	reportedStatus *cache.Cache
}

func (m *Monitor) getStats() <-chan haproxy.Stat {
	c := make(chan haproxy.Stat)
	go m.readStats(c)
	return c
}

func (m *Monitor) readStats(c chan<- haproxy.Stat) {
	defer close(c)

	count := 0

	h := &haproxy.Monitor{
		SocketPath: haproxy.HaproxySock,
	}

	for {
		// Sleep up front.  This way if this program gets restarted really fast we don't spam cattle
		time.Sleep(2 * time.Second)

		stats, err := h.Stats()
		currentCount := 0
		if err != nil {
			log.Errorf("Failed to read stats: %v", err)
			continue
		}

		for _, stat := range stats {
			if strings.HasPrefix(stat[ServerName], Prefix) {
				currentCount++
				c <- stat
			}
		}

		if currentCount != count {
			count = currentCount
			log.Infof("Monitoring %d backends", count)
		}

	}
}

func (m *Monitor) processStat(stat haproxy.Stat) {
	serverName := strings.TrimPrefix(stat[ServerName], Prefix)
	currentStatus := stat[Status]

	previousStatus, _ := m.reportedStatus.Get(serverName)
	if strings.HasPrefix(currentStatus, "UP ") {
		// do nothing on partial UP
		return
	}

	if currentStatus == "UP" && previousStatus != "UP" && previousStatus != "INIT" {
		currentStatus = "INIT"
	}

	update := true
	if previousStatus != currentStatus {
		err := m.reportStatus(serverName, currentStatus)
		if err != nil {
			log.Errorf("Failed to report status %s=%s: %v", serverName, currentStatus, err)
			update = false
		}
	}

	if update {
		m.reportedStatus.Set(serverName, currentStatus, cache.DefaultExpiration)
	}
}

func (m *Monitor) reportStatus(serverName, currentStatus string) error {
	_, err := m.client.ServiceEvent.Create(&client.ServiceEvent{
		HealthcheckUuid:   serverName,
		ReportedHealth:    currentStatus,
		ExternalTimestamp: time.Now().Unix(),
	})

	if err != nil {
		return err
	}

	log.Infof("%s=%s", serverName, currentStatus)
	return nil
}
