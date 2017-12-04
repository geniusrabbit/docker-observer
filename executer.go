//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type dockerExecuter struct {
	kill    chan bool
	timeout time.Duration
	router  Router
}

// ExecuteMessage contains all processing data
type ExecuteMessage struct {
	Action        string            `json:"action"`
	Scope         string            `json:"scope"`
	Services      []SwarmService    `json:"services"`
	AllServices   []SwarmService    `json:"all_services"`
	Containers    []DockerContainer `json:"containers"`
	AllContainers []DockerContainer `json:"all_containers"`
	route         Route
}

// ListBase of affected items
func (msg ExecuteMessage) ListBase() (list []interface{}) {
	for _, it := range msg.Containers {
		list = append(list, it)
	}
	for _, it := range msg.Services {
		list = append(list, it)
	}
	return
}

// AllListBase of all items
func (msg ExecuteMessage) AllListBase() (list []interface{}) {
	for _, it := range msg.AllContainers {
		list = append(list, it)
	}
	for _, it := range msg.AllServices {
		list = append(list, it)
	}
	return
}

// NewExecuter object
func NewExecuter(router Router) ContainerEventer {
	return &dockerExecuter{
		kill:    make(chan bool, 1),
		timeout: 10 * time.Second,
		router:  router,
	}
}

// Event processor
func (e *dockerExecuter) Event(msg *ExecuteMessage) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), e.timeout)
		eContainers []*ExecuteMessage
	)

	if len(msg.Containers) < 1 {
		msg.Containers = msg.AllContainers
	}
	if len(msg.Services) < 1 {
		msg.Services = msg.AllServices
	}

	// Collect containers by routes
	for _, it := range msg.ListBase() {
		if routes := e.router.Routes(msg.Action, msg.Scope, it); len(routes) > 0 {
			for _, route := range routes {
				var tpr *ExecuteMessage
				for _, pr := range eContainers {
					if pr.route == route {
						tpr = pr
					}
				}

				if tpr == nil {
					tpr = &ExecuteMessage{Action: msg.Action, Scope: msg.Scope, route: route}
					eContainers = append(eContainers, tpr)
				}

				switch v := it.(type) {
				case DockerContainer:
					tpr.Containers = append(tpr.Containers, v)
				case SwarmService:
					tpr.Services = append(tpr.Services, v)
				}

				if len(msg.AllContainers) > 0 {
					tpr.AllContainers = msg.AllContainers
				}
				if len(msg.AllServices) > 0 {
					tpr.AllServices = msg.AllServices
				}
			} // end for
		}
	}

	go func() {
		select {
		case <-ctx.Done():
		case <-e.kill:
			cancel()
		}
	}()

	for _, pr := range eContainers {
		if err := pr.route.Exec(ctx, pr); err != nil {
			e.Error(err)
		}
	}
}

func (e *dockerExecuter) Error(err error) {
	log.Errorf("Event: %v", err)
}

func (e *dockerExecuter) Kill() {
	e.kill <- true
}
