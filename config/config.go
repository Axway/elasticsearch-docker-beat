// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

// Config dbeat config
type Config struct {
	RESTPort                   int                          `config:"rest_port"`
	Period                     time.Duration                `config:"period"`
	DockerURL                  string                       `config:"docker_url"`
	TLS                        bool                         `config:"tls"`
	CaPath                     string                       `config:"ca_path"`
	CertPath                   string                       `config:"cert_path"`
	KeyPath                    string                       `config:"key_path"`
	Logs                       bool                         `config:"logs"`
	LogsDateSavePeriod         int                          `config:"logs_position_save_period"`
	Net                        bool                         `config:"net"`
	Memory                     bool                         `config:"memory"`
	IO                         bool                         `config:"io"`
	CPU                        bool                         `config:"cpu"`
	LogsMultilineMaxSize       int                          `config:"logs_multiline_max_size"`
	LogsMultiline              map[string]map[string]string `config:"logs_multiline"`
	CustomLabels               []string                     `config:"custom_labels"`
	ExcludedContainers         []string                     `config:"excluded_containers"`
	ExcludedServices           []string                     `config:"excluded_services"`
	ExcludedStacks             []string                     `config:"excluded_stacks"`
	LogsJSONOnly               bool                         `config:"logs_json_only"`
	LogsJSONFilters            map[string]map[string]string `config:"logs_json_filters"`
	LogsPlainFilters           []string                     `config:"logs_plain_filters"`
	LogsPlainFiltersContainers map[string][]string          `config:"logs_plain_filters_containers"`
	LogsPlainFiltersServices   map[string][]string          `config:"logs_plain_filters_services"`
	LogsPlainFiltersStacks     map[string][]string          `config:"logs_plain_filters_stacks"`
}

// MLConfig multiline config struct
type MLConfig struct {
	Activated bool
	Pattern   string
	Negate    bool
	Append    bool
}

// JSONFilter json filter config struct
type JSONFilter struct {
	Name      string
	Pattern   string
	Negate    bool
	Activated bool
}

//DefaultConfig dbeat default config
var DefaultConfig = Config{
	RESTPort:                   3000,
	Period:                     10 * time.Second,
	DockerURL:                  "unix:///var/run/docker.sock",
	Logs:                       true,
	LogsDateSavePeriod:         10,
	Net:                        true,
	Memory:                     true,
	IO:                         true,
	CPU:                        true,
	LogsMultilineMaxSize:       100000,
	LogsMultiline:              make(map[string]map[string]string),
	CustomLabels:               make([]string, 0),
	ExcludedContainers:         make([]string, 0),
	ExcludedServices:           make([]string, 0),
	ExcludedStacks:             make([]string, 0),
	LogsJSONOnly:               false,
	LogsJSONFilters:            make(map[string]map[string]string),
	LogsPlainFilters:           make([]string, 0),
	LogsPlainFiltersContainers: make(map[string][]string),
	LogsPlainFiltersServices:   make(map[string][]string),
	LogsPlainFiltersStacks:     make(map[string][]string),
}
