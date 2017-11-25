//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravitational/configure"
	"github.com/hashicorp/hcl"
	defaults "github.com/mcuadros/go-defaults"
)

type configInterface interface {
	ConfigFile() string
}

// LoadFile from config
func LoadFile(cfg configInterface, file string) error {
	f, err := os.Open(file)
	if nil != err {
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".json":
		return json.NewDecoder(bytes.NewBuffer(data)).Decode(cfg)
	case ".yml", ".yaml":
		return configure.ParseYAML(data, cfg)
	case ".hcl":
		return hcl.Unmarshal(data, cfg)
	}
	return fmt.Errorf("Unsupported config ext: %s", ext)
}

// LoadCommandLine by args
func LoadCommandLine(cfg configInterface, commands []string) error {
	return configure.ParseCommandLine(cfg, commands)
}

// Defaults set if empty
func Defaults(cfg configInterface) error {
	defaults.SetDefaults(cfg)
	return nil
}

// Load data from file
func Load(cfg configInterface) (err error) {
	// Set defaults for config
	Defaults(cfg)

	// parse command line arguments
	if err = LoadCommandLine(cfg, os.Args[1:]); nil != err {
		return
	}

	// parse YAML
	if filepath := cfg.ConfigFile(); len(filepath) > 0 {
		if _, err = os.Stat(filepath); !os.IsNotExist(err) {
			if err = LoadFile(cfg, filepath); nil != err {
				return
			}
		}
	}

	// parse environment variables
	if err = configure.ParseEnv(cfg); nil != err {
		return
	}

	if err = LoadCommandLine(cfg, os.Args[1:]); nil != err { // parse command line arguments
		return
	}

	return
}
