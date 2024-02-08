package core

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"sync"
)

type validationFunc func() error
type descriptionFunc func() string
type defaultValueFunc func() any
type generateValueFunc func(configItem *ConfigItem, newValue string) any
type dynamicReloadHookFunc func(configItem *ConfigItem, newValue any) error

type ConfigItem struct {
	key   string
	value any

	defaultValue      defaultValueFunc
	generateValue     generateValueFunc
	validation        validationFunc
	description       descriptionFunc
	dynamicReloadHook dynamicReloadHookFunc

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
	fs *pflag.FlagSet

	configPath                 []string
	configType                 string
	configName                 string
	configFileNotFoundHandling string
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

	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)

	c := &ConfigSet{
		name:                       name,
		configItemMap:              make(map[string]*ConfigItem),
		vp:                         vp,
		fs:                         fs,
		configFileNotFoundHandling: IGNORE,
	}
	return c, nil
}

func (c *ConfigSet) Register(item *ConfigItem) error {
	//todo check
	c.configItemMap[item.key] = item
	return nil
}

func (c *ConfigSet) String() string {
	return fmt.Sprintf("configPath=%s, configType=%s, configName=%s, configFileNotFoundHandling=%s",
		c.configPath, c.configType, c.configName, c.configFileNotFoundHandling)
}

func (c *ConfigSet) LoadAndWatchConfigFile() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.vp.SetConfigName(c.configName)
	c.vp.SetConfigType(c.configType)
	for _, p := range c.configPath {
		c.vp.AddConfigPath(p)
	}
	err := c.vp.ReadInConfig()
	if err != nil {
		switch c.configFileNotFoundHandling {
		case IGNORE:
			//log.Infof(c.ctx, "ViperConfig: %c", c)
			//log.Infof(c.ctx, "read config file error, err: %c", err)
		case ERROR:
			//log.Errorf(c.ctx, "ViperConfig: %c", c)
			//log.Errorf(c.ctx, "read config file error, err: %c", err)
		case EXIT:
			//log.Errorf(c.ctx, "ViperConfig: %c", c)
			//log.Errorf(c.ctx, "read config file error, err: %c", err)
			os.Exit(2)
		}
	}
	c.loadConfigFileAtStartup()
	c.startWatch()
}

func (c *ConfigSet) loadConfigFileAtStartup() {
	c.reloadMu.Lock()
	defer c.reloadMu.Unlock()
	for _, ci := range c.configItemMap {
		ci.value = ci.defaultValue()
	}
	for _, sectionAndKey := range c.vp.AllKeys() {
		configItem, exists := getConfigItem(c, sectionAndKey)
		if !exists {
			continue
		}
		newValue := c.vp.GetString(sectionAndKey)
		if configItem.dynamicReloadHook != nil {
			configItem.dynamicReloadHook(configItem, newValue)
		}
		configItem.value = configItem.generateValue(configItem, newValue)
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
		if configItem.dynamicReloadHook != nil {
			configItem.dynamicReloadHook(configItem, newValue)
		}
		configItem.value = configItem.generateValue(configItem, newValue)
	}
}

func (c *ConfigSet) startWatch() {
	c.vp.OnConfigChange(func(e fsnotify.Event) {
		c.reloadConfigs()
	})
	c.vp.WatchConfig()
}
