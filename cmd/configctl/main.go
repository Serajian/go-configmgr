package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Serajian/go-configmgr/configmgr"
)

func main() {
	action := flag.String("action", "show", "action: show | validate")
	envKey := flag.String("env", "APP_ENV", "profile environment key")
	baseConf := flag.String("conf", "config.yaml", "base config file (yaml/json/.env)")
	flag.Parse()

	cm := configmgr.NewConfigManager()

	if err := cm.LoadWithProfile(*envKey, *baseConf); err != nil {
		log.Fatal(err)
	}

	switch *action {
	case "show":
		data, _ := cm.ToJSON()
		fmt.Println(string(data))

	default:
		log.Fatalf("unknown action: %s", *action)
	}
}
