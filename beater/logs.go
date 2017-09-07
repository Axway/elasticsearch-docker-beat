package beater

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/elastic/beats/libbeat/common"
)

const containersDateDir = "/containers"

// verify all containers to open logs stream if not already done
func (a *dbeat) updateLogsStream() {
	for ID, data := range a.containers {
		if data.logsStream == nil || data.logsReadError {
			lastTimeID := a.getLastTimeID(ID)
			if lastTimeID == "" {
				fmt.Printf("open logs stream from the begining on container %s\n", data.name)
			} else {
				fmt.Printf("open logs stream from time_id=%s on container %s\n", lastTimeID, data.name)
			}
			stream, err := a.openLogsStream(ID, lastTimeID)
			if err != nil {
				fmt.Printf("Error opening logs stream on container: %s\n", data.name)
			} else {
				data.logsStream = stream
				go a.startReadingLogs(ID, data)
			}
		}
	}
}

// open a logs container stream
func (a *dbeat) openLogsStream(ID string, lastTimeID string) (io.ReadCloser, error) {
	containerLogsOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	}
	if lastTimeID != "" {
		containerLogsOptions.Since = lastTimeID
	}
	return a.dockerClient.ContainerLogs(context.Background(), ID, containerLogsOptions)
}

// get last timestamp if exist
func (a *dbeat) getLastTimeID(ID string) string {
	data, err := ioutil.ReadFile(path.Join(containersDateDir, ID))
	if err != nil {
		return ""
	}
	return string(data)
}

// stream reading loop
func (a *dbeat) startReadingLogs(ID string, data *ContainerData) {
	stream := data.logsStream
	reader := bufio.NewReader(stream)
	data.lastDateSaveTime = time.Now()
	fmt.Printf("start reading logs on container: %s\n", data.name)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("close logs stream on container %s (%v)\n", data.name, err)
			data.logsReadError = true
			stream.Close()
			a.removeContainer(ID)
			return
		}
		if len(line) <= 39 {
			//fmt.Printf("invalid log: [%s]\n", line)
			continue
		}

		date := line[8:38]
		slog := strings.TrimSuffix(line[39:], "\n")
		timestamp, err := time.Parse("2006-01-02T15:04:05.000000000Z", date)
		if err != nil {
			timestamp = time.Now()
		}

		event := common.MapStr{
			"@timestamp":        common.Time(timestamp),
			"type":              "logs",
			"container_id":      ID,
			"container_name":    data.name,
			"container_state":   data.state,
			"service_name":      data.serviceName,
			"service_id":        data.serviceID,
			"task_id":           data.taskID,
			"stack_name":        data.stackName,
			"node_id":           data.nodeID,
			"role":              data.role,
			"axway-target-flow": data.axwayTargetFlow,
			"message":           slog,
		}
		a.client.PublishEvent(event)
		a.periodicDateSave(data, date)
	}
}

func (a *dbeat) periodicDateSave(data *ContainerData, date string) {
	now := time.Now()
	if now.Sub(data.lastDateSaveTime).Seconds() >= float64(a.logsSavedDatePeriod) {
		ioutil.WriteFile(path.Join(containersDateDir, data.ID), []byte(date), 0666)
		data.lastDateSaveTime = now
	}
}

// close all logs stream
func (a *dbeat) closeLogsStreams() {
	for _, data := range a.containers {
		if data.logsStream != nil {
			data.logsStream.Close()
		}
	}
}
