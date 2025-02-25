package chart

import (
	"github.com/imdario/mergo"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"strings"
)

type Component struct {
	version       string
	name          string
	profile       string
	namespace     string
	configuration map[string]interface{}
}

func (c *Component) Configuration() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range c.configuration {
		if err := mergo.Merge(&result, c.convertToNestedMap(key, value), mergo.WithOverride); err != nil {
			return nil, err
		}
	}
	return result, nil
}

//convertToNestedMap converts a key with dot-notation into a nested map (e.g. a.b.c=value become [a:[b:[c:value]]])
func (c *Component) convertToNestedMap(key string, value interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	tokens := strings.Split(key, ".")
	lastNestedMap := result
	for depth, token := range tokens {
		switch depth {
		case len(tokens) - 1: //last token reached, stop nesting
			lastNestedMap[token] = value
		default:
			lastNestedMap[token] = make(map[string]interface{})
			lastNestedMap = lastNestedMap[token].(map[string]interface{})
		}
	}
	return result
}

type ComponentBuilder struct {
	component *Component
}

func NewComponentBuilder(version, name string) *ComponentBuilder {
	return &ComponentBuilder{
		&Component{
			version:       version,
			name:          name,
			configuration: make(map[string]interface{}),
		},
	}
}

func (cb *ComponentBuilder) WithProfile(profile string) *ComponentBuilder {
	cb.component.profile = profile
	return cb
}

func (cb *ComponentBuilder) WithNamespace(namespace string) *ComponentBuilder {
	cb.component.namespace = namespace
	return cb
}

func (cb *ComponentBuilder) WithConfiguration(config []reconciler.Configuration) *ComponentBuilder {
	for _, kvEntry := range config {
		cb.component.configuration[kvEntry.Key] = kvEntry.Value
	}
	return cb
}

func (cb *ComponentBuilder) Build() *Component {
	return cb.component
}
