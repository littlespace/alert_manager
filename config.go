package alert_manager

import (
	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/reporting"
	"github.com/mitchellh/mapstructure"
	"time"
)

type AgentConfig struct {
	ApiAddr             string        `mapstructure:"api_addr"`
	StatsExportInterval time.Duration `mapstructure:"stats_export_interval"`
}

type DbConfig struct {
	Addr, Username, Password string
	DbName                   string `mapstructure:"db_name"`
	Timeout                  int
}

type Config struct {
	Agent    *AgentConfig
	Db       *DbConfig
	Reporter *reporting.InfluxReporter
}

func (c *Config) UnmarshalTOML(data interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{DecodeHook: mapstructure.StringToTimeDurationHookFunc()}
	d, _ := data.(map[string]interface{})
	for key, value := range d {
		v, _ := value.(map[string]interface{})
		switch key {
		case "agent":
			a := &AgentConfig{}
			decoderConfig.Result = a
			decoder, _ := mapstructure.NewDecoder(decoderConfig)
			if err := decoder.Decode(v); err != nil {
				return err
			}
			c.Agent = a
		case "db":
			d := &DbConfig{}
			decoderConfig.Result = d
			decoder, _ := mapstructure.NewDecoder(decoderConfig)
			if err := decoder.Decode(v); err != nil {
				return err
			}
			c.Db = d
		case "reporter":
			r := &reporting.InfluxReporter{}
			decoderConfig.Result = r
			decoder, _ := mapstructure.NewDecoder(decoderConfig)
			if err := decoder.Decode(v); err != nil {
				return err
			}
			c.Reporter = r
		case "listeners":
			for lisKey, lisValue := range v {
				lv, _ := lisValue.(map[string]interface{})
				if listener, ok := Listeners[lisKey]; ok {
					decoderConfig.Result = listener
					decoder, _ := mapstructure.NewDecoder(decoderConfig)
					if err := decoder.Decode(lv); err != nil {
						return err
					}
				}
			}
		case "outputs":
			for rKey, rValue := range v {
				rv, _ := rValue.(map[string]interface{})
				if receiver, ok := Outputs[rKey]; ok {
					decoderConfig.Result = receiver
					decoder, _ := mapstructure.NewDecoder(decoderConfig)
					if err := decoder.Decode(rv); err != nil {
						return err
					}
				}
			}
		case "processors":
			for rKey, rValue := range v {
				rv, _ := rValue.(map[string]interface{})
				if receiver, ok := Processors[rKey]; ok {
					decoderConfig.Result = receiver
					decoder, _ := mapstructure.NewDecoder(decoderConfig)
					if err := decoder.Decode(rv); err != nil {
						return err
					}
				}
			}
		case "transforms":
			for tKey, tValue := range v {
				tv, _ := tValue.(map[string]interface{})
				for _, xform := range handler.Transforms {
					if xform.Name() == tKey {
						decoderConfig.Result = xform
						decoder, _ := mapstructure.NewDecoder(decoderConfig)
						if err := decoder.Decode(tv); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func NewConfig(configFile string) *Config {
	config := &Config{}
	if _, err := toml.DecodeFile(configFile, config); err != nil {
		// failure to parse config is considered a fatal error
		glog.Fatalf("Error decoding config file: %v", err)
	}
	return config
}
