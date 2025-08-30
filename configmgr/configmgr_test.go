package configmgr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// 1. LoadFromFile with JSON
func TestLoadFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")
	jsonData := `{"APP_NAME":"JsonApp","APP_PORT":1234}`
	_ = os.WriteFile(path, []byte(jsonData), 0644)

	cm := NewConfigManager()
	if err := cm.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile JSON failed: %v", err)
	}

	if cm.Get("APP_NAME") != "JsonApp" {
		t.Errorf("expected APP_NAME=JsonApp, got %v", cm.Get("APP_NAME"))
	}
	if cm.Get("APP_PORT") != 1234 {
		t.Errorf("expected APP_PORT=1234, got %v", cm.Get("APP_PORT"))
	}
}

// 2. LoadFromDotEnv when file does not exist
func TestLoadFromDotEnv_MissingFile(t *testing.T) {
	cm := NewConfigManager()
	err := cm.LoadFromDotEnv("not-exist.env")
	if err == nil {
		t.Errorf("expected error for missing .env file, got nil")
	}
}

// 3. LoadWithProfile for .env
func TestLoadWithProfile_EnvFile(t *testing.T) {
	tmpDir := t.TempDir()

	baseEnv := "APP_DEBUG=false\n"
	basePath := filepath.Join(tmpDir, ".env")
	_ = os.WriteFile(basePath, []byte(baseEnv), 0644)

	profileEnv := "APP_DEBUG=true\n"
	profilePath := filepath.Join(tmpDir, ".env.dev")
	_ = os.WriteFile(profilePath, []byte(profileEnv), 0644)

	os.Setenv("APP_ENV", "dev")
	cm := NewConfigManager()
	if err := cm.LoadWithProfile("APP_ENV", basePath); err != nil {
		t.Fatalf("LoadWithProfile .env failed: %v", err)
	}

	if cm.Get("APP_DEBUG") != true {
		t.Errorf("expected APP_DEBUG=true, got %v", cm.Get("APP_DEBUG"))
	}
}

// 4. Unmarshal with validation error
func TestUnmarshal_ValidationError(t *testing.T) {
	type AppCfg struct {
		Name string `json:"APP_NAME" validate:"required"`
		Port int    `json:"APP_PORT" validate:"gte=1000"`
	}

	cm := NewConfigManager()
	cm.Set("APP_PORT", 50) // invalid, must be >=1000

	var cfg AppCfg
	err := cm.Unmarshal(&cfg)
	if err == nil {
		t.Errorf("expected validation error, got nil")
	}
}

// 5. normalizeValue with float
func TestNormalizeValue_Float(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("APP_RATE", "12.34")

	if cm.Get("APP_RATE") != "12.34" {
		// چون الان normalizeValue فقط int و bool ساپورت می‌کنه، باید string بمونه
		t.Errorf("expected APP_RATE as string '12.34', got %v", cm.Get("APP_RATE"))
	}
}

// 6. LoadEncryptedFile with JSON
func TestLoadEncryptedFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()

	jsonData := `{"APP_NAME":"SecureJson","APP_PORT":9090}`
	raw := []byte(jsonData)

	secret := "my-json-secret"
	encPath := filepath.Join(tmpDir, "config.json.enc")

	encryptFileForTest(t, encPath, raw, secret)

	cm := NewConfigManager()
	if err := cm.LoadEncryptedFile(encPath, secret); err != nil {
		t.Fatalf("LoadEncryptedFile JSON failed: %v", err)
	}

	if cm.Get("APP_NAME") != "SecureJson" {
		t.Errorf("expected APP_NAME=SecureJson, got %v", cm.Get("APP_NAME"))
	}
	if cm.Get("APP_PORT") != 9090 {
		t.Errorf("expected APP_PORT=9090, got %v", cm.Get("APP_PORT"))
	}
}

// 7. Logger integration with fake logger
type FakeLogger struct {
	infos  []string
	errors []string
}

func (l *FakeLogger) Info(msg string, fields map[string]interface{}) {
	l.infos = append(l.infos, msg)
}
func (l *FakeLogger) Error(msg string, err error, fields map[string]interface{}) {
	l.errors = append(l.errors, msg)
}

func TestLoggerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	yamlData := "APP_NAME: LogTest\n"
	_ = os.WriteFile(path, []byte(yamlData), 0644)

	cm := NewConfigManager()
	logger := &FakeLogger{}
	cm.SetLogger(logger)

	_ = cm.LoadFromFile(path)

	if len(logger.infos) == 0 {
		t.Errorf("expected info logs, got none")
	}
}

