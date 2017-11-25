//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"reflect"
	"strings"
	"text/template"

	"github.com/demdxx/gocast"
	"github.com/docker/docker/api/types"
)

// TplRouter errors
var (
	ErrTplRouteInvalidSource = errors.New("Invalid template source")
	ErrTplRouteInvalidTarget = errors.New("Invalid template target")
)

// TplRoute processor
type TplRoute struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Exec route object
func (r *TplRoute) Exec(ctx context.Context, action string, containers, allContainers []types.ContainerJSON) error {
	data, err := r.prepareTmp(map[string]interface{}{
		"action":        action,
		"containers":    containers,
		"allcontainers": allContainers,
	})

	if err == nil {
		err = ioutil.WriteFile(r.Target, data, 0666)
	}
	return err
}

// Validate route data
func (r *TplRoute) Validate() error {
	if r.Source == "" {
		return ErrTplRouteInvalidSource
	}
	if r.Target == "" {
		return ErrTplRouteInvalidTarget
	}
	return nil
}

func (r *TplRoute) prepareTmp(data interface{}) (_ []byte, err error) {
	var (
		tpl    = template.New("tpl").Funcs(tplFuncs)
		fldata []byte
		buf    bytes.Buffer
	)

	if fldata, err = ioutil.ReadFile(r.Source); err != nil {
		return nil, err
	}

	if tpl, err = tpl.Parse(string(fldata)); err != nil {
		return
	}

	err = tpl.Execute(&buf, data)
	return buf.Bytes(), err
}

var tplFuncs = template.FuncMap{
	"json": func(v interface{}) string {
		data, err := json.Marshal(v)
		if err != nil {
			return `{"error":"` + strings.Replace(err.Error(), `"`, `\"`, -1) + `"}`
		}
		return string(data)
	},
	"jsonbeauty": func(v interface{}) string {
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return `{"error":"` + strings.Replace(err.Error(), `"`, `\"`, -1) + `"}`
		}
		return string(data)
	},
	"indexor": func(data interface{}, key, def string) (rs string) {
		if mp, _ := gocast.ToStringMap(data, "json", false); mp != nil {
			if rs, _ = mp[key]; rs != "" {
				return rs
			}
		}
		return def
	},
	"first": func(input interface{}) interface{} {
		if input == nil {
			return nil
		}
		arr := reflect.ValueOf(input)
		if arr.Len() == 0 {
			return nil
		}
		return arr.Index(0).Interface()
	},
	"coalesce": func(input ...interface{}) interface{} {
		for _, v := range input {
			if v != nil {
				return v
			}
		}
		return nil
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}

		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
}
