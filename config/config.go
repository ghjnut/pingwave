package config

import (
	//"github.com/ghjnut/pingwave"

	//"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"io/ioutil"
)

// struct for the global config
type Config struct {
	TargetGroups []TargetGroup `hcl:"target_group"`
}

type TargetGroup struct {
	Name     string   `hcl:",key"`
	Prefix   string   `hcl:"prefix"`
	Interval int      `hcl:"interval"`
	Targets  []Target `hcl:"target"`
}

type Target struct {
	Address string `hcl:"address"`
	Name    string `hcl:",key"`
}

// parse the config
func Parse(configpath string) (*Config, error) {
	configfile, err := ioutil.ReadFile(configpath)
	if err != nil {
		return nil, err
	}

	// TODO make this recursive
	hclParseTree, err := hcl.Parse(string(configfile))
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := hcl.DecodeObject(&config, hclParseTree); err != nil {
		return nil, err
	}

	return config, nil
}
