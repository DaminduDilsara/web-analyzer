package web_analyzer_utils

import (
	"context"
	"net/url"
)

type WebAnalyzerUtils interface {
	DetectHTMLVersion(ctx context.Context, body string) string
	IsLinksAccessible(ctx context.Context, links []string, base *url.URL) int
}
