package services

import (
	"context"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/response_dtos"
	"github.com/DaminduDilsara/web-analyzer/internal/web_analyzer_utils"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
)

const webAnalyzerServiceLogPrefix = "web_analyzer_service_impl"

var typesOfHeadings = [...]string{"h1", "h2", "h3", "h4", "h5", "h6"}

type webAnalyzerServiceImpl struct {
	logger           log_utils.LoggerInterface
	webAnalyzerUtils web_analyzer_utils.WebAnalyzerUtils
}

func NewWebAnalyzerService(
	logger log_utils.LoggerInterface,
	webAnalyzerUtils web_analyzer_utils.WebAnalyzerUtils,
) WebAnalyzerService {
	return &webAnalyzerServiceImpl{
		logger:           logger,
		webAnalyzerUtils: webAnalyzerUtils,
	}
}

// AnalyzeUrl - analyze the given url and return UrlAnalyzerResponse as response
// - HTMLVersion - version of the web page
// - Title - title of web page
// - Headings - count of each heading type h1, h2, h3, h4, h5, h6
// - InternalLinks - count of internal links
// - ExternalLinks - count of external links
// - InaccessibleLinks - count of inaccessible links
// - LoginForm - if a login form present (true or false)
func (w *webAnalyzerServiceImpl) AnalyzeUrl(ctx context.Context, parsedURL *url.URL) (*response_dtos.UrlAnalyzerResponse, error) {

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		w.logger.ErrorWithContext(ctx, "Unable to fetch data from the url", err, log_utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}
	defer resp.Body.Close()

	if (resp.StatusCode < http.StatusOK) || (resp.StatusCode >= http.StatusMultipleChoices) {
		err = fmt.Errorf("unexpected HTTP status code: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		w.logger.ErrorWithContext(ctx, "unexpected HTTP status code", err, log_utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		w.logger.ErrorWithContext(ctx, "response cannot parse to html", err, log_utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}

	htmlText, err := doc.Html()
	if err != nil {
		w.logger.ErrorWithContext(ctx, "cannot extract html text from document", err, log_utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}

	result := response_dtos.UrlAnalyzerResponse{
		Headings:    make(map[string]int),
		Title:       doc.Find("title").First().Text(),
		HTMLVersion: w.webAnalyzerUtils.DetectHTMLVersion(ctx, htmlText),
		LoginForm:   doc.Find("form input[type='password']").Length() > 0,
	}

	// Count headings h1 to h6
	for _, heading := range typesOfHeadings {
		result.Headings[heading] = doc.Find(heading).Length()
	}

	var allLinks []string

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript") {
			return
		}

		link := strings.TrimSpace(href)
		isInternal := strings.HasPrefix(link, "/") && !strings.HasPrefix(link, "//") || strings.Contains(link, parsedURL.Host)

		if isInternal {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}
		allLinks = append(allLinks, link)
	})

	w.logger.InfoWithContext(ctx, fmt.Sprintf("found %v links totally", len(allLinks)), log_utils.SetLogFile(webAnalyzerServiceLogPrefix))

	result.InaccessibleLinks = w.webAnalyzerUtils.IsLinksAccessible(ctx, allLinks, parsedURL)

	return &result, nil

}
