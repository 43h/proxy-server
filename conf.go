package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const confFile = "conf.yaml"

type Config struct {
	Listen string `yaml:"listen"`
}

var ConfigParam Config = Config{""}

func checkConfFile() bool {
	if _, err := os.Stat(confFile); os.IsNotExist(err) {
		fmt.Println("conf.yaml does not exist")
		return false
	} else {
		fmt.Println("conf.yaml exists")
		return true
	}
}

func loadConf() bool {
	data, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		fmt.Println(err)
		return false
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		fmt.Println(config)
		ConfigParam = config
	}

	return true
}

func initConf() bool {
	if checkConfFile() == false {
		return false
	}

	if loadConf() == false {
		return false
	}
	return true
}
