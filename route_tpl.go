//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

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
)

// TplRouter errors
var (
	ErrTplRouteInvalidSource = errors.New("Invalid template source")
	ErrTplRouteInvalidTarget = errors.New("Invalid template target")
)

// TplRoute processor
type TplRoute struct {
	Each         bool   `json:"each"` // Do process containers one by one
	Condition    string `json:"condition"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	Filter       filter `json:"filter"`
	conditionTpl *template.Template
	sourceTpl    *template.Template
	targetTpl    *template.Template
}

// Exec route object
func (r *TplRoute) Exec(ctx context.Context, msg *ExecuteMessage) (err error) {
	var (
		b       bool
		data    []byte
		dataCtx = map[string]interface{}{
			"message":       msg,
			"action":        msg.Action,
			"items":         msg.ListBase(),
			"allitems":      msg.AllListBase(),
			"containers":    msg.Containers,
			"allcontainers": msg.AllContainers,
			"services":      msg.Services,
			"allservices":   msg.AllServices,
			"config":        &Config,
		}
	)

	if err = r.Filter.do(msg); err != nil {
		return
	}

	if r.Each {
		for _, it := range msg.ListBase() {
			switch it.(type) {
			case DockerContainer:
				dataCtx["container"] = it
			case SwarmService:
				dataCtx["service"] = it
			}
			if b, err = r.condition(dataCtx); b {
				if data, err = r.prepareTmp(dataCtx); err == nil {
					err = ioutil.WriteFile(r.target(dataCtx), data, 0666)
				}
			}
			if err != nil {
				break
			}
		}
	} else if b, err = r.condition(dataCtx); b {
		if data, err = r.prepareTmp(dataCtx); err == nil {
			err = ioutil.WriteFile(r.target(dataCtx), data, 0666)
		}
	}
	return
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

	if fldata, err = ioutil.ReadFile(r.source(data)); err != nil {
		return nil, err
	}

	if tpl, err = tpl.Parse(string(fldata)); err != nil {
		return
	}

	err = tpl.Execute(&buf, data)
	return buf.Bytes(), err
}

func (r *TplRoute) source(ctx interface{}) string {
	if r.sourceTpl == nil {
		r.sourceTpl, _ = template.New("src").Funcs(tplFuncs).Parse(r.Source)
	}
	if r.sourceTpl != nil {
		var buf bytes.Buffer
		r.sourceTpl.Execute(&buf, ctx)
		return buf.String()
	}
	return r.Source
}

func (r *TplRoute) target(ctx interface{}) string {
	if r.targetTpl == nil {
		r.targetTpl, _ = template.New("trg").Funcs(tplFuncs).Parse(r.Target)
	}
	if r.targetTpl != nil {
		var buf bytes.Buffer
		r.targetTpl.Execute(&buf, ctx)
		return buf.String()
	}
	return r.Target
}

func (r *TplRoute) condition(ctx interface{}) (_ bool, err error) {
	if r.Condition == "" {
		return true, nil
	}

	if r.conditionTpl == nil {
		r.conditionTpl, err = template.New("cond").Funcs(tplFuncs).Parse(r.Condition)
	}

	if err != nil {
		return
	}

	return doTemplateCondition(r.conditionTpl, ctx)
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
				if s, ok := v.(string); !ok || s != "" {
					return v
				}
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
	"is_service": func(it interface{}) bool {
		switch it.(type) {
		case SwarmService, *SwarmService:
			return true
		}
		return false
	},
	"split": func(item interface{}, sep string) []string {
		if s, _ := item.(string); len(s) > 0 {
			return strings.Split(s, sep)
		}
		return nil
	},
	"join": func(list []string, sep string) string {
		return strings.Join(list, sep)
	},
}
