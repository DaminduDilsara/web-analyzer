package services

import (
	"context"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/responseDtos"
	"github.com/DaminduDilsara/web-analyzer/internal/utils"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const webAnalyzerServiceLogPrefix = "web_analyzer_service_impl"

var typesOfHeadings = [...]string{"h1", "h2", "h3", "h4", "h5", "h6"}

type webAnalyzerServiceImpl struct {
	logger utils.LoggerInterface
}

func NewWebAnalyzerService(logger utils.LoggerInterface) WebAnalyzerService {
	return &webAnalyzerServiceImpl{
		logger: logger,
	}
}

func (w *webAnalyzerServiceImpl) AnalyzeUrl(ctx context.Context, parsedURL *url.URL) (*responseDtos.UrlAnalyzerResponse, error) {

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		w.logger.ErrorWithContext(ctx, "Unable to fetch URL", err, utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}
	defer resp.Body.Close()

	if (resp.StatusCode < http.StatusOK) || (resp.StatusCode >= http.StatusMultipleChoices) {
		err = fmt.Errorf("unexpected HTTP status code: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		w.logger.ErrorWithContext(ctx, "unexpected HTTP status code", err, utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		w.logger.ErrorWithContext(ctx, "response cannot parse to html", err, utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}

	htmlText, err := doc.Html()
	if err != nil {
		w.logger.ErrorWithContext(ctx, "cannot extract html text from document", err, utils.SetLogFile(webAnalyzerServiceLogPrefix))
		return nil, err
	}

	result := responseDtos.UrlAnalyzerResponse{
		Headings:    make(map[string]int),
		Title:       doc.Find("title").First().Text(),
		HTMLVersion: detectHTMLVersion(htmlText),
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

	result.InaccessibleLinks = isLinksAccessible(allLinks, parsedURL)

	return &result, nil

}

func detectHTMLVersion(body string) string {
	body = strings.ToLower(body)

	switch {
	case strings.Contains(body, "<!doctype html>"):
		return "HTML 5"
	case strings.Contains(body, "-//w3c//dtd html 4.01 frameset//en"):
		return "HTML 4.01 Frameset"
	case strings.Contains(body, "-//w3c//dtd html 4.01 transitional//en"):
		return "HTML 4.01 Transitional"
	case strings.Contains(body, "-//w3c//dtd html 4.01//en"):
		return "HTML 4.01 Strict"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.1//en"):
		return "XHTML 1.1"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.0 frameset//en"):
		return "XHTML 1.0 Frameset"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.0 transitional//en"):
		return "XHTML 1.0 Transitional"
	case strings.Contains(body, "-//w3c//dtd xhtml 1.0 strict//en"):
		return "XHTML 1.0 Strict"
	default:
		return "Unknown"
	}
}

func isLinksAccessible(links []string, base *url.URL) int {
	const maxWorkers = 20

	workers := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	resultChan := make(chan int, len(links))

	for _, link := range links {
		wg.Add(1)
		workers <- struct{}{} // acquire worker

		go func(link string) {
			defer wg.Done()
			defer func() { <-workers }() // release worker

			fullURL := normalizeURL(link, base)

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

func normalizeURL(link string, base *url.URL) string {
	if strings.HasPrefix(link, "//") {
		return base.Scheme + ":" + link
	}
	if strings.HasPrefix(link, "/") {
		return base.Scheme + "://" + base.Host + link
	}
	return link
}
