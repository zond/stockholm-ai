package web

import (
	"github.com/zond/stockholm-ai/ai"
	"github.com/zond/stockholm-ai/hub/common"
	myAi "ai"
	"net/http"
)

func init() {
	http.Handle("/", ai.HTTPHandlerFunc(common.GAELoggerFactory, myAi.Broken{}))
}
