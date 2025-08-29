package main

import (
	"fmt"
	"log"

	"github.com/Serajian/go-configmgr/configmgr"
)

// AppConfig example
type AppConfig struct {
	Name  string `json:"APP_NAME"`
	Port  int    `json:"APP_PORT"`
	Debug bool   `json:"APP_DEBUG"`
}

// DatabaseConfig example
type DatabaseConfig struct {
	Host string `json:"DB_HOST"`
	Port int    `json:"DB_PORT" validate:"max=10000,min=1"`
	User string `json:"DB_USER" validate:"required" default:"root"`
	Pass string `json:"DB_PASS"`
	Name string `json:"DB_NAME" default:"postgres" validate:"required"`
}

func main() {
	fmt.Println("Hello World")

	cm := configmgr.NewConfigManager()

	// load .env
	if err := cm.LoadFromDotEnv(".env"); err != nil {
		log.Fatal(err)
	}
	if err := cm.LoadFromFile("config.yaml"); err != nil {
		log.Fatal(err)
	}

	// parse AppConfig
	var appCfg AppConfig
	if err := cm.Unmarshal(&appCfg); err != nil {
		log.Fatal(err)
	}

	// parse DatabaseConfig
	var dbCfg DatabaseConfig
	if err := cm.Unmarshal(&dbCfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App Config: %+v\n", appCfg)
	//fmt.Printf("%+v\n", cm.GetAll())
	fmt.Printf("Database Config: %+v\n", dbCfg)
}
