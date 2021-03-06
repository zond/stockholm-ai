package main

import (
	"net/http"

	"github.com/zond/stockholm-ai/ai"
	"github.com/zond/stockholm-ai/hub/common"
	"google.golang.org/appengine"

	myAi "github.com/zond/stockholm-ai/broken/ai"
)

func main() {
	http.Handle("/", ai.HTTPHandlerFunc(common.GAELoggerFactory, myAi.Broken{}))
	appengine.Main()
}
