package beater

import (
	"time"

	"github.com/docker/docker/api/types"
	"github.com/elastic/beats/libbeat/common"
)

// IOStats IO stats
type IOStats struct {
	Time   time.Time
	Reads  uint64
	Writes uint64
	Totals uint64
}

// IOStatsDiff diff between two IOStats
type IOStatsDiff struct {
	Duration int64
	Reads    int64
	Writes   int64
	Totals   int64
}

// publish one IO metrics event
func (a *dbeat) publishIOMetrics(stats *types.StatsJSON, data *ContainerData) {
	io := a.newIOStats(stats)
	if data.previousIOStats == nil {
		data.previousIOStats = io
		return
	}
	diff := a.newIODiff(io, data.previousIOStats)
	if diff == nil {
		return
	}
	data.previousIOStats = io
	event := common.MapStr{
		"@timestamp":           common.Time(stats.Read),
		"type":                 "io",
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
		"blkio": common.MapStr{
			"read":  diff.Reads,
			"write": diff.Writes,
			"total": diff.Totals,
		},
	}
	for labelName, labelValue := range data.customLabelsMap {
		event[labelName] = labelValue
	}
	a.client.PublishEvent(event)
}

// create new io stats
func (a *dbeat) newIOStats(stats *types.StatsJSON) *IOStats {
	var io = &IOStats{Time: stats.Read}
	for _, s := range stats.BlkioStats.IoServicedRecursive {
		if s.Op == "Read" {
			io.Reads += s.Value
		} else if s.Op == "Write" {
			io.Writes += s.Value
		} else if s.Op == "Total" {
			io.Totals += s.Value
		}
	}
	return io
}

// create a new io diff computing difference between two io stats
func (a *dbeat) newIODiff(newIO *IOStats, previousIO *IOStats) *IOStatsDiff {
	diff := &IOStatsDiff{Duration: int64(newIO.Time.Sub(previousIO.Time).Seconds())}
	if diff.Duration <= 0 {
		return nil
	}
	diff.Reads = int64(newIO.Reads - previousIO.Reads)
	diff.Writes = int64(newIO.Writes - previousIO.Writes)
	diff.Totals = int64(newIO.Totals - previousIO.Totals)
	return diff
}
