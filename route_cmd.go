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
	"fmt"
	"io"
	"os/exec"
	"text/template"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

// CmdRoute errors
var (
	ErrCmdRouteInvalidCommand = errors.New("Invalid CMD value")
)

// CmdRoute processor
type CmdRoute struct {
	Cmd string `json:"cmd"`
	tpl *template.Template
}

// Exec route object
func (r *CmdRoute) Exec(ctx context.Context, action string, containers, allContainers []types.ContainerJSON) error {
	var (
		data []byte
		info = map[string]interface{}{
			"action":        action,
			"containers":    containers,
			"allcontainers": allContainers,
			"config":        &Config,
		}
		cmd      = r.prepareCmd(info)
		buf, err = toJSON(info)
	)

	if err == nil {
		if data, err = r.exeCmd(ctx, buf, cmd); len(data) > 0 {
			fmt.Println(string(data))
		}
	}
	return err
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

	return _cmd.CombinedOutput()
}

func (r *CmdRoute) prepareCmd(data interface{}) string {
	if r.tpl == nil {
		var err error
		r.tpl, err = template.New("x").Parse(r.Cmd)
		fmt.Println(err)
	}

	var buf bytes.Buffer
	r.tpl.Execute(&buf, data)
	return buf.String()
}

func toJSON(data interface{}) (buf *bytes.Buffer, err error) {
	buf = new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(data)
	return
}
