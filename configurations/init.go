package configurations

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	AppConfig *AppConfigurations `yaml:"app_config"`
	LogConfig *LogConfigurations `yaml:"log_config"`
}

func LoadConfigurations() *Config {
	var configs Config

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Printf("Loading config yaml err: %v. Loading configs using env variables", err)
		//configs = loadConfigFromEnv()  // write this function if env variable config loading required
	} else {
		err = yaml.Unmarshal(yamlFile, &configs)
		if err != nil {
			log.Fatalf("App Config Unmarshal error: %v", err)
		}
	}

	return &configs
}
