//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import (
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

// DockerContainer declaration
type DockerContainer types.ContainerJSON

// Label value by key
func (dc DockerContainer) Label(key string) (v string) {
	if dc.Config != nil {
		v, _ = dc.Config.Labels[key]
	}
	return
}

// LabelEq for value
func (dc DockerContainer) LabelEq(key, vl string) bool {
	return dc.Label(key) == vl
}

// LabelHasTag in the list
func (dc DockerContainer) LabelHasTag(key, vl string) bool {
	v := dc.Label(key)
	for _, t := range strings.Split(v, ",") {
		if t == vl {
			return true
		}
	}
	return false
}

// FirstPort for net
func (dc DockerContainer) FirstPort() string {
	for _, binds := range dc.NetworkSettings.Ports {
		if len(binds) > 0 {
			return binds[0].HostPort
		}
	}
	return ""
}

// SwarmService info
type SwarmService struct {
	Service swarm.Service `json:"service"`
	Tasks   []swarm.Task  `json:"tasks,omitempty"`
}

// Count of tasks
func (ss SwarmService) Count() int {
	return len(ss.Tasks)
}

// LiveCount tasks
func (ss SwarmService) LiveCount() (count int) {
	for _, t := range ss.Tasks {
		if t.Status.State == swarm.TaskStateRunning {
			count++
		}
	}
	return
}

// Label by key
func (ss SwarmService) Label(key string) (v string) {
	v, _ = ss.Service.Spec.Labels[key]
	return
}

// LabelEq for value
func (ss SwarmService) LabelEq(key, vl string) bool {
	return ss.Label(key) == vl
}

// LabelHasTag in the list
func (ss SwarmService) LabelHasTag(key, vl string) bool {
	v := ss.Label(key)
	for _, t := range strings.Split(v, ",") {
		if t == vl {
			return true
		}
	}
	return false
}
