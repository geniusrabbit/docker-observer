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
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/template"

	log "github.com/sirupsen/logrus"
)

// CmdRoute errors
var (
	ErrCmdRouteInvalidCommand = errors.New("Invalid CMD value")
)

// CmdRoute processor
type CmdRoute struct {
	Each         bool   `json:"each"` // Do process containers one by one
	Condition    string `json:"condition"`
	Daemon       bool   `json:"daemon"`
	Cmd          string `json:"cmd"`
	Filter       filter `json:"filter"`
	conditionTpl *template.Template
	tpl          *template.Template
}

// Exec route object
func (r *CmdRoute) Exec(ctx context.Context, msg *ExecuteMessage) (err error) {
	var (
		b       bool
		buf     *bytes.Buffer
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
				if buf, err = toJSON(dataCtx); err != nil {
					break
				}
				if data, err = r.exeCmd(ctx, buf, r.prepareCmd(dataCtx)); len(data) > 0 {
					fmt.Println(string(data))
				}
			}
			if err != nil {
				break
			}
		}
	} else if b, err = r.condition(dataCtx); b {
		buf, err = toJSON(dataCtx)
		if data, err = r.exeCmd(ctx, buf, r.prepareCmd(dataCtx)); len(data) > 0 {
			fmt.Println(string(data))
		}
	}
	return
}

// Validate route data
func (r *CmdRoute) Validate() error {
	if r.Cmd == "" {
		return ErrCmdRouteInvalidCommand
	}
	return nil
}

func (r *CmdRoute) exeCmd(ctx context.Context, data io.Reader, cmd string) (out []byte, err error) {
	log.Debug("> Exec: " + cmd)

	_cmd := exec.CommandContext(ctx, "bash", "-c", cmd)
	if data != nil {
		_cmd.Stdin = data
	}

	if r.Daemon {
		go func() {
			_cmd.Stderr = os.Stderr
			if err = _cmd.Run(); err != nil {
				log.WithFields(log.Fields{"type": "background", "cmd": cmd}).Error(err)
			}
		}()
		return
	}

	return _cmd.CombinedOutput()
}

func (r *CmdRoute) prepareCmd(data interface{}) string {
	if r.tpl == nil {
		var err error
		if r.tpl, err = template.New("cmd").Funcs(tplFuncs).Parse(r.Cmd); r.tpl == nil {
			return "<error:" + err.Error() + ">"
		}
	}

	var buf bytes.Buffer
	r.tpl.Execute(&buf, data)
	return buf.String()
}

func (r *CmdRoute) condition(ctx interface{}) (_ bool, err error) {
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

func toJSON(data interface{}) (buf *bytes.Buffer, err error) {
	buf = new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(data)
	return
}
