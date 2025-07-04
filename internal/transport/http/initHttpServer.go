package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/controllers"
	"github.com/DaminduDilsara/web-analyzer/internal/transport/http/engines"
	"log"
	"net/http"
	"time"
)

var engine http.Server
var srvMetrics http.Server

func InitServer(
	appConf *configurations.AppConfigurations,
	controllerV1 *controllers.ControllerV1,
) {

	engine = http.Server{
		Addr:         fmt.Sprintf(":%v", appConf.AppPort),
		Handler:      engines.NewEngine(controllerV1).GetEngine(),
		WriteTimeout: time.Second * time.Duration(appConf.WriteTimeout),
		ReadTimeout:  time.Second * time.Duration(appConf.ReadTimeOut),
		IdleTimeout:  time.Second * time.Duration(appConf.IdleTimeout),
	}

	srvMetrics = http.Server{
		Addr:         fmt.Sprintf(":%v", appConf.MetricPort),
		Handler:      engines.NewMetricsHttpEngine().GetMetricsEngine(),
		WriteTimeout: time.Second * time.Duration(appConf.WriteTimeout),
		ReadTimeout:  time.Second * time.Duration(appConf.ReadTimeOut),
		IdleTimeout:  time.Second * time.Duration(appConf.IdleTimeout),
	}

	go func() {
		if err := engine.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf(fmt.Sprintf("Failed to start default web server : %v", err))
		}
	}()
	log.Println(fmt.Sprintf("Starting default web server under port : %v", appConf.AppPort))

	go func() {
		if err := srvMetrics.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf(fmt.Sprintf("Failed to start the metrics web server : %v", err))
		}
	}()
	log.Println(fmt.Sprintf("Starting metrics web server under port : %v", appConf.MetricPort))
}

func Shutdown() {

	if err := engine.Shutdown(context.Background()); err != nil {
		log.Fatal(fmt.Sprintf("Failed to stop the default web server : %v", err))
	}

	if err := srvMetrics.Shutdown(context.Background()); err != nil {
		log.Fatal(fmt.Sprintf("Failed to stop the metrics web server : %v", err))
	}
}
