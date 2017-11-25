//
// @project docker-observer 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package main

import "regexp"

type RouteFilter struct {
	Actions []string
	RegExp  *regexp.Regexp
}

func NewFilter(actions []string, regExp string) (fl RouteFilter, err error) {
	fl.Actions = actions
	if regExp != "" {
		fl.RegExp, err = regexp.Compile(regExp)
	}
	return
}

func (f RouteFilter) Test(action, name string) bool {
	return f.TestAction(action) && (f.RegExp == nil || f.RegExp.Match([]byte(name)))
}

func (f RouteFilter) TestAction(action string) bool {
	for _, act := range f.Actions {
		if act == action {
			return true
		}
	}
	return len(f.Actions) == 0
}
