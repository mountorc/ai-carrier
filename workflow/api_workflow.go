package workflow

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	httpmount "github.com/xmzail/ai-carrier-dev/mounts/http"
	api "github.com/xmzail/ai-carrier-dev/sdk/mountcore-sdk/go"
)

//go:embed api_workflow.json
var apiWorkflowData []byte

func getHandlerMap() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"ProxyHandler":             wrapHTTPHandler(httpmount.ProxyHandler),
		"VectorSearchHandler":      wrapHTTPHandler(VectorSearchHandler),
		"ExecuteFlowHandler":       wrapHTTPHandler(ExecuteFlowHandler),
		"GetFlowsHandler":          wrapHTTPHandler(GetFlowsHandler),
		"ExecuteFlowByIDHandler":   wrapHTTPHandler(ExecuteFlowByIDHandler),
		"PluginExecuteHandler":     wrapHTTPHandler(PluginExecuteHandler),
		"PluginListHandler":        wrapHTTPHandler(PluginListHandler),
		"DocsPublicListHandler":    wrapHTTPHandler(DocsPublicListHandler),
		"DocsPublicGetHandler":     wrapHTTPHandler(DocsPublicGetHandler),
		"CustomNodesConfigHandler": wrapHTTPHandler(CustomNodesConfigHandler),
		"ExecutionRecordHandler":   wrapHTTPHandler(ExecutionRecordHandler),
		"GetAutosetHandler":        wrapHTTPHandler(GetAutosetHandler),
		"GetWorkflowListHandler":   wrapHTTPHandler(GetWorkflowListHandler),
		"HomeHandler":              wrapHTTPHandler(HomeHandler),
	}
}

func wrapHTTPHandler(handler http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c.Writer, c.Request)
	}
}

func RegisterRoutes(router gin.IRouter) {
	if err := api.RegisterRoutesFromJSONData(router, apiWorkflowData, "workflow/api_workflow.json", getHandlerMap()); err != nil {
		log.Fatalf("Failed to register workflow routes: %v", err)
	}
}

func Init() error {
	if err := InitAutosetDB(); err != nil {
		log.Printf("Warning: failed to initialize Autoset DB: %v", err)
	} else {
		log.Println("Autoset DB initialized successfully")
	}
	return nil
}

func Close() {
	CloseAutosetDB()
}
