package configmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ConfigManager is the core configuration manager.
type ConfigManager struct {
	data map[string]interface{}
}

// NewConfigManager creates a new ConfigManager instance.
func NewConfigManager() *ConfigManager {
	return &ConfigManager{data: make(map[string]interface{})}
}

// LoadFromFile loads configuration from a JSON or YAML file.
func (cm *ConfigManager) LoadFromFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(path))
	tmp := make(map[string]interface{})

	switch ext {
	case ".json":
		if err = json.Unmarshal(raw, &tmp); err != nil {
			return err
		}
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(raw, &tmp); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	for k, v := range tmp {
		cm.data[normalizeKey(k)] = normalizeValue(v)
	}

	return nil
}

// LoadFromDotEnv loads variables from a .env file into both system env and cm.data.
func (cm *ConfigManager) LoadFromDotEnv(path string) error {
	if path == "" {
		path = ".env"
	}
	envMap, err := godotenv.Read(path)
	if err != nil {
		return err
	}
	for k, v := range envMap {
		_ = os.Setenv(k, v)
		cm.data[normalizeKey(k)] = normalizeValue(v)
	}
	return nil
}

// LoadFromSysEnv loads a single environment variable into cm.data.
func (cm *ConfigManager) LoadFromSysEnv(key string) {
	if val, ok := os.LookupEnv(key); ok {
		cm.data[normalizeKey(key)] = normalizeKey(val)
	}
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
