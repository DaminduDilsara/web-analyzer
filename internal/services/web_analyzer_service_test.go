package services

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/DaminduDilsara/web-analyzer/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// mockRoundTripper used to mock http.Get
func mockRoundTripper(resp *http.Response, err error) func() {
	original := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return resp, err
	})
	return func() { http.DefaultClient.Transport = original }
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}

func TestAnalyzeUrl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type testCase struct {
		name         string
		mockResp     *http.Response
		mockErr      error
		mockUtilsFn  func(*mocks.MockWebAnalyzerUtils)
		expectError  bool
		expectResult bool
	}

	html := `<!DOCTYPE html><html><head><title>Test Page</title></head><body><h1>Header</h1><a href="/internal">Internal</a><a href="http://external.com">External</a><form><input type='password'/></form></body></html>`
	parsedURL, _ := url.Parse("http://test.com")
	ctx := context.Background()

	cases := []testCase{
		{
			name: "Happy path",
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(html)),
			},
			mockUtilsFn: func(m *mocks.MockWebAnalyzerUtils) {
				m.EXPECT().DetectHTMLVersion(ctx, gomock.Any()).Return("HTML 5")
				m.EXPECT().IsLinksAccessible(ctx, []string{"/internal", "http://external.com"}, parsedURL).Return(0)
			},
			expectResult: true,
		},
		{
			name:        "HTTP error",
			mockErr:     assert.AnError,
			mockUtilsFn: func(m *mocks.MockWebAnalyzerUtils) {},
			expectError: true,
		},
		{
			name: "Non-200 status",
			mockResp: &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(bytes.NewBufferString("not found")),
			},
			mockUtilsFn: func(m *mocks.MockWebAnalyzerUtils) {},
			expectError: true,
		},
		{
			name: "Goquery parse error",
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(errorReader{}),
			},
			mockUtilsFn: func(m *mocks.MockWebAnalyzerUtils) {},
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUtils := mocks.NewMockWebAnalyzerUtils(ctrl)
			logger := log_utils.InitConsoleLogger()

			// Apply mock expectations
			if tc.mockUtilsFn != nil {
				tc.mockUtilsFn(mockUtils)
			}

			// Mock HTTP
			undo := mockRoundTripper(tc.mockResp, tc.mockErr)
			defer undo()

			service := NewWebAnalyzerService(logger, mockUtils)
			result, err := service.AnalyzeUrl(ctx, parsedURL)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else if tc.expectResult {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Test Page", result.Title)
			}
		})
	}
}
