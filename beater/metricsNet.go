package beater

import (
	"time"

	"github.com/docker/docker/api/types"
	"github.com/elastic/beats/libbeat/common"
)

// NetStats net stats
type NetStats struct {
	Time      time.Time
	RxBytes   uint64
	RxDropped uint64
	RxErrors  uint64
	RxPackets uint64
	TxBytes   uint64
	TxDropped uint64
	TxErrors  uint64
	TxPackets uint64
}

// NetStatsDiff diff between two IOStats
type NetStatsDiff struct {
	Duration  int64
	RxBytes   int64
	RxDropped int64
	RxErrors  int64
	RxPackets int64
	TxBytes   int64
	TxDropped int64
	TxErrors  int64
	TxPackets int64
}

// publish one net metrics event
func (a *dbeat) publishNetMetrics(stats *types.StatsJSON, data *ContainerData) {
	net := a.newNetStats(stats)
	if data.previousNetStats == nil {
		data.previousNetStats = net
		return
	}
	diff := a.newNetDiff(net, data.previousNetStats)
	if diff == nil {
		return
	}
	data.previousNetStats = net
	event := common.MapStr{
		"@timestamp":           common.Time(stats.Read),
		"type":                 "net",
		"container_id":         data.ID,
		"container_name":       data.name,
		"container_short_name": data.shortName,
		"container_state":      data.state,
		"service_name":         data.serviceName,
		"service_id":           data.serviceID,
		"stack_name":           data.stackName,
		"host_ip":              data.hostIP,
		"hostname":             data.hostname,
		"beat.name":            dbeatName,
		"net": common.MapStr{
			"totalBytes": diff.RxBytes + diff.TxBytes,
			"rxBytes":    diff.RxBytes,
			"rxDropped":  diff.RxDropped,
			"rxErrors":   diff.RxErrors,
			"rxPackets":  diff.RxPackets,
			"txBytes":    diff.TxBytes,
			"txDropped":  diff.TxDropped,
			"txErrors":   diff.TxErrors,
			"txPackets":  diff.TxPackets,
		},
	}
	for labelName, labelValue := range data.customLabelsMap {
		event[labelName] = labelValue
	}
	a.client.PublishEvent(event)
}

// create a new net stats
func (a *dbeat) newNetStats(stats *types.StatsJSON) *NetStats {
	var net = &NetStats{Time: stats.Read}
	for _, netStats := range stats.Networks {
		net.RxBytes += netStats.RxBytes
		net.RxDropped += netStats.RxDropped
		net.RxErrors += netStats.RxErrors
		net.RxPackets += netStats.RxPackets
		net.TxBytes += netStats.TxBytes
		net.TxDropped += netStats.TxDropped
		net.TxErrors += netStats.TxErrors
		net.TxPackets += netStats.TxPackets
	}
	return net
}

// create a new net diff computing difference between two net stats
func (a *dbeat) newNetDiff(newNet *NetStats, previousNet *NetStats) *NetStatsDiff {
	diff := &NetStatsDiff{Duration: int64(newNet.Time.Sub(previousNet.Time).Seconds())}
	if diff.Duration <= 0 {
		return nil
	}
	diff.RxBytes = int64(newNet.RxBytes - previousNet.RxBytes)
	diff.RxDropped = int64(newNet.RxDropped - previousNet.RxDropped)
	diff.RxErrors = int64(newNet.RxErrors - previousNet.RxErrors)
	diff.RxPackets = int64(newNet.RxPackets - previousNet.RxPackets)
	diff.TxBytes = int64(newNet.TxBytes - previousNet.TxBytes)
	diff.TxDropped = int64(newNet.TxDropped - previousNet.TxDropped)
	diff.TxErrors = int64(newNet.TxErrors - previousNet.TxErrors)
	diff.TxPackets = int64(newNet.TxPackets - previousNet.TxPackets)
	return diff
}
