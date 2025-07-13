package controllers

import (
	"context"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/request_dtos"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/response_dtos"
	"github.com/DaminduDilsara/web-analyzer/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"regexp"
)

const webAnalyzerControllerLogPrefix = "web_analyzer_controller"

type ControllerV1 struct {
	webAnalyzerService services.WebAnalyzerService
	logger             log_utils.LoggerInterface
}

func NewControllerV1(
	webAnalyzerService services.WebAnalyzerService,
	logger log_utils.LoggerInterface,
) *ControllerV1 {
	return &ControllerV1{
		webAnalyzerService: webAnalyzerService,
		logger:             logger,
	}
}

// AnalyzeController - extracts the url from request body and validate it first.
// then send to the web analyzer service for analyzing it and return the response
func (con *ControllerV1) AnalyzeController(c *gin.Context) {

	ctx := context.WithValue(context.Background(), "requestId", uuid.New().String())
	var jsonBody request_dtos.UrlAnalyzerRequest

	if err := c.BindJSON(&jsonBody); err != nil || jsonBody.Url == "" {
		con.logger.ErrorWithContext(ctx, "invalid or missing json body", err, log_utils.SetLogFile(webAnalyzerControllerLogPrefix))
		errorResponse := response_dtos.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid or missing json body",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	inputURL := jsonBody.Url

	con.logger.InfoWithContext(ctx, fmt.Sprintf("got new request url %v", inputURL), log_utils.SetLogFile(webAnalyzerControllerLogPrefix))

	urlParseRegex := regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)

	parsedURL, err := url.ParseRequestURI(inputURL) // checking if the URL is in valid format
	if err != nil || !parsedURL.IsAbs() || !urlParseRegex.MatchString(parsedURL.Host) {
		con.logger.ErrorWithContext(ctx, "failed to parse url", err, log_utils.SetLogFile(webAnalyzerControllerLogPrefix))
		con.logger.EndOfLog()
		errorResponse := response_dtos.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "failed to parse url",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	result, err := con.webAnalyzerService.AnalyzeUrl(ctx, parsedURL)
	if err != nil || !parsedURL.IsAbs() {
		con.logger.ErrorWithContext(ctx, "failed to analyze url", err, log_utils.SetLogFile(webAnalyzerControllerLogPrefix))
		con.logger.EndOfLog()
		errorResponse := response_dtos.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("failed to analyze url: %v error: %v", inputURL, err.Error()),
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	con.logger.InfoWithContext(ctx, fmt.Sprintf("successfully analyzed url %v", inputURL), log_utils.SetLogFile(webAnalyzerControllerLogPrefix))
	con.logger.EndOfLog()
	c.JSON(http.StatusOK, result)
}
