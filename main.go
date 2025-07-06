package main

import (
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/controllers"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/DaminduDilsara/web-analyzer/internal/services"
	"github.com/DaminduDilsara/web-analyzer/internal/transport/http"
	"github.com/DaminduDilsara/web-analyzer/internal/web_analyzer_utils"
	"os"
)

func main() {

	sig := make(chan os.Signal)

	conf := configurations.LoadConfigurations()

	logger := log_utils.InitLogger("web-analyzer", conf.LogConfig)
	logger.Info("starting web-analyzer service")

	webAnalyzerUtils := web_analyzer_utils.NewWebAnalyzerUtils(logger, conf.WebAnalyzerConfig)

	webAnalyzerService := services.NewWebAnalyzerService(logger, webAnalyzerUtils)

	controller := controllers.NewControllerV1(webAnalyzerService, logger)

	http.InitServer(logger, conf.AppConfig, controller)

	select {
	case <-sig:
		logger.Info("Application is shutting down..")

		http.Shutdown(logger)
		os.Exit(0)
	}
}
