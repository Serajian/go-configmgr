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

## 🌍 Real-world Examples
## Example A: simple
### Example A:1: Just from ENV

```go
package main

import (
	"fmt"
	"log"

	"github.com/Serajian/go-configmgr/configmgr"
)

func main() {
	cm := configmgr.NewConfigManager()

	if err := cm.LoadFromDotEnv(".env"); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	fmt.Println("DB_HOST =", cm.Get("DB_HOST"))
	fmt.Println("DB_PORT =", cm.Get("DB_PORT"))
}

```
---
### Example A:2: Just from YML

```go
package main

import (
	"fmt"
	"log"

	"github.com/Serajian/go-configmgr/configmgr"
)

type DatabaseConfig struct {
	Host string `json:"DB_HOST"`
	Port int    `json:"DB_PORT"`
}

func main() {
	cm := configmgr.NewConfigManager()

	if err := cm.LoadFromFile("config.yaml"); err != nil {
		log.Fatalf("failed to load config.yaml: %v", err)
	}

	var dbCfg DatabaseConfig
	if err := cm.Unmarshal(&dbCfg); err != nil {
		log.Fatalf("failed to unmarshal db config: %v", err)
	}

	fmt.Printf("Database config: %+v\n", dbCfg)
}

```
---
### Example A:3: mix

```go
package main

import (
	"fmt"
	"log"

	"github.com/Serajian/go-configmgr/configmgr"
)

func main() {
	cm := configmgr.NewConfigManager()

	if err := cm.LoadFromFile("config.yaml"); err != nil {
		log.Fatal(err)
	}

	if err := cm.LoadFromDotEnv(".env"); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final config: %+v\n", cm.GetAll())
}

```
---
### Example A:4: fast test by NewTestConfig

```go
package main

import (
	"fmt"
	"log"

	"github.com/Serajian/go-configmgr/configmgr"
)

type AppConfig struct {
	Name string `json:"APP_NAME" default:"MyService"`
	Port int    `json:"APP_PORT" default:"8080"`
}

func main() {
	cm := configmgr.NewTestConfig(map[string]string{
		"APP_NAME": "TestApp",
		"APP_PORT": "9999",
	})

	var cfg AppConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App config: %+v\n", cfg)
}
```
---
## Example B: case
### Example B:1: Microservice with Gin

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Serajian/go-configmgr/configmgr"
)

type AppConfig struct {
	Name string `json:"APP_NAME" default:"UserService" validate:"required"`
	Port int    `json:"APP_PORT" default:"8080" validate:"gte=1000,lte=9999"`
}

func main() {
	cm := configmgr.NewConfigManager()
	if err := cm.LoadWithProfile("APP_ENV", "config.yaml"); err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	var cfg AppConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": cfg.Name})
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("starting %s on %s", cfg.Name, addr)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}

```
### Example B:2: Worker Service (Kafka Consumer)

```go
package main

import (
	"fmt"
	"log"

	"github.com/Serajian/go-configmgr/configmgr"
)

type WorkerConfig struct {
	BrokerAddr string `json:"BROKER_ADDR" default:"localhost:9092" validate:"required"`
	Topic      string `json:"BROKER_TOPIC" default:"events"`
	GroupID    string `json:"BROKER_GROUP_ID" default:"worker"`
}

func main() {
	cm := configmgr.NewConfigManager()
	if err := cm.LoadWithProfile("APP_ENV", "config.yaml"); err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	var cfg WorkerConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	log.Printf("worker starting... broker=%s topic=%s group=%s",
		cfg.BrokerAddr, cfg.Topic, cfg.GroupID)

	// TODO: connect to Kafka consumer and start processing...
	fmt.Println("worker running...")
}

```
### Example B:3: Secure Configs (Encrypted)

```go
package main

import (
	"log"
	"os"

	"github.com/Serajian/go-configmgr/configmgr"
)

func main() {
	cm := configmgr.NewConfigManager()
	secret := os.Getenv("CONFIG_SECRET_KEY")
	if secret == "" {
		log.Fatal("missing CONFIG_SECRET_KEY")
	}

	if err := cm.LoadEncryptedFile("config.yaml.enc", secret); err != nil {
		log.Fatalf("failed to load encrypted config: %v", err)
	}

	log.Println("secure config loaded:", cm.GetAll())
}

```
---






