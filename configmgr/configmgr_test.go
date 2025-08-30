package configmgr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name  string `json:"APP_NAME" default:"MyService" validate:"required"`
	Port  int    `json:"APP_PORT" default:"8080" validate:"gte=1000,lte=9999"`
	Debug bool   `json:"APP_DEBUG" default:"false"`
}

func TestConfigManager(t *testing.T) {
	cm := NewConfigManager()

	envContent := "APP_NAME=TestApp\nAPP_PORT=9090\nAPP_DEBUG=true\n"
	envPath := filepath.Join(os.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(envPath) }()

	// لود از .env
	if err := cm.LoadFromDotEnv(envPath); err != nil {
		t.Fatalf("LoadFromDotEnv failed: %v", err)
	}

	// تست Get
	if cm.Get("APP_NAME") != "TestApp" {
		t.Errorf("expected APP_NAME=TestApp, got %v", cm.Get("APP_NAME"))
	}
	if cm.Get("APP_PORT") != 9090 {
		t.Errorf("expected APP_PORT=9090, got %v", cm.Get("APP_PORT"))
	}
	if cm.Get("APP_DEBUG") != true {
		t.Errorf("expected APP_DEBUG=true, got %v", cm.Get("APP_DEBUG"))
	}

	// تست Set
	cm.Set("CUSTOM_KEY", "customValue")
	if cm.Get("CUSTOM_KEY") != "customValue" {
		t.Errorf("expected CUSTOM_KEY=customValue, got %v", cm.Get("CUSTOM_KEY"))
	}

	// تست Unmarshal
	var appCfg AppConfig
	if err := cm.Unmarshal(&appCfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if appCfg.Name != "TestApp" {
		t.Errorf("expected Name=TestApp, got %v", appCfg.Name)
	}
	if appCfg.Port != 9090 {
		t.Errorf("expected Port=9090, got %v", appCfg.Port)
	}
	if appCfg.Debug != true {
		t.Errorf("expected Debug=true, got %v", appCfg.Debug)
	}
}

func TestUnmarshalWithDefaultsAndValidation(t *testing.T) {
	cm := NewConfigManager()

	envContent := ""
	envPath := filepath.Join(os.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(envPath)

	if err := cm.LoadFromDotEnv(envPath); err != nil {
		t.Fatalf("LoadFromDotEnv failed: %v", err)
	}

	var cfg AppConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if cfg.Name != "MyService" {
		t.Errorf("expected default Name=MyService, got %v", cfg.Name)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default Port=8080, got %v", cfg.Port)
	}
	if cfg.Debug != false {
		t.Errorf("expected default Debug=false, got %v", cfg.Debug)
	}

	cm.Set("APP_NAME", "EnvService")
	cm.Set("APP_PORT", 9090)
	cm.Set("APP_DEBUG", true)

	var cfg2 AppConfig
	if err := cm.Unmarshal(&cfg2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if cfg2.Name != "EnvService" {
		t.Errorf("expected Name=EnvService, got %v", cfg2.Name)
	}
	if cfg2.Port != 9090 {
		t.Errorf("expected Port=9090, got %v", cfg2.Port)
	}
	if cfg2.Debug != true {
		t.Errorf("expected Debug=true, got %v", cfg2.Debug)
	}

	cm.Set("APP_PORT", 100)
	var cfg3 AppConfig
	if err := cm.Unmarshal(&cfg3); err == nil {
		t.Error("expected validation error for APP_PORT < 1000, got nil")
	}
}

type ProfileConfig struct {
	AppName string `json:"APP_NAME"`
	Port    int    `json:"APP_PORT"`
}

func TestLoadWithProfile(t *testing.T) {
	tmpDir := t.TempDir()

	baseConfig := `
APP_NAME: BaseApp
APP_PORT: 8080
`
	basePath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(basePath, []byte(baseConfig), 0644); err != nil {
		t.Fatal(err)
	}

	devConfig := `
APP_PORT: 3000
`
	devPath := filepath.Join(tmpDir, "config-dev.yaml")
	if err := os.WriteFile(devPath, []byte(devConfig), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("APP_ENV", "dev")
	cm := NewConfigManager()
	if err := cm.LoadWithProfile("APP_ENV", basePath); err != nil {
		t.Fatalf("LoadWithProfile failed: %v", err)
	}

	var cfg ProfileConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if cfg.AppName != "BaseApp" {
		t.Errorf("expected AppName=BaseApp, got %v", cfg.AppName)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected Port=3000 (override by dev), got %v", cfg.Port)
	}

	os.Setenv("APP_ENV", "prod") // config-prod.yaml
	cm2 := NewConfigManager()
	if err := cm2.LoadWithProfile("APP_ENV", basePath); err != nil {
		t.Fatalf("LoadWithProfile failed: %v", err)
	}

	var cfg2 ProfileConfig
	if err := cm2.Unmarshal(&cfg2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if cfg2.AppName != "BaseApp" {
		t.Errorf("expected AppName=BaseApp, got %v", cfg2.AppName)
	}
	if cfg2.Port != 8080 {
		t.Errorf("expected Port=8080 (base only), got %v", cfg2.Port)
	}
}

type ProfileTestConfig struct {
	AppName string `json:"APP_NAME"`
	Port    int    `json:"APP_PORT"`
	Debug   bool   `json:"APP_DEBUG"`
}

func TestLoadWithProfile_YAML(t *testing.T) {
	tmpDir := t.TempDir()

	// base config.yaml
	base := `
APP_NAME: BaseApp
APP_PORT: 8080
`
	basePath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(basePath, []byte(base), 0644); err != nil {
		t.Fatal(err)
	}

	// profile config-dev.yaml
	profile := `
APP_PORT: 3000
`
	profilePath := filepath.Join(tmpDir, "config-dev.yaml")
	if err := os.WriteFile(profilePath, []byte(profile), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("APP_ENV", "dev")
	cm := NewConfigManager()
	if err := cm.LoadWithProfile("APP_ENV", basePath); err != nil {
		t.Fatalf("LoadWithProfile failed: %v", err)
	}

	var cfg ProfileTestConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if cfg.AppName != "BaseApp" {
		t.Errorf("expected AppName=BaseApp, got %v", cfg.AppName)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected Port=3000 (override by dev), got %v", cfg.Port)
	}
}

func TestLoadWithProfile_Env(t *testing.T) {
	tmpDir := t.TempDir()

	// base .env
	baseEnv := "APP_DEBUG=false\n"
	basePath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(basePath, []byte(baseEnv), 0644); err != nil {
		t.Fatal(err)
	}

	// profile .env.dev
	profileEnv := "APP_DEBUG=true\n"
	profilePath := filepath.Join(tmpDir, ".env.dev")
	if err := os.WriteFile(profilePath, []byte(profileEnv), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("APP_ENV", "dev")
	cm := NewConfigManager()
	if err := cm.LoadWithProfile("APP_ENV", basePath); err != nil {
		t.Fatalf("LoadWithProfile failed: %v", err)
	}

	var cfg ProfileTestConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if cfg.Debug != true {
		t.Errorf("expected Debug=true from .env.dev, got %v", cfg.Debug)
	}
}

// helper: encrypt YAML with AES-GCM and save as .enc
func encryptFile(t *testing.T, path string, data []byte, secret string) {
	key := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		t.Fatal(err)
	}

	// prepend nonce to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	enc := base64.StdEncoding.EncodeToString(ciphertext)

	if err = os.WriteFile(path, []byte(enc), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadEncryptedFile_GCM(t *testing.T) {
	tmpDir := t.TempDir()

	// sample config data
	configData := map[string]interface{}{
		"APP_NAME": "SecureApp",
		"APP_PORT": 7070,
	}
	raw, err := yaml.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	secret := "super-secret-key"
	encPath := filepath.Join(tmpDir, "config.yaml.enc")

	// encrypt and save
	encryptFile(t, encPath, raw, secret)

	// load encrypted config
	cm := NewConfigManager()
	if err := cm.LoadEncryptedFile(encPath, secret); err != nil {
		t.Fatalf("LoadEncryptedFile failed: %v", err)
	}

	if cm.Get("APP_NAME") != "SecureApp" {
		t.Errorf("expected APP_NAME=SecureApp, got %v", cm.Get("APP_NAME"))
	}
	if cm.Get("APP_PORT") != 7070 {
		t.Errorf("expected APP_PORT=7070, got %v", cm.Get("APP_PORT"))
	}
}
func encryptFileForTest(t *testing.T, path string, data []byte, secret string) {
	t.Helper() // mark this as a test helper

	// derive key (AES-256)
	key := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("failed to generate nonce: %v", err)
	}

	// prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	enc := base64.StdEncoding.EncodeToString(ciphertext)
	if err := os.WriteFile(path, []byte(enc), 0644); err != nil {
		t.Fatalf("failed to write encrypted file: %v", err)
	}
}
func TestEncryptedConfigWrongKey(t *testing.T) {
	tmpDir := t.TempDir()
	data := "APP_NAME: SecureApp\n"
	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte(data), 0644)

	secret := "correct-key"
	encPath := filepath.Join(tmpDir, "config.yaml.enc")

	// Encrypt با کلید درست
	raw, _ := os.ReadFile(basePath)
	encryptFileForTest(t, encPath, raw, secret)

	// حالا با کلید اشتباه تست می‌کنیم
	cm := NewConfigManager()
	err := cm.LoadEncryptedFile(encPath, "wrong-key")
	if err == nil {
		t.Errorf("expected error with wrong key, got nil")
	}
}

func TestNewTestConfig(t *testing.T) {
	cm := NewTestConfig(map[string]string{
		"APP_NAME": "TestApp",
		"APP_PORT": "1234",
	})

	if cm.Get("APP_NAME") != "TestApp" {
		t.Errorf("expected APP_NAME=TestApp, got %v", cm.Get("APP_NAME"))
	}
	if cm.Get("APP_PORT") != 1234 { // ✅ حالا int رو چک می‌کنیم
		t.Errorf("expected APP_PORT=1234, got %v", cm.Get("APP_PORT"))
	}
}

func TestConflict_FileVsEnv(t *testing.T) {
	tmpDir := t.TempDir()

	// base yaml
	yamlData := `
APP_PORT: 8080
`
	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte(yamlData), 0644)

	// .env
	envPath := filepath.Join(tmpDir, ".env")
	_ = os.WriteFile(envPath, []byte("APP_PORT=3000\n"), 0644)

	cm := NewConfigManager()
	_ = cm.LoadFromFile(basePath)
	_ = cm.LoadFromDotEnv(envPath)

	if cm.Get("APP_PORT") != 3000 {
		t.Errorf("expected APP_PORT=3000, got %v", cm.Get("APP_PORT"))
	}
}

func TestConflict_ProfileOverride(t *testing.T) {
	tmpDir := t.TempDir()

	baseData := "DB_HOST: localhost\n"
	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte(baseData), 0644)

	profileData := "DB_HOST: dev.db.local\n"
	profilePath := filepath.Join(tmpDir, "config-dev.yaml")
	_ = os.WriteFile(profilePath, []byte(profileData), 0644)

	os.Setenv("APP_ENV", "dev")
	cm := NewConfigManager()
	_ = cm.LoadWithProfile("APP_ENV", basePath)

	if cm.Get("DB_HOST") != "dev.db.local" {
		t.Errorf("expected DB_HOST=dev.db.local, got %v", cm.Get("DB_HOST"))
	}
}

func TestMissingProfileFile(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte("APP_NAME: BaseApp\n"), 0644)

	os.Setenv("APP_ENV", "staging") // config-staging.yaml doesn't exist
	cm := NewConfigManager()
	err := cm.LoadWithProfile("APP_ENV", basePath)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cm.Get("APP_NAME") != "BaseApp" {
		t.Errorf("expected APP_NAME=BaseApp, got %v", cm.Get("APP_NAME"))
	}
}

func TestSystemEnvOverride(t *testing.T) {
	os.Setenv("APP_NAME", "FromSys")

	cm := NewConfigManager()
	cm.Set("APP_NAME", "FromFile")
	cm.LoadFromSysEnv("APP_NAME")

	if cm.Get("APP_NAME") != "FROMSYS" {
		t.Errorf("expected APP_NAME=FromSys, got %v", cm.Get("APP_NAME"))
	}
}

func TestNormalizeKey(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("app_name", "MyApp")

	if cm.Get("APP_NAME") != "MyApp" {
		t.Errorf("expected APP_NAME=MyApp, got %v", cm.Get("APP_NAME"))
	}
}

func TestTypeConversion(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("APP_PORT", "1234")
	cm.Set("APP_DEBUG", "true")

	if cm.Get("APP_PORT") != 1234 {
		t.Errorf("expected int 1234, got %v", cm.Get("APP_PORT"))
	}
	if cm.Get("APP_DEBUG") != true {
		t.Errorf("expected bool true, got %v", cm.Get("APP_DEBUG"))
	}
}

func TestEncryptedConfigWrongKey2(t *testing.T) {
	// از تست اصلی encryptFile استفاده می‌کنیم (کلید درست)
	tmpDir := t.TempDir()
	data := "APP_NAME: SecureApp\n"
	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte(data), 0644)

	secret := "correct-key"
	encPath := filepath.Join(tmpDir, "config.yaml.enc")

	// Encrypt فایل
	raw, _ := os.ReadFile(basePath)
	encryptFileForTest(t, encPath, raw, secret)

	// با کلید اشتباه لود می‌کنیم
	cm := NewConfigManager()
	err := cm.LoadEncryptedFile(encPath, "wrong-key")
	if err == nil {
		t.Errorf("expected error with wrong key, got nil")
	}
}

func TestEmptyEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	_ = os.WriteFile(envPath, []byte(""), 0644) // empty file

	cm := NewConfigManager()
	err := cm.LoadFromDotEnv(envPath)
	if err != nil {
		t.Errorf("expected no error on empty env file, got %v", err)
	}
}

func TestMultipleLoadOrder(t *testing.T) {
	tmpDir := t.TempDir()

	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte("APP_DEBUG: false\n"), 0644)

	overridePath := filepath.Join(tmpDir, "config.local.yaml")
	_ = os.WriteFile(overridePath, []byte("APP_DEBUG: true\n"), 0644)

	cm := NewConfigManager()
	_ = cm.LoadFromFile(basePath)
	_ = cm.LoadFromFile(overridePath)

	if cm.Get("APP_DEBUG") != true {
		t.Errorf("expected APP_DEBUG=true, got %v", cm.Get("APP_DEBUG"))
	}
}

func TestUnmarshalDefaults(t *testing.T) {
	type AppCfg struct {
		Name string `json:"APP_NAME" default:"MyService"`
		Port int    `json:"APP_PORT" default:"9090"`
	}

	cm := NewConfigManager()
	var cfg AppCfg
	_ = cm.Unmarshal(&cfg)

	if cfg.Name != "MyService" {
		t.Errorf("expected default Name=MyService, got %v", cfg.Name)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected default Port=9090, got %v", cfg.Port)
	}
}
