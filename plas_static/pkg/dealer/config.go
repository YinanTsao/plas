package dealer

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Users       []string           `yaml:"users"`
	Deployments []DeploymentConfig `yaml:"deployments"`
}

// Deployments setup
type DeploymentConfig struct {
	DeploymentName     string  `yaml:"deploymentName"`
	Namespace          string  `yaml:"namespace"`
	ServiceTime        float64 `yaml:"serviceTime"`
	SLO                int     `yaml:"SLO"`
	ServiceJaeger      string  `yaml:"serviceJaeger"`
	RPSPerUser         float64 `yaml:"rpsPerUser"`
	OperationJaeger    string  `yaml:"operationJaeger"`
	IntervalMonitoring int     `yaml:"intervalMonitoring"`
	OptiPref           int     `yaml:"optiPref"`
	SLOVioToleration   float64 `yaml:"SLOVioToleration"`
}

func ReadConfig(filename string) (*Config, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) GetUserIPs() []string {
	return c.Users
}
