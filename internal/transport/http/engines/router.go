package engines

import (
	"github.com/DaminduDilsara/web-analyzer/internal/controllers"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Engine struct {
	controller *controllers.ControllerV1
}

func NewEngine(
	controller *controllers.ControllerV1,
) *Engine {
	return &Engine{
		controller: controller,
	}
}

func (e *Engine) GetEngine() *gin.Engine {
	engine := gin.New()

	pprof.Register(engine)

	engine.GET("/ping", func(context *gin.Context) {
		context.String(http.StatusOK, "pong")
	})

	v1Group := engine.Group("/api/v1")
	{
		v1Group.POST("analyze", e.controller.AnalyzeController)
	}

	return engine
}
