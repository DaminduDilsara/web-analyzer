package web_analyzer_utils

import (
	"context"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
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
				w.logger.InfoWithContext(ctx, fmt.Sprintf("unable to head the url %v. response status %v", fullURL, resp.Status), log_utils.SetLogFile(webAnalyzerUtilsLogPrefix))
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
