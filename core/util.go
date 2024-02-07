package core

import "strings"

// todo support get by alias
func getConfigItem(c *ConfigSet, sectionAndKey string) (*ConfigItem, bool) {
	key := sectionAndKey
	if strings.Contains(sectionAndKey, ".") {
		// remove section from key
		key = strings.SplitN(sectionAndKey, ".", 2)[1]
	}

	item, exists := c.configItemMap[key]
	return item, exists
}
