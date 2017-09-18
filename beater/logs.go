package beater

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"regexp"
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
		} else {
			if data.lastLog != "" && time.Now().Sub(data.lastLogTime).Seconds() >= 3 {
				a.publishEvent(data, data.lastLogTimestamp, data.lastLog)
				data.lastLog = ""
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

		now := time.Now()
		data.sdate = line[8:38]
		slog := strings.TrimSuffix(line[39:], "\n")
		timestamp, err := time.Parse("2006-01-02T15:04:05.000000000Z", data.sdate)
		if err != nil {
			timestamp = now
		}

		if !data.mlConfig.Activated {
			a.publishEvent(data, timestamp, slog)
		} else {
			a.groupEvent(data, timestamp, slog)
		}

	}
}

//group event logs concidering the logsMultiline setting
func (a *dbeat) groupEvent(data *ContainerData, timestamp time.Time, slog string) {
	toBeGrouped := false
	if data.mlConfig.Pattern != "" {
		if matched, err := regexp.MatchString(data.mlConfig.Pattern, slog); err == nil {
			if data.mlConfig.Negate {
				toBeGrouped = !matched
			} else {
				toBeGrouped = matched
			}
		}
	}
	if !toBeGrouped {
		//log has not to be grouped, then we send the last group and save the log in lastlog for future possible append
		if data.lastLog != "" {
			a.publishEvent(data, data.lastLogTimestamp, data.lastLog)
		}
		data.lastLog = slog
		// set timestamp of the current group to be able to send it if container stop
		data.lastLogTimestamp = timestamp
		//set time of the last group update to be able to send it if one period of time is exceeded
		data.lastLogTime = time.Now()
	} else {
		//log has to be grouped, if the group size become too big, the group is sent anyway
		if len(data.lastLog)+len(slog) > a.config.LogsMultilineMaxSize {
			a.publishEvent(data, data.lastLogTimestamp, data.lastLog)
			data.lastLog = ""
		}
		//the log is append to the group if the group if not empty
		if data.lastLog == "" {
			data.lastLog = slog
			// set timestamp of the current group to be able to send it if container stop
			data.lastLogTimestamp = timestamp
			//set time of the last group update to be able to send it if one period of time is exceeded
			data.lastLogTime = time.Now()
		} else {
			if data.mlConfig.Append {
				data.lastLog = data.lastLog + "\n" + slog
			} else {
				data.lastLog = slog + "\n" + data.lastLog
			}
		}
	}
}

//send one log event
func (a *dbeat) publishEvent(data *ContainerData, timestamp time.Time, slog string) {
	event := common.MapStr{
		"@timestamp":        common.Time(timestamp),
		"type":              "logs",
		"container_id":      data.ID,
		"container_name":    data.name,
		"container_state":   data.state,
		"service_name":      data.serviceName,
		"service_id":        data.serviceID,
		"task_id":           data.taskID,
		"stack_name":        data.stackName,
		"node_id":           data.nodeID,
		"host_ip":           data.hostIP,
		"hostname":          data.hostname,
		"axway-target-flow": data.axwayTargetFlow,
		"message":           slog,
	}
	for labelName, labelValue := range data.customLabelsMap {
		event[labelName] = labelValue
	}
	a.nbLogs++
	a.client.PublishEvent(event)
	a.periodicDateSave(data)
}

// periodically save the current log date for the container
func (a *dbeat) periodicDateSave(data *ContainerData) {
	now := time.Now()
	if now.Sub(data.lastDateSaveTime).Seconds() >= float64(a.logsSavedDatePeriod) {
		ioutil.WriteFile(path.Join(containersDateDir, data.ID), []byte(data.sdate), 0666)
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
