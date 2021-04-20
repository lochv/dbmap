package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type config struct {
	Workers     int    `yaml:"workers,omitempty"`
	Iface       string `yaml:"iface,omitempty"`
	ReceiveMode string `yaml:"receive_mode,omitempty"`
	ReportMode  string `yaml:"report_mode,omitempty"`
	TestIp      string `yaml:"test_ip,omitempty"`
	SourcePort  int    `yaml:"source_port,omitempty"`
	InputFile   string `yaml:"input_file,omitempty"`
	OutputFile  string `yaml:"output_file,omitempty"`
	Ports       []int  `yaml:"ports,omitempty"`
}

var Conf = config{}

func init() {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic("Need config.yaml")
	}
	err = yaml.Unmarshal(yamlFile, &Conf)
	if err != nil {
		panic(err.Error())
	}
}
