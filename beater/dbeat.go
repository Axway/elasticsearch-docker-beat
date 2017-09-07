package beater

import (
	"fmt"
	"log"
	"time"

	"github.com/Axway/elasticsearch-docker-beat/config"
	"github.com/docker/docker/client"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
)

// dbeat the amp beat struct
type dbeat struct {
	done                chan struct{}
	config              config.Config
	client              publisher.Client
	dockerClient        *client.Client
	eventStreamReading  bool
	containers          map[string]*ContainerData
	lastUpdate          time.Time
	logsSavedDatePeriod int
}

// New Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}
	log.Printf("Config: %+v\n", config)
	bt := &dbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

// Run dbeat main loop
func (bt *dbeat) Run(b *beat.Beat) error {
	logp.Info("starting dbeat")
	fmt.Printf("config: %v\n", bt.config)

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	err := bt.start(&bt.config)
	if err != nil {
		log.Fatal(err)
	}
	logp.Info("dbeat is running! Hit CTRL-C to stop it.")

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		bt.tick()
	}
}

// Stop dbeat stop
func (bt *dbeat) Stop() {
	bt.client.Close()
	bt.Close()
	close(bt.done)
}
