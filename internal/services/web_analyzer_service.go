package services

import (
	"context"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/responseDtos"
	"net/url"
)

type WebAnalyzerService interface {
	AnalyzeUrl(ctx context.Context, parsedURL *url.URL) (*responseDtos.UrlAnalyzerResponse, error)
}
