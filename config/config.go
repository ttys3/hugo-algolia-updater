package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

var Cfg Config

// use a single instance of Validate, it caches struct info
var validate = validator.New()

type Config struct {
	Algolia struct {
		Index    string `yaml:"index" validate:"ascii"`
		AdminKey string `yaml:"admin-key" validate:"required"`
		AppID    string `yaml:"app-id" validate:"required"`
	} `yaml:"algolia" validate:"required"`

	Http struct {
		Proxy string `yaml:"proxy"`
	} `yaml:"http"`

	Segment struct {
		Dict struct {
			Path     string `yaml:"path"`
			StopPath string `yaml:"stop-path"`
		} `yaml:"dict"`
	} `yaml:"segment"`
}

func (c *Config) Load(cfgfile string) error {
	zap.S().Infof("begin reading config: %s", cfgfile)
	content, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		return fmt.Errorf("reading config file %s failed, err=%w", cfgfile, err)
	}
	if err := yaml.Unmarshal(content, c); err != nil {
		return fmt.Errorf("unmarshal config file %s failed, err=%w", cfgfile, err)
	}
	return nil
}

func (c *Config) Validate() error {
	err := validate.Struct(c)
	if err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			return validateErr
		}
		return err
	}
	return nil
}
