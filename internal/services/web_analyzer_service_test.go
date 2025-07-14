package services

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/DaminduDilsara/web-analyzer/custom_errors"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/response_dtos"
	"github.com/DaminduDilsara/web-analyzer/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// mockHTTPClient creates a mock HTTP client
func mockHTTPClient(resp *http.Response, err error) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return resp, err
		}),
		Timeout: 50 * time.Second,
	}
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
		name              string
		mockResp          *http.Response
		mockErr           error
		mockUtilsFn       func(*mocks.MockWebAnalyzerUtils)
		expectError       bool
		expectResult      *response_dtos.UrlAnalyzerResponse
		expectCustomError *custom_errors.CustomError
	}

	html := ``
	parsedURL, _ := url.Parse("http://test.test")
	ctx := context.Background()

	expectedHeadings := map[string]int{
		"h1": 1,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}

	expectedResponse := &response_dtos.UrlAnalyzerResponse{
		HTMLVersion:       "HTML 5",
		Title:             "Test Page",
		Headings:          expectedHeadings,
		InternalLinks:     1,
		ExternalLinks:     1,
		InaccessibleLinks: 0,
		LoginForm:         true,
	}

	cases := []testCase{
		{
			name: "Happy path",
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(html)),
			},
			mockUtilsFn: func(m *mocks.MockWebAnalyzerUtils) {
				m.EXPECT().DetectHTMLVersion(ctx, gomock.Any()).Return("HTML 5")
				m.EXPECT().DetectPageTitle(ctx, gomock.Any()).Return("Test Page")
				m.EXPECT().DetectLoginForm(ctx, gomock.Any()).Return(true)
				m.EXPECT().DetectHeaders(ctx, gomock.Any(), typesOfHeadings).Return(expectedHeadings)
				m.EXPECT().DetectLinks(ctx, gomock.Any(), "test.test").Return(1, 1, []string{"/internal", "http://external.test"})
				m.EXPECT().IsLinksAccessible(ctx, []string{"/internal", "http://external.test"}, parsedURL).Return(0)
			},
			expectResult:      expectedResponse,
			expectError:       false,
			expectCustomError: nil,
		},
		{
			name:              "HTTP error (no such host)",
			mockErr:           &url.Error{Op: "Get", URL: "http://test.test", Err: fmt.Errorf("no such host")},
			mockUtilsFn:       func(m *mocks.MockWebAnalyzerUtils) {},
			expectError:       true,
			expectCustomError: &custom_errors.CustomError{Code: 404, Message: "server not found for the given url or domain does not exist"},
		},
		{
			name:              "HTTP error (none url.Error)",
			mockErr:           fmt.Errorf("some generic error"),
			mockUtilsFn:       func(m *mocks.MockWebAnalyzerUtils) {},
			expectError:       true,
			expectCustomError: &custom_errors.CustomError{Code: 502, Message: "failed to connect to the given server"},
		},
		{
			name: "404 error",
			mockResp: &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(bytes.NewBufferString("not found")),
			},
			mockUtilsFn:       func(m *mocks.MockWebAnalyzerUtils) {},
			expectError:       true,
			expectCustomError: &custom_errors.CustomError{Code: 404, Message: "unexpected HTTP status code"},
		},
		{
			name: "500 error",
			mockResp: &http.Response{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewBufferString("internal server error")),
			},
			mockUtilsFn:       func(m *mocks.MockWebAnalyzerUtils) {},
			expectError:       true,
			expectCustomError: &custom_errors.CustomError{Code: 500, Message: "unexpected HTTP status code"},
		},
		{
			name: "Goquery parse error",
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(errorReader{}),
			},
			mockUtilsFn:       func(m *mocks.MockWebAnalyzerUtils) {},
			expectError:       true,
			expectCustomError: &custom_errors.CustomError{Code: 500, Message: "response cannot parse to html"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUtils := mocks.NewMockWebAnalyzerUtils(ctrl)
			logger := log_utils.InitConsoleLogger()

			if tc.mockUtilsFn != nil {
				tc.mockUtilsFn(mockUtils)
			}

			mockClient := mockHTTPClient(tc.mockResp, tc.mockErr)

			service := NewWebAnalyzerServiceWithClient(logger, mockUtils, mockClient)
			result, customErr := service.AnalyzeUrl(ctx, parsedURL)

			if tc.expectError {
				assert.Nil(t, result)
				assert.NotNil(t, customErr)
				if tc.expectCustomError != nil {
					ce, ok := customErr.(*custom_errors.CustomError)
					if !ok {
						t.Fatalf("error should be of type *CustomError, got %T: %v", customErr, customErr)
					}
					assert.Equal(t, tc.expectCustomError.Code, ce.Code)
					assert.Equal(t, tc.expectCustomError.Message, ce.Message)
				}
			} else {
				assert.Nil(t, customErr)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectResult.HTMLVersion, result.HTMLVersion)
				assert.Equal(t, tc.expectResult.Title, result.Title)
				assert.Equal(t, tc.expectResult.Headings, result.Headings)
				assert.Equal(t, tc.expectResult.InternalLinks, result.InternalLinks)
				assert.Equal(t, tc.expectResult.ExternalLinks, result.ExternalLinks)
				assert.Equal(t, tc.expectResult.InaccessibleLinks, result.InaccessibleLinks)
				assert.Equal(t, tc.expectResult.LoginForm, result.LoginForm)
			}
		})
	}
}
