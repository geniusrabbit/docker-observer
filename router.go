//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

// Router basic interface
type Router interface {
	Add(pattern, scope string, actions []string, route Route) error
	Routes(action, scope string, obj interface{}) []Route
}

// Route interface
type Route interface {
	Exec(ctx context.Context, msg *ExecuteMessage) error
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
func (r *router) Add(pattern, scope string, actions []string, route Route) error {
	fl, err := NewFilter(actions, scope, pattern)
	r.routes = append(r.routes, routeItem{
		filter: fl,
		route:  route,
	})
	return err
}

// Routes for container
func (r *router) Routes(action, scope string, obj interface{}) (list []Route) {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, rt := range r.routes {
		var name string
		switch o := obj.(type) {
		case types.ContainerJSON:
			name = o.Name
		case swarm.Service:
			name = o.Spec.Name
		}

		if rt.filter.Test(action, scope, name) {
			list = append(list, rt.route)
		}
	}
	return
}
