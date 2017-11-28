//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

// Observer description
type Observer interface {
	Run() error
	Stop()
}

// ContainerEventer processor
type ContainerEventer interface {
	Event(msg *ExecuteMessage)
	Error(err error)
}

// Service container observer
type baseObserver struct {
	sync.Mutex
	ContainerEventer
	inProcess bool
	script    []*ExecuteMessage
	executer  *time.Ticker
	ticker    *time.Ticker
	docker    *client.Client
}

// New for current docker container
func New(eventer ContainerEventer, host, version string, httpClient *http.Client, httpHeader map[string]string) (Observer, error) {
	client, err := client.NewClient(
		def(host, client.DefaultDockerHost),
		def(version, api.DefaultVersion),
		httpClient,
		httpHeader,
	)

	if err != nil {
		return nil, err
	}

	return &baseObserver{
		ContainerEventer: eventer,
		docker:           client,
	}, nil
}

// Run baseObserver
func (o *baseObserver) Run() error {
	o.Stop()

	messages, errors := o.docker.Events(context.Background(), types.EventsOptions{})
	o.ticker = time.NewTicker(10 * time.Second)
	o.executer = time.NewTicker(1 * time.Second)

	// Do refresh state at begining
	o.refreshAll("init")

	for {
		select {
		case msg := <-messages:
			switch msg.Type {
			case events.ContainerEventType:
				containers, err := o.containerInspectList()

				if err != nil {
					o.ContainerEventer.Error(err)
				}

				for _, cnt := range containers {
					if cnt.ID == msg.Actor.ID {
						o.scriptActionPair(msg, func(pair *ExecuteMessage) {
							pair.Containers = append(pair.Containers, cnt)
							pair.AllContainers = containers
						})
						break
					}
				} // end for
			case events.ServiceEventType:
				services, err := o.serviceInspectList()

				if err != nil {
					o.ContainerEventer.Error(err)
				}

				for _, srv := range services {
					if srv.ID == msg.Actor.ID {
						o.scriptActionPair(msg, func(pair *ExecuteMessage) {
							pair.Services = append(pair.Services, srv)
							pair.AllServices = services
						})
						break
					}
				} // end for
			}
		case err := <-errors:
			if err != nil {
				o.ContainerEventer.Error(err)
			}
		case <-o.executer.C:
			if len(o.script) > 0 {
				o.Lock()
				script := o.script
				o.script = nil
				o.Unlock()

				for _, pair := range script {
					go o.ContainerEventer.Event(pair)
				}
			}
		case <-o.ticker.C:
			o.refreshAll()
		}
	}
	return nil
}

func (o *baseObserver) Stop() {
	if o.ticker != nil {
		o.ticker.Stop()
		o.ticker = nil
	}
	if o.executer != nil {
		o.executer.Stop()
		o.executer = nil
	}
}

// Docker client
func (o *baseObserver) Docker() *client.Client {
	return o.docker
}

func (o *baseObserver) refreshAll(action ...string) {
	if o.goInProcess() {
		return
	}
	defer o.outOfProcess()

	var act = "refresh"
	if len(action) > 0 {
		act = action[0]
	}

	containers, err := o.containerInspectList()
	{
		if err != nil {
			log.Errorf("Refresh container list: %v", err)
			return
		}

		o.ContainerEventer.Event(&ExecuteMessage{
			Action:        act,
			Scope:         "local",
			AllContainers: containers,
		})
	}

	services, err := o.serviceInspectList()
	{
		if err != nil {
			log.Errorf("Refresh services list: %v", err)
			return
		}

		o.ContainerEventer.Event(&ExecuteMessage{
			Action:      act,
			Scope:       "swarm",
			AllServices: services,
		})
	}
}

func (o *baseObserver) containerInspectList() (list []types.ContainerJSON, err error) {
	var containers []types.Container

	containers, err = o.docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, cnt := range containers {
		if c, err := o.docker.ContainerInspect(context.Background(), cnt.ID); err == nil {
			list = append(list, c)
		}
	}

	return
}

func (o *baseObserver) serviceInspectList() ([]swarm.Service, error) {
	return o.docker.ServiceList(context.Background(), types.ServiceListOptions{})
}

func (o *baseObserver) scriptActionPair(msg events.Message, f func(pair *ExecuteMessage)) {
	o.Lock()
	var pair *ExecuteMessage
	for _, pr := range o.script {
		if pr.Action == msg.Action && pr.Scope == msg.Scope {
			pair = pr
		}
	}

	if pair == nil {
		pair = &ExecuteMessage{Action: msg.Action, Scope: msg.Scope}
		o.script = append(o.script, pair)
	}

	f(pair)

	o.Unlock()
}

func (o *baseObserver) goInProcess() bool {
	o.Lock()
	defer o.Unlock()

	if o.inProcess {
		return true
	}

	o.inProcess = true
	return false
}

func (o *baseObserver) outOfProcess() {
	o.Lock()
	o.inProcess = false
	o.Unlock()
}

func def(v, def string) string {
	if len(v) > 0 {
		return v
	}
	return def
}
