package services

import (
	"context"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/response_dtos"
	"net/url"
)

type WebAnalyzerService interface {
	AnalyzeUrl(ctx context.Context, parsedURL *url.URL) (*response_dtos.UrlAnalyzerResponse, error)
}
