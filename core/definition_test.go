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
		defaultValue: func() any {
			return 100
		},
	})
	c.LoadAndWatchConfigFile()

	ci := c.configItemMap["foo"]
	assert.Equal(t, ci.value, "bar")
}
