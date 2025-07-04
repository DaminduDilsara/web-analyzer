package engines

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHttpEngine struct {
}

func NewMetricsHttpEngine() *MetricsHttpEngine {
	return &MetricsHttpEngine{}
}

func (m *MetricsHttpEngine) GetMetricsEngine() *gin.Engine {
	engine := gin.New()

	engine.GET("/metrics", func(context *gin.Context) {
		promhttp.Handler().ServeHTTP(context.Writer, context.Request)
	})
	return engine
}
