package web

import (
	"net/http"

	"github.com/zond/stockholm-ai/ai"
	"github.com/zond/stockholm-ai/hub/common"

	myAi "github.com/zond/stockholm-ai/simpleton/ai"
)

func init() {
	http.Handle("/", ai.HTTPHandlerFunc(common.GAELoggerFactory, myAi.Simpleton{}))
}
