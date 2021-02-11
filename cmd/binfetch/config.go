package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	S3Region string `toml:"S3Region"`
	S3Bucket string `toml:"S3Bucket"`
}

func mustParseConfig() *Config {

	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting homedir: %+v\n", err)
		os.Exit(1)
	}
	configPath := fmt.Sprintf("%s/.binfetch.toml", homedir)

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing toml config: %+v\n", err)
		os.Exit(1)
	}

	return &config
}
