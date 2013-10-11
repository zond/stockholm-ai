package randomizer

import (
	ai "github.com/zond/stockholm-ai/ai"
	common "github.com/zond/stockholm-ai/hub/common"
	randomizerAi "github.com/zond/stockholm-ai/randomizer/ai"
	"net/http"
)

func init() {
	http.Handle("/", ai.HTTPHandlerFunc(common.GAELoggerFactory, randomizerAi.Randomizer{}))
}
