package configmgr

// Logger is a simple interface for logging inside ConfigManager.
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
}

func (cm *ConfigManager) SetLogger(l Logger) {
	cm.logger = l
}
