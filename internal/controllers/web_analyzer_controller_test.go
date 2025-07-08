package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
		setupMock      func(*mocks.MockWebAnalyzerService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "Valid Request",
			requestBody: request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			setupMock: func(mockService *mocks.MockWebAnalyzerService) {
				parsedURL, _ := url.ParseRequestURI("http://example.com")
				mockResp := &response_dtos.UrlAnalyzerResponse{Title: "Example"}
				mockService.EXPECT().AnalyzeUrl(gomock.Any(), parsedURL).Return(mockResp, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &response_dtos.UrlAnalyzerResponse{Title: "Example"},
		},
		{
			name:        "Invalid JSON",
			requestBody: "not-json",
			setupMock:   func(mockService *mocks.MockWebAnalyzerService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   &response_dtos.ErrorResponse{Message: "invalid or missing json body"},
		},
		{
			name:        "Invalid URL",
			requestBody: request_dtos.UrlAnalyzerRequest{Url: "not-a-url"},
			setupMock:   func(mockService *mocks.MockWebAnalyzerService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   &response_dtos.ErrorResponse{Message: "failed to parse url"},
		},
		{
			name:        "Service Error",
			requestBody: request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			setupMock: func(mockService *mocks.MockWebAnalyzerService) {
				parsedURL, _ := url.ParseRequestURI("http://example.com")
				mockService.EXPECT().AnalyzeUrl(gomock.Any(), parsedURL).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   &response_dtos.ErrorResponse{Message: "something is wrong, please try again later"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockWebAnalyzerService(ctrl)
			controller := NewControllerV1(mockService, logger)

			tt.setupMock(mockService)

			var jsonBody []byte
			if str, ok := tt.requestBody.(string); ok {
				jsonBody = []byte(str)
			} else {
				jsonBody, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			controller.AnalyzeController(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp interface{}
			if tt.expectedBody != nil {
				resp = tt.expectedBody
				json.NewDecoder(w.Body).Decode(resp)
			}
		})
	}
}
