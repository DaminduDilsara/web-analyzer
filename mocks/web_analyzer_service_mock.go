// Code generated by MockGen. DO NOT EDIT.
// Source: internal/services/web_analyzer_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	url "net/url"
	reflect "reflect"

	response_dtos "github.com/DaminduDilsara/web-analyzer/internal/schemas/response_dtos"
	gomock "github.com/golang/mock/gomock"
)

// MockWebAnalyzerService is a mock of WebAnalyzerService interface.
type MockWebAnalyzerService struct {
	ctrl     *gomock.Controller
	recorder *MockWebAnalyzerServiceMockRecorder
}

// MockWebAnalyzerServiceMockRecorder is the mock recorder for MockWebAnalyzerService.
type MockWebAnalyzerServiceMockRecorder struct {
	mock *MockWebAnalyzerService
}

// NewMockWebAnalyzerService creates a new mock instance.
func NewMockWebAnalyzerService(ctrl *gomock.Controller) *MockWebAnalyzerService {
	mock := &MockWebAnalyzerService{ctrl: ctrl}
	mock.recorder = &MockWebAnalyzerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebAnalyzerService) EXPECT() *MockWebAnalyzerServiceMockRecorder {
	return m.recorder
}

// AnalyzeUrl mocks base method.
func (m *MockWebAnalyzerService) AnalyzeUrl(ctx context.Context, parsedURL *url.URL) (*response_dtos.UrlAnalyzerResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AnalyzeUrl", ctx, parsedURL)
	ret0, _ := ret[0].(*response_dtos.UrlAnalyzerResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AnalyzeUrl indicates an expected call of AnalyzeUrl.
func (mr *MockWebAnalyzerServiceMockRecorder) AnalyzeUrl(ctx, parsedURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AnalyzeUrl", reflect.TypeOf((*MockWebAnalyzerService)(nil).AnalyzeUrl), ctx, parsedURL)
}
