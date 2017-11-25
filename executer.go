//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package main

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

type containerPair struct {
	route         Route
	containers    []types.ContainerJSON
	allContainers []types.ContainerJSON
}

type dockerExecuter struct {
	kill    chan bool
	timeout time.Duration
	router  Router
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
func (e *dockerExecuter) Event(action string, containers, allContainers []types.ContainerJSON) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), e.timeout)
		eContainers []*containerPair
	)

	if len(containers) < 1 {
		containers = allContainers
	}

	// Collect containers by routes
	for _, cnt := range containers {
		if routes := e.router.Routes(action, &cnt); len(routes) > 0 {
			for _, route := range routes {
				var tpr *containerPair
				for _, pr := range eContainers {
					if pr.route == route {
						tpr = pr
					}
				}
				if tpr == nil {
					tpr = &containerPair{route: route}
					eContainers = append(eContainers, tpr)
				}

				tpr.containers = append(tpr.containers, cnt)
				tpr.allContainers = allContainers
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
		if err := pr.route.Exec(ctx, action, pr.containers, pr.allContainers); err != nil {
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
