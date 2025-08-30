package configmgr

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// ToJSON returns config as pretty JSON.
func (cm *ConfigManager) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cm.GetAll(), "", "  ")
}

// ToYAML returns config as YAML.
func (cm *ConfigManager) ToYAML() ([]byte, error) {
	return yaml.Marshal(cm.GetAll())
}
