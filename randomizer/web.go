package main

import (
	ai "github.com/zond/stockholm-ai/ai"
	common "github.com/zond/stockholm-ai/hub/common"
	examples "github.com/zond/stockholm-ai/hub/examples"
	"net/http"
)

func init() {
	http.Handle("/", ai.HTTPHandlerFunc(common.GAELoggerFactory, examples.Randomizer{}))
}
