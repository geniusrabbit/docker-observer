//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	fatalError(Config.Load())

	log.SetLevel(log.InfoLevel)
	if Config.Debug {
		log.SetLevel(log.DebugLevel)
		go func() { log.Println(http.ListenAndServe(":6060", nil)) }()
	}

	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05 MST",
	})

	if Config.Debug {
		fmt.Println(Config.String())
	}
}

func main() {
	observer, err := newObserver()
	fatalError(err)
	fatalError(observer.Run())
}

func newObserver() (Observer, error) {
	var (
		router   = NewRouter()
		executor = NewExecuter(router)
	)

	for _, rt := range Config.Routes {
		var route Route
		if rt.Cmd != "" {
			if rt.Source != "" || rt.Target != "" {
				fatalError(fmt.Errorf("Router can't combine command and template"))
			}
			route = &CmdRoute{Cmd: rt.Cmd}
		} else {
			route = &TplRoute{
				Source: strings.Replace(rt.Source, "{{basedir}}", Config.BaseDir, -1),
				Target: strings.Replace(rt.Target, "{{basedir}}", Config.BaseDir, -1),
			}
		}

		fatalError(route.Validate())
		router.Add(rt.NamePattern, rt.Actions, route)
	}

	return New(
		executor,
		Config.Docker.Host,
		Config.Docker.Version,
		nil, nil)
}

func fatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
