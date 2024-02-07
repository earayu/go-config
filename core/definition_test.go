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
		defaultValueFunc: func() any {
			return "bar"
		},
	})
	c.LoadAndWatchConfigFile()

	ci := c.configItemMap["foo"]
	assert.Equal(t, ci.value, "baz")
}
