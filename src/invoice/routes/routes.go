package routes

import (
	"net/http"
	"regexp"
	"strings"
)

type Routes struct {
	Known map[string][]*Route
}

type Route struct {
	Pattern    *regexp.Regexp
	Handler    func(w http.ResponseWriter, r *http.Request, args []string)
	Middleware []func(w http.ResponseWriter, r *http.Request) bool
}

func (self *Routes) Add(method string, pattern string,
	handler func(w http.ResponseWriter, r *http.Request, args []string),
	middleware ...func(w http.ResponseWriter, r *http.Request) bool) {

	if self.Known == nil {
		self.Known = map[string][]*Route{}
	}

	routes, ok := self.Known[method]
	if !ok {
		routes = []*Route{}
	}

	pattern = strings.TrimRight(pattern, "/")
	pattern = "^" + pattern + "/?$"

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	routes = append(routes, &Route{re, handler, middleware})

	self.Known[method] = routes
}

func (self *Routes) Get(pattern string,
	handler func(w http.ResponseWriter, r *http.Request, args []string),
	middleware ...func(w http.ResponseWriter, r *http.Request) bool) {

	self.Add("GET", pattern, handler, middleware...)
}

func (self *Routes) Post(pattern string,
	handler func(w http.ResponseWriter, r *http.Request, args []string),
	middleware ...func(w http.ResponseWriter, r *http.Request) bool) {

	self.Add("POST", pattern, handler, middleware...)
}
