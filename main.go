package main

import (
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/controllers"
	"github.com/DaminduDilsara/web-analyzer/internal/transport/http"
	"log"
	"os"
)

func main() {
	log.Println("starting web-analyzer")

	sig := make(chan os.Signal)

	conf := configurations.LoadConfigurations()

	controller := controllers.NewControllerV1()

	http.InitServer(conf.AppConfig, controller)

	select {
	case <-sig:
		log.Println("Application is shutting down..")

		http.Shutdown()
		os.Exit(0)
	}
}
