package web_analyzer_utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/DaminduDilsara/web-analyzer/configurations"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
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
