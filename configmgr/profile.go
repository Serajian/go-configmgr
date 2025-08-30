package configmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadWithProfile loads a base config file and, if available, a profile-specific file
// based on the value of a given environment variable.
//
// Supported file types: .json, .yaml, .yml, .env
//
// Behavior:
//   - If envKey is not set, only the baseFile is loaded.
//   - If envKey=dev and baseFile=config.yaml, then config.yaml + config-dev.yaml are loaded.
//   - If envKey=prod and baseFile=config.json, then config.json + config-prod.json are loaded.
//   - If envKey=dev and baseFile=.env, then .env + .env.dev are loaded.
//   - If profile-specific file does not exist, only the baseFile is used.
//
// Example:
//
//	os.Setenv("APP_ENV", "dev")
//	cm.LoadWithProfile("APP_ENV", "config.yaml") // loads config.yaml + config-dev.yaml
//	cm.LoadWithProfile("APP_ENV", ".env")        // loads .env + .env.dev
func (cm *ConfigManager) LoadWithProfile(envKey, baseFile string) error {
	ext := strings.ToLower(filepath.Ext(baseFile))

	// determine profile (dev, staging, prod, etc.)
	env := os.Getenv(envKey)
	if env == "" {
		if ext == ".env" {
			return cm.LoadFromDotEnv(baseFile)
		}
		return cm.LoadFromFile(baseFile)
	}

	switch ext {
	case ".json", ".yaml", ".yml":
		// base
		if err := cm.LoadFromFile(baseFile); err != nil {
			return err
		}
		// profile
		name := strings.TrimSuffix(baseFile, ext)
		profileFile := fmt.Sprintf("%s-%s%s", name, env, ext)
		if _, err := os.Stat(profileFile); err == nil {
			return cm.LoadFromFile(profileFile)
		}
		return nil

	case ".env":
		// base
		if err := cm.LoadFromDotEnv(baseFile); err != nil {
			return err
		}
		// profile
		profileFile := fmt.Sprintf("%s.%s", baseFile, env) // e.g. .env.dev
		if _, err := os.Stat(profileFile); err == nil {
			return cm.LoadFromDotEnv(profileFile)
		}
		return nil

	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}
}
