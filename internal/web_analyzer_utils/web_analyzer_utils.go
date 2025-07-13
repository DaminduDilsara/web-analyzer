package web_analyzer_utils

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"net/url"
)

type WebAnalyzerUtils interface {
	DetectHTMLVersion(ctx context.Context, body string) string
	DetectPageTitle(ctx context.Context, doc *goquery.Document) string
	DetectLoginForm(ctx context.Context, doc *goquery.Document) bool
	DetectHeaders(ctx context.Context, doc *goquery.Document, typesOfHeadings [6]string) map[string]int
	DetectLinks(ctx context.Context, doc *goquery.Document, host string) (int, int, []string)
	IsLinksAccessible(ctx context.Context, links []string, base *url.URL) int
}
