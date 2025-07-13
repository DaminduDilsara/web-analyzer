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
	"time"
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

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(parsedURL.String())
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

	htmlVersion := w.webAnalyzerUtils.DetectHTMLVersion(ctx, htmlText)

	pageTitle := w.webAnalyzerUtils.DetectPageTitle(ctx, doc)

	isLoginFormExist := w.webAnalyzerUtils.DetectLoginForm(ctx, doc)

	headingData := w.webAnalyzerUtils.DetectHeaders(ctx, doc, typesOfHeadings)

	internalLinks, externalLinks, allLinks := w.webAnalyzerUtils.DetectLinks(ctx, doc, parsedURL.Host)

	inaccessibleLinks := w.webAnalyzerUtils.IsLinksAccessible(ctx, allLinks, parsedURL)

	result := response_dtos.UrlAnalyzerResponse{
		HTMLVersion:       htmlVersion,
		Title:             pageTitle,
		Headings:          headingData,
		InternalLinks:     internalLinks,
		ExternalLinks:     externalLinks,
		InaccessibleLinks: inaccessibleLinks,
		LoginForm:         isLoginFormExist,
	}

	return &result, nil
}
