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
		expectedStatus int
		mockSetup      func(*mocks.MockWebAnalyzerService)
	}{
		{
			name:           "Valid Request",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			expectedStatus: http.StatusOK,
			mockSetup: func(s *mocks.MockWebAnalyzerService) {
				urlParsed, _ := url.ParseRequestURI("http://example.com")
				s.EXPECT().AnalyzeUrl(gomock.Any(), urlParsed).Return(&response_dtos.UrlAnalyzerResponse{Title: "Example"}, nil)
			},
		},
		{
			name:           "Invalid JSON",
			requestBody:    "not-json",
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(s *mocks.MockWebAnalyzerService) {},
		},
		{
			name:           "Bad URL",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "bad_url"},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(s *mocks.MockWebAnalyzerService) {},
		},
		{
			name:           "Service Error",
			requestBody:    request_dtos.UrlAnalyzerRequest{Url: "http://example.com"},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func(s *mocks.MockWebAnalyzerService) {
				urlParsed, _ := url.ParseRequestURI("http://example.com")
				s.EXPECT().AnalyzeUrl(gomock.Any(), urlParsed).Return(nil, errors.New("internal error"))
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
		})
	}
}