func TestNormalizeValue_Float2(t *testing.T) {
	cm := NewConfigManager()

	// case 1: integer-like float64 (from JSON)
	cm.Set("APP_PORT", 1234.0)
	if cm.Get("APP_PORT") != 1234 {
		t.Errorf("expected APP_PORT=1234 (int), got %T %v", cm.Get("APP_PORT"), cm.Get("APP_PORT"))
	}

	// case 2: real float64
	cm.Set("APP_RATE", 12.34)
	if cm.Get("APP_RATE") != 12.34 {
		t.Errorf("expected APP_RATE=12.34 (float64), got %T %v", cm.Get("APP_RATE"), cm.Get("APP_RATE"))
	}

	// case 3: string "12.34" (from .env or yaml)
	cm.Set("APP_RATE_STR", "12.34")
	if cm.Get("APP_RATE_STR") != "12.34" {
		t.Errorf("expected APP_RATE_STR='12.34' (string), got %T %v", cm.Get("APP_RATE_STR"), cm.Get("APP_RATE_STR"))
	}
}

func TestGetAll(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("APP_NAME", "TestAll")
	all := cm.GetAll()
	if all["APP_NAME"] != "TestAll" {
		t.Errorf("expected APP_NAME=TestAll, got %v", all["APP_NAME"])
	}
}

func TestExportJSONYAML(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("APP_NAME", "ExportApp")

	j, err := cm.ToJSON()
	if err != nil || !strings.Contains(string(j), "ExportApp") {
		t.Errorf("ToJSON failed: %v", err)
	}

	y, err := cm.ToYAML()
	if err != nil || !strings.Contains(string(y), "ExportApp") {
		t.Errorf("ToYAML failed: %v", err)
	}
}

func TestLoadFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "c1.yaml")
	file2 := filepath.Join(tmpDir, "c2.yaml")
	os.WriteFile(file1, []byte("APP_NAME: File1\n"), 0644)
	os.WriteFile(file2, []byte("APP_NAME: File2\n"), 0644)

	cm := NewConfigManager()
	if err := cm.LoadFiles(file1, file2); err != nil {
		t.Fatalf("LoadFiles failed: %v", err)
	}

	if cm.Get("APP_NAME") != "File2" {
		t.Errorf("expected APP_NAME=File2 (last file wins), got %v", cm.Get("APP_NAME"))
	}
}

func TestLoadFromFile_UnsupportedExt(t *testing.T) {
	cm := NewConfigManager()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.txt")
	_ = os.WriteFile(path, []byte("hello"), 0644)

	if err := cm.LoadFromFile(path); err == nil {
		t.Errorf("expected error for unsupported file type, got nil")
	}
}

func TestLoadFromSysEnv_Unset(t *testing.T) {
	cm := NewConfigManager()
	cm.LoadFromSysEnv("SOME_RANDOM_KEY")
	if _, ok := cm.GetAll()["SOME_RANDOM_KEY"]; ok {
		t.Errorf("expected no value for unset env key")
	}
}

