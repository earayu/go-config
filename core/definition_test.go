package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfigSet(t *testing.T) {
	c, err := NewConfigSet("test1", []string{"../test"}, "ini", "test1.ini")
	assert.Nil(t, err)

	c.Register(&ConfigItem{
		key: "foo",
		defaultValue: func() any {
			return "bar"
		},
	})
	c.Register(&ConfigItem{
		key: "count",
		validation: func(newValue string) error {
			return nil
		},
		defaultValue: func() any {
			return 100
		},
		generateValue: func(configItem *ConfigItem, newValue string) any {
			return newValue
		},
		description: func() string {
			return "desc"
		},
		dynamicReloadHook: func(configItem *ConfigItem, newValue any) error {
			return nil
		},
	})
	c.LoadAndWatchConfigFile()

	ci := c.configItemMap["foo"]
	assert.Equal(t, ci.value, "bar")
}
