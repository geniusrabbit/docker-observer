//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

// DockerContainer declaration
type DockerContainer types.ContainerJSON

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
