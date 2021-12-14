package config

import (
	"fmt"
	"io/ioutil"

	"github.com/sgaunet/chaospg/postgresctl"
	"gopkg.in/yaml.v2"
)

func ReadyamlConfigFile(filename string) (postgresctl.DbConfig, error) {
	var yamlConfig postgresctl.DbConfig

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return yamlConfig, err
	}

	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
		return yamlConfig, err
	}

	return yamlConfig, err
}
