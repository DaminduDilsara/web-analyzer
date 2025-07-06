package configurations

type AppConfigurations struct {
	AppPort      int `yaml:"app_port"`
	MetricPort   int `yaml:"metric_port"`
	WriteTimeout int `yaml:"write_timeout"`
	ReadTimeOut  int `yaml:"read_time_out"`
	IdleTimeout  int `yaml:"idle_timeout"`
}
