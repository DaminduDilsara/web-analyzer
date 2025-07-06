package main

import (
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/controllers"
	"github.com/DaminduDilsara/web-analyzer/internal/services"
	"github.com/DaminduDilsara/web-analyzer/internal/transport/http"
	"github.com/DaminduDilsara/web-analyzer/internal/utils"
	"log"
	"os"
)

func main() {
	log.Println("starting web-analyzer")

	sig := make(chan os.Signal)

	conf := configurations.LoadConfigurations()

	logger := utils.InitLogger("web-analyzer", conf.LogConfig)

	webAnalyzerService := services.NewWebAnalyzerService(logger)

	controller := controllers.NewControllerV1(webAnalyzerService, logger)

	http.InitServer(conf.AppConfig, controller)

	select {
	case <-sig:
		log.Println("Application is shutting down..")

		http.Shutdown()
		os.Exit(0)
	}
}
