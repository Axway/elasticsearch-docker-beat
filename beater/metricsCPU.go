package beater

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/elastic/beats/libbeat/common"
)

// CPUStats cpu stats
type CPUStats struct {
	Time                 time.Time
	PerCPUUsage          []uint64
	TotalUsage           uint64
	UsageInKernelmode    uint64
	UsageInUsermode      uint64
	PreTime              time.Time
	PrePerCPUUsage       []uint64
	PreTotalUsage        uint64
	PreUsageInKernelmode uint64
	PreUsageInUsermode   uint64
}

// CPUStatsDiff diff between two cpu stats
type CPUStatsDiff struct {
	Duration          uint64
	PerCPUUsage       common.MapStr
	TotalUsage        float64
	UsageInKernelmode float64
	UsageInUsermode   float64
}

// publish one cpu event
func (a *dbeat) publishCPUMetrics(stats *types.StatsJSON, data *ContainerData) {
	cpu := a.newCPUStats(stats)
	diff := a.newCPUDiff(cpu)
	if diff == nil {
		return
	}
	event := common.MapStr{
		"@timestamp":           common.Time(stats.Read),
		"type":                 "cpu",
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
		"cpu": common.MapStr{
			"percpuUsage":       diff.PerCPUUsage,
			"totalUsage":        diff.TotalUsage,
			"usageInKernelmode": diff.UsageInKernelmode,
			"usageInUsermode":   diff.UsageInUsermode,
		},
	}
	for labelName, labelValue := range data.customLabelsMap {
		event[labelName] = labelValue
	}
	a.client.PublishEvent(event)
}

// build a new cpu metrics stats
func (a *dbeat) newCPUStats(stats *types.StatsJSON) *CPUStats {
	var cpu = &CPUStats{
		Time:                 stats.Read,
		PerCPUUsage:          stats.CPUStats.CPUUsage.PercpuUsage,
		TotalUsage:           stats.CPUStats.CPUUsage.TotalUsage,
		UsageInKernelmode:    stats.CPUStats.CPUUsage.UsageInKernelmode,
		UsageInUsermode:      stats.CPUStats.CPUUsage.UsageInUsermode,
		PreTime:              stats.PreRead,
		PrePerCPUUsage:       stats.PreCPUStats.CPUUsage.PercpuUsage,
		PreTotalUsage:        stats.PreCPUStats.CPUUsage.TotalUsage,
		PreUsageInKernelmode: stats.PreCPUStats.CPUUsage.UsageInKernelmode,
		PreUsageInUsermode:   stats.PreCPUStats.CPUUsage.UsageInUsermode,
	}
	return cpu
}

// build a new diff computing difference between two cpu stats
func (a *dbeat) newCPUDiff(cpu *CPUStats) *CPUStatsDiff {
	diff := &CPUStatsDiff{Duration: uint64(cpu.Time.Sub(cpu.PreTime).Seconds())}
	if diff.Duration <= 0 {
		return nil
	}
	ret := common.MapStr{}
	if cap(cpu.PerCPUUsage) == cap(cpu.PrePerCPUUsage) {
		for index := range cpu.PerCPUUsage {
			name := fmt.Sprintf("cpu%d", index)
			ret[name] = a.calculateLoad(cpu.PerCPUUsage[index], cpu.PrePerCPUUsage[index], diff.Duration)
		}
	}
	diff.PerCPUUsage = ret
	diff.TotalUsage = a.calculateLoad(cpu.TotalUsage, cpu.PreTotalUsage, diff.Duration)
	diff.UsageInKernelmode = a.calculateLoad(cpu.UsageInKernelmode, cpu.UsageInKernelmode, diff.Duration)
	diff.UsageInUsermode = a.calculateLoad(cpu.UsageInUsermode, cpu.UsageInUsermode, diff.Duration)
	return diff
}

// compute cpu usage concidering event duration
func (a *dbeat) calculateLoad(oldValue uint64, newValue uint64, duration uint64) float64 {
	value := int64(oldValue - newValue)
	if value < 0 || duration == 0 {
		return float64(0)
	}
	return float64(value) / (float64(duration) * float64(1000000000))
}
