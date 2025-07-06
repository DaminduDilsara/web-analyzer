package controllers

import (
	"context"
	"fmt"
	app_errors "github.com/DaminduDilsara/web-analyzer/internal/errors"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/requestDtos"
	"github.com/DaminduDilsara/web-analyzer/internal/services"
	"github.com/DaminduDilsara/web-analyzer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/url"
)

const webAnalyzerControllerLogPrefix = "web_analyzer_controller"

type ControllerV1 struct {
	webAnalyzerService services.WebAnalyzerService
	logger             utils.LoggerInterface
}

func NewControllerV1(
	webAnalyzerService services.WebAnalyzerService,
	logger utils.LoggerInterface,
) *ControllerV1 {
	return &ControllerV1{
		webAnalyzerService: webAnalyzerService,
		logger:             logger,
	}
}

func (con *ControllerV1) AnalyzeController(c *gin.Context) {

	ctx := context.WithValue(context.Background(), "requestId", uuid.New().String())
	var jsonBody requestDtos.UrlAnalyzerRequest

	if err := c.BindJSON(&jsonBody); err != nil || jsonBody.Url == "" {
		con.logger.ErrorWithContext(ctx, "invalid or missing json body", err, utils.SetLogFile(webAnalyzerControllerLogPrefix))
		c.JSON(http.StatusBadRequest, app_errors.ErrorBadRequest)
		return
	}
	inputURL := jsonBody.Url

	con.logger.InfoWithContext(ctx, fmt.Sprintf("got new request url %v", inputURL), utils.SetLogFile(webAnalyzerControllerLogPrefix))

	parsedURL, err := url.ParseRequestURI(inputURL) // checking if the URL is in valid format
	if err != nil || !parsedURL.IsAbs() {
		c.JSON(http.StatusBadRequest, app_errors.ErrorBadRequest)
		return
	}

	result, err := con.webAnalyzerService.AnalyzeUrl(ctx, parsedURL)
	if err != nil || !parsedURL.IsAbs() {
		c.JSON(http.StatusInternalServerError, app_errors.ErrorInternal)
		return
	}

	c.JSON(http.StatusOK, result)
}
