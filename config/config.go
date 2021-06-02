package config

type Config struct {
	Algolia struct {
		Index string `yaml:"index"`
		Key   string `yaml:"key"`
		AppID string `yaml:"appid"`
	}

	Http struct {
		Proxy string `yaml:"http-proxy"`
	}

	Segment struct {
		Dict struct {
			Path     string `yaml:"path"`
			StopPath string `yaml:"stop-path"`
		}
	}
}
