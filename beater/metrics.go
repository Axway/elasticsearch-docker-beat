package beater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
)

// verify all containers to open metrics stream if not already done
func (a *dbeat) updateMetricsStream() {
	for ID, data := range a.containers {
		if data.metricsStream == nil || data.metricsReadError {
			streamb, err := a.dockerClient.ContainerStats(context.Background(), ID, true)
			if err != nil {
				fmt.Printf("Error opening metrics stream on container: %s\n", data.name)
			} else {
				fmt.Printf("open metrics stream on container: %s\n", data.name)
				data.metricsStream = streamb.Body
				go a.startReadingMetrics(ID, data)
			}
		}
	}
}

// open a metrics container stream
func (a *dbeat) startReadingMetrics(ID string, data *ContainerData) {
	stream := data.metricsStream
	fmt.Printf("start reading metrics on container: %s\n", data.name)
	decoder := json.NewDecoder(stream)
	stats := new(types.StatsJSON)
	for err := decoder.Decode(stats); err != io.EOF; err = decoder.Decode(stats) {
		if err != nil {
			fmt.Printf("close metrics stream on container %s (%v)\n", data.name, err)
			data.metricsReadError = true
			stream.Close()
			a.removeContainer(ID)
			return
		}
		if a.config.Memory {
			a.publishMemMetrics(stats, data)
		}
		if a.config.IO {
			a.publishIOMetrics(stats, data)
		}
		if a.config.Net {
			a.publishNetMetrics(stats, data)
		}
		if a.config.CPU {
			a.publishCPUMetrics(stats, data)
		}
		if a.config.Memory || a.config.IO || a.config.Net || a.config.CPU {
			a.nbMetrics++
		}
	}
}

// close all metrics streams
func (a *dbeat) closeMetricsStreams() {
	for _, data := range a.containers {
		if data.metricsStream != nil {
			data.metricsStream.Close()
		}
	}
}
