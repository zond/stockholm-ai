package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/zond/stockholm-ai/ai"
	"github.com/zond/stockholm-ai/hub/common"
	"github.com/zond/stockholm-ai/hub/models"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"

	brokenAi "github.com/zond/stockholm-ai/broken/ai"
	aiCommon "github.com/zond/stockholm-ai/common"
	randomizerAi "github.com/zond/stockholm-ai/randomizer/ai"
	simpletonAi "github.com/zond/stockholm-ai/simpleton/ai"
)

var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var jsModelTemplates = template.Must(template.New("jsModelTemplates").ParseGlob("templates/js/models/*.js"))
var jsCollectionTemplates = template.Must(template.New("jsCollectionTemplates").ParseGlob("templates/js/collections/*.js"))
var jsViewTemplates = template.Must(template.New("jsViewTemplates").ParseGlob("templates/js/views/*.js"))
var _Templates = template.Must(template.New("_Templates").ParseGlob("templates/_/*.html"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = template.Must(template.New("cssTemplates").ParseGlob("templates/css/*.css"))

func allCSS(c common.Context) {
	c.SetContentType("text/css; charset=UTF-8", true)
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
		rendered = strings.Replace(rendered, "\\", "\\\\", -1)
		rendered = strings.Replace(rendered, "'", "\\'", -1)
		rendered = strings.Replace(rendered, "\n", "\\n", -1)
		fmt.Fprint(c.Resp, rendered)
		fmt.Fprintln(c.Resp, "');")
		fmt.Fprintln(c.Resp, "  $('head').append(n);")
	}
	fmt.Fprintln(c.Resp, "})();")
}

func allJS(c common.Context) {
	c.SetContentType("application/javascript; charset=UTF-8", true)
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
	c.SetContentType("text/html; charset=UTF-8", false)
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

func getAIErrors(c common.Context) {
	if c.Authenticated() {
		if ai := models.GetAIById(c, common.MustDecodeKey(c.Vars["ai_id"])); ai != nil && ai.Owner == c.User.Email {
			c.RenderJSON(ai.GetErrors(c))
		}
	}
}

func getGames(c common.Context) {
	limit := aiCommon.TryParseInt(c.Req.URL.Query().Get("limit"), 10)
	offset := aiCommon.TryParseInt(c.Req.URL.Query().Get("offset"), 0)
	c.RenderJSON(models.GetGamePage(c, offset, limit))
}

func getGame(c common.Context) {
	c.RenderJSON(models.GetGameById(c, common.MustDecodeKey(c.Vars["game_id"])))
}

func getTurn(c common.Context) {
	c.RenderJSON(models.GetGivenTurnByParent(c, common.MustDecodeKey(c.Vars["game_id"]), aiCommon.MustParseInt(c.Vars["turn_ordinal"])))
}

func createGame(c common.Context) {
	if c.Authenticated() {
		var game models.Game
		aiCommon.MustDecodeJSON(c.Req.Body, &game)
		if len(game.Players) > 0 {
			c.RenderJSON(game.Save(c))
		}
	}
}

func createAI(c common.Context) {
	if c.Authenticated() {
		var ai models.AI
		aiCommon.MustDecodeJSON(c.Req.Body, &ai)
		if ai.Name != "" && ai.URL != "" {
			ai.Owner = c.User.Email
			ai.Id = nil
			c.RenderJSON(ai.Save(c))
		}
	}
}

func deleteAI(c common.Context) {
	if c.Authenticated() {
		if ai := models.GetAIById(c, common.MustDecodeKey(c.Vars["ai_id"])); ai != nil && ai.Owner == c.User.Email {
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

func handleStatic(router *mux.Router, dir string) {
	static, err := os.Open(dir)
	if err != nil {
		panic(err)
	}
	children, err := static.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	for _, fil := range children {
		cpy := fil
		router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
			return strings.HasSuffix(r.URL.Path, cpy)
		}).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".css") {
				common.SetContentType(w, "text/css; charset=UTF-8", true)
			} else if strings.HasSuffix(r.URL.Path, ".js") {
				common.SetContentType(w, "application/javascript; charset=UTF-8", true)
			} else if strings.HasSuffix(r.URL.Path, ".png") {
				common.SetContentType(w, "image/png", true)
			} else if strings.HasSuffix(r.URL.Path, ".gif") {
				common.SetContentType(w, "image/gif", true)
			} else if strings.HasSuffix(r.URL.Path, ".woff") {
				common.SetContentType(w, "application/font-woff", true)
			} else if strings.HasSuffix(r.URL.Path, ".ttf") {
				common.SetContentType(w, "font/truetype", true)
			} else {
				common.SetContentType(w, "application/octet-stream", true)
			}
			if in, err := os.Open(filepath.Join("static", cpy)); err != nil {
				w.WriteHeader(500)
			} else {
				defer in.Close()
				if _, err := io.Copy(w, in); err != nil {
					w.WriteHeader(500)
				}
			}
		})
	}
}

func main() {
	router := mux.NewRouter()
	router.Path("/js/{ver}/all.js").HandlerFunc(handler(allJS))
	router.Path("/css/{ver}/all.css").HandlerFunc(handler(allCSS))

	router.Path("/user").MatcherFunc(wantsJSON).HandlerFunc(handler(getUser))
	router.Path("/login").MatcherFunc(wantsHTML).HandlerFunc(handler(login))
	router.Path("/logout").MatcherFunc(wantsHTML).HandlerFunc(handler(logout))

	gamesRouter := router.PathPrefix("/games").MatcherFunc(wantsJSON).Subrouter()

	gameRouter := gamesRouter.PathPrefix("/{game_id}").Subrouter()

	turnsRouter := gameRouter.PathPrefix("/turns").Subrouter()
	turnRouter := turnsRouter.PathPrefix("/{turn_ordinal}").Subrouter()
	turnRouter.Methods("GET").HandlerFunc(handler(getTurn))

	gameRouter.Methods("GET").HandlerFunc(handler(getGame))

	gamesRouter.Methods("GET").HandlerFunc(handler(getGames))
	gamesRouter.Methods("POST").HandlerFunc(handler(createGame))

	aisRouter := router.PathPrefix("/ais").MatcherFunc(wantsJSON).Subrouter()

	aiRouter := aisRouter.PathPrefix("/{ai_id}").Subrouter()

	aiErrorsRouter := aiRouter.Path("/errors").Subrouter()
	aiErrorsRouter.Methods("GET").HandlerFunc(handler(getAIErrors))

	aiRouter.Methods("DELETE").HandlerFunc(handler(deleteAI))

	aisRouter.Methods("GET").HandlerFunc(handler(getAIs))
	aisRouter.Methods("POST").HandlerFunc(handler(createAI))

	router.Path("/examples/randomizer").Methods("POST").Handler(ai.HTTPHandlerFunc(common.GAELoggerFactory, randomizerAi.Randomizer{}))
	router.Path("/examples/simpleton").Methods("POST").Handler(ai.HTTPHandlerFunc(common.GAELoggerFactory, simpletonAi.Simpleton{}))
	router.Path("/examples/broken").Methods("POST").Handler(ai.HTTPHandlerFunc(common.GAELoggerFactory, brokenAi.Broken{}))

	handleStatic(router, "static")

	router.PathPrefix("/").MatcherFunc(wantsHTML).HandlerFunc(handler(index))
	http.Handle("/", router)
	appengine.Main()
}
