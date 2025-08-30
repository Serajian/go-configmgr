package configmgr

import (
	"os"

	"github.com/joho/godotenv"
)

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
	if cm.logger != nil {
		cm.logger.Info("loaded env", map[string]interface{}{"path": path})
	}
	return nil
}

// LoadFromSysEnv loads a single environment variable into cm.data.
func (cm *ConfigManager) LoadFromSysEnv(key string) {
	if val, ok := os.LookupEnv(key); ok {
		cm.data[normalizeKey(key)] = normalizeKey(val)
	}
	if cm.logger != nil {
		cm.logger.Info("loaded system env", map[string]interface{}{"key": key})
	}
}
