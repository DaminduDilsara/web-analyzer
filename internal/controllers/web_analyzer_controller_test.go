package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaminduDilsara/web-analyzer/custom_errors"
	"github.com/DaminduDilsara/web-analyzer/internal/log_utils"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/request_dtos"
	"github.com/DaminduDilsara/web-analyzer/internal/schemas/response_dtos"
	"github.com/DaminduDilsara/web-analyzer/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzeController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := log_utils.InitConsoleLogger()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
		mockSetup      func(*mocks.MockWebAnalyzerService)
	}{
		{
			name:           "Valid Request",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"html_version":       "HTML5",
				"title":              "Example",
				"headings":           map[string]interface{}{},
				"internal_links":     float64(0),
				"external_links":     float64(0),
				"inaccessible_links": float64(0),
				"login_form":         false,
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {
				s.EXPECT().AnalyzeUrl(gomock.Any(), gomock.Any()).Return(&response_dtos.UrlAnalyzerResponse{
					HTMLVersion:       "HTML5",
					Title:             "Example",
					Headings:          map[string]int{},
					InternalLinks:     0,
					ExternalLinks:     0,
					InaccessibleLinks: 0,
					LoginForm:         false,
				}, nil)
			},
		},
		{
			name:           "Invalid JSON",
			requestBody:    "not-json",
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusBadRequest),
				"message": "invalid or missing json body",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {},
		},
		{
			name:           "Bad URL",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "bad_url"},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusBadRequest),
				"message": "failed to parse url",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {},
		},
		{
			name: "Missing URL Field",
			requestBody: struct {
				Foo string `json:"foo"`
			}{Foo: "bar"},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusBadRequest),
				"message": "invalid or missing json body",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {},
		},
		{
			name:           "Service internal Error",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusInternalServerError),
				"message": "internal error",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {
				s.EXPECT().AnalyzeUrl(gomock.Any(), gomock.Any()).Return(nil, custom_errors.NewCustomError(http.StatusInternalServerError, "internal error", nil))
			},
		},
		{
			name:           "Service Error",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://notfound.com"},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusNotFound),
				"message": "server not found for the given url or domain does not exist",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {
				s.EXPECT().AnalyzeUrl(gomock.Any(), gomock.Any()).Return(nil, custom_errors.NewCustomError(http.StatusNotFound, "server not found for the given url or domain does not exist", nil))
			},
		},
		{
			name:           "Service Error",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusInternalServerError),
				"message": "failed to analyze url: http://example.com error: some generic error",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {
				s.EXPECT().AnalyzeUrl(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some generic error"))
			},
		},
		{
			name:           "Invalid Host in URL",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://invalid_host"},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusBadRequest),
				"message": "failed to parse url",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {},
		},
		{
			name:           "Service Error not found",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"code":    float64(http.StatusNotFound),
				"message": "unexpected HTTP status code",
			},
			mockSetup: func(s *mocks.MockWebAnalyzerService) {
				s.EXPECT().AnalyzeUrl(gomock.Any(), gomock.Any()).Return(nil, custom_errors.NewCustomError(http.StatusNotFound, "unexpected HTTP status code", nil))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockWebAnalyzerService(ctrl)
			tt.mockSetup(mockService)

			controller := NewControllerV1(mockService, logger)

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			controller.AnalyzeController(c)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var resp map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				for k, v := range tt.expectedBody {
					assert.Equal(t, v, resp[k], "field %s mismatch", k)
				}
			}
		})
	}
}
