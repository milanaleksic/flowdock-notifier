package main

import (
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

const (
	configurationName = "config.toml"
)

type config struct {
	FlowdockToken string
	MyUsername    string
}

func readConfig() config {
	var conf config
	if _, err := toml.DecodeFile(path.Join(path.Dir(os.Args[0]), configurationName), &conf); err != nil {
		if _, err := toml.DecodeFile(configurationName, &conf); err != nil {
			log.Fatalf("Failure while parsing configuration file %s: %v",
				path.Join(path.Base(os.Args[0]), configurationName), err)
		}
	}
	return conf
}
