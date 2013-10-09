package web

import (
	"appengine"
	"appengine/user"
	"bytes"
	"common"
	"fmt"
	"github.com/gorilla/mux"
	"models"
	"net/http"
	"regexp"
	"strings"
	"text/template"
)

var spaceRegexp = regexp.MustCompile("\\s+")

var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var jsModelTemplates = template.Must(template.New("jsModelTemplates").ParseGlob("templates/js/models/*.js"))
var jsCollectionTemplates = template.Must(template.New("jsCollectionTemplates").ParseGlob("templates/js/collections/*.js"))
var jsViewTemplates = template.Must(template.New("jsViewTemplates").ParseGlob("templates/js/views/*.js"))
var _Templates = template.Must(template.New("_Templates").ParseGlob("templates/_/*.html"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = template.Must(template.New("cssTemplates").ParseGlob("templates/css/*.css"))

func allCSS(c common.Context) {
	if !appengine.IsDevAppServer() {
		c.Resp.Header().Set("Cache-Control", "public, max-age=864000")
	}
	c.Resp.Header().Set("Content-Type", "text/css; charset=UTF-8")
	renderText(c, cssTemplates, "bootstrap.min.css")
	renderText(c, cssTemplates, "bootstrap-theme.min.css")
	renderText(c, cssTemplates, "bootstrap-multiselect.css")
	renderText(c, cssTemplates, "common.css")
}

func renderText(c common.Context, templates *template.Template, template string) {
	if err := templates.ExecuteTemplate(c.Resp, template, c); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}

func render_Templates(c common.Context) {
	fmt.Fprintln(c.Resp, "(function() {")
	fmt.Fprintln(c.Resp, "  var n;")
	var buf *bytes.Buffer
	var rendered string
	for _, templ := range _Templates.Templates() {
		fmt.Fprintf(c.Resp, "  n = $('<script type=\"text/template\" id=\"%v_underscore\"></script>');\n", strings.Split(templ.Name(), ".")[0])
		fmt.Fprintf(c.Resp, "  n.text('")
		buf = new(bytes.Buffer)
		if err := templ.Execute(buf, c); err != nil {
			panic(err)
		}
		rendered = string(buf.Bytes())
		rendered = spaceRegexp.ReplaceAllString(rendered, " ")
		rendered = strings.Replace(rendered, "\\", "\\\\", -1)
		rendered = strings.Replace(rendered, "'", "\\'", -1)
		fmt.Fprint(c.Resp, rendered)
		fmt.Fprintln(c.Resp, "');")
		fmt.Fprintln(c.Resp, "  $('head').append(n);")
	}
	fmt.Fprintln(c.Resp, "})();")
}

func allJS(c common.Context) {
	if !appengine.IsDevAppServer() {
		c.Resp.Header().Set("Cache-Control", "public, max-age=864000")
	}
	c.Resp.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	renderText(c, jsTemplates, "jquery-2.0.3.min.js")
	renderText(c, jsTemplates, "underscore-min.js")
	renderText(c, jsTemplates, "backbone-min.js")
	renderText(c, jsTemplates, "bootstrap.min.js")
	renderText(c, jsTemplates, "bootstrap-multiselect.js")
	renderText(c, jsTemplates, "viz.js")
	renderText(c, jsTemplates, "tinycolor.js")
	render_Templates(c)
	for _, templ := range jsModelTemplates.Templates() {
		if err := templ.Execute(c.Resp, c); err != nil {
			panic(err)
		}
	}
	for _, templ := range jsCollectionTemplates.Templates() {
		if err := templ.Execute(c.Resp, c); err != nil {
			panic(err)
		}
	}
	for _, templ := range jsViewTemplates.Templates() {
		if err := templ.Execute(c.Resp, c); err != nil {
			panic(err)
		}
	}
	renderText(c, jsTemplates, "app.js")
}

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "text/html"
}

func index(c common.Context) {
	if !appengine.IsDevAppServer() {
		c.Resp.Header().Set("Cache-Control", "public, max-age=864000")
	}
	c.Resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
	renderText(c, htmlTemplates, "index.html")
}

func getUser(c common.Context) {
	c.RenderJSON(c.User)
}

func login(c common.Context) {
	url, err := user.LoginURL(c.Context, c.Req.URL.Scheme+c.Req.URL.Host)
	if err != nil {
		panic(err)
	}
	c.Resp.Header().Set("Location", url)
	c.Resp.WriteHeader(302)
}

func logout(c common.Context) {
	url, err := user.LogoutURL(c.Context, c.Req.URL.Scheme+c.Req.URL.Host)
	if err != nil {
		panic(err)
	}
	c.Resp.Header().Set("Location", url)
	c.Resp.WriteHeader(302)
}

func getAIs(c common.Context) {
	c.RenderJSON(models.GetAllAIs(c))
}

func getGames(c common.Context) {
	c.RenderJSON(models.GetAllGames(c))
}

func getGame(c common.Context) {
	c.RenderJSON(models.GetGameById(c, common.MustDecodeKey(c.Vars["game_id"])))
}

func createGame(c common.Context) {
	if c.Authenticated() {
		var game models.Game
		common.MustDecodeJSON(c.Req.Body, &game)
		if len(game.Players) > 0 {
			c.RenderJSON(game.Save(c))
		}
	}
}

func createAI(c common.Context) {
	if c.Authenticated() {
		var ai models.AI
		common.MustDecodeJSON(c.Req.Body, &ai)
		if ai.Name != "" && ai.URL != "" {
			ai.Owner = c.User.Email
			ai.Id = nil
			c.RenderJSON(ai.Save(c))
		}
	}
}

func deleteAI(c common.Context) {
	if c.Authenticated() {
		if ai := models.GetAIById(c, common.MustDecodeKey(c.Vars["ai_id"])); ai.Owner == c.User.Email {
			ai.Delete(c)
		}
	}
}

func handler(f func(c common.Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := common.Context{
			Context: appengine.NewContext(r),
			Req:     r,
			Resp:    w,
			Vars:    mux.Vars(r),
		}
		c.User = user.Current(c)
		c.Version = appengine.VersionID(c.Context)
		f(c)
	}
}

func init() {
	router := mux.NewRouter()
	router.Path("/js/{ver}/all.js").HandlerFunc(handler(allJS))
	router.Path("/css/{ver}/all.css").HandlerFunc(handler(allCSS))

	router.Path("/user").MatcherFunc(wantsJSON).HandlerFunc(handler(getUser))
	router.Path("/login").MatcherFunc(wantsHTML).HandlerFunc(handler(login))
	router.Path("/logout").MatcherFunc(wantsHTML).HandlerFunc(handler(logout))

	gamesRouter := router.PathPrefix("/games").MatcherFunc(wantsJSON).Subrouter()

	gameRouter := gamesRouter.PathPrefix("/{game_id}").Subrouter()
	gameRouter.Methods("GET").HandlerFunc(handler(getGame))

	gamesRouter.Methods("GET").HandlerFunc(handler(getGames))
	gamesRouter.Methods("POST").HandlerFunc(handler(createGame))

	aisRouter := router.PathPrefix("/ais").MatcherFunc(wantsJSON).Subrouter()

	aiRouter := aisRouter.PathPrefix("/{ai_id}").Subrouter()
	aiRouter.Methods("DELETE").HandlerFunc(handler(deleteAI))

	aisRouter.Methods("GET").HandlerFunc(handler(getAIs))
	aisRouter.Methods("POST").HandlerFunc(handler(createAI))

	router.PathPrefix("/").MatcherFunc(wantsHTML).HandlerFunc(handler(index))
	http.Handle("/", router)
}
