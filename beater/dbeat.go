package beater

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Axway/elasticsearch-docker-beat/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
)

const (
	sfalse = "false"
	strue  = "true"
)

// dbeat the amp beat struct
type dbeat struct {
	done                chan struct{}
	config              config.Config
	client              publisher.Client
	dockerClient        *client.Client
	eventStreamReading  bool
	containers          map[string]*ContainerData
	hostname            string
	hostIP              string
	lastUpdate          time.Time
	logsSavedDatePeriod int
	nbLogs              int
	nbMetrics           int
	MLDefault           *config.MLConfig
	MLStackMap          map[string]*config.MLConfig
	MLServiceMap        map[string]*config.MLConfig
	MLContainerMap      map[string]*config.MLConfig
	JSONFiltersMap      map[string]*config.JSONFilter
	PlainFiltersMap     []string
	beaterStarted       bool
}

// New Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	fmt.Println("dbeat version 0.0.3 b12")
	bt := &dbeat{
		done:            make(chan struct{}),
		MLStackMap:      make(map[string]*config.MLConfig),
		MLServiceMap:    make(map[string]*config.MLConfig),
		MLContainerMap:  make(map[string]*config.MLConfig),
		JSONFiltersMap:  make(map[string]*config.JSONFilter),
		PlainFiltersMap: make([]string, 0),
	}
	dconfig := config.DefaultConfig
	if err := cfg.Unpack(&dconfig); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}
	bt.config = dconfig
	return bt, nil
}

// set log multiline configuration
func (bt *dbeat) setMLConfig() {
	for mlName, mlMap := range bt.config.LogsMultiline {
		ml := &config.MLConfig{Activated: true, Negate: false, Append: true}
		applyOn := ""
		if mlMap != nil {
			for name, value := range mlMap {
				if strings.ToLower(name) == "activated" {
					if strings.ToLower(value) == sfalse {
						ml.Activated = false
					}
				}
				if strings.ToLower(name) == "applyon" {
					applyOn = strings.ToLower(value)
				}
				if strings.ToLower(name) == "pattern" {
					ml.Pattern = value
				}
				if strings.ToLower(name) == "negate" {
					if strings.ToLower(value) == strue {
						ml.Negate = true
					}
				}
				if strings.ToLower(name) == "append" {
					if strings.ToLower(value) == sfalse {
						ml.Append = false
					}
				}
				if strings.ToLower(mlName) == "default" {
					bt.MLDefault = ml
				} else if applyOn == "container" {
					bt.MLContainerMap["/"+mlName] = ml
				} else if applyOn == "service" {
					bt.MLServiceMap["/"+mlName] = ml
				} else if applyOn == "stack" {
					bt.MLStackMap["/"+mlName] = ml
				}
			}
		}
		fmt.Printf("ML apply on %s name=%s: %+v\n", applyOn, mlName, ml)
	}
}

// set log json filter configuration
func (bt *dbeat) setJSONFilterConfig() {
	fmt.Printf("json filter config: %+v\n", bt.config.LogsJSONFilters)
	for attributeName, filterMap := range bt.config.LogsJSONFilters {
		filter := &config.JSONFilter{Activated: true, Negate: false}
		filter.Name = "\"" + attributeName + "\":"
		if filterMap != nil {
			for name, value := range filterMap {
				if strings.ToLower(name) == "activated" {
					if strings.ToLower(value) == sfalse {
						filter.Activated = false
					}
				}
				if strings.ToLower(name) == "pattern" {
					filter.Pattern = value
				}
				if strings.ToLower(name) == "negate" {
					if strings.ToLower(value) == strue {
						filter.Negate = true
					}
				}
			}
		}
		bt.JSONFiltersMap[attributeName] = filter
		fmt.Printf("JSON Filter apply for attribut %s: %+v\n", attributeName, filter)
	}
}

// set custom label configuration using env variable
func (bt *dbeat) setCLConfig() {
	if cs := os.Getenv("CUSTOM_LABELS"); cs != "" {
		fmt.Printf("Custom labels: %s\n", cs)
		bt.config.CustomLabels = make([]string, 0)
		list := strings.Split(cs, ",")
		for _, pattern := range list {
			bt.config.CustomLabels = append(bt.config.CustomLabels, strings.TrimSpace(pattern))
		}
	}
}

// set custom label configuration using env variable
func (bt *dbeat) setExcludedConfig() {
	if cs := os.Getenv("EXCLUDED_CONTAINERS"); cs != "" {
		fmt.Printf("Excluded containers: %s\n", cs)
		bt.config.ExcludedContainers = make([]string, 0)
		list := strings.Split(cs, ",")
		for _, pattern := range list {
			bt.config.ExcludedContainers = append(bt.config.ExcludedContainers, strings.TrimSpace(pattern))
		}
	}
	if cs := os.Getenv("EXCLUDED_SERVICES"); cs != "" {
		fmt.Printf("Excluded service: %s\n", cs)
		bt.config.ExcludedServices = make([]string, 0)
		list := strings.Split(cs, ",")
		for _, pattern := range list {
			bt.config.ExcludedServices = append(bt.config.ExcludedServices, strings.TrimSpace(pattern))
		}
	}
	if cs := os.Getenv("EXCLUDED_STACKS"); cs != "" {
		fmt.Printf("Excluded stacks: %s\n", cs)
		bt.config.ExcludedStacks = make([]string, 0)
		list := strings.Split(cs, ",")
		for _, pattern := range list {
			bt.config.ExcludedStacks = append(bt.config.ExcludedStacks, strings.TrimSpace(pattern))
		}
	}
}

// Run dbeat main loop
func (bt *dbeat) Run(b *beat.Beat) error {
	logp.Info("starting dbeat")
	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	err := bt.start(&bt.config)
	if err != nil {
		log.Fatal(err)
	}
	bt.setMLConfig()
	bt.setCLConfig()
	bt.setExcludedConfig()
	bt.setJSONFilterConfig()
	bt.initAPI()
	fmt.Printf("Config: %+v\n", bt.config)
	if info, errc := bt.dockerClient.Info(context.Background()); errc == nil {
		bt.hostname = info.Name
	}
	bt.hostIP = bt.getHostIP()
	bt.beaterStarted = true
	bt.containers = make(map[string]*ContainerData)
	ContainerListOptions := types.ContainerListOptions{All: true}
	containers, err := bt.dockerClient.ContainerList(context.Background(), ContainerListOptions)
	if err != nil {
		return err
	}
	//bt.eventStreamReading = true
	for _, cont := range containers {
		bt.addContainer(cont.ID)
	}
	bt.lastUpdate = time.Now()
	fmt.Println("dbeat is running! Hit CTRL-C to stop it.")

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		bt.tick()
	}
}

func (bt *dbeat) getHostIP() string {
	return bt.getHTTPString("http://169.254.169.254/latest/meta-data/local-ipv4")
}

func (bt *dbeat) getHTTPString(url string) string {
	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	res, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}
	return string(data)
}

// Stop dbeat stop
func (bt *dbeat) Stop() {
	bt.client.Close()
	bt.Close()
	close(bt.done)
}
