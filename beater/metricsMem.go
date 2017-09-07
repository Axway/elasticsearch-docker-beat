package beater

import (
	"github.com/docker/docker/api/types"
	"github.com/elastic/beats/libbeat/common"
)

// publish one memory metrics event
func (a *dbeat) publishMemMetrics(stats *types.StatsJSON, data *ContainerData) {
	event := common.MapStr{
		"@timestamp":           common.Time(stats.Read),
		"type":                 "mem",
		"container_id":         data.ID,
		"container_name":       data.name,
		"container_short_name": data.shortName,
		"container_state":      data.state,
		"service_name":         data.serviceName,
		"service_id":           data.serviceID,
		"task_id":              data.taskID,
		"stack_name":           data.stackName,
		"memory": common.MapStr{
			"failcnt":  int64(stats.MemoryStats.Failcnt),
			"limit":    int64(stats.MemoryStats.Limit),
			"maxUsage": int64(stats.MemoryStats.MaxUsage),
			"usage":    int64(stats.MemoryStats.Usage),
			"usage_p":  a.getMemUsage(stats),
		},
	}
	a.client.PublishEvent(event)
}

// compute memory usage
func (a *dbeat) getMemUsage(stats *types.StatsJSON) float64 {
	if stats.MemoryStats.Limit == 0 {
		return 0
	}
	return float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit)
}
