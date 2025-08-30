package configmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	if cm.logger != nil {
		cm.logger.Info("load_from_file_success", map[string]interface{}{"path": path})
	}
	return nil
}

// LoadFiles loads multiple config files in order.
// Later files override earlier ones.
func (cm *ConfigManager) LoadFiles(paths ...string) error {
	for _, path := range paths {
		if err := cm.LoadFromFile(path); err != nil {
			return err
		}
	}
	return nil
}
