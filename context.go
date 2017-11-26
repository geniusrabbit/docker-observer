//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import (
	"encoding/json"
	"path/filepath"
)

type docker struct {
	Host    string `json:"host" yaml:"host" env:"DOCKER_HOST"`
	Version string `json:"version" yaml:"version" env:"DOCKER_API_VERSION"`
}

type route struct {
	NamePattern string   `json:"name_pattern" yaml:"name_pattern"`
	Scope       string   `json:"scope" yaml:"scope"`
	Actions     []string `json:"actions" yaml:"actions"`
	Each        bool     `json:"each" yaml:"each"` // Process container one by one
	Condition   string   `json:"condition" yaml:"condition"`
	Cmd         string   `json:"cmd,omitempty" yaml:"cmd"`
	Source      string   `json:"source,omitempty" yaml:"source"`
	Target      string   `json:"target,omitempty" yaml:"target"`
}

type config struct {
	Debug    bool    `cli:"debug" env:"DEBUG"`
	Filepath string  `cli:"config" default:"main.conf"`
	BaseDir  string  `cli:"basedir"`
	Docker   docker  `json:"docker" yaml:"docker"`
	Routes   []route `json:"routes" yaml:"routes"`
}

// String method of Stringer interface
func (c *config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func (c *config) Load() error {
	return Load(c)
}

// Path for dir
func (c *config) Path(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.BaseDir, path)
}

// ConfigFile path
func (c *config) ConfigFile() string {
	return c.Path(c.Filepath)
}

var Config config
