package server

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
)

import (
	"github.com/nathanwinther/totp"
	"invoice/company"
	"invoice/config"
	"invoice/routes"
	"invoice/session"
	"invoice/timesheet"
)

type Handler struct {
	Routes   *routes.Routes
	Template *template.Template
}

func Run(host string, port string) {
	h := new(Handler)
	h.Routes = new(routes.Routes)

	token := "[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{2,}"
	key := "[0-9]{4}-[0-9]{2}-[0-9]{2}"

	h.Routes.Get(fmt.Sprintf("/invoice"), h.getIndex, session.Check)
	h.Routes.Get(fmt.Sprintf("/invoice/(%s)", token), h.getTimesheet, session.Check)
	h.Routes.Get(fmt.Sprintf("/invoice/(%s)/(%s)", token, key), h.getTimesheet, session.Check)
	h.Routes.Get(fmt.Sprintf("/invoice/history/(%s)", token), h.getHistory, session.Check)
	h.Routes.Get(fmt.Sprintf("/invoice/preview/(%s)", token), h.getPreview, session.Check)
	h.Routes.Get(fmt.Sprintf("/invoice/verify"), h.getVerify)
	h.Routes.Get(fmt.Sprintf("/invoice/verify/([^/]*)"), h.getVerify)

	h.Routes.Post(fmt.Sprintf("/invoice/(%s)", token), h.postTimesheet, session.Check)
	h.Routes.Post(fmt.Sprintf("/invoice/(%s)/(%s)", token, key), h.postTimesheet, session.Check)
	h.Routes.Post(fmt.Sprintf("/invoice/close/(%s)", token), h.postClose, session.Check)
	h.Routes.Post(fmt.Sprintf("/invoice/verify"), h.postVerify)
	h.Routes.Post(fmt.Sprintf("/invoice/verify/([^/]*)"), h.postVerify)

	template, err := template.ParseGlob("template/*.*")
	if err != nil {
		panic(err)
	}
	h.Template = template

	listen := fmt.Sprintf("%s:%s", host, port)

	bind, err := net.Listen("tcp", listen)
	if err != nil {
		panic(err)
	}

	http.Handle("/", h)

	err = http.Serve(bind, nil)
	if err != nil {
		panic(err)
	}
}

func (self *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	notfound := func() {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Add("Content-Type", "text/plain")
		fmt.Fprintln(w, "404 Not Found")
		if !config.PRODUCTION {
			fmt.Printf("404: %s\n", r.URL.Path)
		}
	}

	// Handle method?
	routes, ok := self.Routes.Known[r.Method]
	if !ok {
		notfound()
		return
	}

	for _, route := range routes {
		match := route.Pattern.FindStringSubmatch(r.URL.Path)
		if match != nil {
			// Handle middleware
			for _, middleware := range route.Middleware {
				if !middleware(w, r) {
					return
				}
			}
			// Middleware passed, run handler
			route.Handler(w, r, match[1:])
			return
		}
	}

	notfound()
}

func (self *Handler) getHistory(w http.ResponseWriter, r *http.Request, args []string) {
	uuid := args[0]

	items, err := timesheet.Items(uuid)
	if err != nil {
		self.fatal(w, err)
		return
	}

	m := map[string]interface{}{
		"S3":      config.S3_WEBSITE,
		"History": items,
	}

	err = self.Template.ExecuteTemplate(w, "history.html", m)
	if err != nil {
		self.fatal(w, err)
		return
	}
}

func (self *Handler) getPreview(w http.ResponseWriter, r *http.Request, args []string) {
	uuid := args[0]

	c, err := company.Load(uuid, "")
	if err != nil {
		self.fatal(w, err)
		return
	}

	html, err := c.Html()
	if err != nil {
		self.fatal(w, err)
		return
	}

	fmt.Fprint(w, html.String())

}

func (self *Handler) getIndex(w http.ResponseWriter, r *http.Request, args []string) {
	items, err := company.Items()
	if err != nil {
		self.fatal(w, err)
		return
	}

	m := map[string]interface{}{
		"Company": items,
	}

	err = self.Template.ExecuteTemplate(w, "index.html", m)
	if err != nil {
		self.fatal(w, err)
		return
	}
}

func (self *Handler) getTimesheet(w http.ResponseWriter, r *http.Request, args []string) {
	uuid := args[0]

	key := ""
	if len(args) == 2 {
		key = args[1]
	}

	c, err := company.Load(uuid, key)
	if err != nil {
		self.fatal(w, err)
		return
	}

	hours := make([]int, 25)
	hours[c.Timesheet.Entries[c.Timesheet.Selected].Hours] = 1

	m := map[string]interface{}{
		"Company": c,
		"Hours":   hours,
	}

	err = self.Template.ExecuteTemplate(w, "timesheet.html", m)
	if err != nil {
		self.fatal(w, err)
		return
	}
}

func (self *Handler) fatal(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintln(w, "500 Internal server error")
	if !config.PRODUCTION {
		fmt.Fprintln(w, err)
		fmt.Println(err)
	}
}

func (self *Handler) getVerify(w http.ResponseWriter, r *http.Request, args []string) {
	m := map[string]interface{}{}

	err := self.Template.ExecuteTemplate(w, "verify.html", m)
	if err != nil {
		self.fatal(w, err)
		return
	}
}

func (self *Handler) postClose(w http.ResponseWriter, r *http.Request, args []string) {
	uuid := args[0]

	c, err := company.Load(uuid, "")
	if err != nil {
		self.fatal(w, err)
		return
	}

	err = c.CloseTimesheet()
	if err != nil {
		self.fatal(w, err)
		return
	}

	c.Timesheet, err = timesheet.New(c.Timesheet)
	if err != nil {
		self.fatal(w, err)
		return
	}

	err = c.Save()
	if err != nil {
		self.fatal(w, err)
		return
	}

	http.Redirect(w, r, "/invoice/"+uuid, http.StatusFound)
}

func (self *Handler) postTimesheet(w http.ResponseWriter, r *http.Request, args []string) {
	uuid := args[0]
	key := r.FormValue("key")
	hours := r.FormValue("hours")

	c, err := company.Load(uuid, key)
	if err != nil {
		self.fatal(w, err)
		return
	}

	c.Timesheet.SetHours(key, hours)
	err = c.Save()
	if err != nil {
		self.fatal(w, err)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusFound)
}

func (self *Handler) postVerify(w http.ResponseWriter, r *http.Request, args []string) {
	token := strings.TrimSpace(r.FormValue("token"))

	ok := totp.VerifyCode(config.TOTP_SECRET, token, 2)
	if ok {
		session.New(w)

		boomerang := "/invoice"
		if len(args) >= 1 {
			b, err := base64.RawURLEncoding.DecodeString(args[0])
			if err == nil {
				boomerang = string(b)
			}
		}

		http.Redirect(w, r, boomerang, http.StatusFound)

		return
	}

	m := map[string]interface{}{
		"Message": "Invalid Token",
	}

	err := self.Template.ExecuteTemplate(w, "verify.html", m)
	if err != nil {
		self.fatal(w, err)
		return
	}
}
