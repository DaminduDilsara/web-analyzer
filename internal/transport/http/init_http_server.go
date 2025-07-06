package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/controllers"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/DaminduDilsara/web-analyzer/internal/transport/http/engines"
	"net/http"
	"time"
)

const webAnalyzerWebServerLogPrefix = "init_http_server"

var engine http.Server
var srvMetrics http.Server

// InitServer - initialize the web servers
func InitServer(
	logger log_utils.LoggerInterface,
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
			logger.Fatal("failed to start default web server", err, log_utils.SetLogFile(webAnalyzerWebServerLogPrefix))
		}
	}()
	logger.Info(fmt.Sprintf("Starting default web server under port : %v", appConf.AppPort), log_utils.SetLogFile(webAnalyzerWebServerLogPrefix))
	go func() {
		if err := srvMetrics.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("failed to start the metrics web server", err, log_utils.SetLogFile(webAnalyzerWebServerLogPrefix))
		}
	}()
	logger.Info(fmt.Sprintf("Starting metrics web server under port : %v", appConf.MetricPort), log_utils.SetLogFile(webAnalyzerWebServerLogPrefix))
}

// Shutdown - shutdown the web servers
func Shutdown(logger log_utils.LoggerInterface) {

	if err := engine.Shutdown(context.Background()); err != nil {
		logger.Fatal("failed to stop the default web server", err, log_utils.SetLogFile(webAnalyzerWebServerLogPrefix))
	}

	if err := srvMetrics.Shutdown(context.Background()); err != nil {
		logger.Fatal("failed to stop the metrics web server", err, log_utils.SetLogFile(webAnalyzerWebServerLogPrefix))
	}
}
