package web_analyzer_utils

import (
	"context"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const webAnalyzerUtilsLogPrefix = "web_analyzer_utils_impl"

type webAnalyzerUtilsImpl struct {
	logger            log_utils.LoggerInterface
	webAnalyzerConfig *configurations.WebAnalyzerConfigurations
}

func NewWebAnalyzerUtils(
	logger log_utils.LoggerInterface,
	webAnalyzerConfig *configurations.WebAnalyzerConfigurations,
) WebAnalyzerUtils {
	return &webAnalyzerUtilsImpl{
		logger:            logger,
		webAnalyzerConfig: webAnalyzerConfig,
	}
}

// DetectHTMLVersion - detects the html version of the web page using version strings
func (w *webAnalyzerUtilsImpl) DetectHTMLVersion(ctx context.Context, body string) string {
	body = strings.ToLower(body)

	htmlVersion := ""

	switch {
	case strings.Contains(body, "<!doctype html>"):
		htmlVersion = "HTML 5"
	case strings.Contains(body, "-//w3c//dtd html 4.01 frameset//en"):
		htmlVersion = "HTML 4.01 Frameset"
	case strings.Contains(body, "-//w3c//dtd html 4.01 transitional//en"):
		htmlVersion = "HTML 4.01 Transitional"
	case strings.Contains(body, "-//w3c//dtd html 4.01//en"):
		htmlVersion = "HTML 4.01 Strict"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.1//en"):
		htmlVersion = "XHTML 1.1"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.0 frameset//en"):
		htmlVersion = "XHTML 1.0 Frameset"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.0 transitional//en"):
		htmlVersion = "XHTML 1.0 Transitional"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.0 strict//en"):
		htmlVersion = "XHTML 1.0 Strict"
	default:
		htmlVersion = "Unknown"
	}

	w.logger.InfoWithContext(ctx, fmt.Sprintf("identified the document html version as %v", htmlVersion), log_utils.SetLogFile(webAnalyzerUtilsLogPrefix))

	return htmlVersion
}

// DetectPageTitle - detects the title of the web page
func (w *webAnalyzerUtilsImpl) DetectPageTitle(ctx context.Context, doc *goquery.Document) string {
	pageTitle := strings.TrimSpace(doc.Find("title").First().Text())

	w.logger.InfoWithContext(ctx, fmt.Sprintf("identified the document title as %v", pageTitle), log_utils.SetLogFile(webAnalyzerUtilsLogPrefix))

	return pageTitle
}

// DetectLoginForm - detects if there's a login page exist in the web page
// since there are many possibilities to use a username field in a login page, only the field type is checked here
func (w *webAnalyzerUtilsImpl) DetectLoginForm(ctx context.Context, doc *goquery.Document) bool {
	isLoginPageExist := false

	doc.Find("form").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.Find("input[type='password']").Length() > 0 {
			isLoginPageExist = true
			return true
		}
		return false
	})

	w.logger.InfoWithContext(ctx, fmt.Sprintf("identified the document containing a login form as %v", isLoginPageExist), log_utils.SetLogFile(webAnalyzerUtilsLogPrefix))

	return isLoginPageExist
}

// DetectHeaders - detects the number of each header type given in the typesOfHeadings
func (w *webAnalyzerUtilsImpl) DetectHeaders(ctx context.Context, doc *goquery.Document, typesOfHeadings [6]string) map[string]int {
	headers := make(map[string]int)

	for _, heading := range typesOfHeadings {
		headers[heading] = doc.Find(heading).Length()
	}

	w.logger.InfoWithContext(ctx, fmt.Sprintf("identified headers %v", headers), log_utils.SetLogFile(webAnalyzerUtilsLogPrefix))

	return headers
}

// DetectLinks - detects the internal, external link counts and returns an array of all existing links in the web page
func (w *webAnalyzerUtilsImpl) DetectLinks(ctx context.Context, doc *goquery.Document, host string) (int, int, []string) {
	var allLinks []string
	internalLinks, externalLinks := 0, 0

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript") {
			return
		}

		link := strings.TrimSpace(href)
		isInternal := strings.HasPrefix(link, "/") && !strings.HasPrefix(link, "//") || strings.Contains(link, host)

		if isInternal {
			internalLinks++
		} else {
			externalLinks++
		}
		allLinks = append(allLinks, link)
	})

	w.logger.InfoWithContext(ctx, fmt.Sprintf("identified the internal link count as %v and external link count as %v", internalLinks, externalLinks), log_utils.SetLogFile(webAnalyzerUtilsLogPrefix))
	w.logger.InfoWithContext(ctx, fmt.Sprintf("found %v links totally", len(allLinks)), log_utils.SetLogFile(webAnalyzerUtilsLogPrefix))

	return internalLinks, externalLinks, allLinks
}

// IsLinksAccessible - checks a list of links and returns the count of inaccessible links
// uses a worker group of size webAnalyzerConfig.MaxLinkAccessCheckerWorkerCount to keep
// the number of go routines from increasing uncontrollably
func (w *webAnalyzerUtilsImpl) IsLinksAccessible(ctx context.Context, links []string, base *url.URL) int {

	workers := make(chan struct{}, w.webAnalyzerConfig.MaxLinkAccessCheckerWorkerCount)
	var wg sync.WaitGroup
	resultChan := make(chan int, len(links))

	for _, link := range links {
		wg.Add(1)
		workers <- struct{}{} // acquire worker

		go func(link string) {
			defer wg.Done()
			defer func() { <-workers }() // release worker

			fullURL := w.normalizeURL(link, base)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Head(fullURL)
			if (err != nil) || (resp.StatusCode < http.StatusOK) || (resp.StatusCode >= http.StatusMultipleChoices) {
				resultChan <- 1
			} else {
				resultChan <- 0
			}
			if resp != nil {
				resp.Body.Close()
			}
		}(link)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	inaccessibleLinkCount := 0
	for result := range resultChan {
		inaccessibleLinkCount += result
	}

	return inaccessibleLinkCount
}

// check the prefix of links and normalize them by adding host name for internal links
func (w *webAnalyzerUtilsImpl) normalizeURL(link string, base *url.URL) string {
	if strings.HasPrefix(link, "//") {
		return base.Scheme + ":" + link
	}
	if strings.HasPrefix(link, "/") {
		return base.Scheme + "://" + base.Host + link
	}
	return link
}
