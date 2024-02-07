package core

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/appengine/log"
	"os"
	"sync"
)

type validationFunc func() error
type descriptionFunc func() string
type defaultValueFunc func() any
type setValueFunc func(newValue any) error

type ConfigItem struct {
	key   string
	value any

	defaultValueFunc defaultValueFunc
	validationFunc   validationFunc
	descriptionFunc  descriptionFunc
	setValueFunc     setValueFunc

	alias         map[string]bool
	dynamicReload bool
}

const (
	IGNORE = "IGNORE"
	ERROR  = "ERROR"
	EXIT   = "EXIT"
)

type ConfigSet struct {
	mu       sync.Mutex
	reloadMu sync.Mutex

	ctx context.Context

	name          string
	configItemMap map[string]*ConfigItem

	vp *viper.Viper
	Fs *pflag.FlagSet

	ConfigPath                 []string
	ConfigType                 string
	ConfigName                 string
	ConfigFileNotFoundHandling string
}

func NewConfigSet(
	name string,
	configPath []string,
	configType string,
	configName string) (*ConfigSet, error) {

	vp := viper.New()
	vp.SetConfigFile(configName)
	vp.SetConfigType(configType)
	for _, p := range configPath {
		vp.AddConfigPath(p)
	}

	c := &ConfigSet{
		name:                       name,
		configItemMap:              make(map[string]*ConfigItem),
		vp:                         vp,
		ConfigFileNotFoundHandling: IGNORE,
	}
	return c, nil
}

func (c *ConfigSet) String() string {
	return fmt.Sprintf("ConfigPath=%s, ConfigType=%s, ConfigName=%s, ConfigFileNotFoundHandling=%s",
		c.ConfigPath, c.ConfigType, c.ConfigName, c.ConfigFileNotFoundHandling)
}

func (c *ConfigSet) LoadAndWatchConfigFile() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.vp.SetConfigName(c.ConfigName)
	c.vp.SetConfigType(c.ConfigType)
	for _, p := range c.ConfigPath {
		c.vp.AddConfigPath(p)
	}
	err := c.vp.ReadInConfig()
	if err != nil {
		switch c.ConfigFileNotFoundHandling {
		case IGNORE:
			log.Infof(c.ctx, "ViperConfig: %c", c)
			log.Infof(c.ctx, "read config file error, err: %c", err)
		case ERROR:
			log.Errorf(c.ctx, "ViperConfig: %c", c)
			log.Errorf(c.ctx, "read config file error, err: %c", err)
		case EXIT:
			log.Errorf(c.ctx, "ViperConfig: %c", c)
			log.Errorf(c.ctx, "read config file error, err: %c", err)
			os.Exit(2)
		}
	}
	c.loadConfigFileAtStartup()
	c.startWatch()
}

func (c *ConfigSet) loadConfigFileAtStartup() {
	c.reloadMu.Lock()
	defer c.reloadMu.Unlock()
	for _, sectionAndKey := range c.vp.AllKeys() {
		configItem, exists := getConfigItem(c, sectionAndKey)
		if !exists {
			continue
		}
		newValue := c.vp.GetString(sectionAndKey)
		configItem.setValueFunc(newValue)
	}
}

func (c *ConfigSet) reloadConfigs() {
	c.reloadMu.Lock()
	defer c.reloadMu.Unlock()
	for _, sectionAndKey := range c.vp.AllKeys() {
		configItem, exists := getConfigItem(c, sectionAndKey)
		if !exists {
			continue
		}
		newValue := c.vp.GetString(sectionAndKey)
		configItem.setValueFunc(newValue)
	}
}

func (c *ConfigSet) startWatch() {
	c.vp.OnConfigChange(func(e fsnotify.Event) {
		c.reloadConfigs()
	})
	c.vp.WatchConfig()
}
