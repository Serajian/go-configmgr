package configmgr

// NewTestConfig creates a ConfigManager with preloaded key-value pairs.
// Useful for unit tests without needing files or environment variables.
func NewTestConfig(kv map[string]string) *ConfigManager {
	cm := NewConfigManager()
	for k, v := range kv {
		cm.Set(k, v)
	}
	return cm
}
