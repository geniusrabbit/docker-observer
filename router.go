//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package main

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types"
)

// Router basic interface
type Router interface {
	Add(pattern string, actions []string, route Route) error
	Routes(action string, container *types.ContainerJSON) []Route
}

// Route interface
type Route interface {
	Exec(ctx context.Context, action string, containers, allContainers []types.ContainerJSON) error
	Validate() error
}

type routeItem struct {
	filter RouteFilter
	route  Route
}

type router struct {
	mx     sync.Mutex
	routes []routeItem
}

// NewRouter object
func NewRouter() Router {
	return &router{}
}

// Add route executer
func (r *router) Add(pattern string, actions []string, route Route) error {
	fl, err := NewFilter(actions, pattern)
	r.routes = append(r.routes, routeItem{
		filter: fl,
		route:  route,
	})
	return err
}

// Routes for container
func (r *router) Routes(action string, container *types.ContainerJSON) (list []Route) {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, rt := range r.routes {
		if rt.filter.Test(action, container.Name) {
			list = append(list, rt.route)
		}
	}
	return
}
