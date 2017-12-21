//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"text/template"
)

var (
	errInvalidConditionTemplate = errors.New("Invalid condition template")
)

type filter struct {
	Service      string `json:"service"`
	Container    string `json:"container"`
	serviceTpl   *template.Template
	containerTpl *template.Template
}

func (f filter) do(msg *ExecuteMessage) error {
	if f.serviceTpl == nil && strings.TrimSpace(f.Service) != "" {
		f.serviceTpl, _ = template.New("cond").Funcs(tplFuncs).Parse(f.Service)
	}

	if f.containerTpl == nil && strings.TrimSpace(f.Container) != "" {
		f.containerTpl, _ = template.New("cond").Funcs(tplFuncs).Parse(f.Container)
	}

	var (
		services    []SwarmService
		ccontainers []DockerContainer
	)

	if f.serviceTpl != nil {
		for _, srv := range msg.Services {
			if b, err := doTemplateCondition(f.serviceTpl, srv); b && err == nil {
				services = append(services, srv)
			} else if err != nil {
				return err
			}
		}
		msg.Services = services
	}

	if f.containerTpl != nil {
		for _, con := range msg.Containers {
			if b, err := doTemplateCondition(f.containerTpl, con); b && err == nil {
				ccontainers = append(ccontainers, con)
			} else if err != nil {
				return err
			}
		}
		msg.Containers = ccontainers
	}

	return nil
}

func doTemplateCondition(tpl *template.Template, ctx interface{}) (b bool, err error) {
	if tpl == nil {
		return false, errInvalidConditionTemplate
	}

	var (
		buf bytes.Buffer
		v   int64
	)

	if err = tpl.Execute(&buf, ctx); err != nil {
		return false, err
	}

	b, err = strconv.ParseBool(buf.String())
	if !b && err != nil {
		v, err = strconv.ParseInt(buf.String(), 10, 64)
		b = v > 0
	}
	return
}
