package core

type validationFunc func() error
type descriptionFunc func() string
type defaultValueFunc func() any

type ConfigDefinition struct {
	key   string
	value any

	defaultValueFunc defaultValueFunc
	validationFunc   validationFunc
	descriptionFunc  descriptionFunc

	alias     map[string]bool
	hotReload bool
}
