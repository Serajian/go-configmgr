package configmgr

import (
	"strconv"
	"strings"
)

// ConfigManager is the core configuration manager.
type ConfigManager struct {
	data   map[string]interface{}
	logger Logger
}

// NewConfigManager creates a new ConfigManager instance.
func NewConfigManager() *ConfigManager {
	return &ConfigManager{data: make(map[string]interface{})}
}

// Get returns a raw value from config data.
func (cm *ConfigManager) Get(key string) interface{} {
	return cm.data[normalizeKey(key)]
}

// Set sets a config value manually.
func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.data[normalizeKey(key)] = normalizeValue(value)
}

// GetAll returns all config data.
func (cm *ConfigManager) GetAll() map[string]interface{} {
	return cm.data
}

// normalizeKey ensures all keys are stored in uppercase.
func normalizeKey(key string) string {
	return strings.ToUpper(key)
}

func normalizeValue(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		s := strings.TrimSpace(str)

		// try parse int
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}

		// try parse bool
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}

		// default: return as string
		return s
	}
	return value
}
