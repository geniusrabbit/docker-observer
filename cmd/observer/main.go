//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	observer "github.com/geniusrabbit/docker-observer"
	log "github.com/sirupsen/logrus"
)

func init() {
	fatalError(observer.Config.Load())

	log.SetLevel(log.InfoLevel)
	if observer.Config.Debug {
		log.SetLevel(log.DebugLevel)
		go func() { log.Println(http.ListenAndServe(":6060", nil)) }()
	}

	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05 MST",
	})

	if observer.Config.Debug {
		fmt.Println(observer.Config.String())
	}
}

func main() {
	observer, err := newObserver()
	fatalError(err)
	fatalError(observer.Run())
}

func newObserver() (observer.Observer, error) {
	var (
		router   = observer.NewRouter()
		executor = observer.NewExecuter(router)
	)

	for _, rt := range observer.Config.Routes {
		var route observer.Route
		if rt.Cmd != "" {
			if rt.Source != "" || rt.Target != "" {
				fatalError(fmt.Errorf("Router can't combine command and template"))
			}
			route = &observer.CmdRoute{
				Each:      rt.Each,
				Daemon:    rt.Daemon,
				Condition: rt.Condition,
				Cmd:       rt.Cmd,
				Filter:    rt.Filter,
			}
		} else {
			route = &observer.TplRoute{
				Each:      rt.Each,
				Condition: rt.Condition,
				Source:    rt.Source,
				Target:    rt.Target,
				Filter:    rt.Filter,
			}
		}

		fatalError(route.Validate())
		router.Add(rt.NamePattern, rt.Scope, rt.Actions, route)
	}

	return observer.New(
		executor,
		observer.Config.Docker.Host,
		observer.Config.Docker.Version,
		nil, nil)
}

func fatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
