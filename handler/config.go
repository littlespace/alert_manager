package handler

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

type Outs []struct {
	Severity string
	SendTo   []string `yaml:"send_to"`
}

func (o Outs) Get(sev string) []string {
	for _, i := range o {
		if i.Severity == sev {
			return i.SendTo
		}
	}
	return []string{}
}

type GeneralConfig struct {
	DefaultOutputs        Outs          `yaml:"default_outputs"`
	ClearHolddownInterval time.Duration `yaml:"clear_holddown_interval"`
}

type AlertConfig struct {
	Name   string
	Config struct {
		Scope            string
		Severity         string
		Tags             []string
		Source           string
		AutoExpire       *bool         `yaml:"auto_expire"`
		ExpireAfter      time.Duration `yaml:"expire_after"`
		AutoClear        *bool         `yaml:"auto_clear"`
		NotifyOnClear    bool          `yaml:"notify_on_clear"`
		NotifyDelay      time.Duration `yaml:"notify_delay"`
		NotifyRemind     time.Duration `yaml:"notify_remind"`
		DisableNotify    bool          `yaml:"disable_notify"`
		Outputs          Outs
		AggregationRules []string `yaml:"aggregation_rules"`
		EscalationRules  []struct {
			After      time.Duration
			EscalateTo string `yaml:"escalate_to"`
		} `yaml:"escalation_rules"`
	}
}

type AggregationRuleConfig struct {
	Name         string
	Window       time.Duration
	PollInterval time.Duration `yaml:"poll_interval"`
	Alert        AlertConfig
}

type SuppressionRuleConfig struct {
	Name           string
	Duration       time.Duration
	Reason         string
	MatchCondition string `yaml:"match_condition"`
	Matches        map[string]interface{}
}

type InhibitRuleConfig struct {
	Name     string
	Delay    time.Duration
	SrcMatch struct {
		Alert string
		Label string
	} `yaml:"source_match"`
	TargetMatches []struct {
		Alert string
		Label string
	} `yaml:"target_matches"`
}

type configs struct {
	GeneralConfig          GeneralConfig           `yaml:"general_config"`
	AlertConfig            []AlertConfig           `yaml:"alert_config"`
	AggregationRuleConfigs []AggregationRuleConfig `yaml:"aggregation_rules"`
	SuppressionRuleConfigs []SuppressionRuleConfig `yaml:"suppression_rules"`
	InhibitRuleConfigs     []InhibitRuleConfig     `yaml:"inhibit_rules"`
}

func readConfig(file string) (configs, error) {
	absPath, _ := filepath.Abs(file)
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		glog.Errorf("Unable to read config file: %v", err)
		return configs{}, err
	}
	c := configs{}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		glog.Errorf("Unable to decode yaml: %v", err)
		return configs{}, err
	}
	return c, nil
}

type ConfigHandler struct {
	file          string
	generalConfig GeneralConfig
	alertConfigs  map[string]AlertConfig
	aggRules      map[string]AggregationRuleConfig
	suppRules     map[string]SuppressionRuleConfig
	inhibitRules  map[string]InhibitRuleConfig
	sync.Mutex
}

func NewConfigHandler(file string) *ConfigHandler {
	c := &ConfigHandler{
		file:         file,
		alertConfigs: make(map[string]AlertConfig),
		aggRules:     make(map[string]AggregationRuleConfig),
		suppRules:    make(map[string]SuppressionRuleConfig),
		inhibitRules: make(map[string]InhibitRuleConfig),
	}
	c.LoadConfig()
	return c
}

var Config *ConfigHandler

func (c *ConfigHandler) LoadConfig() {
	c.Lock()
	defer c.Unlock()
	if c.file == "" {
		return
	}
	glog.Infof("Loading configs from file: %s", c.file)
	configs, err := readConfig(c.file)
	if err != nil {
		glog.Fatalf("Unable to load config file : %v", err)
	}
	c.generalConfig = configs.GeneralConfig
	for _, config := range configs.AlertConfig {
		c.alertConfigs[config.Name] = config
	}
	for _, rule := range configs.AggregationRuleConfigs {
		c.aggRules[rule.Name] = rule
		c.alertConfigs[rule.Alert.Name] = rule.Alert
	}
	for _, rule := range configs.SuppressionRuleConfigs {
		c.suppRules[rule.Name] = rule
	}
	for _, rule := range configs.InhibitRuleConfigs {
		c.inhibitRules[rule.Name] = rule
	}
}

func (c *ConfigHandler) GetGeneralConfig() GeneralConfig {
	c.Lock()
	defer c.Unlock()
	return c.generalConfig
}

func (c *ConfigHandler) GetConfiguredAlerts() []AlertConfig {
	c.Lock()
	defer c.Unlock()
	configs := []AlertConfig{}
	for _, config := range c.alertConfigs {
		configs = append(configs, config)
	}
	return configs
}

func (c *ConfigHandler) GetAggRules() []AggregationRuleConfig {
	c.Lock()
	defer c.Unlock()
	rules := []AggregationRuleConfig{}
	for _, rule := range c.aggRules {
		rules = append(rules, rule)
	}
	return rules
}

func (c *ConfigHandler) GetSuppressionRules() []SuppressionRuleConfig {
	c.Lock()
	defer c.Unlock()
	rules := []SuppressionRuleConfig{}
	for _, rule := range c.suppRules {
		rules = append(rules, rule)
	}
	return rules
}

func (c *ConfigHandler) GetInhibitRules() []InhibitRuleConfig {
	c.Lock()
	defer c.Unlock()
	rules := []InhibitRuleConfig{}
	for _, rule := range c.inhibitRules {
		rules = append(rules, rule)
	}
	return rules
}

func (c *ConfigHandler) GetAlertConfig(name string) (AlertConfig, bool) {
	c.Lock()
	defer c.Unlock()
	config, ok := c.alertConfigs[name]
	return config, ok
}

func (c *ConfigHandler) GetAggregationRuleConfig(name string) (AggregationRuleConfig, bool) {
	c.Lock()
	defer c.Unlock()
	rule, ok := c.aggRules[name]
	return rule, ok
}

func (c *ConfigHandler) GetInhibitRuleConfig(name string) (InhibitRuleConfig, bool) {
	c.Lock()
	defer c.Unlock()
	rule, ok := c.inhibitRules[name]
	return rule, ok
}