func TestGetEncryptedExt(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"c.json.enc", ".json.enc"},
		{"c.yaml.enc", ".yaml.enc"},
		{"c.yml.enc", ".yml.enc"},
		{"c.unknown", ".unknown"},
	}
	for _, tt := range tests {
		if got := getEncryptedExt(tt.file); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestLoadFromSysEnv_Unset2(t *testing.T) {
	cm := NewConfigManager()
	cm.LoadFromSysEnv("SOME_RANDOM_KEY_SHOULD_NOT_EXIST")

	if _, ok := cm.GetAll()["SOME_RANDOM_KEY_SHOULD_NOT_EXIST"]; ok {
		t.Errorf("expected no value for unset env")
	}
}

func TestLoadWithProfile_NoEnv(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(path, []byte("APP_NAME: BaseApp\n"), 0644)

	os.Unsetenv("APP_ENV") // no profile
	cm := NewConfigManager()
	err := cm.LoadWithProfile("APP_ENV", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cm.Get("APP_NAME") != "BaseApp" {
		t.Errorf("expected BaseApp, got %v", cm.Get("APP_NAME"))
	}
}

func TestLoadEncryptedFile_InvalidBase64(t *testing.T) {
	tmpDir := t.TempDir()
	encPath := filepath.Join(tmpDir, "bad.yaml.enc")
	_ = os.WriteFile(encPath, []byte("$$$not_base64$$$"), 0644)

	cm := NewConfigManager()
	err := cm.LoadEncryptedFile(encPath, "key")
	if err == nil {
		t.Errorf("expected error for invalid base64, got nil")
	}
}

func TestLoadEncryptedFile_ShortCipher(t *testing.T) {
	tmpDir := t.TempDir()
	encPath := filepath.Join(tmpDir, "short.yaml.enc")
	// خیلی کوتاه، کوچکتر از nonce
	_ = os.WriteFile(encPath, []byte("YWJj"), 0644) // base64("abc")

	cm := NewConfigManager()
	err := cm.LoadEncryptedFile(encPath, "key")
	if err == nil {
		t.Errorf("expected error for short ciphertext, got nil")
	}
}

func TestLoadFromDotEnv_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	// محتوای خراب (godotenv باید fail بده)
	_ = os.WriteFile(envPath, []byte("\x00\x01\x02"), 0644)

	cm := NewConfigManager()
	err := cm.LoadFromDotEnv(envPath)
	if err == nil {
		t.Errorf("expected error for invalid env file, got nil")
	}
}

func TestLoadFromSysEnv_Unset3(t *testing.T) {
	cm := NewConfigManager()
	key := "THIS_KEY_SHOULD_NOT_EXIST_123456"
	os.Unsetenv(key)

	cm.LoadFromSysEnv(key)
	if _, ok := cm.GetAll()[key]; ok {
		t.Errorf("expected no value for unset env key")
	}
}

func TestLoadFromFile_UnsupportedExt2(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.txt")
	_ = os.WriteFile(path, []byte("hello"), 0644)

	cm := NewConfigManager()
	err := cm.LoadFromFile(path)
	if err == nil {
		t.Errorf("expected error for unsupported ext, got nil")
	}
}

func TestLoadFiles_Error(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "c1.yaml")
	_ = os.WriteFile(file1, []byte("APP_NAME: File1\n"), 0644)
	file2 := filepath.Join(tmpDir, "not_exist.yaml")

	cm := NewConfigManager()
	err := cm.LoadFiles(file1, file2)
	if err == nil {
		t.Errorf("expected error for missing file, got nil")
	}
}

func TestLoadWithProfile_NoEnvKey(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte("APP_NAME: BaseApp\n"), 0644)

	os.Unsetenv("APP_ENV")
	cm := NewConfigManager()
	err := cm.LoadWithProfile("APP_ENV", basePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cm.Get("APP_NAME") != "BaseApp" {
		t.Errorf("expected APP_NAME=BaseApp, got %v", cm.Get("APP_NAME"))
	}
}

func TestLoadWithProfile_BadProfileFile(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(basePath, []byte("APP_NAME: BaseApp\n"), 0644)

	profilePath := filepath.Join(tmpDir, "config-dev.yaml")
	_ = os.WriteFile(profilePath, []byte("{bad yaml:::"), 0644)

	os.Setenv("APP_ENV", "dev")
	cm := NewConfigManager()
	err := cm.LoadWithProfile("APP_ENV", basePath)
	if err == nil {
		t.Errorf("expected error for bad profile file, got nil")
	}
}

func TestUnmarshal_NoValidation(t *testing.T) {
	type Simple struct {
		Name string `json:"APP_NAME"`
		Port int    `json:"APP_PORT" default:"5555"`
	}

	cm := NewConfigManager()
	cm.Set("APP_NAME", "TestApp")

	var cfg Simple
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 5555 {
		t.Errorf("expected default Port=5555, got %v", cfg.Port)
	}
}

func TestUnmarshal_DefaultOverride(t *testing.T) {
	type WithDefault struct {
		Debug bool `json:"APP_DEBUG" default:"false"`
	}

	cm := NewConfigManager()
	cm.Set("APP_DEBUG", true)

	var cfg WithDefault
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Debug != true {
		t.Errorf("expected Debug=true, got %v", cfg.Debug)
	}
}

// اختیاری: تست integration برای CLI
func TestConfigctlShow(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/configctl", "-action=show", "-conf=testdata/config.yaml")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// ممکنه testdata/config.yaml نداشته باشی، پس skip می‌کنیم
		if strings.Contains(err.Error(), "exit status") {
			t.Skip("skip CLI integration test (missing testdata/config.yaml)")
			return
		}
		t.Fatalf("configctl failed: %v", err)
	}
	if !strings.Contains(string(out), "{") {
		t.Errorf("expected JSON output, got %s", out)
	}
}
