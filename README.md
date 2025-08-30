# go-configmgr

A lightweight and extensible configuration manager for Go projects.  
Supports loading configuration from multiple sources (YAML, JSON, `.env`, system env),  
profile-based overrides (dev/prod/staging), defaults, validation, and exporting.

---

## ✨ Features
- Load configuration from:
    - JSON / YAML files
    - `.env` files
    - System environment variables
- Profile-based overrides (e.g. `config-dev.yaml`, `.env.prod`)
- Default values via struct tags (`default:"value"`)
- Validation via [go-playground/validator](https://github.com/go-playground/validator)
- Normalize keys to uppercase for consistency
- Export config to JSON/YAML
- Testing utilities (`NewTestConfig`)
- Pluggable logging
- Simple CLI (`configctl`) to inspect configs

---

## 🚀 Installation

```bash
go get github.com/Serajian/go-configmgr/configmgr
```
---

## 🛠 Usage

### 1. Load from files and .env

```go
cm := configmgr.NewConfigManager()

// Load with profile support
_ = cm.LoadWithProfile("APP_ENV", "config.yaml")
_ = cm.LoadWithProfile("APP_ENV", ".env")
```
#### If APP_ENV=dev, this loads:

* config.yaml + config-dev.yaml
* .env + .env.dev
---
### 2. Struct mapping with defaults and validation
```go
type AppConfig struct {
    Name  string `json:"APP_NAME" default:"MyService" validate:"required"`
    Port  int    `json:"APP_PORT" default:"8080" validate:"gte=1000,lte=9999"`
    Debug bool   `json:"APP_DEBUG" default:"false"`
}

var cfg AppConfig
if err := cm.Unmarshal(&cfg); err != nil {
    log.Fatalf("validation failed: %v", err)
}

fmt.Printf("%+v\n", cfg)
```
---
### 3. Export config
```go
fmt.Println(string(cm.ToJSON()))
fmt.Println(string(cm.ToYAML()))
```
---
### 4. Testing utility
```go
cm := configmgr.NewTestConfig(map[string]string{
    "APP_NAME": "TestApp",
    "APP_PORT": "1234",
})

var cfg AppConfig
_ = cm.Unmarshal(&cfg)
```
---
### 5. CLI (configctl)
* A simple CLI tool to inspect configs.
```bash
go run ./cmd/configctl -action=show -conf=config.yaml -env=APP_ENV
```
output:
```json
{
  "APP_NAME": "MyApp",
  "APP_PORT": 8080,
  "DB_HOST": "prod.db.server"
}
```
---

### 🔒 Encrypted Configs
Supports loading encrypted configs (.yaml.enc, .json.enc) using AES-GCM.
Encryption key provided via env (e.g. CONFIG_SECRET_KEY).
```go
err := cm.LoadEncryptedFile("config.yaml.enc", os.Getenv("CONFIG_SECRET_KEY"))
```
---









