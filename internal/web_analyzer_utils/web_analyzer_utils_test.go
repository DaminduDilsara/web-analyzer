package web_analyzer_utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/PuerkitoBio/goquery"
)

func TestDetectHTMLVersion(t *testing.T) {
	logger := log_utils.InitConsoleLogger()
	config := &configurations.WebAnalyzerConfigurations{MaxLinkAccessCheckerWorkerCount: 1}
	utils := NewWebAnalyzerUtils(logger, config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML 5",
			input:    "<!DOCTYPE html>",
			expected: "HTML 5",
		},
		{
			name:     "HTML 4.01 Frameset",
			input:    "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Frameset//EN\">",
			expected: "HTML 4.01 Frameset",
		},
		{
			name:     "HTML 4.01 Transitional",
			input:    "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\">",
			expected: "HTML 4.01 Transitional",
		},
		{
			name:     "HTML 4.01 Strict",
			input:    "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\">",
			expected: "HTML 4.01 Strict",
		},
		{
			name:     "XHTML 1.1",
			input:    "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.1//EN\">",
			expected: "XHTML 1.1",
		},
		{
			name:     "XHTML 1.0 Frameset",
			input:    "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Frameset//EN\">",
			expected: "XHTML 1.0 Frameset",
		},
		{
			name:     "XHTML 1.0 Transitional",
			input:    "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\">",
			expected: "XHTML 1.0 Transitional",
		},
		{
			name:     "XHTML 1.0 Strict",
			input:    "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\">",
			expected: "XHTML 1.0 Strict",
		},
		{
			name:     "Unknown Version",
			input:    "<html><head></head><body></body></html>",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.DetectHTMLVersion(context.Background(), tt.input)
			if result != tt.expected {
				t.Errorf("DetectHTMLVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectPageTitle(t *testing.T) {
	logger := log_utils.InitConsoleLogger()
	config := &configurations.WebAnalyzerConfigurations{MaxLinkAccessCheckerWorkerCount: 1}
	utils := NewWebAnalyzerUtils(logger, config)

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Sample Title",
			html:     "<html><head><title>Test Page</title></head><body></body></html>",
			expected: "Test Page",
		},
		{
			name:     "No Title",
			html:     "<html><head></head><body></body></html>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := utils.DetectPageTitle(context.Background(), doc)
			if result != tt.expected {
				t.Errorf("DetectPageTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectLoginForm(t *testing.T) {
	logger := log_utils.InitConsoleLogger()
	config := &configurations.WebAnalyzerConfigurations{MaxLinkAccessCheckerWorkerCount: 1}
	utils := NewWebAnalyzerUtils(logger, config)

	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name:     "Login Form with Password Field",
			html:     "<html><body><form><input type='text' name='username'><input type='password' name='password'></form></body></html>",
			expected: true,
		},
		{
			name:     "Multiple Forms - One with Password",
			html:     "<html><head></head><body><form><input type='text' /></form><form><input type='password' /></form></body></html>",
			expected: true,
		},
		{
			name:     "No Login Form",
			html:     "<html><body><form><input type='text' name='search'></form></body></html>",
			expected: false,
		},
		{
			name:     "Multiple Password Fields",
			html:     "<html><body><form><input type='password'><input type='password'></form></body></html>",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := utils.DetectLoginForm(context.Background(), doc)
			if result != tt.expected {
				t.Errorf("DetectLoginForm() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectHeaders(t *testing.T) {
	logger := log_utils.InitConsoleLogger()
	config := &configurations.WebAnalyzerConfigurations{MaxLinkAccessCheckerWorkerCount: 1}
	utils := NewWebAnalyzerUtils(logger, config)

	tests := []struct {
		name            string
		html            string
		typesOfHeadings [6]string
		expected        map[string]int
	}{
		{
			name:            "All Header Types Present",
			html:            "<html><body><h1>Title</h1><h2>Subtitle</h2><h3>Section</h3><h4>Subsection</h4><h5>Sub-subsection</h5><h6>Detail</h6></body></html>",
			typesOfHeadings: [6]string{"h1", "h2", "h3", "h4", "h5", "h6"},
			expected:        map[string]int{"h1": 1, "h2": 1, "h3": 1, "h4": 1, "h5": 1, "h6": 1},
		},
		{
			name:            "Multiple Headers of Same Type",
			html:            "<html><body><h1>Title 1</h1><h1>Title 2</h1><h2>Subtitle</h2></body></html>",
			typesOfHeadings: [6]string{"h1", "h2", "h3", "h4", "h5", "h6"},
			expected:        map[string]int{"h1": 2, "h2": 1, "h3": 0, "h4": 0, "h5": 0, "h6": 0},
		},
		{
			name:            "No Headers",
			html:            "<html><body><div>Content</div></body></html>",
			typesOfHeadings: [6]string{"h1", "h2", "h3", "h4", "h5", "h6"},
			expected:        map[string]int{"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := utils.DetectHeaders(context.Background(), doc, tt.typesOfHeadings)

			// Compare maps
			if len(result) != len(tt.expected) {
				t.Errorf("DetectHeaders() returned map with %d keys, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("DetectHeaders()[%s] = %v, want %v", key, result[key], expectedValue)
				}
			}
		})
	}
}

func TestDetectLinks(t *testing.T) {
	logger := log_utils.InitConsoleLogger()
	config := &configurations.WebAnalyzerConfigurations{MaxLinkAccessCheckerWorkerCount: 1}
	utils := NewWebAnalyzerUtils(logger, config)

	tests := []struct {
		name             string
		html             string
		host             string
		expectedInternal int
		expectedExternal int
		expectedLinks    []string
	}{
		{
			name:             "Internal Links Only",
			html:             "<html><body><a href='/page1'>Link 1</a><a href='/page2'>Link 2</a></body></html>",
			host:             "example.com",
			expectedInternal: 2,
			expectedExternal: 0,
			expectedLinks:    []string{"/page1", "/page2"},
		},
		{
			name:             "External Links Only",
			html:             "<html><body><a href='https://google.com'>Google</a><a href='http://github.com'>GitHub</a></body></html>",
			host:             "example.com",
			expectedInternal: 0,
			expectedExternal: 2,
			expectedLinks:    []string{"https://google.com", "http://github.com"},
		},
		{
			name:             "Mixed Internal and External Links",
			html:             "<html><body><a href='/internal'>Internal</a><a href='https://external.com'>External</a><a href='//example.com/protocol-relative'>Protocol Relative</a></body></html>",
			host:             "example.com",
			expectedInternal: 2,
			expectedExternal: 1,
			expectedLinks:    []string{"/internal", "https://external.com", "//example.com/protocol-relative"},
		},
		{
			name:             "No Links",
			html:             "<html><body><div>Content</div></body></html>",
			host:             "example.com",
			expectedInternal: 0,
			expectedExternal: 0,
			expectedLinks:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			internal, external, links := utils.DetectLinks(context.Background(), doc, tt.host)

			if internal != tt.expectedInternal {
				t.Errorf("DetectLinks() internal = %v, want %v", internal, tt.expectedInternal)
			}

			if external != tt.expectedExternal {
				t.Errorf("DetectLinks() external = %v, want %v", external, tt.expectedExternal)
			}

			if len(links) != len(tt.expectedLinks) {
				t.Errorf("DetectLinks() returned %d links, want %d", len(links), len(tt.expectedLinks))
			}

			// Check if all expected links are present
			for _, expectedLink := range tt.expectedLinks {
				found := false
				for _, link := range links {
					if link == expectedLink {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("DetectLinks() missing expected link: %s", expectedLink)
				}
			}
		})
	}
}

func TestIsLinksAccessible(t *testing.T) {
	logger := log_utils.InitConsoleLogger()
	config := &configurations.WebAnalyzerConfigurations{MaxLinkAccessCheckerWorkerCount: 2}
	utils := NewWebAnalyzerUtils(logger, config)

	// Accessible server (returns 200)
	accessibleSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer accessibleSrv.Close()

	// Inaccessible server (returns 404)
	inaccessibleSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer inaccessibleSrv.Close()

	// Server that times out
	timeoutSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate timeout by not responding
		select {}
	}))
	defer timeoutSrv.Close()

	accessibleURL, _ := url.Parse(accessibleSrv.URL)
	inaccessibleURL, _ := url.Parse(inaccessibleSrv.URL)

	tests := []struct {
		name     string
		links    []string
		base     *url.URL
		expected int
	}{
		{
			name:     "All Accessible",
			links:    []string{accessibleSrv.URL, accessibleSrv.URL + "/foo"},
			base:     accessibleURL,
			expected: 0,
		},
		{
			name:     "All Inaccessible",
			links:    []string{inaccessibleSrv.URL, inaccessibleSrv.URL + "/bar"},
			base:     inaccessibleURL,
			expected: 2,
		},
		{
			name:     "Mixed Accessible and Inaccessible",
			links:    []string{accessibleSrv.URL, inaccessibleSrv.URL},
			base:     accessibleURL,
			expected: 1,
		},
		{
			name:     "Relative and Protocol-Relative URLs",
			links:    []string{"/foo", "//" + accessibleURL.Host + "/bar", inaccessibleSrv.URL},
			base:     accessibleURL,
			expected: 1, // only inaccessibleSrv.URL is inaccessible
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := utils.IsLinksAccessible(context.Background(), tt.links, tt.base)
			if count != tt.expected {
				t.Errorf("IsLinksAccessible() = %v, want %v", count, tt.expected)
			}
		})
	}
}
